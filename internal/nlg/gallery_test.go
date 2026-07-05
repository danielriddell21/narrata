package nlg

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// galleryPersonas covers the tone/energy/humour spread.
var galleryPersonas = []struct {
	name  string
	style Style
}{
	{"narrator", Style{Tone: "calm", Energy: "medium", Humour: "none", Verbosity: "short"}},
	{"dungeon_master", Style{Tone: "dramatic", Energy: "high", Humour: "light", Verbosity: "short"}},
	{"funny_narrator", Style{Tone: "dry", Energy: "medium", Humour: "witty", Verbosity: "short"}},
	{"executive_briefing", Style{Tone: "professional", Energy: "medium", Humour: "none", Verbosity: "brief"}},
	{"sports_commentator", Style{Tone: "excited", Energy: "high", Humour: "light", Verbosity: "short"}},
	{"sci_fi_computer", Style{Tone: "calm", Energy: "low", Humour: "none", Verbosity: "short"}},
}

// galleryEvents exercises each event shape plus units and an example template.
var galleryEvents = []Task{
	{Event: "washing_machine_done", Data: map[string]any{"room": "utility room", "cycle": "cottons"}},
	{Event: "door_opened", Data: map[string]any{"room": "garage"}},
	{Event: "service_degraded", Data: map[string]any{"service": "payments-api", "cpu": 96, "latency_ms": 950}},
	{Event: "disk_full", Data: map[string]any{"disk": "/dev/sda", "usage": 99}},
	{Event: "server_offline", Data: map[string]any{"server": "web-2", "region": "eu-west"}},
	{Event: "player_low_health", Data: map[string]any{"player": "Ari", "health": 8, "enemy": "Bone Dragon"}},
	{Event: "goal_scored", Data: map[string]any{"team": "Rovers", "minute": 89}},
	{Event: "backup_uploaded", Data: map[string]any{"file": "db.sql", "size_mb": 512}},
	{Event: "player_died", Data: map[string]any{"player": "Ari", "enemy": "Bone Dragon"},
		Examples: map[string]string{"player_died": "{player} has fallen to the {enemy}."}},
}

// TestGallery renders every persona x event and diffs against a golden file, so
// output-quality changes are visible in one review. Run `go test ./internal/nlg
// -run Gallery -update` to refresh the golden after an intended change.
func TestGallery(t *testing.T) {
	c := mustClient(t)

	var b strings.Builder
	for _, p := range galleryPersonas {
		fmt.Fprintf(&b, "## %s (%s)\n", p.name, p.style.Tone)
		style := p.style
		for _, ev := range galleryEvents {
			ev.Style = &style
			ev.Seed = 1
			res, err := c.Generate(context.Background(), ev)
			if err != nil {
				t.Fatalf("%s/%s: %v", p.name, ev.Event, err)
			}
			fmt.Fprintf(&b, "%-22s %s\n", ev.Event, res.Text)
		}
		b.WriteString("\n")
	}
	got := b.String()

	golden := filepath.Join("testdata", "gallery.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with -update to create): %v", err)
	}
	if got != string(want) {
		t.Errorf("gallery output changed; run `go test ./internal/nlg -run Gallery -update` to review.\n--- got ---\n%s", got)
	}
}
