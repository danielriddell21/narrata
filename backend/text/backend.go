// Package text defines the local text-generation backend interface and its
// implementations. Backends are selected by name and sit behind the Backend
// interface so the runtime stays model-agnostic.
package text

import (
	"context"
	"fmt"
)

// Backend is the local text-generation contract. Implementations must be safe
// for the Engine's concurrency model (the Engine serialises/limits calls via
// MaxConcurrent; a backend may add its own guarding).
type Backend interface {
	// Generate produces text for the given prompt, honouring opts and ctx
	// cancellation.
	Generate(ctx context.Context, prompt string, opts GenerateOptions) (Result, error)
	// Close releases any resources held by the backend.
	Close() error
}

// GenerateOptions are per-call sampling parameters. Zero values mean "use the
// backend default".
type GenerateOptions struct {
	// Temperature controls sampling randomness.
	Temperature float32
	// TopP controls nucleus sampling.
	TopP float32
	// MaxTokens caps the number of generated tokens.
	MaxTokens int
	// ContextTokens is the context-window size hint.
	ContextTokens int
}

// Result is a backend's generation output.
type Result struct {
	// Text is the generated text.
	Text string
	// Tokens is the number of tokens generated, when known.
	Tokens int
}

// Options configures backend construction. It is populated by the engine from
// the host's text configuration.
type Options struct {
	// Backend names the implementation: "mock", "template", or "llama.cpp".
	Backend string
	// ModelPath is the path to a model file (used by "llama.cpp").
	ModelPath string
	// ContextTokens is the model context window. Zero uses the backend default.
	ContextTokens int
	// Temperature controls sampling randomness. Zero uses the backend default.
	Temperature float32
	// TopP controls nucleus sampling. Zero uses the backend default.
	TopP float32
	// MaxTokens caps generated tokens. Zero uses the backend default.
	MaxTokens int
}

// New constructs a backend by name. An empty name defaults to the deterministic
// mock backend so apps can run without a model file.
func New(o Options) (Backend, error) {
	switch o.Backend {
	case "", "mock":
		return NewMock(), nil
	case "template":
		return NewTemplate(), nil
	case "native", "grammar":
		return NewNative(), nil
	case "born":
		return newBorn(o)
	case "llama.cpp", "llama":
		return newLlama(o)
	default:
		return nil, fmt.Errorf("text: unknown backend %q", o.Backend)
	}
}
