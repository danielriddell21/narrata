package text

import (
	"context"
	"strings"
	"testing"

	"github.com/danielriddell21/narrata/internal/prompt"
)

func nativePrompt(tone, event string, data []string) string {
	return prompt.Build(prompt.Input{
		PersonaName: "P",
		Tone:        tone,
		Event:       event,
		DataLines:   data,
	})
}

func TestNativeRegistered(t *testing.T) {
	b, err := New(Options{Backend: "native"})
	if err != nil {
		t.Fatalf("New(native): %v", err)
	}
	if _, ok := b.(*Native); !ok {
		t.Fatalf("expected *Native, got %T", b)
	}
	if _, err := New(Options{Backend: "grammar"}); err != nil {
		t.Fatalf("New(grammar): %v", err)
	}
}

func TestNativeDeterministic(t *testing.T) {
	p := nativePrompt("dramatic", "player_low_health", []string{"enemy: Bone Dragon", "health: 8"})
	n := NewNative()
	a, _ := n.Generate(context.Background(), p, GenerateOptions{})
	b, _ := n.Generate(context.Background(), p, GenerateOptions{})
	if a.Text != b.Text {
		t.Fatalf("not deterministic: %q vs %q", a.Text, b.Text)
	}
	if !strings.Contains(a.Text, "Bone Dragon") {
		t.Fatalf("expected data in output: %q", a.Text)
	}
	if a.Tokens == 0 {
		t.Fatal("expected non-zero token count")
	}
}

func TestNativeVariesAcrossEvents(t *testing.T) {
	n := NewNative()
	seen := map[string]bool{}
	for _, ev := range []string{"door_opened", "service_degraded", "goal_scored", "boot_complete"} {
		r, _ := n.Generate(context.Background(), nativePrompt("dry", ev, []string{"x: 1"}), GenerateOptions{})
		seen[r.Text] = true
	}
	if len(seen) < 3 {
		t.Fatalf("expected varied output across events, got %d distinct: %v", len(seen), seen)
	}
}

func TestNativeRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewNative().Generate(ctx, nativePrompt("calm", "x", nil), GenerateOptions{}); err == nil {
		t.Fatal("expected cancellation error")
	}
}
