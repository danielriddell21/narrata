package nlg

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func mustClient(t *testing.T, opts ...Option) *Client {
	t.Helper()
	c, err := New(opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNarrateDeterministicAndUsesData(t *testing.T) {
	c := mustClient(t, WithPersona(Persona{
		ID:    "dm",
		Style: Style{Tone: "dramatic", Verbosity: "short"},
	}))
	task := Task{
		Persona: "dm",
		Event:   "player_low_health",
		Data:    map[string]any{"player": "Ari", "health": 8, "enemy": "Bone Dragon"},
	}
	a, err := c.Generate(context.Background(), task)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := c.Generate(context.Background(), task)
	if a.Text != b.Text {
		t.Fatalf("not deterministic: %q vs %q", a.Text, b.Text)
	}
	if a.Text == "" {
		t.Fatal("empty output")
	}
	// A salient field (name) should surface.
	if !strings.Contains(a.Text, "Ari") && !strings.Contains(a.Text, "Bone Dragon") {
		t.Fatalf("expected a salient field in output: %q", a.Text)
	}
}

func TestConstraintsBoundWords(t *testing.T) {
	c := mustClient(t)
	res, err := c.Generate(context.Background(), Task{
		Event:       "many_fields",
		Data:        map[string]any{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6},
		Constraints: Constraints{MaxWords: 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Fields(res.Text)); n > 4 {
		t.Fatalf("word count %d > 4: %q", n, res.Text)
	}
}

func TestVariesAcrossEvents(t *testing.T) {
	c := mustClient(t, WithPersona(Persona{ID: "n", Style: Style{Tone: "dry"}}))
	seen := map[string]bool{}
	for _, ev := range []string{"door_opened", "service_degraded", "goal_scored", "boot_complete"} {
		r, _ := c.Generate(context.Background(), Task{Persona: "n", Event: ev, Data: map[string]any{"x": 1}})
		seen[r.Text] = true
	}
	if len(seen) < 3 {
		t.Fatalf("expected varied output, got %d distinct: %v", len(seen), seen)
	}
}

func TestInlineStyleOverride(t *testing.T) {
	c := mustClient(t)
	st := Style{Tone: "excited"}
	res, err := c.Generate(context.Background(), Task{Event: "goal_scored", Style: &st, Seed: 1})
	if err != nil || res.Text == "" {
		t.Fatalf("res=%q err=%v", res.Text, err)
	}
}

func TestPersonasJSON(t *testing.T) {
	doc := `{"default":"narrator","personas":[
	  {"id":"narrator","name":"Narrator","style":{"tone":"calm","verbosity":"short"},
	   "rules":["clear"],"constraints":{"max_words":20,"max_sentences":1}}]}`
	c := mustClient(t, WithPersonasJSON([]byte(doc)))
	if _, ok := c.Persona("narrator"); !ok {
		t.Fatal("persona not loaded")
	}
	// Default persona applies when Task omits one.
	res, err := c.Generate(context.Background(), Task{Event: "door_opened", Data: map[string]any{"room": "garage"}})
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Fields(res.Text)); n > 20 {
		t.Fatalf("persona max_words not applied: %d words", n)
	}
}

func TestDescribeLeadsWithSubject(t *testing.T) {
	c := mustClient(t)
	res, err := c.Generate(context.Background(), Task{
		Intent: Describe,
		Event:  "player",
		Data:   map[string]any{"player": "Ari", "health": 8, "enemy": "Bone Dragon"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.Text, "Ari") {
		t.Fatalf("describe should lead with the subject: %q", res.Text)
	}
}

func TestHumourGatingSuppressesColourfulOpeners(t *testing.T) {
	c := mustClient(t)
	for i := 0; i < 200; i++ {
		res, _ := c.Generate(context.Background(), Task{
			Event: "x", Data: map[string]any{"a": 1},
			Style: &Style{Tone: "witty", Humour: "none"}, Seed: uint64(i + 1),
		})
		for opener := range wittyOpeners {
			if strings.HasPrefix(res.Text, strings.TrimRight(opener, " —, ")) &&
				opener != "" {
				t.Fatalf("colourful opener %q leaked with humour=none: %q", opener, res.Text)
			}
		}
	}
}

func TestOpenEndedNeedsModel(t *testing.T) {
	c := mustClient(t)
	if _, err := c.Generate(context.Background(), Task{Intent: Summarize, Event: "x"}); !errors.Is(err, ErrNeedsModel) {
		t.Fatalf("err = %v, want ErrNeedsModel", err)
	}
}

type stubBackend struct{ out string }

func (s stubBackend) Generate(_ context.Context, _ string) (string, error) { return s.out, nil }

func TestFallbackRoutes(t *testing.T) {
	c := mustClient(t, WithFallback(stubBackend{out: "from model"}))
	res, err := c.Generate(context.Background(), Task{Intent: Classify, Event: "x", Hint: "classify this"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "from model" {
		t.Fatalf("expected fallback output, got %q", res.Text)
	}
}

func TestCancellation(t *testing.T) {
	c := mustClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Generate(ctx, Task{Event: "x"}); err == nil {
		t.Fatal("expected cancellation error")
	}
}
