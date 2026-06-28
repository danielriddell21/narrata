package narrata

import (
	"github.com/danielriddell21/narrata/backend/text"
	"github.com/danielriddell21/narrata/backend/tts"
)

// Public backend interface and value-type aliases. These let host apps and
// custom backends refer to narrata.TextBackend / narrata.TTSBackend etc. while
// the implementations live in the backend subpackages.
type (
	// TextBackend is the local text-generation contract.
	TextBackend = text.Backend
	// GenerateOptions are per-call text sampling parameters.
	GenerateOptions = text.GenerateOptions
	// TextResult is a text backend's output.
	TextResult = text.Result

	// TTSBackend is the optional speech-synthesis contract.
	TTSBackend = tts.Backend
	// SpeakOptions are per-call voice parameters.
	SpeakOptions = tts.SpeakOptions
	// AudioResult is a TTS backend's output.
	AudioResult = tts.Result
)
