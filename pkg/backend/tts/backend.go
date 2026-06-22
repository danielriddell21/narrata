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
	// Speak synthesises audio for text, honouring opts and ctx cancellation.
	Speak(ctx context.Context, text string, opts SpeakOptions) (Result, error)
	// Close releases any resources held by the backend.
	Close() error
}

// SpeakOptions are per-call voice parameters. Zero values mean "use the
// backend/persona default".
type SpeakOptions struct {
	// VoiceID names the voice to synthesise with.
	VoiceID string
	// Speed is the speaking-rate multiplier (1.0 is normal).
	Speed float32
	// Pitch is the pitch multiplier (1.0 is normal).
	Pitch float32
	// SampleRate overrides the output sample rate in Hz.
	SampleRate int
}

// Result is synthesised audio.
type Result struct {
	// Audio holds the encoded audio bytes.
	Audio []byte
	// Format names the audio encoding, e.g. "wav".
	Format string
	// SampleRate is the output sample rate in Hz.
	SampleRate int
}

// Options configures backend construction. It is populated by the engine from
// the host's TTS configuration.
type Options struct {
	// Backend names the implementation: "mock" or "kokoro".
	Backend string
	// ModelPath is the path to the TTS model file.
	ModelPath string
	// DefaultVoice is used when a persona declares no voice.
	DefaultVoice string
	// SampleRate is the output sample rate in Hz. Zero uses the backend default.
	SampleRate int
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
