package narrata

import (
	"fmt"
	"strings"
)

// GeneratePersona turns a short free-text description into a [Persona] scaffold,
// inferring style, voice, and baseline constraints from keywords in the
// description. It is a deterministic authoring helper for tools and tests: it
// adds no agent-like capability and performs no inference.
//
// The result is a starting point meant to be reviewed and hand-edited; callers
// typically write it to personas.json or a personas/*.json file. An error is
// returned only when the description is empty or yields an invalid persona.
func GeneratePersona(description string) (Persona, error) {
	desc := strings.TrimSpace(description)
	if desc == "" {
		return Persona{}, fmt.Errorf("%w: description is required", ErrInvalidRequest)
	}

	lower := strings.ToLower(desc)
	tone, energy, humour := inferStyle(lower)

	p := Persona{
		ID:          slugify(desc, 4),
		Name:        titleCase(slugify(desc, 4)),
		Description: capitalizeFirst(desc) + ".",
		Style: Style{
			Tone:      tone,
			Energy:    energy,
			Humour:    humour,
			Verbosity: "short",
		},
		Voice: Voice{ID: voiceFor(tone), Speed: 1.0, Pitch: 1.0},
		Rules: inferRules(tone, humour),
		Constraints: PersonaConstraints{
			MaxWords:       25,
			MaxSentences:   2,
			AllowProfanity: false,
			AllowMarkdown:  false,
		},
	}
	if err := validatePersona(p); err != nil {
		return Persona{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	return p, nil
}

// inferStyle derives tone, energy, and humour from keywords.
func inferStyle(lower string) (tone, energy, humour string) {
	tone, energy, humour = "neutral", "medium", "none"

	switch {
	case containsAny(lower, "pirate", "playful", "cheeky"):
		tone, humour = "playful", "light"
	case containsAny(lower, "dramatic", "epic", "fantasy", "dungeon", "rpg", "storyteller"):
		tone, energy = "dramatic", "high"
	case containsAny(lower, "calm", "zen", "soothing", "computer", "ship", "sci-fi", "robot", "butler"):
		tone, energy = "calm", "low"
	case containsAny(lower, "professional", "business", "executive", "corporate"):
		tone = "professional"
	case containsAny(lower, "news", "reporter", "anchor"):
		tone = "neutral"
	case containsAny(lower, "sport", "commentator", "match", "game show"):
		tone, energy = "excited", "high"
	}

	if containsAny(lower, "funny", "witty", "dry", "comic", "joke", "humorous", "sarcastic") {
		humour = "witty"
	}
	return tone, energy, humour
}

// inferRules produces a small rule set tuned to the inferred style.
func inferRules(tone, humour string) []string {
	rules := []string{
		"Keep the underlying information clear.",
		"Do not invent facts that are not present in the data.",
		"Keep the output concise.",
	}
	if tone != "neutral" {
		rules = append(rules, fmt.Sprintf("Match a %s style.", tone))
	}
	if humour != "none" {
		rules = append(rules, "Be light or witty, but never obscure the key information.")
	}
	return rules
}

// voiceFor maps a tone to a sensible default voice id.
func voiceFor(tone string) string {
	switch tone {
	case "dramatic":
		return "deep_storyteller"
	case "excited":
		return "energetic_clear"
	case "calm":
		return "synthetic_calm"
	case "playful":
		return "british_dry"
	default:
		return "neutral_clear"
	}
}

// slugify lowercases description, keeps word characters, drops common stopwords,
// and joins up to maxWords words with underscores.
func slugify(s string, maxWords int) string {
	fields := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	stop := map[string]bool{
		"for": true, "a": true, "an": true, "the": true, "of": true,
		"to": true, "and": true, "with": true, "in": true, "on": true,
	}
	var kept []string
	for _, f := range fields {
		if stop[f] {
			continue
		}
		kept = append(kept, f)
		if len(kept) == maxWords {
			break
		}
	}
	if len(kept) == 0 {
		return "persona"
	}
	return strings.Join(kept, "_")
}

// titleCase turns an underscore slug into a space-separated title.
func titleCase(slug string) string {
	parts := strings.Split(slug, "_")
	for i, p := range parts {
		parts[i] = capitalizeFirst(p)
	}
	return strings.Join(parts, " ")
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
