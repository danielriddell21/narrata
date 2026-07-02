package text

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/danielriddell21/narrata/internal/prompt"
)

// Native is an experimental, pure-Go text backend. It composes persona-flavoured
// narration from the event and data using seeded grammar templates — no model
// file and no cgo. On this branch it is the first step toward replacing the
// llama.cpp backend with native Go generation.
//
// Output is deterministic for a given (event, data, tone) but varies across
// different events and tones, giving more life than the Template backend
// without any inference. See docs/exp-native-go.md for the roadmap beyond
// simple templates (phrase corpus, Markov, optional tiny embedded model).
type Native struct{}

// NewNative returns a Native backend.
func NewNative() *Native { return &Native{} }

// Generate composes a short, tone-flavoured line from the prompt's event, data,
// and style.
func (n *Native) Generate(ctx context.Context, p string, _ GenerateOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("native backend: %w", err)
	}
	event, dataLines := prompt.ParseEventData(p)
	tone := parseTone(p)
	seed := seedFor(tone, event, dataLines)

	text := compose(tone, humanizeEvent(event), dataLines, seed)
	return Result{Text: text, Tokens: len(strings.Fields(text))}, nil
}

// Close is a no-op.
func (n *Native) Close() error { return nil }

// parseTone extracts the persona tone from the prompt's "Style: tone=..." line,
// defaulting to "neutral".
func parseTone(p string) string {
	i := strings.Index(p, "tone=")
	if i < 0 {
		return "neutral"
	}
	rest := p[i+len("tone="):]
	for j, r := range rest {
		if r == ',' || r == '\n' {
			return strings.TrimSpace(rest[:j])
		}
	}
	return strings.TrimSpace(rest)
}

// seedFor derives a stable seed from the tone, event, and data so output is
// deterministic per input but varies across inputs.
func seedFor(tone, event string, dataLines []string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(tone))
	h.Write([]byte{0})
	h.Write([]byte(event))
	for _, l := range dataLines {
		h.Write([]byte{0})
		h.Write([]byte(l))
	}
	return h.Sum64()
}

// openersByTone gives each tone a small pool of opening phrases. An empty string
// means "no opener".
var openersByTone = map[string][]string{
	"dramatic":     {"", "And so, ", "Mark this — ", "At last, "},
	"calm":         {"", "Note: ", "For your awareness, "},
	"dry":          {"", "Well. ", "Naturally, ", "Of course, "},
	"witty":        {"", "Well. ", "Naturally, ", "Delightful — "},
	"excited":      {"", "Here we go — ", "Big one — ", "Get this — "},
	"professional": {"", "Update: ", "Status: "},
	"neutral":      {"", "Update: "},
	"clear":        {"", "Heads up — ", "Update: "},
	"playful":      {"", "Well now, ", "Fancy that — "},
	"polite":       {"", "If I may — ", "A small note — "},
}

// compose assembles opener + event + data clause into a single line.
func compose(tone, eventPhrase string, dataLines []string, seed uint64) string {
	openers, ok := openersByTone[tone]
	if !ok {
		openers = openersByTone["neutral"]
	}
	opener := openers[seed%uint64(len(openers))]
	clause := dataClause(dataLines, seed/7)

	line := opener + eventPhrase + clause + "."
	return capitalise(line)
}

// dataClause renders the data lines in one of a few shapes for variety.
func dataClause(dataLines []string, seed uint64) string {
	if len(dataLines) == 0 {
		return ""
	}
	joined := strings.Join(dataLines, ", ")
	switch seed % 3 {
	case 0:
		return " — " + joined
	case 1:
		return " (" + joined + ")"
	default:
		return ": " + joined
	}
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
