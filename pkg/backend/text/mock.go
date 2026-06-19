package text

import (
	"context"
	"strings"

	"github.com/danielriddell21/narrata/internal/prompt"
)

// Mock is a deterministic, dependency-free text backend. It produces stable
// output derived from the event and data embedded in the prompt, which makes
// it ideal for tests, golden files, and running examples without a model.
type Mock struct{}

// NewMock returns a Mock backend.
func NewMock() *Mock { return &Mock{} }

// Generate returns a deterministic narration line of the form:
//
//	<event phrase> (key=value; key=value)
//
// derived purely from the prompt's Event/Data block.
func (m *Mock) Generate(ctx context.Context, p string, _ GenerateOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	event, dataLines := prompt.ParseEventData(p)

	var b strings.Builder
	b.WriteString(humanizeEvent(event))
	if len(dataLines) > 0 {
		pairs := make([]string, 0, len(dataLines))
		for _, line := range dataLines {
			pairs = append(pairs, strings.Replace(line, ": ", "=", 1))
		}
		b.WriteString(" (")
		b.WriteString(strings.Join(pairs, "; "))
		b.WriteString(")")
	}
	b.WriteString(".")

	text := b.String()
	return Result{Text: text, Tokens: len(strings.Fields(text))}, nil
}

// Close is a no-op.
func (m *Mock) Close() error { return nil }

// humanizeEvent turns a snake_case event name into a readable phrase.
func humanizeEvent(event string) string {
	if event == "" {
		return "Event"
	}
	words := strings.FieldsFunc(event, func(r rune) bool { return r == '_' || r == '-' })
	if len(words) == 0 {
		return event
	}
	words[0] = strings.ToUpper(words[0][:1]) + words[0][1:]
	return strings.Join(words, " ")
}
