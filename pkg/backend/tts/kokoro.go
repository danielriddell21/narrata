//go:build kokoro

// This file implements the Kokoro/ONNX text-to-speech backend. It is compiled
// only with the "kokoro" build tag and requires ONNX Runtime (headers +
// libonnxruntime). See pkg/backend/tts/README.md for build instructions.
//
// This is an experimental/feasibility implementation. Kokoro model exports vary
// (input/output tensor names, the style-vector layout, and the grapheme/phoneme
// vocabulary differ between releases), so the constants below may need adjusting
// to match a specific kokoro.onnx file. The C shim in the cgo preamble wraps the
// ONNX Runtime C API so the Go side stays small.

package tts

/*
#cgo CFLAGS: -I${SRCDIR}/../../../third_party/onnxruntime/include
#cgo LDFLAGS: -L${SRCDIR}/../../../third_party/onnxruntime/lib -lonnxruntime
#include <stdlib.h>
#include <string.h>
#include "onnxruntime_c_api.h"

static const OrtApi* g_ort = NULL;

static int kk_init() {
    g_ort = OrtGetApiBase()->GetApi(ORT_API_VERSION);
    return g_ort != NULL ? 0 : -1;
}

typedef struct {
    OrtEnv* env;
    OrtSession* session;
    OrtSessionOptions* opts;
} kk_ctx;

static kk_ctx* kk_open(const char* model_path) {
    if (g_ort == NULL && kk_init() != 0) return NULL;
    kk_ctx* c = (kk_ctx*)calloc(1, sizeof(kk_ctx));
    if (c == NULL) return NULL;
    if (g_ort->CreateEnv(ORT_LOGGING_LEVEL_WARNING, "narrata-kokoro", &c->env) != NULL) { free(c); return NULL; }
    g_ort->CreateSessionOptions(&c->opts);
    if (g_ort->CreateSession(c->env, model_path, c->opts, &c->session) != NULL) {
        if (c->opts) g_ort->ReleaseSessionOptions(c->opts);
        if (c->env) g_ort->ReleaseEnv(c->env);
        free(c);
        return NULL;
    }
    return c;
}

static void kk_close(kk_ctx* c) {
    if (c == NULL) return;
    if (c->session) g_ort->ReleaseSession(c->session);
    if (c->opts) g_ort->ReleaseSessionOptions(c->opts);
    if (c->env) g_ort->ReleaseEnv(c->env);
    free(c);
}

// kk_run runs the model with: tokens (int64[1,n]), style (float[1,style_len]),
// speed (float[1]). It writes up to out_cap float samples into out and returns
// the number of samples produced, or a negative error code.
static long kk_run(kk_ctx* c,
                   const char* tok_name, const long long* tokens, long n_tokens,
                   const char* style_name, const float* style, long style_len,
                   const char* speed_name, float speed,
                   const char* out_name,
                   float* out, long out_cap) {
    OrtMemoryInfo* mem = NULL;
    if (g_ort->CreateCpuMemoryInfo(OrtArenaAllocator, OrtMemTypeDefault, &mem) != NULL) return -1;

    OrtValue* tok_t = NULL; OrtValue* style_t = NULL; OrtValue* speed_t = NULL;
    int64_t tok_shape[2] = {1, n_tokens};
    int64_t style_shape[2] = {1, style_len};
    int64_t speed_shape[1] = {1};
    long rc = -2;

    if (g_ort->CreateTensorWithDataAsOrtValue(mem, (void*)tokens, n_tokens*sizeof(long long),
            tok_shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &tok_t) != NULL) goto done;
    if (g_ort->CreateTensorWithDataAsOrtValue(mem, (void*)style, style_len*sizeof(float),
            style_shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT, &style_t) != NULL) goto done;
    if (g_ort->CreateTensorWithDataAsOrtValue(mem, (void*)&speed, sizeof(float),
            speed_shape, 1, ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT, &speed_t) != NULL) goto done;

    {
        const char* in_names[3]  = {tok_name, style_name, speed_name};
        const OrtValue* in_vals[3] = {tok_t, style_t, speed_t};
        const char* out_names[1] = {out_name};
        OrtValue* out_val = NULL;
        if (g_ort->Run(c->session, NULL, in_names, in_vals, 3, out_names, 1, &out_val) != NULL) { rc = -3; goto done; }

        float* data = NULL;
        if (g_ort->GetTensorMutableData(out_val, (void**)&data) != NULL) { g_ort->ReleaseValue(out_val); rc = -4; goto done; }
        struct OrtTensorTypeAndShapeInfo* info = NULL;
        g_ort->GetTensorTypeAndShape(out_val, &info);
        size_t count = 0;
        g_ort->GetTensorShapeElementCount(info, &count);
        g_ort->ReleaseTensorTypeAndShapeInfo(info);
        long n = (long)count;
        if (n > out_cap) n = out_cap;
        memcpy(out, data, n*sizeof(float));
        g_ort->ReleaseValue(out_val);
        rc = n;
    }

done:
    if (tok_t) g_ort->ReleaseValue(tok_t);
    if (style_t) g_ort->ReleaseValue(style_t);
    if (speed_t) g_ort->ReleaseValue(speed_t);
    if (mem) g_ort->ReleaseMemoryInfo(mem);
    return rc;
}
*/
import "C"

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// Tensor input/output names and style length for the targeted Kokoro export.
// Adjust these to match your kokoro.onnx if it differs.
const (
	kokoroInputTokens = "input_ids"
	kokoroInputStyle  = "style"
	kokoroInputSpeed  = "speed"
	kokoroOutputAudio = "waveform"
	kokoroStyleLen    = 256
	kokoroMaxSamples  = 24000 * 30 // 30s cap
	kokoroSampleRate  = 24000
)

// Kokoro is an ONNX-backed TTS backend.
type Kokoro struct {
	mu         sync.Mutex
	ctx        *C.kk_ctx
	sampleRate int
}

func newKokoro(o Options) (Backend, error) {
	if strings.TrimSpace(o.ModelPath) == "" {
		return nil, fmt.Errorf("kokoro: ModelPath is required")
	}
	cPath := C.CString(o.ModelPath)
	defer C.free(unsafe.Pointer(cPath))

	ctx := C.kk_open(cPath)
	if ctx == nil {
		return nil, fmt.Errorf("kokoro: failed to open ONNX session for %q", o.ModelPath)
	}
	rate := o.SampleRate
	if rate <= 0 {
		rate = kokoroSampleRate
	}
	return &Kokoro{ctx: ctx, sampleRate: rate}, nil
}

// Speak synthesises audio for text. Grapheme tokenisation here is a placeholder;
// production use should phonemise text (e.g. via espeak-ng) to match the model's
// vocabulary. The audio bytes are returned as a 16-bit PCM WAV.
func (k *Kokoro) Speak(ctx context.Context, text string, opts SpeakOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	k.mu.Lock()
	defer k.mu.Unlock()

	tokens := graphemeTokens(text)
	if len(tokens) == 0 {
		return Result{}, fmt.Errorf("kokoro: empty token sequence")
	}
	style := make([]C.float, kokoroStyleLen) // zero style vector by default
	out := make([]C.float, kokoroMaxSamples)

	speed := C.float(1.0)
	if opts.Speed > 0 {
		speed = C.float(opts.Speed)
	}

	cTok := C.CString(kokoroInputTokens)
	cStyle := C.CString(kokoroInputStyle)
	cSpeed := C.CString(kokoroInputSpeed)
	cOut := C.CString(kokoroOutputAudio)
	defer func() {
		C.free(unsafe.Pointer(cTok))
		C.free(unsafe.Pointer(cStyle))
		C.free(unsafe.Pointer(cSpeed))
		C.free(unsafe.Pointer(cOut))
	}()

	n := C.kk_run(k.ctx,
		cTok, (*C.longlong)(unsafe.Pointer(&tokens[0])), C.long(len(tokens)),
		cStyle, &style[0], C.long(kokoroStyleLen),
		cSpeed, speed,
		cOut, &out[0], C.long(len(out)))
	if n < 0 {
		return Result{}, fmt.Errorf("kokoro: inference failed (code %d)", int(n))
	}

	pcm := floatsToPCM(out[:int(n)])
	return Result{
		Audio:      encodeWAV(pcm, k.sampleRate),
		Format:     "wav",
		SampleRate: k.sampleRate,
	}, nil
}

// Close releases the ONNX session.
func (k *Kokoro) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.ctx != nil {
		C.kk_close(k.ctx)
		k.ctx = nil
	}
	return nil
}

// graphemeTokens is a placeholder tokeniser mapping bytes to ids. Replace with a
// phonemiser matching your Kokoro export for intelligible speech.
func graphemeTokens(text string) []C.longlong {
	tokens := make([]C.longlong, 0, len(text))
	for _, b := range []byte(text) {
		tokens = append(tokens, C.longlong(b))
	}
	return tokens
}

func floatsToPCM(samples []C.float) []int16 {
	pcm := make([]int16, len(samples))
	for i, s := range samples {
		v := float64(s)
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		pcm[i] = int16(v * 32767)
	}
	return pcm
}
