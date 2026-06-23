// Package prompt builds compact model prompts from a persona and a request.
// It is a pure leaf package (stdlib only) so the public narrata package can
// import it without an import cycle.
//
// The prompt format is deliberately stable: it keeps persona rules separate
// from host data, repeatedly frames the task as narration rather than chat,
// and exposes the event/data in a simple block that offline backends (mock,
// template) can parse deterministically while an LLM reads it as prose.
package prompt

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Input is the flattened view of persona + request needed to build a prompt.
type Input struct {
	PersonaName string
	Description string
	Tone        string
	Energy      string
	Humour      string
	Verbosity   string
	Rules       []string

	MaxWords      int
	MaxSentences  int
	AllowMarkdown bool

	Event       string
	DataLines   []string // pre-compacted "key: value" lines, in stable order
	Instruction string
}

// Markers used in the prompt body. Offline backends rely on these.
const (
	eventPrefix = "Event: "
	dataHeader  = "Data:"
	dataBullet  = "- "
)

// Build assembles the prompt string.
func Build(in Input) string {
	var b strings.Builder

	name := in.PersonaName
	if name == "" {
		name = "Narrator"
	}
	b.WriteString("You are ")
	b.WriteString(name)
	b.WriteString(".")
	if in.Description != "" {
		b.WriteString(" ")
		b.WriteString(in.Description)
	}
	b.WriteString("\n")

	b.WriteString("Style: ")
	b.WriteString(styleLine(in))
	b.WriteString("\n")

	if len(in.Rules) > 0 {
		b.WriteString("Rules:\n")
		for _, r := range in.Rules {
			b.WriteString(dataBullet)
			b.WriteString(r)
			b.WriteString("\n")
		}
	}

	if c := constraintLine(in); c != "" {
		b.WriteString("Constraints: ")
		b.WriteString(c)
		b.WriteString("\n")
	}

	surface := "plain prose"
	if in.AllowMarkdown {
		surface = "markdown"
	}
	b.WriteString("Narrate the event below as a short line of ")
	b.WriteString(surface)
	b.WriteString(". Do not output JSON, quotes, labels, or commentary. Output only the narration.\n")

	b.WriteString(eventPrefix)
	b.WriteString(sanitize(in.Event))
	b.WriteString("\n")

	b.WriteString(dataHeader)
	b.WriteString("\n")
	for _, line := range in.DataLines {
		b.WriteString(dataBullet)
		b.WriteString(line)
		b.WriteString("\n")
	}

	if in.Instruction != "" {
		b.WriteString("Instruction: ")
		b.WriteString(sanitize(in.Instruction))
		b.WriteString("\n")
	}

	b.WriteString("Narration:")
	return b.String()
}

func styleLine(in Input) string {
	parts := []string{
		"tone=" + orDefault(in.Tone, "neutral"),
		"energy=" + orDefault(in.Energy, "medium"),
		"humour=" + orDefault(in.Humour, "none"),
		"verbosity=" + orDefault(in.Verbosity, "short"),
	}
	return strings.Join(parts, ", ")
}

func constraintLine(in Input) string {
	var parts []string
	if in.MaxWords > 0 {
		parts = append(parts, "max_words="+strconv.Itoa(in.MaxWords))
	}
	if in.MaxSentences > 0 {
		parts = append(parts, "max_sentences="+strconv.Itoa(in.MaxSentences))
	}
	return strings.Join(parts, ", ")
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// CompactData flattens an arbitrary JSON-serialisable value into stable,
// sorted "key: value" lines suitable for prompting. Nested objects are
// dotted (e.g. "player.health: 8"). Non-object values render as a single
// "value: ..." line. The output is deterministic for golden tests.
func CompactData(data any) []string {
	if data == nil {
		return nil
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return []string{"value: " + fmt.Sprintf("%v", data)}
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return []string{"value: " + string(raw)}
	}
	var lines []string
	flatten("", decoded, &lines)
	sort.Strings(lines)
	return lines
}

func flatten(prefix string, v any, out *[]string) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flatten(key, val, out)
		}
	case []any:
		for i, val := range t {
			key := strconv.Itoa(i)
			if prefix != "" {
				key = prefix + "." + key
			}
			flatten(key, val, out)
		}
	default:
		key := prefix
		if key == "" {
			key = "value"
		}
		*out = append(*out, sanitize(key)+": "+scalar(t))
	}
}

func scalar(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		return sanitize(t)
	case float64:
		// JSON numbers decode to float64; render integers cleanly.
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// sanitize neutralizes untrusted text before it enters a prompt. Event data,
// keys, the event name, and any instruction may originate from untrusted
// sources, so newlines and tabs are collapsed to spaces and other control
// characters are dropped. This prevents injected content from forging new
// prompt lines (for example a fake "Narration:" directive) or breaking the
// offline backends' Event/Data parsing.
func sanitize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		switch {
		case r == '\n', r == '\r', r == '\t', r == ' ':
			if !prevSpace {
				b.WriteByte(' ')
			}
			prevSpace = true
		case r < 0x20 || r == 0x7f:
			// Drop other control characters entirely.
		default:
			b.WriteRune(r)
			prevSpace = false
		}
	}
	return strings.TrimSpace(b.String())
}

// ParseEventData extracts the event name and data lines from a prompt built by
// Build. Offline backends use it to produce deterministic narration without a
// model. The returned data lines retain their "key: value" form.
func ParseEventData(prompt string) (event string, dataLines []string) {
	inData := false
	for _, line := range strings.Split(prompt, "\n") {
		switch {
		case strings.HasPrefix(line, eventPrefix):
			event = strings.TrimSpace(strings.TrimPrefix(line, eventPrefix))
			inData = false
		case line == dataHeader:
			inData = true
		case inData && strings.HasPrefix(line, dataBullet):
			dataLines = append(dataLines, strings.TrimPrefix(line, dataBullet))
		case inData:
			inData = false
		}
	}
	return event, dataLines
}
