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

> **Chosen direction:** zero-weights procedural generation (grammar + runtime
> statistics + constrained decoding), evolving `native` into a standalone
> pure-Go library. Full design: [design-native-nlg.md](./design-native-nlg.md).

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

### Outcome

We **did not** integrate Born as a backend — its dependency weight is at odds
with Narrata's "few dependencies" goal, and the direction we chose ([the `nlg`
library](./design-native-nlg.md)) needs no model at all. A `born` backend seam
was prototyped and then removed.

Born stays a **reference** for the day we want a lean, in-house pure-Go *model*
path: borrow its primitive list above (GGUF K-quant dequant, RMSNorm/RoPE/SwiGLU,
KV-cache, the sampler set) rather than the framework. Until then, hosts that need
a real model can plug one in via `nlg`'s `WithFallback`.

The cgo backends (`llama.cpp`, `kokoro`) have been **removed** on this branch —
the runtime is now pure Go end to end. Hosts that need a real model plug one in
via `nlg.WithFallback`.

## TTS parity (later, same principle)

- Pure-Go **formant synth** (retro/robotic — fits sci-fi/game aesthetics), or
- host-embedded phrase-bank audio for a fixed callout set.
- The `mock` TTS backend remains for the optional speech path.
