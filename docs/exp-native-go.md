# Experiment: native-Go generation (no cgo / no external models)

**Branch:** `exp/pure-go-backends` · **Status:** exploratory

## Goal

Move Narrata's text generation off the cgo `llama.cpp` path and do it entirely
in Go, so a host can build a **single self-contained binary** — no native
libraries, no external model files, no cgo toolchain. The backend interface
(`backend/text.Backend`) already makes this a drop-in: new backends register by
name in `backend/text.New` with no engine changes.

This is a deliberate trade of *general LLM quality* for *zero dependencies and
instant, deterministic output* — a good fit for Narrata's short, bounded
narration (a line or two per event), and especially for games/automation that
want a self-contained build.

## What's here now: the `native` backend

`backend/text/native.go` — a pure-Go, tone-aware grammar backend selected with
`Text.Backend = "native"` (alias `"grammar"`). It:

- parses the event, data, and persona `tone` from the prompt,
- seeds a stable RNG from those inputs (deterministic per input, varied across
  inputs),
- composes an opener + event phrase + data clause, flavoured by tone.

Example (`enemy_killed`, combo 3):

```
dungeon_master      Enemy killed (combo: 3, enemy: Cacodemon).
funny_narrator      Naturally, Enemy killed — combo: 3, enemy: Cacodemon.
sports_commentator  Get this — Enemy killed (combo: 3, enemy: Cacodemon).
```

It's a clear step up from `template` (tone-aware, varied) while staying pure Go.

## Roadmap (increasing quality, all pure Go)

1. **Grammar++ (current).** Richer per-tone phrase pools, verb/subject slots
   derived from event names, light data selection (surface the most salient
   field rather than dumping all).
2. **Phrase corpus + Markov.** Ship a small embedded phrase corpus per tone
   (`//go:embed`) and stitch with a low-order Markov/template hybrid for more
   natural variation. Still tiny (KB), deterministic-seedable.
3. **Optional real model, pure Go (via Born or an in-house subset).** For hosts
   wanting genuine LLM output with **no cgo and a single binary**, run a small
   quantised GGUF through a pure-Go transformer. See the Born evaluation below.
   Perf-bound today (no BLAS; Go 1.26's experimental `simd` helps). Kept opt-in
   behind its own backend name so the default stays zero-dependency.

Guiding rule: **Narrata bundles no model.** Pure-Go backends (grammar/corpus)
are the zero-asset default; if a host wants a model, it embeds the bytes in its
*own* binary via a `[]byte`/`fs.FS` model source (a small additive API on
`TextConfig`) and stays self-contained. Keeps Narrata light and license-clean.

## Evaluated: [born-ml/born](https://github.com/born-ml/born)

Born is a **pure-Go, zero-CGO** deep-learning/inference framework — the same
"models are born production-ready, single binary, no Python/cgo" ethos we want.
It already implements the hard parts we'd otherwise build:

- GGUF loading with K-quant dequant (Q4_K/Q5_K/Q6_K/Q8_0, F16/F32),
- transformer inference: RMSNorm, RoPE/ALiBi, SwiGLU, GQA, KV-cache, Flash-
  Attention-2, speculative decoding,
- sampling: temperature / top-k / top-p / min-p / repetition penalty, streaming.

Usage is a clean fit for our `text.Backend`:

```go
be := cpu.New()
model, _ := llama.LoadGGUF("tinyllama-1.1b.Q8_0.gguf", be) // defer model.Release()
gen := generate.NewTextGenerator(model, tok, generate.SamplingConfig{Temperature: 0.7, TopP: 0.9, TopK: 40})
out, _ := gen.Generate(prompt, generate.GenerateConfig{MaxTokens: 100})
```

**The catch — it's pure Go but not lean.** `go.mod` (module
`github.com/born-ml/born`, Go 1.26) pulls a heavy tree even for CPU use: the
`gogpu/wgpu` + `gogpu/naga` + `go-webgpu/*` WebGPU stack, `tiktoken-go` +
`dlclark/regexp2`, `google/uuid`, `golang.org/x/sys`, `yaml.v3`. So it satisfies
"no cgo / single binary" but not "few dependencies."

### Recommended stance

- **Ideas to borrow directly** if we ever build a lean in-house path: their
  primitive list above (GGUF K-quant dequant, RMSNorm/RoPE/SwiGLU, KV-cache,
  the sampler set). That's the exact shopping list for a minimal pure-Go engine.
- **Integration**: add a `born` backend **behind a `born` build tag** plus an
  optional module require — same gating pattern as the cgo backends, but pure
  Go, so it's strictly better than the `llama.cpp` cgo path (no native lib, no
  compiler). The default build stays zero-dependency; hosts that want real
  local inference opt in with `-tags born` and accept Born's dependency weight.
- **Deprecate `llama.cpp`** once `born` is proven: a pure-Go, single-binary
  backend removes the cgo/native-lib burden entirely.

This gives a clean spectrum: `native` (zero-dep default) → `born` (opt-in, pure
Go, real model, single binary) → `llama.cpp` (legacy cgo, max hardware perf).

A tag-gated seam is scaffolded in `backend/text/born.go` (real, behind
`-tags born`) and `backend/text/born_stub.go` (default). The real file is a
**spike** — verify its calls against Born's current API and pick the correct
tokenizer for the model (the quickstart's tiktoken/gpt-4 is not right for a
LLaMA GGUF) before relying on it.

## TTS parity (later, same principle)

- Pure-Go **formant synth** (retro/robotic — fits sci-fi/game aesthetics), or
- host-embedded phrase-bank audio for a fixed callout set.
- Neural (Kokoro/ONNX) stays the opt-in high-quality path, not a dependency.

## Relationship to `llama.cpp` / `kokoro`

Those cgo backends stay build-tagged and available for hosts that want maximum
quality and can accept the native-lib build. The intent of this branch is to
make the **pure-Go path the recommended default**, and eventually let a
self-contained binary be the out-of-the-box experience.
