// Command monitoring demonstrates Narrata narrating dashboard/service alerts.
// It uses the default deterministic backend, so it runs with no external files.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/danielriddell21/narrata"
)

func main() {
	ctx := context.Background()

	// No Text.Backend configured -> deterministic mock backend.
	engine, err := narrata.New(narrata.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	alerts := []map[string]any{
		{"service": "payments-api", "latency_ms": 950, "cpu": 96},
		{"service": "auth", "latency_ms": 120, "cpu": 40},
	}

	for _, alert := range alerts {
		res, err := engine.Generate(ctx, narrata.Request{
			PersonaID: "executive_briefing",
			Event:     "service_degraded",
			Data:      alert,
			Output:    narrata.OutputText,
			Constraints: narrata.Constraints{
				MaxSentences: 2,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[%s] %s\n", res.PersonaID, res.Text)
	}
}
