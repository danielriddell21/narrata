package narrata

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPersonasLoadAndValidate(t *testing.T) {
	r := newRegistry()
	if err := r.loadDefaults(); err != nil {
		t.Fatalf("loadDefaults: %v", err)
	}
	personas := r.List()
	if len(personas) != 9 {
		t.Fatalf("default persona count = %d, want 9", len(personas))
	}
	want := []string{
		"narrator", "home_announcer", "funny_narrator", "dungeon_master",
		"executive_briefing", "newsreader", "sports_commentator",
		"sci_fi_computer", "robot_butler",
	}
	for _, id := range want {
		if _, ok := r.Get(id); !ok {
			t.Errorf("missing default persona %q", id)
		}
	}
	for _, p := range personas {
		if err := validatePersona(p); err != nil {
			t.Errorf("default persona %q invalid: %v", p.ID, err)
		}
	}
	if r.defaultID() != "narrator" {
		t.Fatalf("defaultID = %q, want narrator", r.defaultID())
	}
}

func TestLoadFileMergesAndOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "personas.json")
	doc := `{"version":"0.1","default":"pirate","personas":[
		{"id":"pirate","name":"Pirate","description":"Arr.","rules":["Talk like a pirate."],
		 "constraints":{"max_words":20,"max_sentences":2}}]}`
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	e := newTestEngine(t, Config{PersonasPath: path})
	if _, ok := e.Personas().Get("pirate"); !ok {
		t.Fatal("custom persona not loaded")
	}
	// Bundled defaults still present.
	if _, ok := e.Personas().Get("narrator"); !ok {
		t.Fatal("default persona missing after merge")
	}
	res, err := e.Generate(context.Background(), Request{Event: "x"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.PersonaID != "pirate" {
		t.Fatalf("default persona = %q, want pirate", res.PersonaID)
	}
}

func TestLoadDirOnePersonaPerFile(t *testing.T) {
	dir := t.TempDir()
	bare := `{"id":"butler","name":"Butler","description":"Polite.","rules":["Be polite."],
		"constraints":{"max_words":20,"max_sentences":1}}`
	if err := os.WriteFile(filepath.Join(dir, "butler.json"), []byte(bare), 0o644); err != nil {
		t.Fatal(err)
	}
	e := newTestEngine(t, Config{PersonasDir: dir})
	if _, ok := e.Personas().Get("butler"); !ok {
		t.Fatal("dir persona not loaded")
	}
}

func TestSchemaVersionRejectsIncompatibleMajor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.json")
	doc := `{"version":"9.0","personas":[{"id":"x","name":"X","description":"d","rules":["r"],
		"constraints":{"max_words":10,"max_sentences":1}}]}`
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := New(Config{PersonasPath: path}); err == nil {
		t.Fatal("expected schema version error")
	}
}

func TestSchemaVersionAcceptsSameMajor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.json")
	doc := `{"version":"0.5","personas":[{"id":"x2","name":"X","description":"d","rules":["r"],
		"constraints":{"max_words":10,"max_sentences":1}}]}`
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := New(Config{PersonasPath: path}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvalidPersonaRejected(t *testing.T) {
	r := newRegistry()
	err := r.Save(Persona{ID: "x"}) // missing name/description/rules
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSaveAndList(t *testing.T) {
	r := newRegistry()
	if err := r.loadDefaults(); err != nil {
		t.Fatal(err)
	}
	before := len(r.List())
	p := Persona{
		ID: "custom", Name: "Custom", Description: "desc",
		Rules:       []string{"rule"},
		Constraints: PersonaConstraints{MaxWords: 10, MaxSentences: 1},
	}
	if err := r.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if len(r.List()) != before+1 {
		t.Fatalf("List size = %d, want %d", len(r.List()), before+1)
	}
	// Re-saving updates in place, not appends.
	if err := r.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if len(r.List()) != before+1 {
		t.Fatalf("re-save changed count to %d", len(r.List()))
	}
}
