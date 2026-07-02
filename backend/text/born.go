//go:build born

// This file implements a pure-Go GGUF inference backend on top of
// github.com/born-ml/born. It is compiled only with the "born" build tag, which
// also pulls in the Born module and its dependency tree. Unlike llama_cpp.go it
// needs NO cgo and NO native library — a host can produce a single self-
// contained binary.
//
// SPIKE: the calls below follow Born's quickstart. Verify them against Born's
// current API before relying on this, and choose the correct tokenizer for the
// model — the quickstart's tiktoken/gpt-4 tokenizer is wrong for a LLaMA GGUF,
// which carries its own vocabulary. See docs/exp-native-go.md.
package text

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/born-ml/born/backend/cpu"
	"github.com/born-ml/born/generate"
	"github.com/born-ml/born/models/llama"
	"github.com/born-ml/born/tokenizer"
)

// bornBackend runs a GGUF model through Born's pure-Go CPU backend. Generation
// is guarded by a mutex because the generator holds a single KV-cache.
type bornBackend struct {
	mu        sync.Mutex
	model     *llama.Model
	gen       *generate.TextGenerator
	maxTokens int
}

func newBorn(o Options) (Backend, error) {
	if strings.TrimSpace(o.ModelPath) == "" {
		return nil, fmt.Errorf("born: ModelPath is required")
	}

	be := cpu.New()
	model, err := llama.LoadGGUF(o.ModelPath, be)
	if err != nil {
		return nil, fmt.Errorf("born: loading %q: %w", o.ModelPath, err)
	}

	// TODO: prefer the tokenizer embedded in the GGUF; gpt-4 tiktoken is a
	// placeholder that will mis-tokenise LLaMA models.
	tok, err := tokenizer.NewTikTokenForModel("gpt-4")
	if err != nil {
		model.Release()
		return nil, fmt.Errorf("born: tokenizer: %w", err)
	}

	temp := o.Temperature
	if temp == 0 {
		temp = 0.7
	}
	topP := o.TopP
	if topP == 0 {
		topP = 0.9
	}
	maxTokens := o.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 64
	}

	gen := generate.NewTextGenerator(model, tok, generate.SamplingConfig{
		Temperature: temp,
		TopP:        topP,
		TopK:        40,
	})

	return &bornBackend{model: model, gen: gen, maxTokens: maxTokens}, nil
}

func (b *bornBackend) Generate(ctx context.Context, prompt string, opts GenerateOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	maxTokens := b.maxTokens
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	out, err := b.gen.Generate(prompt, generate.GenerateConfig{MaxTokens: maxTokens})
	if err != nil {
		return Result{}, fmt.Errorf("born: generate: %w", err)
	}
	text := strings.TrimSpace(out)
	return Result{Text: text, Tokens: len(strings.Fields(text))}, nil
}

func (b *bornBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.model != nil {
		b.model.Release()
		b.model = nil
	}
	return nil
}
