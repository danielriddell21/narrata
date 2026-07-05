// Command game demonstrates Narrata producing short, persona-shaped narration
// for game events. It uses the pure-Go native backend, so it runs locally with
// no model file.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/danielriddell21/narrata"
)

type player struct {
	Name   string `json:"player"`
	Health int    `json:"health"`
	Enemy  string `json:"enemy"`
}

func main() {
	ctx := context.Background()

	engine, err := narrata.New(narrata.Config{
		Text: narrata.TextConfig{Backend: "native"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	p := player{Name: "Ari", Health: 8, Enemy: "Bone Dragon"}

	res, err := engine.Generate(ctx, narrata.Request{
		PersonaID: "dungeon_master",
		Event:     "player_low_health",
		Data:      p,
		Output:    narrata.OutputText,
		Constraints: narrata.Constraints{
			MaxWords: 20,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// In a real game: game.ShowSubtitle(res.Text)
	fmt.Println(res.Text)
}
