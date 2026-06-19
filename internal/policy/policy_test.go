package policy

import "testing"

func TestMaxWords(t *testing.T) {
	r := Apply("one two three four five", Options{MaxWords: 3})
	if r.Text != "one two three." {
		t.Fatalf("got %q", r.Text)
	}
}

func TestMaxSentences(t *testing.T) {
	r := Apply("First sentence. Second one. Third here.", Options{MaxSentences: 2})
	if r.Text != "First sentence. Second one." {
		t.Fatalf("got %q", r.Text)
	}
}

func TestStripSurroundingQuotes(t *testing.T) {
	for _, in := range []string{`"hello world"`, "'hello world'", "“hello world”"} {
		if r := Apply(in, Options{}); r.Text != "hello world" {
			t.Fatalf("input %q -> %q", in, r.Text)
		}
	}
}

func TestStripMarkdown(t *testing.T) {
	r := Apply("**Bold** and `code` and # heading", Options{})
	if want := "Bold and code and heading"; r.Text != want {
		t.Fatalf("got %q, want %q", r.Text, want)
	}
}

func TestMarkdownKeptWhenAllowed(t *testing.T) {
	r := Apply("**Bold**", Options{AllowMarkdown: true})
	if r.Text != "**Bold**" {
		t.Fatalf("got %q", r.Text)
	}
}

func TestUnderscoresPreserved(t *testing.T) {
	r := Apply("latency_ms is high", Options{})
	if r.Text != "latency_ms is high" {
		t.Fatalf("got %q", r.Text)
	}
}

func TestStripCodeFence(t *testing.T) {
	r := Apply("```\nhello\n```", Options{})
	if r.Text != "hello" {
		t.Fatalf("got %q", r.Text)
	}
}

func TestProfanityMasked(t *testing.T) {
	r := Apply("this is shit", Options{})
	if r.Text != "this is s***" {
		t.Fatalf("got %q", r.Text)
	}
	allowed := Apply("this is shit", Options{AllowProfanity: true})
	if allowed.Text != "this is shit" {
		t.Fatalf("got %q", allowed.Text)
	}
}

func TestSilence(t *testing.T) {
	r := Apply("   ", Options{AllowSilence: true})
	if !r.Silent || r.Text != "" {
		t.Fatalf("expected silent, got %+v", r)
	}
	r2 := Apply("   ", Options{AllowSilence: false})
	if r2.Silent {
		t.Fatal("did not expect silence when not allowed")
	}
}
