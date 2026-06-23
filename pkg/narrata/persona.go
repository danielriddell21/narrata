package narrata

// SchemaVersion is the persona schema version this build understands. A
// personas file may declare a "version"; a file whose major version differs is
// rejected on load so schema changes can evolve safely.
const SchemaVersion = "0.1"

// Persona is a configurable behavioural and stylistic profile. Personas shape
// narration only: they have no memory, tools, goals, plans, or autonomy.
type Persona struct {
	// ID is the stable, machine-readable identifier used to select the persona.
	ID string `json:"id"`
	// Name is the human-readable persona name.
	Name string `json:"name"`
	// Description is a short explanation of the persona's intended use.
	Description string `json:"description"`
	// Style describes tone and delivery.
	Style Style `json:"style"`
	// Voice describes TTS voice preferences. Optional.
	Voice Voice `json:"voice"`
	// Rules are behavioural instructions injected into the prompt.
	Rules []string `json:"rules"`
	// Constraints are the output limits applied unless a [Request] overrides them.
	Constraints PersonaConstraints `json:"constraints"`
	// EventPolicy holds reserved rendering preferences (see roadmap Phase 5).
	EventPolicy PersonaEventPolicy `json:"event_policy"`
	// Examples are optional few-shot pairs for stronger consistency.
	Examples []Example `json:"examples,omitempty"`
}

// Style describes a persona's tone and delivery.
type Style struct {
	// Tone is the overall manner, e.g. "calm" or "dramatic".
	Tone string `json:"tone"`
	// Energy is the delivery intensity, e.g. "low", "medium", or "high".
	Energy string `json:"energy"`
	// Humour is the humour level, e.g. "none", "light", or "witty".
	Humour string `json:"humour"`
	// Verbosity is the preferred length, e.g. "short" or "brief".
	Verbosity string `json:"verbosity"`
}

// Voice describes a persona's TTS voice preferences.
type Voice struct {
	// ID names the backend voice to use.
	ID string `json:"id"`
	// Speed is the speaking-rate multiplier (1.0 is normal).
	Speed float32 `json:"speed"`
	// Pitch is the pitch multiplier (1.0 is normal).
	Pitch float32 `json:"pitch"`
}

// PersonaConstraints are the persona-level output limits, mirroring the JSON
// schema. They are applied unless a [Request] overrides them.
type PersonaConstraints struct {
	// MaxWords caps the output word count. Zero means unlimited.
	MaxWords int `json:"max_words"`
	// MaxSentences caps the output sentence count. Zero means unlimited.
	MaxSentences int `json:"max_sentences"`
	// AllowProfanity permits profanity in the output when true.
	AllowProfanity bool `json:"allow_profanity"`
	// AllowMarkdown permits markdown in the output when true.
	AllowMarkdown bool `json:"allow_markdown"`
}

// PersonaEventPolicy is the schema-level event policy. It is reserved for
// post-MVP rendering decisions and remains rendering policy, not autonomy.
type PersonaEventPolicy struct {
	// DefaultImportance is the baseline importance for the persona's events.
	DefaultImportance string `json:"default_importance"`
	// CooldownSeconds is the minimum gap between spoken renderings.
	CooldownSeconds int `json:"cooldown_seconds"`
	// AllowSilence permits the persona to render nothing for an event.
	AllowSilence bool `json:"allow_silence"`
	// Intensity is the rendering intensity, e.g. "medium".
	Intensity string `json:"intensity"`
}

// Example is an optional few-shot pair for stronger persona consistency.
type Example struct {
	// Event is the example event name.
	Event string `json:"event"`
	// Output is the desired narration for that event.
	Output string `json:"output"`
}

// PersonaStore is the registry contract for personas.
type PersonaStore interface {
	// Get returns the persona with the given id and reports whether it exists.
	Get(id string) (Persona, bool)
	// List returns all registered personas.
	List() []Persona
	// Save validates and upserts a persona.
	Save(persona Persona) error
}

// personaFile is the on-disk personas.json shape.
type personaFile struct {
	Version  string    `json:"version"`
	Default  string    `json:"default"`
	Personas []Persona `json:"personas"`
}
