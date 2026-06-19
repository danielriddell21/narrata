// Package tts defines the optional text-to-speech backend interface and its
// implementations. TTS is pluggable and entirely optional: the text engine
// works with no TTS backend present.
package tts

import (
	"context"
	"fmt"
)

// Backend is the speech-synthesis contract.
type Backend interface {
	Speak(ctx context.Context, text string, opts SpeakOptions) (Result, error)
	Close() error
}

// SpeakOptions are per-call voice parameters. Zero values mean "use the
// backend/persona default".
type SpeakOptions struct {
	VoiceID    string
	Speed      float32
	Pitch      float32
	SampleRate int
}

// Result is synthesised audio.
type Result struct {
	Audio      []byte
	Format     string
	SampleRate int
}

// Options configures backend construction.
type Options struct {
	Backend      string
	ModelPath    string
	DefaultVoice string
	SampleRate   int
}

// New constructs a TTS backend by name. An empty name defaults to the
// deterministic mock backend.
func New(o Options) (Backend, error) {
	switch o.Backend {
	case "", "mock":
		return NewMock(o.SampleRate), nil
	case "kokoro":
		return newKokoro(o)
	default:
		return nil, fmt.Errorf("tts: unknown backend %q", o.Backend)
	}
}
