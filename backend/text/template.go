package text

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielriddell21/narrata/internal/prompt"
)

// Template is a pure-Go backend for constrained systems. It composes a readable
// sentence from the event and data without any inference. Output is
// deterministic but more prose-like than the Mock backend.
type Template struct{}

// NewTemplate returns a Template backend.
func NewTemplate() *Template { return &Template{} }

// Generate composes a single natural-language line from the prompt's event and
// data block, e.g. "Washing machine done — room: utility room, cycle: cottons."
func (t *Template) Generate(ctx context.Context, p string, _ GenerateOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("template backend: %w", err)
	}
	event, dataLines := prompt.ParseEventData(p)

	var b strings.Builder
	b.WriteString(humanizeEvent(event))
	if len(dataLines) > 0 {
		b.WriteString(" — ")
		b.WriteString(strings.Join(dataLines, ", "))
	}
	b.WriteString(".")

	text := b.String()
	return Result{Text: text, Tokens: len(strings.Fields(text))}, nil
}

// Close is a no-op.
func (t *Template) Close() error { return nil }
