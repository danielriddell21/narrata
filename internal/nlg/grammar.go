package nlg

import (
	"hash/fnv"
	"strings"
)

// rng is a small deterministic PRNG (splitmix64) so generation is reproducible
// from a seed without depending on math/rand's global state.
type rng struct{ s uint64 }

func newRNG(seed uint64) *rng { return &rng{s: seed} }

func (r *rng) next() uint64 {
	r.s += 0x9E3779B97F4A7C15
	z := r.s
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

func (r *rng) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next() % uint64(n))
}

func (r *rng) pick(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return ss[r.intn(len(ss))]
}

// seedFor derives a stable seed from the task's identity.
func seedFor(t Task, style Style) uint64 {
	if t.Seed != 0 {
		return t.Seed
	}
	h := fnv.New64a()
	write := func(s string) { h.Write([]byte(s)); h.Write([]byte{0}) }
	write(t.Persona)
	write(string(t.Intent))
	write(t.Event)
	write(style.Tone)
	for _, f := range flattenData(t.Data) {
		write(f.Key)
		write(valueString(f.Value))
	}
	for _, f := range t.Fields {
		write(f.Key)
		write(valueString(f.Value))
	}
	return h.Sum64()
}

// openersByTone gives each tone a small pool of opening phrases; "" = no opener.
var openersByTone = map[string][]string{
	"dramatic":     {"", "", "And so, ", "Mark this — ", "At last, "},
	"calm":         {"", "", "Note: ", "For your awareness, "},
	"dry":          {"", "", "Well. ", "Naturally, ", "Of course, "},
	"witty":        {"", "", "Well. ", "Naturally, ", "Delightful — "},
	"excited":      {"", "Here we go — ", "Big one — ", "Get this — "},
	"professional": {"", "", "Update: ", "Status: "},
	"neutral":      {"", "", "Update: "},
	"clear":        {"", "", "Heads up — ", "Update: "},
	"playful":      {"", "Well now, ", "Fancy that — "},
	"polite":       {"", "If I may — ", "A small note — "},
	"aggressive":   {"", "Boom — ", "There it is — "},
}

// gateOpeners drops colourful openers when humour is off so the voice stays plain.
func gateOpeners(pool []string, style Style) []string {
	if style.Humour == "" || style.Humour == "none" {
		plain := pool[:0:0]
		for _, o := range pool {
			if !wittyOpeners[o] {
				plain = append(plain, o)
			}
		}
		if len(plain) > 0 {
			return plain
		}
	}
	return pool
}

// wittyOpeners are colourful openers suppressed when humour is off.
var wittyOpeners = map[string]bool{
	"Delightful — ": true, "Fancy that — ": true, "Well now, ": true,
	"Of course, ": true, "Naturally, ": true, "Boom — ": true, "There it is — ": true,
}

// compose builds a single line: opener + event clause + data clause, bounded by
// cons. The opener pool is resolved and gated by the caller.
func compose(openers []string, eventPhrase string, phrases []string, r *rng, cons Constraints) string {
	opener := r.pick(openers)
	phrases = fitPhrases(opener+eventPhrase, phrases, cons)

	var dc string
	if len(phrases) > 0 {
		joined := strings.Join(phrases, ", ")
		switch r.intn(3) {
		case 0:
			dc = " — " + joined
		case 1:
			dc = " (" + joined + ")"
		default:
			dc = ": " + joined
		}
	}

	return applyBudget(capitalise(opener+eventPhrase+dc+"."), cons)
}

// fitPhrases keeps only the leading detail phrases that fit within the word
// budget (counting a separator each), so details are dropped whole rather than
// truncated mid-phrase. With no budget it keeps them all.
func fitPhrases(base string, phrases []string, cons Constraints) []string {
	if cons.MaxWords <= 0 {
		return phrases
	}
	used := len(strings.Fields(base))
	kept := phrases[:0:0]
	for _, p := range phrases {
		cost := len(strings.Fields(p)) + 1 // + separator
		if used+cost > cons.MaxWords {
			break
		}
		used += cost
		kept = append(kept, p)
	}
	return kept
}

// applyBudget trims a line to the sentence/word constraints.
func applyBudget(line string, cons Constraints) string {
	if cons.MaxSentences > 0 {
		line = limitSentences(line, cons.MaxSentences)
	}
	if cons.MaxWords > 0 {
		line = limitWords(line, cons.MaxWords)
	}
	return line
}

// humanizeEvent turns a snake/kebab-case event into a readable phrase.
func humanizeEvent(event string) string {
	if event == "" {
		return "Event"
	}
	words := strings.FieldsFunc(event, func(r rune) bool { return r == '_' || r == '-' })
	if len(words) == 0 {
		return event
	}
	words[0] = capitalise(words[0])
	return strings.Join(words, " ")
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// limitSentences keeps the first max sentences. A terminator only ends a
// sentence when it is at the end of the string or followed by whitespace, so
// periods inside tokens (e.g. "report.pdf") do not split the text.
func limitSentences(s string, max int) string {
	count := 0
	for i := 0; i < len(s); i++ {
		if c := s[i]; c == '.' || c == '!' || c == '?' {
			if i == len(s)-1 || s[i+1] == ' ' {
				count++
				if count >= max {
					return strings.TrimSpace(s[:i+1])
				}
			}
		}
	}
	return s
}

func limitWords(s string, max int) string {
	words := strings.Fields(s)
	if len(words) <= max {
		return s
	}
	out := strings.Join(words[:max], " ")
	if !strings.HasSuffix(out, ".") && !strings.HasSuffix(out, "!") && !strings.HasSuffix(out, "?") {
		out += "."
	}
	return out
}
