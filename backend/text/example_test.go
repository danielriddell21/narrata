package text_test

import (
	"context"
	"fmt"
	"log"

	"github.com/danielriddell21/narrata/backend/text"
)

// ExampleNew selects the pure-Go template backend and generates a line from a
// prompt containing an event and data block.
func ExampleNew() {
	b, err := text.New(text.Options{Backend: "template"})
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	prompt := "Event: door_opened\nData:\n- room: garage\nNarration:"
	res, err := b.Generate(context.Background(), prompt, text.GenerateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.Text)
	// Output:
	// Door opened — room: garage.
}
