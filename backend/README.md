# Backends

Narrata's text and TTS backends sit behind interfaces and are selected by name
via `Config.Text.Backend` / `Config.TTS.Backend`.

## Always available (pure Go, no build tags)

| Kind | Name       | Notes |
|------|------------|-------|
| Text | `mock`     | Deterministic; default when no backend is set. Ideal for tests. |
| Text | `template` | Pure-Go sentence composer; a real local generator with no model file. |
| TTS  | `mock`     | Emits a valid 16-bit PCM WAV derived from the text. |

These compile and run everywhere with `go build ./...` — no cgo, no model files.

## Experimental backends (behind build tags)

These are feasibility implementations. They are excluded from the default build
and from CI so the pure-Go path stays green. The C APIs they target evolve
between upstream releases, so expect to adjust signatures/tensor names to match
your local build.

### `llama.cpp` text backend — `-tags llama`

Requires a local [llama.cpp](https://github.com/ggml-org/llama.cpp) build
providing `llama.h` and `libllama`.

```bash
git clone https://github.com/ggml-org/llama.cpp third_party/llama.cpp
cmake -S third_party/llama.cpp -B third_party/llama.cpp/build
cmake --build third_party/llama.cpp/build --config Release

go build -tags llama ./...
```

Then:

```go
narrata.New(narrata.Config{
    Text: narrata.TextConfig{Backend: "llama.cpp", ModelPath: "./models/model.gguf"},
})
```

### `kokoro` TTS backend — `-tags kokoro`

Requires [ONNX Runtime](https://github.com/microsoft/onnxruntime) (headers +
`libonnxruntime`) under `third_party/onnxruntime/` and a Kokoro ONNX export.
Tensor names and the style-vector layout are defined as constants in
`tts/kokoro.go`; adjust them to match your model. The grapheme tokeniser is a
placeholder — phonemise text (e.g. via espeak-ng) for intelligible speech.

```bash
go build -tags kokoro ./...
```

The include/library search paths are set in each file's cgo directives via
`${SRCDIR}` and can be overridden with `CGO_CFLAGS` / `CGO_LDFLAGS`.
