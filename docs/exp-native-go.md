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
3. **Optional tiny embedded model.** For hosts wanting more, embed a small
   quantised model's weights via `//go:embed` and run a pure-Go forward pass.
   Feasible but perf-bound (no BLAS/SIMD today; Go 1.26's experimental `simd`
   package will help). Best kept opt-in behind its own backend name.

Guiding rule: **Narrata bundles no model.** Pure-Go backends (grammar/corpus)
are the zero-asset default; if a host wants a model, it embeds the bytes in its
*own* binary via a `[]byte`/`fs.FS` model source (a small additive API on
`TextConfig`) and stays self-contained. Keeps Narrata light and license-clean.

## TTS parity (later, same principle)

- Pure-Go **formant synth** (retro/robotic — fits sci-fi/game aesthetics), or
- host-embedded phrase-bank audio for a fixed callout set.
- Neural (Kokoro/ONNX) stays the opt-in high-quality path, not a dependency.

## Relationship to `llama.cpp` / `kokoro`

Those cgo backends stay build-tagged and available for hosts that want maximum
quality and can accept the native-lib build. The intent of this branch is to
make the **pure-Go path the recommended default**, and eventually let a
self-contained binary be the out-of-the-box experience.
