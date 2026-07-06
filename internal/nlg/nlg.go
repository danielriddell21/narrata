package nlg

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrNeedsModel = errors.New("nlg: task needs a language model; configure WithFallback")

type Style struct {
	Tone      string
	Energy    string
	Humour    string
	Verbosity string
}

type Constraints struct {
	MaxWords      int
	MaxSentences  int
	AllowMarkdown bool
}

type Persona struct {
	ID          string
	Style       Style
	Rules       []string
	Constraints Constraints // defaults; a Task's Constraints override these
}

type Intent string

const (
	Narrate   Intent = "narrate"
	Describe  Intent = "describe"
	Summarize Intent = "summarize"
	Classify  Intent = "classify"
	Extract   Intent = "extract"
)

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
	// Examples maps an event name to a persona-authored template. When the task's
	// event matches, the template is used (with {field} placeholders filled from
	// Data) instead of the grammar, which lets a persona pin an exact
	// hand-written line to a known event.
	Examples map[string]string
	Seed     uint64 // determinism; 0 derives a seed from the task
}

type Result struct {
	Text  string
	Label string // set for Classify
}

type Backend interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type Client struct {
	personas map[string]Persona
	def      string
	fallback Backend
	openers  map[string][]string // per-tone opener overrides
}

type Option func(*Client) error

func New(opts ...Option) (*Client, error) {
	c := &Client{personas: make(map[string]Persona)}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func WithPersona(ps ...Persona) Option {
	return func(c *Client) error {
		for _, p := range ps {
			c.personas[p.ID] = p
		}
		return nil
	}
}

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

func WithDefaultPersona(id string) Option {
	return func(c *Client) error { c.def = id; return nil }
}

func WithFallback(b Backend) Option {
	return func(c *Client) error { c.fallback = b; return nil }
}

func WithOpeners(tone string, openers []string) Option {
	return func(c *Client) error {
		if c.openers == nil {
			c.openers = make(map[string][]string)
		}
		c.openers[tone] = openers
		return nil
	}
}

func (c *Client) openerPool(style Style) []string {
	pool, ok := c.openers[style.Tone]
	if !ok {
		if pool, ok = openersByTone[style.Tone]; !ok {
			pool = openersByTone["neutral"]
		}
	}
	return gateOpeners(pool, style)
}

func (c *Client) Persona(id string) (Persona, bool) {
	p, ok := c.personas[id]
	return p, ok
}

func (c *Client) Generate(ctx context.Context, t Task) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("nlg: %w", err)
	}
	intent := t.Intent
	if intent == "" {
		intent = Narrate
	}

	persona, style := c.resolve(t)
	cons := mergeConstraints(persona.Constraints, t.Constraints)

	switch intent {
	case Narrate:
		return Result{Text: c.narrate(t, style, cons)}, nil
	case Describe:
		return Result{Text: c.describe(t, style, cons)}, nil
	case Summarize:
		return Result{Text: c.summarize(t, style, cons)}, nil
	case Classify:
		return c.classify(ctx, t)
	case Extract:
		return Result{Text: c.extract(t)}, nil
	default:
		return c.viaFallback(ctx, t)
	}
}

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

func (c *Client) exampleLine(t Task, style Style, cons Constraints) (string, bool) {
	tmpl, ok := t.Examples[t.Event]
	if !ok || tmpl == "" {
		return "", false
	}
	out, filled := renderExample(tmpl, withKinds(fieldsOf(t)))
	if !filled || out == "" {
		return "", false
	}
	return terminate(applyBudget(capitalise(out), cons), style), true
}

func (c *Client) narrate(t Task, style Style, cons Constraints) string {
	if line, ok := c.exampleLine(t, style, cons); ok {
		return line
	}
	r := newRNG(seedFor(t, style))
	fields := withKinds(fieldsOf(t))

	sh, noun, action := classifyEvent(t.Event)
	if sh == shapeGeneric {
		fields = salient(fields, maxFields(style, cons))
		return terminate(compose(c.openerPool(style), humanizeEvent(t.Event), realizeAll(fields), r, cons), style)
	}

	subject, rest, matched := subjectPhrase(noun, fields)
	// For a generic verb action with no matching noun field, a name reads as the
	// actor ("goal_scored" + team Rovers -> "Rovers scored").
	if sh == shapeAction && !matched {
		if name, r2, ok := takeFirstName(fields); ok {
			subject, rest = name, r2
		}
	}
	rest = salient(rest, maxFields(style, cons))

	var clause string
	if sh == shapeAction {
		clause = capitalise(subject + " " + action)
	} else {
		clause = capitalise(clauseFor(sh, action, subject, toneBucket(style.Tone), r))
	}
	// Place/time adjuncts read as part of the action, so attach them to the verb
	// clause ("done in the utility room"); the rest form an em-dash detail list.
	adjuncts, details := partitionAdjuncts(rest)
	if adj := fitPhrases(clause, realizeAll(adjuncts), cons); len(adj) > 0 {
		head, aside := splitTrailingAside(clause)
		clause = head + " " + strings.Join(adj, " ") + aside
	}
	// A verbose voice spells the details out as a second sentence ("... is
	// struggling. CPU is at 96% and latency is 950ms."); others keep the terse
	// em-dash list.
	if isVerbose(style) {
		if s := fitPhrases(clause, statementsOf(details), cons); len(s) > 0 {
			second := capitalise(joinList(s)) + "."
			return terminate(applyBudget(clause+". "+second, cons), style)
		}
	}
	if p := fitPhrases(clause, realizeAll(details), cons); len(p) > 0 {
		clause += " — " + joinList(p)
	}
	return terminate(applyBudget(clause+".", cons), style)
}

func (c *Client) describe(t Task, style Style, cons Constraints) string {
	if line, ok := c.exampleLine(t, style, cons); ok {
		return line
	}
	subject, rest := splitSubject(withKinds(fieldsOf(t)), t.Event)
	rest = salient(rest, maxFields(style, cons))

	// Describe is a snapshot: lead with the subject, no opener.
	line := capitalise(subject)
	if p := fitPhrases(line, realizeAll(rest), cons); len(p) > 0 {
		line += " — " + joinList(p)
	}
	return terminate(applyBudget(line+".", cons), style)
}

func (c *Client) summarize(t Task, style Style, cons Constraints) string {
	fields := salient(withKinds(fieldsOf(t)), summaryMax(style))
	phrases := realizeAll(fields)

	lead := ""
	if t.Event != "" {
		lead = humanizeEvent(t.Event)
	}
	kept := fitPhrases(lead, phrases, cons)

	var line string
	switch {
	case lead != "" && len(kept) > 0:
		line = lead + " — " + joinList(kept)
	case len(kept) > 0:
		line = joinList(kept)
	default:
		line = lead
	}
	return terminate(applyBudget(capitalise(line)+".", cons), style)
}

func summaryMax(style Style) int {
	switch style.Verbosity {
	case "short":
		return 2
	case "brief":
		return 3
	default:
		return 4
	}
}

func (c *Client) classify(ctx context.Context, t Task) (Result, error) {
	if len(t.Labels) == 0 {
		return c.viaFallback(ctx, t)
	}
	var hay strings.Builder
	hay.WriteString(strings.ToLower(t.Event + " " + t.Hint))
	for _, f := range flattenData(t.Data) {
		hay.WriteString(" ")
		hay.WriteString(strings.ToLower(f.Key + " " + valueString(f.Value)))
	}
	text := hay.String()

	best, bestScore := t.Labels[0], -1
	for _, lbl := range t.Labels {
		s := 0
		for _, w := range strings.Fields(strings.ToLower(lbl)) {
			if strings.Contains(text, w) {
				s++
			}
		}
		if s > bestScore {
			best, bestScore = lbl, s
		}
	}
	return Result{Label: best, Text: best}, nil
}

func (c *Client) extract(t Task) string {
	fields := withKinds(fieldsOf(t))
	keys := t.Labels
	var parts []string
	if len(keys) == 0 {
		for _, f := range fields {
			parts = append(parts, f.Key+"="+valueString(f.Value))
		}
		return strings.Join(parts, ", ")
	}
	for _, key := range keys {
		for _, f := range fields {
			if strings.EqualFold(f.Key, key) || strings.EqualFold(humanizeKey(f.Key), key) {
				parts = append(parts, key+"="+valueString(f.Value))
				break
			}
		}
	}
	return strings.Join(parts, ", ")
}

func fieldsOf(t Task) []Field {
	if t.Fields != nil {
		return t.Fields
	}
	return flattenData(t.Data)
}

func realizeAll(fields []Field) []string {
	ps := make([]string, 0, len(fields))
	for _, f := range fields {
		ps = append(ps, realize(f))
	}
	return ps
}

func statementsOf(fields []Field) []string {
	ps := make([]string, 0, len(fields))
	for _, f := range fields {
		ps = append(ps, realizeStatement(f))
	}
	return ps
}

func isVerbose(style Style) bool {
	return style.Verbosity == "verbose" || style.Verbosity == "detailed"
}

func splitSubject(fields []Field, event string) (string, []Field) {
	pick := -1
	if event != "" {
		for i, f := range fields {
			if strings.EqualFold(f.Key, event) {
				pick = i
				break
			}
		}
	}
	if pick < 0 {
		for i, f := range fields {
			if f.Kind == KindName {
				pick = i
				break
			}
		}
	}
	if pick >= 0 {
		f := fields[pick]
		rest := make([]Field, 0, len(fields)-1)
		rest = append(rest, fields[:pick]...)
		rest = append(rest, fields[pick+1:]...)
		return valueString(f.Value), rest
	}
	if event != "" {
		return humanizeEvent(event), fields
	}
	return "It", fields
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
		return Result{}, fmt.Errorf("nlg: fallback: %w", err)
	}
	return Result{Text: out}, nil
}

func maxFields(style Style, cons Constraints) int {
	n := 2
	switch style.Verbosity {
	case "short":
		n = 1
	case "brief":
		n = 2
	case "verbose", "detailed":
		n = 4
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
