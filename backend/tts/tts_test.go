package tts

import (
	"bytes"
	"context"
	"testing"
)

func TestNewSelectsMock(t *testing.T) {
	for _, name := range []string{"", "mock"} {
		b, err := New(Options{Backend: name})
		if err != nil {
			t.Fatalf("New(%q): %v", name, err)
		}
		if b == nil {
			t.Fatalf("New(%q): nil", name)
		}
	}
}

func TestNewUnknownBackend(t *testing.T) {
	if _, err := New(Options{Backend: "bogus"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestMockProducesValidWAV(t *testing.T) {
	m := NewMock(24000)
	r, err := m.Speak(context.Background(), "hello world", SpeakOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Format != "wav" {
		t.Fatalf("format = %q", r.Format)
	}
	if len(r.Audio) < 44 {
		t.Fatalf("audio too small: %d bytes", len(r.Audio))
	}
	if !bytes.HasPrefix(r.Audio, []byte("RIFF")) || !bytes.Contains(r.Audio[:16], []byte("WAVE")) {
		t.Fatal("missing RIFF/WAVE header")
	}
}

func TestMockDeterministic(t *testing.T) {
	m := NewMock(24000)
	a, _ := m.Speak(context.Background(), "same text", SpeakOptions{})
	b, _ := m.Speak(context.Background(), "same text", SpeakOptions{})
	if !bytes.Equal(a.Audio, b.Audio) {
		t.Fatal("mock TTS not deterministic")
	}
}

func TestMockRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewMock(24000).Speak(ctx, "x", SpeakOptions{}); err == nil {
		t.Fatal("expected cancellation error")
	}
}
