package eventpolicy

import (
	"strings"
	"time"
)

type Inputs struct {
	Importance   string
	Urgency      string
	Intensity    string
	AllowSilence bool
	Cooldown     time.Duration
	SinceLast    time.Duration
	HasLast      bool
	MaxWords     int
	MaxSentences int
}

type Decision struct {
	Silent       bool
	Reason       string
	MaxWords     int
	MaxSentences int
	Energy       string // overrides the persona's style energy when non-empty
}

func Decide(in Inputs) Decision {
	d := Decision{MaxWords: in.MaxWords, MaxSentences: in.MaxSentences}

	// Cooldown suppresses output, but only when silence is allowed and a prior
	// render exists within the window.
	if in.AllowSilence && in.Cooldown > 0 && in.HasLast && in.SinceLast < in.Cooldown {
		d.Silent = true
		d.Reason = "cooldown"
		return d
	}

	switch norm(in.Importance) {
	case "low", "minor":
		d.MaxWords = scaleDown(d.MaxWords)
	case "high", "critical", "major":
		d.MaxWords = scaleUp(d.MaxWords)
	}

	// Urgency favours a single, terse sentence.
	switch norm(in.Urgency) {
	case "immediate", "urgent", "high":
		if d.MaxSentences == 0 || d.MaxSentences > 1 {
			d.MaxSentences = 1
		}
	}

	// Intensity maps to the style's energy, which drives dramatic phrasing and
	// punctuation in the generator (there is no prompt to steer).
	switch norm(in.Intensity) {
	case "high", "dramatic":
		d.Energy = "high"
	case "low", "calm", "understated":
		d.Energy = "low"
	}

	return d
}

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func scaleDown(base int) int {
	if base <= 0 {
		return base
	}
	v := base * 3 / 5
	if v < 5 {
		v = 5
	}
	if v > base {
		v = base
	}
	return v
}

func scaleUp(base int) int {
	if base <= 0 {
		return base
	}
	return base * 13 / 10
}
