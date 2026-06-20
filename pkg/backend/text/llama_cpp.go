//go:build llama

// Package-level llama.cpp/GGUF backend. This file is compiled only with the
// "llama" build tag and requires a local llama.cpp build (headers + libllama).
// See pkg/backend/text/README.md for build instructions.
//
// It is an experimental/feasibility implementation: the llama.cpp C API evolves
// between releases, so the cgo declarations below target a recent (2025-era)
// API. If your llama.cpp build differs, adjust the function names/signatures to
// match its llama.h.
package text

/*
#cgo CFLAGS: -I${SRCDIR}/../../../third_party/llama.cpp/include
#cgo LDFLAGS: -L${SRCDIR}/../../../third_party/llama.cpp/build/bin -lllama -lm -lstdc++
#include <stdlib.h>
#include "llama.h"
*/
import "C"

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

var llamaInitOnce sync.Once

// Llama is a llama.cpp-backed text backend. A single context is guarded by a
// mutex because llama contexts are not safe for concurrent decoding.
type Llama struct {
	mu    sync.Mutex
	model *C.struct_llama_model
	ctx   *C.struct_llama_context
	vocab *C.struct_llama_vocab

	maxTokens int
	temp      float32
	topP      float32
}

// newLlama loads a GGUF model and prepares a context.
func newLlama(o Options) (Backend, error) {
	if strings.TrimSpace(o.ModelPath) == "" {
		return nil, fmt.Errorf("llama: ModelPath is required")
	}
	llamaInitOnce.Do(func() { C.llama_backend_init() })

	cPath := C.CString(o.ModelPath)
	defer C.free(unsafe.Pointer(cPath))

	mparams := C.llama_model_default_params()
	model := C.llama_model_load_from_file(cPath, mparams)
	if model == nil {
		return nil, fmt.Errorf("llama: failed to load model %q", o.ModelPath)
	}

	cparams := C.llama_context_default_params()
	if o.ContextTokens > 0 {
		cparams.n_ctx = C.uint32_t(o.ContextTokens)
	}
	ctx := C.llama_init_from_model(model, cparams)
	if ctx == nil {
		C.llama_model_free(model)
		return nil, fmt.Errorf("llama: failed to create context")
	}

	maxTokens := o.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 128
	}
	return &Llama{
		model:     model,
		ctx:       ctx,
		vocab:     C.llama_model_get_vocab(model),
		maxTokens: maxTokens,
		temp:      o.Temperature,
		topP:      o.TopP,
	}, nil
}

// Generate runs greedy/temperature decoding over the prompt.
func (l *Llama) Generate(ctx context.Context, prompt string, opts GenerateOptions) (Result, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	maxTokens := l.maxTokens
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}
	temp := l.temp
	if opts.Temperature > 0 {
		temp = opts.Temperature
	}
	topP := l.topP
	if opts.TopP > 0 {
		topP = opts.TopP
	}

	tokens, err := l.tokenize(prompt)
	if err != nil {
		return Result{}, err
	}

	sampler := l.newSampler(temp, topP)
	defer C.llama_sampler_free(sampler)

	batch := C.llama_batch_get_one(&tokens[0], C.int32_t(len(tokens)))
	if rc := C.llama_decode(l.ctx, batch); rc != 0 {
		return Result{}, fmt.Errorf("llama: decode failed (%d)", int(rc))
	}

	var out strings.Builder
	generated := 0
	for generated < maxTokens {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		tok := C.llama_sampler_sample(sampler, l.ctx, -1)
		if C.llama_vocab_is_eog(l.vocab, tok) {
			break
		}
		out.WriteString(l.tokenToText(tok))
		generated++

		one := tok
		batch = C.llama_batch_get_one(&one, 1)
		if rc := C.llama_decode(l.ctx, batch); rc != 0 {
			return Result{}, fmt.Errorf("llama: decode failed (%d)", int(rc))
		}
	}

	return Result{Text: strings.TrimSpace(out.String()), Tokens: generated}, nil
}

// Close releases the context and model.
func (l *Llama) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.ctx != nil {
		C.llama_free(l.ctx)
		l.ctx = nil
	}
	if l.model != nil {
		C.llama_model_free(l.model)
		l.model = nil
	}
	return nil
}

func (l *Llama) tokenize(text string) ([]C.llama_token, error) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	// Negative return is the (negated) required capacity.
	n := C.llama_tokenize(l.vocab, cText, C.int32_t(len(text)), nil, 0, C.bool(true), C.bool(true))
	capN := int(-n)
	if capN <= 0 {
		capN = int(n)
	}
	if capN <= 0 {
		return nil, fmt.Errorf("llama: tokenization produced no tokens")
	}
	tokens := make([]C.llama_token, capN)
	got := C.llama_tokenize(l.vocab, cText, C.int32_t(len(text)), &tokens[0], C.int32_t(capN), C.bool(true), C.bool(true))
	if got < 0 {
		return nil, fmt.Errorf("llama: tokenization failed")
	}
	return tokens[:int(got)], nil
}

func (l *Llama) tokenToText(tok C.llama_token) string {
	buf := make([]C.char, 256)
	n := C.llama_token_to_piece(l.vocab, tok, &buf[0], C.int32_t(len(buf)), 0, C.bool(false))
	if n <= 0 {
		return ""
	}
	return C.GoStringN(&buf[0], n)
}

func (l *Llama) newSampler(temp, topP float32) *C.struct_llama_sampler {
	params := C.llama_sampler_chain_default_params()
	chain := C.llama_sampler_chain_init(params)
	if temp <= 0 {
		// Greedy decoding is deterministic, which suits short narration.
		C.llama_sampler_chain_add(chain, C.llama_sampler_init_greedy())
		return chain
	}
	if topP > 0 {
		C.llama_sampler_chain_add(chain, C.llama_sampler_init_top_p(C.float(topP), 1))
	}
	C.llama_sampler_chain_add(chain, C.llama_sampler_init_temp(C.float(temp)))
	C.llama_sampler_chain_add(chain, C.llama_sampler_init_dist(C.LLAMA_DEFAULT_SEED))
	return chain
}
