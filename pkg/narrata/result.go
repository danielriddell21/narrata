package narrata

import "time"

// Result is the output of Engine.Generate.
type Result struct {
	// Text is the generated narration. Empty when the mode is speech-only or
	// when the request was rendered silent.
	Text string
	// Audio holds synthesised speech bytes, when requested and available.
	Audio []byte
	// AudioFormat names the encoding of Audio, e.g. "wav".
	AudioFormat string
	// PersonaID is the persona that produced this result.
	PersonaID string
	// Spoken reports whether audio was produced.
	Spoken bool
	// Silent reports whether output policy suppressed the narration.
	Silent bool
	// Duration is the wall-clock time spent generating.
	Duration time.Duration
	// Tokens is the number of tokens generated, when the backend reports it.
	Tokens int
}
