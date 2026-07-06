package narrata

import "time"

// Config configures an [Engine]. The zero value is usable: with no text backend
// specified, the Engine defaults to the deterministic "mock" backend so host
// apps can wire up and test with zero setup.
type Config struct {
	// PersonasPath is an optional path to a single personas.json file.
	PersonasPath string
	// PersonasDir is an optional directory of one-persona-per-file *.json files.
	PersonasDir string
	// DefaultPersona overrides the persona used when a Request omits PersonaID.
	// When empty, the default declared by the loaded persona set is used.
	DefaultPersona string

	// Text configures the local text-generation backend.
	Text TextConfig
	// TTS configures the optional text-to-speech backend. Disabled by default.
	TTS TTSConfig

	// MaxConcurrent bounds in-flight Generate calls. Zero means unbounded.
	MaxConcurrent int
	// Timeout is the default per-request deadline. Zero means no default
	// deadline (the caller's context still applies).
	Timeout time.Duration
}

// TextConfig selects and tunes the text backend.
type TextConfig struct {
	// Backend names the implementation: "mock" (default), "template", or
	// "native" (persona-aware pure-Go generation).
	Backend string
	// ContextTokens bounds the generation context. Zero uses the backend default.
	ContextTokens int
	// Temperature controls sampling randomness. Zero uses the backend default.
	Temperature float32
	// TopP controls nucleus sampling. Zero uses the backend default.
	TopP float32
	// MaxTokens caps generated tokens. Zero uses the backend default.
	MaxTokens int
}

// TTSConfig selects and tunes the optional speech backend.
type TTSConfig struct {
	// Enabled turns speech synthesis on. When false, requests that ask for
	// speech fail with ErrTTSUnavailable and text-only requests are unaffected.
	Enabled bool
	// Backend names the implementation: "mock".
	Backend string
	// DefaultVoice is used when a persona declares no voice.
	DefaultVoice string
	// SampleRate is the output sample rate in Hz. Zero uses the backend default.
	SampleRate int
}
