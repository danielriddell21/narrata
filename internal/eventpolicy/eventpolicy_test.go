package eventpolicy

import (
	"testing"
	"time"
)

func TestDefaultsUnchanged(t *testing.T) {
	d := Decide(Inputs{MaxWords: 20, MaxSentences: 1})
	if d.Silent || d.Directive != "" || d.MaxWords != 20 || d.MaxSentences != 1 {
		t.Fatalf("defaults altered output: %+v", d)
	}
}

func TestCooldownSilences(t *testing.T) {
	d := Decide(Inputs{
		AllowSilence: true,
		Cooldown:     time.Minute,
		SinceLast:    10 * time.Second,
		HasLast:      true,
		MaxWords:     20,
	})
	if !d.Silent || d.Reason != "cooldown" {
		t.Fatalf("expected cooldown silence, got %+v", d)
	}
}

func TestCooldownNotSilencedWhenNotAllowed(t *testing.T) {
	d := Decide(Inputs{
		AllowSilence: false,
		Cooldown:     time.Minute,
		SinceLast:    10 * time.Second,
		HasLast:      true,
	})
	if d.Silent {
		t.Fatal("must not silence when AllowSilence is false")
	}
}

func TestCooldownNotSilencedAfterWindow(t *testing.T) {
	d := Decide(Inputs{
		AllowSilence: true,
		Cooldown:     time.Minute,
		SinceLast:    2 * time.Minute,
		HasLast:      true,
	})
	if d.Silent {
		t.Fatal("must not silence after the cooldown window")
	}
}

func TestCooldownNoPriorRender(t *testing.T) {
	d := Decide(Inputs{AllowSilence: true, Cooldown: time.Minute, HasLast: false})
	if d.Silent {
		t.Fatal("must not silence the first render")
	}
}

func TestImportanceScaling(t *testing.T) {
	low := Decide(Inputs{Importance: "low", MaxWords: 20})
	if low.MaxWords != 12 {
		t.Fatalf("low importance MaxWords = %d, want 12", low.MaxWords)
	}
	high := Decide(Inputs{Importance: "high", MaxWords: 20})
	if high.MaxWords != 26 {
		t.Fatalf("high importance MaxWords = %d, want 26", high.MaxWords)
	}
	// Unlimited stays unlimited.
	if got := Decide(Inputs{Importance: "low", MaxWords: 0}).MaxWords; got != 0 {
		t.Fatalf("unlimited changed to %d", got)
	}
}

func TestUrgencyForcesTerse(t *testing.T) {
	d := Decide(Inputs{Urgency: "immediate", MaxSentences: 3})
	if d.MaxSentences != 1 {
		t.Fatalf("MaxSentences = %d, want 1", d.MaxSentences)
	}
	if d.Directive == "" {
		t.Fatal("expected a terse directive")
	}
}

func TestIntensityDirectives(t *testing.T) {
	if d := Decide(Inputs{Intensity: "high"}); d.Directive == "" {
		t.Fatal("high intensity should add a directive")
	}
	if d := Decide(Inputs{Intensity: "medium"}); d.Directive != "" {
		t.Fatalf("medium intensity should add no directive, got %q", d.Directive)
	}
}
