package eventpolicy

import (
	"testing"
	"time"
)

func TestDefaultsUnchanged(t *testing.T) {
	d := Decide(Inputs{MaxWords: 30, MaxSentences: 2})
	if d.Silent || d.Energy != "" || d.MaxWords != 30 || d.MaxSentences != 2 {
		t.Fatalf("defaults altered: %+v", d)
	}
}

func TestCooldownSilencesWhenAllowed(t *testing.T) {
	in := Inputs{
		AllowSilence: true,
		Cooldown:     time.Minute,
		SinceLast:    10 * time.Second,
		HasLast:      true,
	}
	if d := Decide(in); !d.Silent || d.Reason != "cooldown" {
		t.Fatalf("expected cooldown silence, got %+v", d)
	}

	// Not allowed -> never silent.
	in.AllowSilence = false
	if d := Decide(in); d.Silent {
		t.Fatalf("silenced without AllowSilence: %+v", d)
	}

	// Past the window -> not silent.
	in.AllowSilence, in.SinceLast = true, 2*time.Minute
	if d := Decide(in); d.Silent {
		t.Fatalf("silenced past cooldown: %+v", d)
	}

	// No prior render -> not silent.
	in.SinceLast, in.HasLast = 10*time.Second, false
	if d := Decide(in); d.Silent {
		t.Fatalf("silenced with no prior render: %+v", d)
	}
}

func TestImportanceScalesWords(t *testing.T) {
	base := 20
	if d := Decide(Inputs{Importance: "low", MaxWords: base}); d.MaxWords >= base {
		t.Fatalf("low importance did not scale down: %d", d.MaxWords)
	}
	if d := Decide(Inputs{Importance: "high", MaxWords: base}); d.MaxWords <= base {
		t.Fatalf("high importance did not scale up: %d", d.MaxWords)
	}
	// Unlimited (0) stays unlimited.
	if d := Decide(Inputs{Importance: "low", MaxWords: 0}); d.MaxWords != 0 {
		t.Fatalf("scaled an unlimited budget: %d", d.MaxWords)
	}
}

func TestUrgencyForcesSingleSentence(t *testing.T) {
	if d := Decide(Inputs{Urgency: "urgent", MaxSentences: 3}); d.MaxSentences != 1 {
		t.Fatalf("urgency did not force one sentence: %d", d.MaxSentences)
	}
}

func TestIntensityMapsToEnergy(t *testing.T) {
	if d := Decide(Inputs{Intensity: "high"}); d.Energy != "high" {
		t.Fatalf("high intensity -> energy %q", d.Energy)
	}
	if d := Decide(Inputs{Intensity: "calm"}); d.Energy != "low" {
		t.Fatalf("calm intensity -> energy %q", d.Energy)
	}
}
