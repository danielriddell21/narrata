package narrata_test

import (
	"context"
	"fmt"
	"log"

	"github.com/danielriddell21/narrata/pkg/narrata"
)

// Example shows the minimal flow: construct an Engine with the default
// deterministic backend, narrate an event, and print the text. It needs no
// model file.
func Example() {
	engine, err := narrata.New(narrata.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	res, err := engine.Generate(context.Background(), narrata.Request{
		PersonaID: "narrator",
		Event:     "door_opened",
		Data:      map[string]any{"room": "garage", "time": "22:41"},
		Output:    narrata.OutputText,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.Text)
	// Output:
	// Door opened (room=garage; time=22:41).
}

// ExampleEngine_Generate demonstrates shaping output with a persona and a
// per-request word limit, using the pure-Go template backend.
func ExampleEngine_Generate() {
	engine, err := narrata.New(narrata.Config{
		Text: narrata.TextConfig{Backend: "template"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	res, err := engine.Generate(context.Background(), narrata.Request{
		PersonaID:   "dungeon_master",
		Event:       "player_low_health",
		Data:        map[string]any{"player": "Ari", "health": 8},
		Constraints: narrata.Constraints{MaxWords: 20},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.Text)
	// Output:
	// Player low health — health: 8, player: Ari.
}

// Example_speech requests text plus synthesized speech using the mock TTS
// backend, which produces a valid WAV without a model file.
func Example_speech() {
	engine, err := narrata.New(narrata.Config{
		TTS: narrata.TTSConfig{Enabled: true, Backend: "mock"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	res, err := engine.Generate(context.Background(), narrata.Request{
		PersonaID: "home_announcer",
		Event:     "washing_machine_done",
		Data:      map[string]any{"room": "utility room"},
		Output:    narrata.OutputTextSpeech,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s spoken=%v format=%s\n", res.Text, res.Spoken, res.AudioFormat)
	// Output:
	// Washing machine done (room=utility room). spoken=true format=wav
}
