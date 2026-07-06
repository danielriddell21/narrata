package text

import "context"

// Style carries persona style for structured generation.
type Style struct {
	Tone      string
	Energy    string
	Humour    string
	Verbosity string
}

// StructuredInput carries typed generation inputs, letting a capable backend
// skip the rendered prompt string and its lossy round-trip (numbers stay
// numbers, the full style is available, constraints are explicit).
type StructuredInput struct {
	Event       string
	Data        any
	Instruction string
	Style       Style
	Rules       []string
	// Examples maps an event name to a persona-authored template (with {field}
	// placeholders) used verbatim for that event.
	Examples      map[string]string
	MaxWords      int
	MaxSentences  int
	AllowMarkdown bool
}

// Structured is an optional backend capability. Backends that implement it (the
// native/nlg backend does) receive typed inputs; the engine falls back to the
// prompt-string path for backends that do not.
type Structured interface {
	GenerateStructured(ctx context.Context, in StructuredInput, opts GenerateOptions) (Result, error)
}
