package nlg

import (
	"encoding/json"
	"fmt"
)

// personaFile mirrors Narrata's personas.json so the same files load here.
type personaFile struct {
	Default  string        `json:"default"`
	Personas []personaJSON `json:"personas"`
}

type personaJSON struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Style struct {
		Tone      string `json:"tone"`
		Energy    string `json:"energy"`
		Humour    string `json:"humour"`
		Verbosity string `json:"verbosity"`
	} `json:"style"`
	Rules       []string `json:"rules"`
	Constraints struct {
		MaxWords      int  `json:"max_words"`
		MaxSentences  int  `json:"max_sentences"`
		AllowMarkdown bool `json:"allow_markdown"`
	} `json:"constraints"`
}

// ParsePersonas decodes a personas.json document (or a single persona object)
// into Personas. It accepts the same schema Narrata uses.
func ParsePersonas(data []byte) ([]Persona, string, error) {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, "", fmt.Errorf("nlg: parsing personas: %w", err)
	}

	if _, isDoc := probe["personas"]; isDoc {
		var pf personaFile
		if err := json.Unmarshal(data, &pf); err != nil {
			return nil, "", fmt.Errorf("nlg: parsing personas: %w", err)
		}
		out := make([]Persona, 0, len(pf.Personas))
		for _, p := range pf.Personas {
			out = append(out, p.toPersona())
		}
		return out, pf.Default, nil
	}

	var single personaJSON
	if err := json.Unmarshal(data, &single); err != nil {
		return nil, "", fmt.Errorf("nlg: parsing persona: %w", err)
	}
	return []Persona{single.toPersona()}, "", nil
}

func (p personaJSON) toPersona() Persona {
	return Persona{
		ID: p.ID,
		Style: Style{
			Tone:      p.Style.Tone,
			Energy:    p.Style.Energy,
			Humour:    p.Style.Humour,
			Verbosity: p.Style.Verbosity,
		},
		Rules: p.Rules,
		Constraints: Constraints{
			MaxWords:      p.Constraints.MaxWords,
			MaxSentences:  p.Constraints.MaxSentences,
			AllowMarkdown: p.Constraints.AllowMarkdown,
		},
	}
}
