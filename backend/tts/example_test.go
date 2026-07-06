package tts_test

import (
	"context"
	"fmt"
	"log"

	"github.com/danielriddell21/narrata/backend/tts"
)

// ExampleNew selects the mock TTS backend and synthesizes a short WAV buffer
// in pure Go.
func ExampleNew() {
	b, err := tts.New(tts.Options{Backend: "mock", SampleRate: 24000})
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	res, err := b.Speak(context.Background(), "hello", tts.SpeakOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s %dHz\n", res.Format, res.SampleRate)
	// Output:
	// wav 24000Hz
}
