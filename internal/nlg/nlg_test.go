package nlg

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestPersonaExampleTemplate(t *testing.T) {
	c := mustClient(t)
	res, _ := c.Generate(context.Background(), Task{
		Event:    "player_died",
		Data:     map[string]any{"player": "Ari", "enemy": "Bone Dragon"},
		Examples: map[string]string{"player_died": "{player} has fallen to the {enemy}."},
	})
	if res.Text != "Ari has fallen to the Bone Dragon." {
		t.Fatalf("example template not used: %q", res.Text)
	}
}

func TestExampleFallsThroughWhenPlaceholderMissing(t *testing.T) {
	c := mustClient(t)
	res, _ := c.Generate(context.Background(), Task{
		Event:    "boot_done",
		Data:     map[string]any{"host": "web-1"},
		Examples: map[string]string{"boot_done": "{missing} is ready"},
	})
	if strings.Contains(res.Text, "{missing}") {
		t.Fatalf("emitted an unfilled placeholder: %q", res.Text)
	}
}

func mustClient(t *testing.T, opts ...Option) *Client {
	t.Helper()
	c, err := New(opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNarrateDeterministicAndUsesData(t *testing.T) {
	c := mustClient(t, WithPersona(Persona{
		ID:    "dm",
		Style: Style{Tone: "dramatic", Verbosity: "short"},
	}))
	task := Task{
		Persona: "dm",
		Event:   "player_low_health",
		Data:    map[string]any{"player": "Ari", "health": 8, "enemy": "Bone Dragon"},
	}
	a, err := c.Generate(context.Background(), task)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := c.Generate(context.Background(), task)
	if a.Text != b.Text {
		t.Fatalf("not deterministic: %q vs %q", a.Text, b.Text)
	}
	if a.Text == "" {
		t.Fatal("empty output")
	}
	// A salient field (name) should surface.
	if !strings.Contains(a.Text, "Ari") && !strings.Contains(a.Text, "Bone Dragon") {
		t.Fatalf("expected a salient field in output: %q", a.Text)
	}
}

func TestConstraintsBoundWords(t *testing.T) {
	c := mustClient(t)
	res, err := c.Generate(context.Background(), Task{
		Event:       "many_fields",
		Data:        map[string]any{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6},
		Constraints: Constraints{MaxWords: 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Fields(res.Text)); n > 4 {
		t.Fatalf("word count %d > 4: %q", n, res.Text)
	}
}

func TestVariesAcrossEvents(t *testing.T) {
	c := mustClient(t, WithPersona(Persona{ID: "n", Style: Style{Tone: "dry"}}))
	seen := map[string]bool{}
	for _, ev := range []string{"door_opened", "service_degraded", "goal_scored", "boot_complete"} {
		r, _ := c.Generate(context.Background(), Task{Persona: "n", Event: ev, Data: map[string]any{"x": 1}})
		seen[r.Text] = true
	}
	if len(seen) < 3 {
		t.Fatalf("expected varied output, got %d distinct: %v", len(seen), seen)
	}
}

func TestInlineStyleOverride(t *testing.T) {
	c := mustClient(t)
	st := Style{Tone: "excited"}
	res, err := c.Generate(context.Background(), Task{Event: "goal_scored", Style: &st, Seed: 1})
	if err != nil || res.Text == "" {
		t.Fatalf("res=%q err=%v", res.Text, err)
	}
}

func TestPersonasJSON(t *testing.T) {
	doc := `{"default":"narrator","personas":[
	  {"id":"narrator","name":"Narrator","style":{"tone":"calm","verbosity":"short"},
	   "rules":["clear"],"constraints":{"max_words":20,"max_sentences":1}}]}`
	c := mustClient(t, WithPersonasJSON([]byte(doc)))
	if _, ok := c.Persona("narrator"); !ok {
		t.Fatal("persona not loaded")
	}
	// Default persona applies when Task omits one.
	res, err := c.Generate(context.Background(), Task{Event: "door_opened", Data: map[string]any{"room": "garage"}})
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Fields(res.Text)); n > 20 {
		t.Fatalf("persona max_words not applied: %d words", n)
	}
}

func TestDescribeLeadsWithSubject(t *testing.T) {
	c := mustClient(t)
	res, err := c.Generate(context.Background(), Task{
		Intent: Describe,
		Event:  "player",
		Data:   map[string]any{"player": "Ari", "health": 8, "enemy": "Bone Dragon"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.Text, "Ari") {
		t.Fatalf("describe should lead with the subject: %q", res.Text)
	}
}

func TestHumourGatingSuppressesColourfulOpeners(t *testing.T) {
	c := mustClient(t)
	for i := 0; i < 200; i++ {
		res, _ := c.Generate(context.Background(), Task{
			Event: "x", Data: map[string]any{"a": 1},
			Style: &Style{Tone: "witty", Humour: "none"}, Seed: uint64(i + 1),
		})
		for opener := range wittyOpeners {
			if strings.HasPrefix(res.Text, strings.TrimRight(opener, " —, ")) &&
				opener != "" {
				t.Fatalf("colourful opener %q leaked with humour=none: %q", opener, res.Text)
			}
		}
	}
}

func TestToneBucketsVaryClause(t *testing.T) {
	c := mustClient(t)
	base := Task{Event: "reactor_critical", Data: map[string]any{"reactor": "core-1"}, Seed: 7}
	plain, drama, wry := base, base, base
	plain.Style = &Style{Tone: "calm"}
	drama.Style = &Style{Tone: "dramatic"}
	wry.Style = &Style{Tone: "dry"}

	rp, _ := c.Generate(context.Background(), plain)
	rd, _ := c.Generate(context.Background(), drama)
	rw, _ := c.Generate(context.Background(), wry)
	if rp.Text == rd.Text || rp.Text == rw.Text || rd.Text == rw.Text {
		t.Fatalf("tone buckets did not vary clause: %q / %q / %q", rp.Text, rd.Text, rw.Text)
	}
	if !strings.HasSuffix(rd.Text, "!") {
		t.Fatalf("dramatic should exclaim: %q", rd.Text)
	}
	if strings.HasSuffix(rp.Text, "!") {
		t.Fatalf("calm should not exclaim: %q", rp.Text)
	}
}

func TestVerbLikeEventBecomesClause(t *testing.T) {
	c := mustClient(t)
	res, _ := c.Generate(context.Background(), Task{
		Event: "player_jumped", Data: map[string]any{"player": "Ari"},
	})
	if res.Text != "Ari jumped." {
		t.Fatalf("verb-like event not turned into a clause: %q", res.Text)
	}
}

func TestValueWithPeriodNotSplit(t *testing.T) {
	c := mustClient(t)
	res, _ := c.Generate(context.Background(), Task{
		Event: "file_uploaded", Data: map[string]any{"file": "report.pdf"},
		Constraints: Constraints{MaxSentences: 1},
	})
	if !strings.Contains(strings.ToLower(res.Text), "report.pdf") {
		t.Fatalf("value with a period was split: %q", res.Text)
	}
}

func TestActionUsesNameActor(t *testing.T) {
	c := mustClient(t)
	res, _ := c.Generate(context.Background(), Task{
		Event: "goal_scored", Data: map[string]any{"team": "Rovers", "minute": 89},
	})
	if !strings.HasPrefix(res.Text, "Rovers scored") {
		t.Fatalf("name actor not used: %q", res.Text)
	}
}

func TestSalienceSurfacesNotableQuantity(t *testing.T) {
	c := mustClient(t)
	// One field slot: a near-limit percentage should win over a plain number.
	res, _ := c.Generate(context.Background(), Task{
		Event: "status_report", Data: map[string]any{"cpu": 96, "queue": 3},
		Style: &Style{Verbosity: "short"},
	})
	if !strings.Contains(res.Text, "cpu 96%") {
		t.Fatalf("notable field not surfaced: %q", res.Text)
	}
}

func TestSubjectSubstitution(t *testing.T) {
	c := mustClient(t)
	// A field matching the event's subject noun becomes the sentence subject.
	res, _ := c.Generate(context.Background(), Task{
		Event: "service_degraded",
		Data:  map[string]any{"service": "payments-api", "cpu": 96},
	})
	if !strings.HasPrefix(res.Text, "Payments-api") {
		t.Fatalf("expected subject substitution, got %q", res.Text)
	}
}

func TestShapedSentences(t *testing.T) {
	c := mustClient(t)
	for _, ev := range []string{"washing_machine_done", "door_opened", "server_offline"} {
		res, _ := c.Generate(context.Background(), Task{Event: ev, Seed: 1})
		// A shaped sentence starts with "The " and is a full clause, not "Event (data)".
		if !strings.HasPrefix(res.Text, "The ") || strings.Contains(res.Text, "(") {
			t.Fatalf("event %q not shaped into a sentence: %q", ev, res.Text)
		}
	}
}

func TestPlaceAdjunctAttachesToClause(t *testing.T) {
	c := mustClient(t)
	// A place field attaches to the verb clause ("open in the garage"), not as an
	// em-dash detail ("open — in the garage").
	res, _ := c.Generate(context.Background(), Task{
		Event: "door_opened", Data: map[string]any{"room": "garage"}, Seed: 1,
	})
	if want := "in the garage"; !strings.Contains(res.Text, want) || strings.Contains(res.Text, "— in the") {
		t.Fatalf("place not attached to clause: %q", res.Text)
	}
}

func TestNamedPlaceTakesNoArticle(t *testing.T) {
	c := mustClient(t)
	// A region reads as a named location ("in eu-west"), not a common one
	// ("in the eu-west"), and attaches to the clause rather than as a detail.
	res, _ := c.Generate(context.Background(), Task{
		Event: "server_offline", Data: map[string]any{"server": "web-2", "region": "eu-west"}, Seed: 1,
	})
	if !strings.Contains(res.Text, "in eu-west") || strings.Contains(res.Text, "in the eu-west") {
		t.Fatalf("named place mis-articled: %q", res.Text)
	}
	if strings.Contains(res.Text, "region eu-west") {
		t.Fatalf("region not realised as a place: %q", res.Text)
	}
}

func TestCombatantNameReadsAsRelation(t *testing.T) {
	c := mustClient(t)
	// An enemy name reads as a relation ("facing the Bone Dragon"), attached to
	// the clause rather than joined into the detail list with "and".
	res, _ := c.Generate(context.Background(), Task{
		Event: "player_low_health",
		Data:  map[string]any{"player": "Ari", "health": 8, "enemy": "Bone Dragon"},
		Style: &Style{Verbosity: "brief"}, Seed: 1,
	})
	if !strings.Contains(res.Text, "facing the Bone Dragon") {
		t.Fatalf("enemy not framed as relation: %q", res.Text)
	}
	if strings.Contains(res.Text, "Bone Dragon and") {
		t.Fatalf("enemy garden-paths with a detail: %q", res.Text)
	}
}

func TestAdjunctSlotsBeforeTrailingAside(t *testing.T) {
	c := mustClient(t)
	// A wry template's trailing aside ("done, somehow") must stay at the end when
	// an adjunct is attached, not split the clause ("done, somehow in the room").
	res, _ := c.Generate(context.Background(), Task{
		Event: "washing_machine_done", Data: map[string]any{"room": "utility room"},
		Style: &Style{Tone: "dry", Humour: "witty"}, Seed: 1,
	})
	if strings.Contains(res.Text, "somehow in the") {
		t.Fatalf("aside splits the clause: %q", res.Text)
	}
	if strings.Contains(res.Text, ", somehow") && !strings.HasSuffix(res.Text, ", somehow.") {
		t.Fatalf("aside not relocated to the end: %q", res.Text)
	}
}

func TestBudgetDropsWholeDetails(t *testing.T) {
	c := mustClient(t)
	res, _ := c.Generate(context.Background(), Task{
		Event:       "service_degraded",
		Data:        map[string]any{"service": "payments-api", "cpu": 96, "latency_ms": 950, "region": "eu"},
		Constraints: Constraints{MaxWords: 6},
	})
	if n := len(strings.Fields(res.Text)); n > 6 {
		t.Fatalf("over budget: %d words in %q", n, res.Text)
	}
	// No dangling separators from mid-phrase truncation.
	if strings.Contains(res.Text, ",.") || strings.HasSuffix(res.Text, "—.") ||
		strings.Contains(res.Text, "— .") {
		t.Fatalf("dangling separator in %q", res.Text)
	}
}

func TestOpenEndedNeedsModel(t *testing.T) {
	c := mustClient(t)
	// An unknown/open-ended intent with no fallback needs a model.
	if _, err := c.Generate(context.Background(), Task{Intent: Intent("reason"), Event: "x"}); !errors.Is(err, ErrNeedsModel) {
		t.Fatalf("err = %v, want ErrNeedsModel", err)
	}
}

func TestSummarize(t *testing.T) {
	c := mustClient(t)
	res, err := c.Generate(context.Background(), Task{
		Intent: Summarize,
		Event:  "incident",
		Data:   map[string]any{"service": "payments", "cpu": 96, "latency_ms": 950},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text == "" || !strings.HasPrefix(res.Text, "Incident") {
		t.Fatalf("unexpected summary: %q", res.Text)
	}
}

func TestClassify(t *testing.T) {
	c := mustClient(t)
	res, err := c.Generate(context.Background(), Task{
		Intent: Classify,
		Event:  "disk_full",
		Labels: []string{"ok", "warning", "full"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Label != "full" {
		t.Fatalf("Label = %q, want full", res.Label)
	}
}

func TestExtract(t *testing.T) {
	c := mustClient(t)
	res, err := c.Generate(context.Background(), Task{
		Intent: Extract,
		Data:   map[string]any{"service": "payments", "cpu": 96, "region": "eu"},
		Labels: []string{"service", "cpu"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "service=payments, cpu=96" {
		t.Fatalf("Extract = %q", res.Text)
	}
}

type stubBackend struct{ out string }

func (s stubBackend) Generate(_ context.Context, _ string) (string, error) { return s.out, nil }

func TestFallbackRoutes(t *testing.T) {
	c := mustClient(t, WithFallback(stubBackend{out: "from model"}))
	res, err := c.Generate(context.Background(), Task{Intent: Intent("chat"), Hint: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "from model" {
		t.Fatalf("expected fallback output, got %q", res.Text)
	}
}

func TestWithOpeners(t *testing.T) {
	c := mustClient(t, WithOpeners("neutral", []string{"CUSTOM — "}))
	// "status_report" has no shape keyword and no verb-like word, so it takes
	// the generic opener path.
	res, _ := c.Generate(context.Background(), Task{
		Event: "status_report", Data: map[string]any{"cups": 2},
		Style: &Style{Tone: "neutral", Humour: "light"}, Seed: 1,
	})
	if !strings.HasPrefix(res.Text, "CUSTOM") {
		t.Fatalf("custom opener not used: %q", res.Text)
	}
}

func TestCancellation(t *testing.T) {
	c := mustClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Generate(ctx, Task{Event: "x"}); err == nil {
		t.Fatal("expected cancellation error")
	}
}
