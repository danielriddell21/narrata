// Package policy post-processes generated text to enforce host and persona
// constraints. It is a pure leaf package (stdlib only).
package policy

import (
	"regexp"
	"strings"
)

// Options control post-processing.
type Options struct {
	MaxWords       int
	MaxSentences   int
	AllowMarkdown  bool
	AllowProfanity bool
	AllowSilence   bool
}

// Result is the processed output.
type Result struct {
	Text   string
	Silent bool
}

var (
	// Strip common markdown emphasis/structure when markdown is disallowed.
	// Underscores are intentionally excluded: they appear inside ordinary
	// tokens (e.g. "latency_ms") far more often than as emphasis here.
	mdEmphasis = regexp.MustCompile("[*`#>]+")
	mdLink     = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	multiSpace = regexp.MustCompile(`[ \t]+`)
	sentenceRe = regexp.MustCompile(`[^.!?]+[.!?]+|\S[^.!?]*$`)
)

// profanity is a deliberately small mask list; the goal is a basic gate, not a
// comprehensive filter. Hosts that need more should pre/post-filter themselves.
var profanity = []string{"fuck", "shit", "bastard", "asshole", "bitch"}

// Apply runs the post-processing pipeline and returns the cleaned text.
func Apply(text string, o Options) Result {
	t := strings.TrimSpace(text)
	t = stripCodeFences(t)
	t = stripSurroundingQuotes(t)

	if !o.AllowMarkdown {
		t = mdLink.ReplaceAllString(t, "$1")
		t = mdEmphasis.ReplaceAllString(t, "")
	}

	if !o.AllowProfanity {
		t = maskProfanity(t)
	}

	t = collapseWhitespace(t)

	if o.MaxSentences > 0 {
		t = limitSentences(t, o.MaxSentences)
	}
	if o.MaxWords > 0 {
		t = limitWords(t, o.MaxWords)
	}

	t = strings.TrimSpace(t)
	if t == "" && o.AllowSilence {
		return Result{Text: "", Silent: true}
	}
	return Result{Text: t}
}

func stripCodeFences(t string) string {
	if strings.HasPrefix(t, "```") {
		if i := strings.Index(t[3:], "```"); i >= 0 {
			inner := t[3 : 3+i]
			if nl := strings.IndexByte(inner, '\n'); nl >= 0 {
				inner = inner[nl+1:]
			}
			return strings.TrimSpace(inner)
		}
	}
	return t
}

func stripSurroundingQuotes(t string) string {
	for len(t) >= 2 {
		first, last := t[0], t[len(t)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') ||
			(first == '`' && last == '`') {
			t = strings.TrimSpace(t[1 : len(t)-1])
			continue
		}
		// Smart quotes.
		if strings.HasPrefix(t, "“") && strings.HasSuffix(t, "”") {
			t = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(t, "“"), "”"))
			continue
		}
		break
	}
	return t
}

func collapseWhitespace(t string) string {
	t = strings.ReplaceAll(t, "\r\n", "\n")
	t = strings.ReplaceAll(t, "\n", " ")
	t = multiSpace.ReplaceAllString(t, " ")
	return strings.TrimSpace(t)
}

func maskProfanity(t string) string {
	for _, w := range profanity {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(w) + `\b`)
		t = re.ReplaceAllStringFunc(t, func(m string) string {
			if len(m) <= 1 {
				return m
			}
			return string(m[0]) + strings.Repeat("*", len(m)-1)
		})
	}
	return t
}

func limitSentences(t string, max int) string {
	matches := sentenceRe.FindAllString(t, -1)
	if len(matches) <= max {
		return t
	}
	var kept []string
	for i := 0; i < max && i < len(matches); i++ {
		kept = append(kept, strings.TrimSpace(matches[i]))
	}
	return strings.Join(kept, " ")
}

func limitWords(t string, max int) string {
	words := strings.Fields(t)
	if len(words) <= max {
		return t
	}
	truncated := strings.Join(words[:max], " ")
	// Preserve terminal punctuation if the truncation cut it off.
	if !strings.HasSuffix(truncated, ".") && !strings.HasSuffix(truncated, "!") &&
		!strings.HasSuffix(truncated, "?") {
		truncated += "."
	}
	return truncated
}
