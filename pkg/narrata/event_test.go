package narrata

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEventCooldownSilences(t *testing.T) {
	e := newTestEngine(t, Config{})
	req := Request{
		PersonaID:   "narrator",
		Event:       "ping",
		Data:        map[string]any{"n": 1},
		EventPolicy: EventPolicy{Cooldown: time.Hour, AllowSilence: true},
	}
	first, err := e.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if first.Silent || first.Text == "" {
		t.Fatalf("first render should produce text: %+v", first)
	}
	second, err := e.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !second.Silent || second.Text != "" {
		t.Fatalf("second render within cooldown should be silent: %+v", second)
	}
}

func TestEventCooldownRequiresAllowSilence(t *testing.T) {
	e := newTestEngine(t, Config{})
	req := Request{
		PersonaID:   "narrator",
		Event:       "ping",
		Data:        map[string]any{"n": 1},
		EventPolicy: EventPolicy{Cooldown: time.Hour}, // AllowSilence false
	}
	if _, err := e.Generate(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	second, err := e.Generate(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if second.Silent {
		t.Fatal("must not silence when AllowSilence is false")
	}
}

func TestEventCooldownIsPerEvent(t *testing.T) {
	e := newTestEngine(t, Config{})
	base := EventPolicy{Cooldown: time.Hour, AllowSilence: true}
	if _, err := e.Generate(context.Background(), Request{
		PersonaID: "narrator", Event: "a", Data: map[string]any{"n": 1}, EventPolicy: base,
	}); err != nil {
		t.Fatal(err)
	}
	r, err := e.Generate(context.Background(), Request{
		PersonaID: "narrator", Event: "b", Data: map[string]any{"n": 1}, EventPolicy: base,
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Silent {
		t.Fatal("a different event must not share the cooldown bucket")
	}
}

func TestPersonaCooldownSilences(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.json")
	doc := `{"personas":[{"id":"chatty","name":"Chatty","description":"d","rules":["r"],
		"constraints":{"max_words":20,"max_sentences":1},
		"event_policy":{"cooldown_seconds":3600,"allow_silence":true}}]}`
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	e := newTestEngine(t, Config{PersonasPath: path})
	req := Request{PersonaID: "chatty", Event: "ping", Data: map[string]any{"n": 1}}
	if _, err := e.Generate(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	second, err := e.Generate(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Silent {
		t.Fatal("persona-level cooldown should silence the second render")
	}
}

func TestImportanceTightensWordLimit(t *testing.T) {
	e := newTestEngine(t, Config{Text: TextConfig{Backend: "template"}})
	data := map[string]any{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8}

	normal, err := e.Generate(context.Background(), Request{
		PersonaID: "narrator", Event: "many", Data: data,
	})
	if err != nil {
		t.Fatal(err)
	}
	low, err := e.Generate(context.Background(), Request{
		PersonaID: "narrator", Event: "many2", Data: data,
		EventPolicy: EventPolicy{Importance: "low"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Fields(low.Text)) > len(strings.Fields(normal.Text)) {
		t.Fatalf("low importance should not be longer: low=%q normal=%q", low.Text, normal.Text)
	}
}
