package narrata

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTestEngine(t *testing.T, cfg Config) *Engine {
	t.Helper()
	e, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e
}

func TestGenerateMockDeterministic(t *testing.T) {
	e := newTestEngine(t, Config{})
	req := Request{
		PersonaID: "narrator",
		Event:     "door_opened",
		Data:      map[string]any{"room": "garage", "time": "22:41"},
		Output:    OutputText,
	}
	first, err := e.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	second, err := e.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if first.Text != second.Text {
		t.Fatalf("mock output not deterministic: %q vs %q", first.Text, second.Text)
	}
	if first.Text == "" {
		t.Fatal("expected non-empty text")
	}
	if first.PersonaID != "narrator" {
		t.Fatalf("PersonaID = %q, want narrator", first.PersonaID)
	}
	if !strings.Contains(first.Text, "garage") {
		t.Fatalf("expected data in output, got %q", first.Text)
	}
}

func TestGeneratePersonaNotFound(t *testing.T) {
	e := newTestEngine(t, Config{})
	_, err := e.Generate(context.Background(), Request{
		PersonaID: "does_not_exist",
		Event:     "x",
	})
	if !errors.Is(err, ErrPersonaNotFound) {
		t.Fatalf("err = %v, want ErrPersonaNotFound", err)
	}
}

func TestGenerateInvalidRequest(t *testing.T) {
	e := newTestEngine(t, Config{})
	_, err := e.Generate(context.Background(), Request{}) // nothing to narrate
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

func TestDefaultPersonaResolution(t *testing.T) {
	e := newTestEngine(t, Config{DefaultPersona: "funny_narrator"})
	res, err := e.Generate(context.Background(), Request{Event: "boot"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.PersonaID != "funny_narrator" {
		t.Fatalf("PersonaID = %q, want funny_narrator", res.PersonaID)
	}
}

func TestUnknownDefaultPersonaFails(t *testing.T) {
	_, err := New(Config{DefaultPersona: "nope"})
	if !errors.Is(err, ErrPersonaNotFound) {
		t.Fatalf("err = %v, want ErrPersonaNotFound", err)
	}
}

func TestSpeechRequestedButTTSDisabled(t *testing.T) {
	e := newTestEngine(t, Config{})
	_, err := e.Generate(context.Background(), Request{
		PersonaID: "narrator",
		Event:     "x",
		Output:    OutputSpeech,
	})
	if !errors.Is(err, ErrTTSUnavailable) {
		t.Fatalf("err = %v, want ErrTTSUnavailable", err)
	}
}

func TestTextOnlyWorksWithTTSDisabled(t *testing.T) {
	e := newTestEngine(t, Config{})
	res, err := e.Generate(context.Background(), Request{
		PersonaID: "narrator",
		Event:     "x",
		Output:    OutputText,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.Spoken {
		t.Fatal("did not expect audio")
	}
}

func TestTextSpeechWithMockTTS(t *testing.T) {
	e := newTestEngine(t, Config{
		TTS: TTSConfig{Enabled: true, Backend: "mock", SampleRate: 24000},
	})
	res, err := e.Generate(context.Background(), Request{
		PersonaID: "home_announcer",
		Event:     "washing_machine_done",
		Data:      map[string]any{"room": "utility"},
		Output:    OutputTextSpeech,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.Text == "" {
		t.Fatal("expected text")
	}
	if !res.Spoken || len(res.Audio) == 0 {
		t.Fatal("expected audio")
	}
	if res.AudioFormat != "wav" {
		t.Fatalf("AudioFormat = %q, want wav", res.AudioFormat)
	}
}

func TestConstraintsLimitWords(t *testing.T) {
	e := newTestEngine(t, Config{})
	res, err := e.Generate(context.Background(), Request{
		PersonaID:   "narrator",
		Event:       "many_fields",
		Data:        map[string]any{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6},
		Output:      OutputText,
		Constraints: Constraints{MaxWords: 3},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got := len(strings.Fields(res.Text)); got > 3 {
		t.Fatalf("word count = %d, want <= 3 (text=%q)", got, res.Text)
	}
}

func TestContextCancellation(t *testing.T) {
	e := newTestEngine(t, Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := e.Generate(ctx, Request{PersonaID: "narrator", Event: "x"})
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestConcurrentGenerate(t *testing.T) {
	e := newTestEngine(t, Config{MaxConcurrent: 2})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := e.Generate(context.Background(), Request{
				PersonaID: "narrator",
				Event:     "tick",
				Data:      map[string]any{"n": 1},
			})
			if err != nil {
				t.Errorf("Generate: %v", err)
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent generate deadlocked")
	}
}

func TestDefaultTimeoutApplied(t *testing.T) {
	// A very short default timeout still completes for the fast mock backend.
	e := newTestEngine(t, Config{Timeout: 50 * time.Millisecond})
	_, err := e.Generate(context.Background(), Request{PersonaID: "narrator", Event: "x"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
}

func TestStructuredNativePathUsesFullStyle(t *testing.T) {
	e := newTestEngine(t, Config{Text: TextConfig{Backend: "native"}})
	res, err := e.Generate(context.Background(), Request{
		PersonaID: "dungeon_master", // dramatic, high energy
		Event:     "reactor_critical",
		Data:      map[string]any{"reactor": "core-1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(res.Text, "!") {
		t.Fatalf("expected dramatic exclamation via structured path, got %q", res.Text)
	}
	if !strings.HasPrefix(res.Text, "Core-1") {
		t.Fatalf("expected subject substitution, got %q", res.Text)
	}
}
