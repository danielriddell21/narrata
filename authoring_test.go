package narrata

import (
	"context"
	"strings"
	"testing"
)

func TestGeneratePersonaDeterministicAndValid(t *testing.T) {
	a, err := GeneratePersona("pirate ship captain for server alerts")
	if err != nil {
		t.Fatalf("GeneratePersona: %v", err)
	}
	b, err := GeneratePersona("pirate ship captain for server alerts")
	if err != nil {
		t.Fatalf("GeneratePersona: %v", err)
	}
	if a.ID != b.ID || a.Name != b.Name || a.Style != b.Style {
		t.Fatal("GeneratePersona is not deterministic")
	}
	if err := validatePersona(a); err != nil {
		t.Fatalf("generated persona invalid: %v", err)
	}
	// "for" is a dropped stopword; "pirate" sets a playful tone.
	if a.ID != "pirate_ship_captain_server" {
		t.Fatalf("ID = %q", a.ID)
	}
	if a.Style.Tone != "playful" {
		t.Fatalf("tone = %q, want playful", a.Style.Tone)
	}
}

func TestGeneratePersonaInfersHumour(t *testing.T) {
	p, err := GeneratePersona("dry witty narrator")
	if err != nil {
		t.Fatal(err)
	}
	if p.Style.Humour != "witty" {
		t.Fatalf("humour = %q, want witty", p.Style.Humour)
	}
}

func TestGeneratePersonaUsableInEngine(t *testing.T) {
	p, err := GeneratePersona("calm ship computer")
	if err != nil {
		t.Fatal(err)
	}
	e := newTestEngine(t, Config{})
	if err := e.Personas().Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	res, err := e.Generate(context.Background(), Request{PersonaID: p.ID, Event: "boot"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.TrimSpace(res.Text) == "" {
		t.Fatal("expected narration")
	}
}

func TestGeneratePersonaEmpty(t *testing.T) {
	if _, err := GeneratePersona("   "); err == nil {
		t.Fatal("expected error for empty description")
	}
}
