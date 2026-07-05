package text

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/danielriddell21/narrata/internal/prompt"
	"github.com/danielriddell21/narrata/nlg"
)

// Native is a pure-Go text backend built on the nlg library: it composes
// persona-flavoured narration from the event, data, and style with no model
// file and no cgo. Output is deterministic for a given input.
//
// It is a thin adapter — the generation logic lives in
// github.com/danielriddell21/narrata/nlg, which is usable on its own.
type Native struct{ client *nlg.Client }

// NewNative returns a Native backend.
func NewNative() *Native {
	c, _ := nlg.New() // no options never errors
	return &Native{client: c}
}

// Generate parses the prompt's event/data/tone and delegates to nlg.
func (n *Native) Generate(ctx context.Context, p string, _ GenerateOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("native backend: %w", err)
	}
	event, dataLines := prompt.ParseEventData(p)
	style := parseStyle(p)

	res, err := n.client.Generate(ctx, nlg.Task{
		Intent:      nlg.Narrate,
		Event:       event,
		Fields:      fieldsFromLines(dataLines),
		Style:       &style,
		Constraints: parseConstraints(p),
	})
	if err != nil {
		return Result{}, fmt.Errorf("native backend: %w", err)
	}
	return Result{Text: res.Text, Tokens: len(strings.Fields(res.Text))}, nil
}

// GenerateStructured produces narration from typed inputs, bypassing the prompt
// string so numbers stay numeric and the full persona style is used.
func (n *Native) GenerateStructured(ctx context.Context, in StructuredInput, _ GenerateOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("native backend: %w", err)
	}
	res, err := n.client.Generate(ctx, nlg.Task{
		Intent:      nlg.Narrate,
		Event:       in.Event,
		Data:        in.Data,
		Hint:        in.Instruction,
		Style:       &nlg.Style{Tone: in.Style.Tone, Energy: in.Style.Energy, Humour: in.Style.Humour, Verbosity: in.Style.Verbosity},
		Constraints: nlg.Constraints{MaxWords: in.MaxWords, MaxSentences: in.MaxSentences, AllowMarkdown: in.AllowMarkdown},
	})
	if err != nil {
		return Result{}, fmt.Errorf("native backend: %w", err)
	}
	return Result{Text: res.Text, Tokens: len(strings.Fields(res.Text))}, nil
}

// Close is a no-op.
func (n *Native) Close() error { return nil }

// parseStyle extracts the persona style from the prompt's "Style: tone=...,
// energy=..., humour=..., verbosity=..." line.
func parseStyle(p string) nlg.Style {
	return nlg.Style{
		Tone:      styleField(p, "tone"),
		Energy:    styleField(p, "energy"),
		Humour:    styleField(p, "humour"),
		Verbosity: styleField(p, "verbosity"),
	}
}

// parseConstraints reads the prompt's "Constraints: max_words=..,
// max_sentences=.." line so nlg can trim cleanly (whole details) rather than
// leaving the engine's post-processing to cut mid-phrase.
func parseConstraints(p string) nlg.Constraints {
	return nlg.Constraints{
		MaxWords:     styleFieldInt(p, "max_words"),
		MaxSentences: styleFieldInt(p, "max_sentences"),
	}
}

func styleFieldInt(p, key string) int {
	n, _ := strconv.Atoi(styleField(p, key))
	return n
}

func styleField(p, key string) string {
	i := strings.Index(p, key+"=")
	if i < 0 {
		return ""
	}
	rest := p[i+len(key)+1:]
	for j, r := range rest {
		if r == ',' || r == '\n' {
			return strings.TrimSpace(rest[:j])
		}
	}
	return strings.TrimSpace(rest)
}

// fieldsFromLines turns "key: value" prompt lines back into nlg fields. Types
// are recovered by nlg (numeric strings become quantities, etc.).
func fieldsFromLines(lines []string) []nlg.Field {
	fs := make([]nlg.Field, 0, len(lines))
	for _, l := range lines {
		k, v, ok := strings.Cut(l, ": ")
		if !ok {
			k, v = l, ""
		}
		fs = append(fs, nlg.Field{Key: k, Value: v})
	}
	return fs
}
