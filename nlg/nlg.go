// Package nlg is a pure-Go, dependency-free, persona-aware text generator with
// an LLM-shaped call site. It turns a structured task (an event plus data) into
// a short, in-character line using procedural generation — an authored,
// persona-conditioned grammar plus salience selection and constraint solving.
// There are no trained weights, no model files, and no cgo.
//
// nlg covers the structured subset of "things people use an LLM for"
// (narrate/describe now; summarize/classify/extract planned). For open-ended
// tasks that genuinely need a language model, register a [Backend] with
// [WithFallback]; nlg routes those there rather than faking them, returning
// [ErrNeedsModel] when no fallback is set.
//
// The package imports only the standard library and knows nothing about the
// rest of Narrata, so it can be used on its own or extracted into its own
// module unchanged.
package nlg

import (
	"context"
	"errors"
)

// ErrNeedsModel is returned when a task needs a real language model but no
// fallback [Backend] is configured.
var ErrNeedsModel = errors.New("nlg: task needs a language model; configure WithFallback")

// Style conditions tone and delivery (mirrors a Narrata persona's style).
type Style struct {
	Tone      string
	Energy    string
	Humour    string
	Verbosity string
}

// Constraints bound the output. Zero fields mean "unbounded".
type Constraints struct {
	MaxWords      int
	MaxSentences  int
	AllowMarkdown bool
}

// Persona is a specifiable style profile. Define it in Go or load it from JSON
// with [ParsePersonas] / [WithPersonasJSON].
type Persona struct {
	ID          string
	Style       Style
	Rules       []string
	Constraints Constraints // defaults; a Task's Constraints override these
}

// Intent selects the kind of generation.
type Intent string

// Supported intents. Narrate and Describe are implemented; the others are
// reserved and route to the fallback until implemented.
const (
	Narrate   Intent = "narrate"
	Describe  Intent = "describe"
	Summarize Intent = "summarize"
	Classify  Intent = "classify"
	Extract   Intent = "extract"
)

// Task is the LLM-shaped request: structured in, text out.
type Task struct {
	Persona     string   // persona id; empty uses the default (or inline Style)
	Intent      Intent   // defaults to Narrate
	Event       string   // event/label for narrate/describe
	Data        any      // structured payload (map, struct, JSON value)
	Fields      []Field  // pre-structured fields; used instead of Data when set
	Labels      []string // candidate labels for Classify
	Hint        string   // optional one-off steer; also the fallback prompt
	Style       *Style   // optional inline style, overriding Persona lookup
	Constraints Constraints
	Seed        uint64 // determinism; 0 derives a seed from the task
}

// Result is the generated output.
type Result struct {
	Text  string
	Label string // set for Classify
}

// Backend is the escape hatch for tasks that need a real language model.
// Narrata's llama.cpp / born text backends satisfy this shape.
type Backend interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// Client generates text from tasks. It is safe for concurrent use.
type Client struct {
	personas map[string]Persona
	def      string
	fallback Backend
}

// Option configures a Client.
type Option func(*Client) error

// New builds a Client from options.
func New(opts ...Option) (*Client, error) {
	c := &Client{personas: make(map[string]Persona)}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithPersona registers one or more personas.
func WithPersona(ps ...Persona) Option {
	return func(c *Client) error {
		for _, p := range ps {
			c.personas[p.ID] = p
		}
		return nil
	}
}

// WithPersonasJSON registers personas from a personas.json document (Narrata
// schema). The document's default persona becomes the Client default.
func WithPersonasJSON(data []byte) Option {
	return func(c *Client) error {
		ps, def, err := ParsePersonas(data)
		if err != nil {
			return err
		}
		for _, p := range ps {
			c.personas[p.ID] = p
		}
		if def != "" {
			c.def = def
		}
		return nil
	}
}

// WithDefaultPersona sets the persona used when a Task omits one.
func WithDefaultPersona(id string) Option {
	return func(c *Client) error { c.def = id; return nil }
}

// WithFallback sets the model backend used for open-ended intents.
func WithFallback(b Backend) Option {
	return func(c *Client) error { c.fallback = b; return nil }
}

// Persona returns a registered persona by id.
func (c *Client) Persona(id string) (Persona, bool) {
	p, ok := c.personas[id]
	return p, ok
}

// Generate renders a task into text.
func (c *Client) Generate(ctx context.Context, t Task) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	intent := t.Intent
	if intent == "" {
		intent = Narrate
	}

	persona, style := c.resolve(t)
	cons := mergeConstraints(persona.Constraints, t.Constraints)

	switch intent {
	case Narrate, Describe:
		return Result{Text: c.narrate(t, style, cons)}, nil
	default:
		return c.viaFallback(ctx, t)
	}
}

// resolve picks the effective persona and style for a task.
func (c *Client) resolve(t Task) (Persona, Style) {
	id := t.Persona
	if id == "" {
		id = c.def
	}
	p := c.personas[id] // zero Persona if unknown
	if t.Style != nil {
		return p, *t.Style
	}
	return p, p.Style
}

// narrate builds a line from the task's event and data.
func (c *Client) narrate(t Task, style Style, cons Constraints) string {
	fields := t.Fields
	if fields == nil {
		fields = flattenData(t.Data)
	}
	fields = withKinds(fields)
	fields = salient(fields, maxFields(style, cons))

	phrases := make([]string, 0, len(fields))
	for _, f := range fields {
		phrases = append(phrases, realize(f))
	}

	r := newRNG(seedFor(t, style))
	return compose(style.Tone, humanizeEvent(t.Event), phrases, r, cons)
}

func (c *Client) viaFallback(ctx context.Context, t Task) (Result, error) {
	if c.fallback == nil {
		return Result{}, ErrNeedsModel
	}
	prompt := t.Hint
	if prompt == "" {
		prompt = t.Event
	}
	out, err := c.fallback.Generate(ctx, prompt)
	if err != nil {
		return Result{}, err
	}
	return Result{Text: out}, nil
}

// maxFields chooses how many data fields to surface, from verbosity and budget.
func maxFields(style Style, cons Constraints) int {
	n := 2
	switch style.Verbosity {
	case "short":
		n = 1
	case "brief":
		n = 2
	}
	if cons.MaxWords > 0 && cons.MaxWords < 12 {
		n = 1
	}
	return n
}

func mergeConstraints(base, over Constraints) Constraints {
	if over.MaxWords > 0 {
		base.MaxWords = over.MaxWords
	}
	if over.MaxSentences > 0 {
		base.MaxSentences = over.MaxSentences
	}
	if over.AllowMarkdown {
		base.AllowMarkdown = true
	}
	return base
}
