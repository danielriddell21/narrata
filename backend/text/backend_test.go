package text

import (
	"context"
	"strings"
	"testing"

	"github.com/danielriddell21/narrata/internal/prompt"
)

func buildPrompt() string {
	return prompt.Build(prompt.Input{
		PersonaName: "Narrator",
		Event:       "door_opened",
		DataLines:   []string{"room: garage", "time: 22:41"},
	})
}

func TestNewSelectsBackends(t *testing.T) {
	cases := map[string]string{"": "*text.Mock", "mock": "*text.Mock", "template": "*text.Template"}
	for name := range cases {
		b, err := New(Options{Backend: name})
		if err != nil {
			t.Fatalf("New(%q): %v", name, err)
		}
		if b == nil {
			t.Fatalf("New(%q): nil backend", name)
		}
	}
}

func TestNewUnknownBackend(t *testing.T) {
	if _, err := New(Options{Backend: "bogus"}); err == nil {
		t.Fatal("expected error for unknown backend")
	}
}

func TestLlamaStubUnavailableByDefault(t *testing.T) {
	if _, err := New(Options{Backend: "llama.cpp"}); err == nil {
		t.Fatal("expected llama backend to be unavailable without -tags llama")
	}
}

func TestBornStubUnavailableByDefault(t *testing.T) {
	if _, err := New(Options{Backend: "born"}); err == nil {
		t.Fatal("expected born backend to be unavailable without -tags born")
	}
}

func TestMockDeterministic(t *testing.T) {
	m := NewMock()
	p := buildPrompt()
	a, _ := m.Generate(context.Background(), p, GenerateOptions{})
	b, _ := m.Generate(context.Background(), p, GenerateOptions{})
	if a.Text != b.Text {
		t.Fatalf("not deterministic: %q vs %q", a.Text, b.Text)
	}
	if !strings.Contains(a.Text, "garage") {
		t.Fatalf("expected data in output: %q", a.Text)
	}
}

func TestTemplateOutput(t *testing.T) {
	tb := NewTemplate()
	r, err := tb.Generate(context.Background(), buildPrompt(), GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(r.Text, "Door opened") {
		t.Fatalf("unexpected template output: %q", r.Text)
	}
}

func TestBackendsRespectCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, b := range []Backend{NewMock(), NewTemplate()} {
		if _, err := b.Generate(ctx, buildPrompt(), GenerateOptions{}); err == nil {
			t.Fatalf("%T did not honour cancellation", b)
		}
	}
}
