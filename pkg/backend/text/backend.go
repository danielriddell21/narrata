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
	Generate(ctx context.Context, prompt string, opts GenerateOptions) (Result, error)
	Close() error
}

// GenerateOptions are per-call sampling parameters. Zero values mean "use the
// backend default".
type GenerateOptions struct {
	Temperature   float32
	TopP          float32
	MaxTokens     int
	ContextTokens int
}

// Result is a backend's generation output.
type Result struct {
	Text   string
	Tokens int
}

// Options configures backend construction.
type Options struct {
	Backend       string
	ModelPath     string
	ContextTokens int
	Temperature   float32
	TopP          float32
	MaxTokens     int
}

// New constructs a backend by name. An empty name defaults to the deterministic
// mock backend so apps can run without a model file.
func New(o Options) (Backend, error) {
	switch o.Backend {
	case "", "mock":
		return NewMock(), nil
	case "template":
		return NewTemplate(), nil
	case "llama.cpp", "llama":
		return newLlama(o)
	default:
		return nil, fmt.Errorf("text: unknown backend %q", o.Backend)
	}
}
