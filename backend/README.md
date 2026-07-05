# Backends

Narrata's text and TTS backends sit behind interfaces and are selected by name
via `Config.Text.Backend` / `Config.TTS.Backend`. They are pure Go — no cgo, no
model files, no build tags.

## Text backends

| Name       | Notes |
|------------|-------|
| `native` (alias `grammar`) | Persona-aware pure-Go generation (event-shape sentences, tone-conditioned) built on the [`nlg`](../internal/nlg) library. |
| `template` | Simple pure-Go sentence composer. |
| `mock`     | Deterministic; the default when no backend is set. Ideal for tests. |

## TTS backends

| Name   | Notes |
|--------|-------|
| `mock` | Emits a valid 16-bit PCM WAV derived from the text. Optional; text works without it. |

All backends compile and run everywhere with `go build ./...`.

## Bringing your own model

For tasks that genuinely need a language model, implement the `nlg.Backend`
interface (`Generate(ctx, prompt) (string, error)`) around any inference engine
and register it with `nlg.WithFallback`, or wrap it as a text backend. Narrata
itself ships no model and no cgo.
