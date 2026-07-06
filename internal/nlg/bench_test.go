package nlg

import (
	"context"
	"testing"
)

func benchTask(event string) Task {
	return Task{
		Persona: "p",
		Event:   event,
		Data:    map[string]any{"service": "payments-api", "cpu": 96, "latency_ms": 950},
	}
}

func benchmarkGenerate(b *testing.B, task Task) {
	b.Helper()
	c, _ := New(WithPersona(Persona{ID: "p", Style: Style{Tone: "dry"}}))
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Generate(ctx, task); err != nil {
			b.Fatal(err)
		}
	}
}

// Shaped event (goes through the event-shape templates).
func BenchmarkNarrateShaped(b *testing.B) { benchmarkGenerate(b, benchTask("service_degraded")) }

// Generic event (opener + fragment path).
func BenchmarkNarrateGeneric(b *testing.B) { benchmarkGenerate(b, benchTask("goal_scored")) }

func BenchmarkSummarize(b *testing.B) {
	t := benchTask("incident")
	t.Intent = Summarize
	benchmarkGenerate(b, t)
}
