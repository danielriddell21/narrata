// Command home_automation demonstrates Narrata producing a spoken announcement
// for a smart-home event. It enables the mock TTS backend so the speech path
// runs with no model file.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/danielriddell21/narrata"
)

func main() {
	ctx := context.Background()

	engine, err := narrata.New(narrata.Config{
		Text: narrata.TextConfig{Backend: "native"},
		TTS: narrata.TTSConfig{
			Enabled:      true,
			Backend:      "mock",
			DefaultVoice: "neutral_british",
			SampleRate:   24000,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	res, err := engine.Generate(ctx, narrata.Request{
		PersonaID: "home_announcer",
		Event:     "washing_machine_done",
		Data: map[string]any{
			"room":  "utility room",
			"cycle": "cottons",
		},
		Output: narrata.OutputTextSpeech,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("text:   %s\n", res.Text)
	fmt.Printf("spoken: %v (%d bytes, %s)\n", res.Spoken, len(res.Audio), res.AudioFormat)

	// In a real system: speaker.Play(res.Audio). Here we write the WAV so the
	// example produces a tangible artefact.
	if len(res.Audio) > 0 {
		if err := os.WriteFile("announcement.wav", res.Audio, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Println("wrote announcement.wav")
	}
}
