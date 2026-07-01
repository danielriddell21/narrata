package validate

import (
	"fmt"
	"strings"
)

type PersonaSpec struct {
	ID           string
	Name         string
	Description  string
	Rules        []string
	MaxWords     int
	MaxSentences int
}

func Persona(p PersonaSpec) error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("persona: id is required")
	}
	if strings.ContainsAny(p.ID, " \t\n") {
		return fmt.Errorf("persona %q: id must not contain whitespace", p.ID)
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("persona %q: name is required", p.ID)
	}
	if strings.TrimSpace(p.Description) == "" {
		return fmt.Errorf("persona %q: description is required", p.ID)
	}
	if len(p.Rules) == 0 {
		return fmt.Errorf("persona %q: at least one rule is required", p.ID)
	}
	if p.MaxWords < 0 {
		return fmt.Errorf("persona %q: max_words must not be negative", p.ID)
	}
	if p.MaxSentences < 0 {
		return fmt.Errorf("persona %q: max_sentences must not be negative", p.ID)
	}
	return nil
}

type RequestSpec struct {
	Event          string
	HasData        bool
	HasInstruction bool
}

func Request(r RequestSpec) error {
	if strings.TrimSpace(r.Event) == "" && !r.HasData && !r.HasInstruction {
		return fmt.Errorf("request: one of event, data, or instruction is required")
	}
	return nil
}
