package narrata

import "time"

// OutputMode selects what an Engine returns for a Request.
type OutputMode int

const (
	// OutputText returns generated text only. This is the default.
	OutputText OutputMode = iota
	// OutputSpeech returns audio only (requires TTS enabled).
	OutputSpeech
	// OutputTextSpeech returns both text and audio (requires TTS enabled).
	OutputTextSpeech
	// OutputStreamText is reserved for token streaming; not implemented in MVP
	// and currently behaves like OutputText.
	OutputStreamText
)

// String returns the lowercase wire name of the mode.
func (m OutputMode) String() string {
	switch m {
	case OutputText:
		return "text"
	case OutputSpeech:
		return "speech"
	case OutputTextSpeech:
		return "text_speech"
	case OutputStreamText:
		return "stream_text"
	default:
		return "text"
	}
}

// wantsText reports whether the mode produces text.
func (m OutputMode) wantsText() bool {
	return m == OutputText || m == OutputTextSpeech || m == OutputStreamText
}

// wantsSpeech reports whether the mode produces audio.
func (m OutputMode) wantsSpeech() bool {
	return m == OutputSpeech || m == OutputTextSpeech
}

// Format controls the surface shape of generated text.
type Format int

const (
	// FormatPlain is unadorned prose (default).
	FormatPlain Format = iota
	// FormatMarkdown permits markdown in the output.
	FormatMarkdown
)

// Request is a single narration request. The host supplies an event and its
// data; Narrata returns narration shaped by the selected persona.
type Request struct {
	// PersonaID selects the persona. Empty uses the configured default.
	PersonaID string
	// Event is a short machine-readable name for what happened, e.g.
	// "door_opened" or "service_degraded".
	Event string
	// Data is the structured payload to narrate (map, struct, or any
	// JSON-serialisable value). It is compacted before prompting.
	Data any
	// Instruction is an optional one-off steering note for this request. It
	// never grants assistant/chat behaviour; it only nudges narration.
	Instruction string
	// Output selects text, speech, or both.
	Output OutputMode
	// Constraints override persona constraints for this request.
	Constraints Constraints
	// EventPolicy shapes rendering for this event: importance and urgency adjust
	// the length and terseness, and a cooldown can suppress repeat output when
	// AllowSilence is set.
	EventPolicy EventPolicy
}

// Constraints bound and shape generated text. Zero-valued fields fall back to
// the persona's [PersonaConstraints].
type Constraints struct {
	// MaxWords caps the output word count. Zero defers to the persona.
	MaxWords int
	// MaxSentences caps the output sentence count. Zero defers to the persona.
	MaxSentences int
	// AllowProfanity permits profanity for this request.
	AllowProfanity bool
	// Format selects the output surface (plain prose or markdown).
	Format Format
}

// EventPolicy expresses how an event should be rendered. It is rendering
// policy, not autonomous behaviour: Narrata never schedules or acts on it.
type EventPolicy struct {
	// Importance is the host-assigned importance of the event.
	Importance string
	// Urgency is the host-assigned urgency of the event.
	Urgency string
	// Cooldown is the minimum gap the host suggests between renderings.
	Cooldown time.Duration
	// AllowSilence permits Narrata to render nothing for this event.
	AllowSilence bool
}
