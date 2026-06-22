package prompt

import "testing"

func benchInput() Input {
	return Input{
		PersonaName:  "Narrator",
		Description:  "Clear, neutral narration for general-purpose events.",
		Tone:         "calm",
		Energy:       "medium",
		Humour:       "none",
		Verbosity:    "short",
		Rules:        []string{"Explain the event clearly.", "Do not mention raw JSON."},
		MaxWords:     20,
		MaxSentences: 1,
		Event:        "service_degraded",
		DataLines:    []string{"cpu: 96", "latency_ms: 950", "service: payments-api"},
	}
}

func BenchmarkBuild(b *testing.B) {
	in := benchInput()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Build(in)
	}
}

func BenchmarkCompactData(b *testing.B) {
	data := map[string]any{
		"service": "payments-api",
		"latency": map[string]any{"p50": 120, "p99": 950},
		"cpu":     96,
		"healthy": false,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CompactData(data)
	}
}
