# Narrata Design Rationale

Why Narrata generates narration in house, in pure Go, at runtime, with no
model, no cgo and no external files.

## The decision

Narration in Narrata's domain is short and structured: a line or two derived
from a known event plus a small data payload, shaped by a persona. An authored,
persona-conditioned grammar with salience selection and constraint solving
covers that domain without any learned weights. So the core generates text
procedurally, in Go, and ships as a single self-contained binary that starts
fast and sends no data anywhere.

## Models we evaluated (and why not in the core)

| Option | Verdict |
|--------|---------|
| llama.cpp / GGUF (cgo) | Great quality, but needs a C toolchain, a native lib, and a multi-GB model file. Too heavy for short structured narration, and breaks "self-contained, pure Go". |
| [born-ml/born](https://github.com/born-ml/born) (pure-Go GGUF) | Actually cgo-free, but pulls a large dependency tree (WebGPU stack, tokenizers). Pure Go, not lightweight. |
| Kokoro / ONNX (TTS) | Real voice, but requires ONNX Runtime and a model export. Same weight/toolchain cost. |
| Ollama sidecar | A separate service beside the app, which is the opposite of embedded. |

Each of these reintroduces what Narrata set out to avoid (models, native code,
external files, startup latency) for output that doesn't need them, so none
made it into the core.

## What we build instead

For text, the pure-Go `native` engine in `internal/nlg`: event-shape grammar,
tone-conditioned clauses, salience, constraint-aware expansion, deterministic
output. For TTS, a pure-Go `mock` WAV backend for now; a real pure-Go
synthesiser can follow under the same no-model rule.

## The escape hatch

Open-ended tasks like free-form Q&A or novel prose do need a model. For those,
`nlg.WithFallback` accepts any `Generate(ctx, prompt) (string, error)` backend,
so a host can bring its own model without it becoming a core dependency. If a
lean pure-Go model path is ever wanted, the primitives to borrow (GGUF dequant,
RMSNorm/RoPE/SwiGLU, KV-cache, samplers) are well understood, but that stays
opt-in.

## Limits

Procedural generation restates and dramatizes host data in character. It does
not reason, understand arbitrary free text, or produce novel prose. That
matches Narrata's scope (narration, not chat). Breadth comes from authored
variety, not scale.
