package policy

import "testing"

func BenchmarkApply(b *testing.B) {
	text := "**Payments API** is moving like it has chosen a career in archaeology. " +
		"Latency is near one second, and the `cpu` is at 96%. This is shit."
	opts := Options{MaxWords: 25, MaxSentences: 2}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Apply(text, opts)
	}
}
