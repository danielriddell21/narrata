# Narrata Design Rationale

Why Narrata generates narration **in house, in pure Go, at runtime** — with no
model, no cgo, and no external files.

## The decision

Narration in Narrata's domain is **short and structured**: a line or two derived
from a known event plus a small data payload, shaped by a persona. That domain is
well served by an authored, persona-conditioned grammar with salience selection
and constraint solving — no learned weights required. So the core generates text
procedurally, in Go, and ships as a single self-contained binary that starts
instantly and sends no data anywhere.

## Models we evaluated (and why not in the core)

| Option | Verdict |
|--------|---------|
| llama.cpp / GGUF (cgo) | Great quality, but needs a C toolchain, a native lib, and a multi-GB model file. Too heavy for short structured narration; breaks "self-contained, pure Go". |
| [born-ml/born](https://github.com/born-ml/born) (pure-Go GGUF) | Genuinely cgo-free, but pulls a large dependency tree (WebGPU stack, tokenizers). Pure Go, not lightweight. |
| Kokoro / ONNX (TTS) | Real voice, but requires ONNX Runtime and a model export. Same weight/toolchain cost. |
| Ollama sidecar | A separate service beside the app — the opposite of embedded. |

All were rejected for the **core** because each reintroduces exactly what Narrata
set out to avoid: models, native code, external files, and startup latency, for
output that doesn't need them.

## What we build instead

- **Text:** the pure-Go `native` engine (`internal/nlg`) — event-shape grammar,
  tone-conditioned clauses, salience, constraint-aware expansion, deterministic.
- **TTS:** a pure-Go `mock` WAV backend; a pure-Go synthesiser can follow, same
  no-model principle.

## The escape hatch

Open-ended tasks (free-form Q&A, novel prose) genuinely need a model. For those,
`nlg.WithFallback` accepts any `Generate(ctx, prompt) (string, error)` backend, so
a host can bring its own model **without** it becoming a core dependency. If a
lean pure-Go model path is ever wanted, the primitives to borrow (GGUF dequant,
RMSNorm/RoPE/SwiGLU, KV-cache, samplers) are well understood — but that stays
opt-in.

## Honest limits

Procedural generation restates and dramatizes host data in character. It does not
reason, understand arbitrary free text, or produce novel prose — which is exactly
Narrata's scope (narration, not chat). Breadth comes from authored variety, not
scale.
