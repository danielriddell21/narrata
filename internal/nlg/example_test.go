package nlg_test

import (
	"context"
	"fmt"

	"github.com/danielriddell21/narrata/internal/nlg"
)

// Example shows the LLM-shaped call site: define a persona, then generate an
// in-character line from a structured event — pure Go, no model.
func Example() {
	c, _ := nlg.New(nlg.WithPersona(nlg.Persona{
		ID:    "ops",
		Style: nlg.Style{Tone: "professional", Verbosity: "brief"},
	}))

	res, _ := c.Generate(context.Background(), nlg.Task{
		Persona: "ops",
		Event:   "service_degraded",
		Data:    map[string]any{"service": "payments-api", "cpu": 96, "latency_ms": 950},
		Seed:    1,
	})

	fmt.Println(res.Text)
	// Output:
	// Payments-api is struggling — cpu 96% and latency 950ms.
}

// ExampleClient_Generate_classify shows the keyword classifier picking a label.
func ExampleClient_Generate_classify() {
	c, _ := nlg.New()
	res, _ := c.Generate(context.Background(), nlg.Task{
		Intent: nlg.Classify,
		Event:  "disk_full",
		Labels: []string{"ok", "warning", "full"},
	})
	fmt.Println(res.Label)
	// Output:
	// full
}
