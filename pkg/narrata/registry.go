package narrata

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/danielriddell21/narrata/internal/validate"
)

//go:embed personas.default.json
var defaultPersonasJSON []byte

// registry is an in-memory PersonaStore. It is safe for concurrent use.
type registry struct {
	mu      sync.RWMutex
	byID    map[string]Persona
	order   []string
	defawlt string
}

func newRegistry() *registry {
	return &registry{byID: make(map[string]Persona)}
}

// Get returns the persona with the given ID.
func (r *registry) Get(id string) (Persona, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	return p, ok
}

// List returns all personas in insertion order.
func (r *registry) List() []Persona {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Persona, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.byID[id])
	}
	return out
}

// Save validates and upserts a persona.
func (r *registry) Save(p Persona) error {
	if err := validatePersona(p); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[p.ID]; !exists {
		r.order = append(r.order, p.ID)
	}
	r.byID[p.ID] = p
	return nil
}

// defaultID returns the configured default persona, falling back to the file
// default, then to "narrator", then to any registered persona.
func (r *registry) defaultID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.defawlt != "" {
		if _, ok := r.byID[r.defawlt]; ok {
			return r.defawlt
		}
	}
	if _, ok := r.byID["narrator"]; ok {
		return "narrator"
	}
	if len(r.order) > 0 {
		return r.order[0]
	}
	return ""
}

// loadDefaults loads the personas bundled with the package.
func (r *registry) loadDefaults() error {
	return r.loadBytes(defaultPersonasJSON)
}

// loadBytes merges a personas.json document into the registry.
func (r *registry) loadBytes(data []byte) error {
	var pf personaFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return fmt.Errorf("%w: parsing personas: %v", ErrInvalidRequest, err)
	}
	for _, p := range pf.Personas {
		if err := r.Save(p); err != nil {
			return err
		}
	}
	if pf.Default != "" {
		r.mu.Lock()
		r.defawlt = pf.Default
		r.mu.Unlock()
	}
	return nil
}

// loadFile merges a single personas.json file.
func (r *registry) loadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w: reading personas file %q: %v", ErrInvalidRequest, path, err)
	}
	return r.loadBytes(data)
}

// loadDir merges every *.json file in a directory as one persona per file. The
// file may contain either a bare persona object or a personas.json document.
func (r *registry) loadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("%w: reading personas dir %q: %v", ErrInvalidRequest, dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names) // deterministic load order
	for _, name := range names {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%w: reading persona %q: %v", ErrInvalidRequest, path, err)
		}
		if err := r.loadPersonaFileOrDoc(data, path); err != nil {
			return err
		}
	}
	return nil
}

// loadPersonaFileOrDoc accepts either a personas.json document or a single
// persona object.
func (r *registry) loadPersonaFileOrDoc(data []byte, path string) error {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return fmt.Errorf("%w: parsing persona %q: %v", ErrInvalidRequest, path, err)
	}
	if _, isDoc := probe["personas"]; isDoc {
		return r.loadBytes(data)
	}
	var p Persona
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("%w: parsing persona %q: %v", ErrInvalidRequest, path, err)
	}
	return r.Save(p)
}

// validatePersona maps a Persona onto the pure validator.
func validatePersona(p Persona) error {
	return validate.Persona(validate.PersonaSpec{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		Rules:        p.Rules,
		MaxWords:     p.Constraints.MaxWords,
		MaxSentences: p.Constraints.MaxSentences,
	})
}
