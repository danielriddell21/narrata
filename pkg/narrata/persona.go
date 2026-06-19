package narrata

// Persona is a configurable behavioural and stylistic profile. Personas shape
// narration only: they have no memory, tools, goals, plans, or autonomy.
type Persona struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Style       Style             `json:"style"`
	Voice       Voice             `json:"voice"`
	Rules       []string          `json:"rules"`
	Constraints PersonaConstraints `json:"constraints"`
	EventPolicy PersonaEventPolicy `json:"event_policy"`
	Examples    []Example         `json:"examples,omitempty"`
}

// Style describes tone and delivery.
type Style struct {
	Tone      string `json:"tone"`
	Energy    string `json:"energy"`
	Humour    string `json:"humour"`
	Verbosity string `json:"verbosity"`
}

// Voice describes TTS voice preferences.
type Voice struct {
	ID    string  `json:"id"`
	Speed float32 `json:"speed"`
	Pitch float32 `json:"pitch"`
}

// PersonaConstraints are the persona-level output limits, mirroring the JSON
// schema. They are applied unless a Request overrides them.
type PersonaConstraints struct {
	MaxWords       int  `json:"max_words"`
	MaxSentences   int  `json:"max_sentences"`
	AllowProfanity bool `json:"allow_profanity"`
	AllowMarkdown  bool `json:"allow_markdown"`
}

// PersonaEventPolicy is the schema-level event policy. It is reserved for
// post-MVP rendering decisions and remains rendering policy, not autonomy.
type PersonaEventPolicy struct {
	DefaultImportance string `json:"default_importance"`
	CooldownSeconds   int    `json:"cooldown_seconds"`
	AllowSilence      bool   `json:"allow_silence"`
	Intensity         string `json:"intensity"`
}

// Example is an optional few-shot pair for stronger persona consistency.
type Example struct {
	Event  string `json:"event"`
	Output string `json:"output"`
}

// PersonaStore is the registry contract for personas.
type PersonaStore interface {
	Get(id string) (Persona, bool)
	List() []Persona
	Save(persona Persona) error
}

// personaFile is the on-disk personas.json shape.
type personaFile struct {
	Version  string    `json:"version"`
	Default  string    `json:"default"`
	Personas []Persona `json:"personas"`
}
