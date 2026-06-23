// Package eventpolicy decides how an event should be rendered within
// host-defined boundaries: whether to stay silent, how tightly to bound the
// output, and what stylistic directive to add. It is pure rendering policy, not
// autonomy — it never schedules work, calls out, or decides tasks. It is a
// leaf package (stdlib only) so the public narrata package can import it
// without an import cycle.
package eventpolicy

import (
	"strings"
	"time"
)

// Inputs are the resolved event-policy values for a single request.
type Inputs struct {
	// Importance scales output length ("low", "normal", "high").
	Importance string
	// Urgency favours terse, direct output when high ("immediate"/"urgent").
	Urgency string
	// Intensity adds a stylistic directive ("low", "medium", "high").
	Intensity string
	// AllowSilence permits a cooldown-suppressed (silent) result.
	AllowSilence bool
	// Cooldown is the minimum gap between rendered outputs for this event.
	Cooldown time.Duration
	// SinceLast is the time since the last rendered output for this event.
	// Only meaningful when HasLast is true.
	SinceLast time.Duration
	// HasLast reports whether a prior render exists for this event.
	HasLast bool
	// MaxWords and MaxSentences are the base constraints to adjust. Zero means
	// unlimited and is left unchanged.
	MaxWords     int
	MaxSentences int
}

// Decision is the rendering outcome.
type Decision struct {
	// Silent reports that the event should render nothing.
	Silent bool
	// Reason explains a silent decision (e.g. "cooldown").
	Reason string
	// MaxWords and MaxSentences are the adjusted constraints.
	MaxWords     int
	MaxSentences int
	// Directive is an optional stylistic instruction for the prompt.
	Directive string
}

// Decide computes the rendering decision. With all fields at their zero/default
// values it returns the base constraints unchanged, no directive, and not
// silent, so default requests are unaffected.
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

	switch norm(in.Urgency) {
	case "immediate", "urgent", "high":
		if d.MaxSentences == 0 || d.MaxSentences > 1 {
			d.MaxSentences = 1
		}
		d.Directive = addDirective(d.Directive, "Be terse and direct.")
	}

	switch norm(in.Intensity) {
	case "high", "dramatic":
		d.Directive = addDirective(d.Directive, "Increase the drama, but keep the facts clear.")
	case "low", "calm", "understated":
		d.Directive = addDirective(d.Directive, "Keep it understated.")
	}

	return d
}

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// scaleDown reduces a word limit to ~60%, with a floor of 5. Unlimited (0) is
// left unchanged.
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

// scaleUp raises a word limit to ~130%. Unlimited (0) is left unchanged.
func scaleUp(base int) int {
	if base <= 0 {
		return base
	}
	return base * 13 / 10
}

func addDirective(existing, add string) string {
	if existing == "" {
		return add
	}
	return existing + " " + add
}
