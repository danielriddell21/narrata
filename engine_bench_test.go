package narrata

import (
	"context"
	"testing"
)

func benchRequest() Request {
	return Request{
		PersonaID: "funny_narrator",
		Event:     "service_degraded",
		Data: map[string]any{
			"service":    "payments-api",
			"latency_ms": 950,
			"cpu":        96,
		},
		Output: OutputText,
	}
}

func benchmarkGenerate(b *testing.B, backend string) {
	b.Helper()
	e, err := New(Config{Text: TextConfig{Backend: backend}})
	if err != nil {
		b.Fatal(err)
	}
	defer e.Close()

	ctx := context.Background()
	req := benchRequest()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.Generate(ctx, req); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateMock(b *testing.B)     { benchmarkGenerate(b, "mock") }
func BenchmarkGenerateTemplate(b *testing.B) { benchmarkGenerate(b, "template") }
