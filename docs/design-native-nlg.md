# Design: nlg, a pure-Go persona-aware generation library

**Branch:** `exp/pure-go-backends` · **Status:** implemented

## Summary

`nlg` is a small Go library with an LLM-shaped call site. You can drop it in
wherever you would otherwise call an LLM for a structured task, and give it a
persona to shape the voice. It is pure Go, standard-library only, with no
trained weights, no model files and no cgo. Output is deterministic and fast.

This is procedural generation, not a neural model. The capability comes from an
authored, persona-conditioned grammar plus runtime statistics and constraint
solving rather than learned parameters. It covers the structured subset of
"things people use an LLM for". For the open-ended rest it exposes the same
interface with a pluggable real backend, so it fails clearly instead of
pretending. See [What it can and can't do](#what-it-can-and-cant-do).

## Call site

```go
// Personas are first-class: define in Go, or load the same JSON Narrata uses.
dm := nlg.Persona{
    ID:    "dungeon_master",
    Style: nlg.Style{Tone: "dramatic", Energy: "high", Humour: "light", Verbosity: "short"},
    Constraints: nlg.Constraints{MaxWords: 24, MaxSentences: 2},
}

c, _ := nlg.New(nlg.WithPersona(dm))

out, _ := c.Generate(ctx, nlg.Task{
    Persona: "dungeon_master",
    Intent:  nlg.Narrate,
    Event:   "player_low_health",
    Data:    map[string]any{"player": "Ari", "health": 8, "enemy": "Bone Dragon"},
    Constraints: nlg.Constraints{MaxWords: 20}, // overrides persona default
})
// out.Text: "Ari is on the brink facing the Bone Dragon!"
```

## Public API (sketch)

```go
package nlg

type Style struct{ Tone, Energy, Humour, Verbosity string }

type Constraints struct {
    MaxWords      int
    MaxSentences  int
    AllowMarkdown bool
}

// Persona is a specifiable style profile (definable in Go or loaded from JSON,
// same schema as Narrata personas).
type Persona struct {
    ID          string
    Style       Style
    Rules       []string
    Constraints Constraints // defaults; per-Task Constraints override
}

type Intent string

const (
    Narrate   Intent = "narrate"   // v1: structured data -> a line, in character
    Describe  Intent = "describe"  // v1: describe an entity/state
    Summarize Intent = "summarize" // roadmap
    Classify  Intent = "classify"  // roadmap: map input to Labels
    Extract   Intent = "extract"   // roadmap: pull fields against a schema
)

// Task is the LLM-shaped request: structured in, text out.
type Task struct {
    Persona     string      // persona id (empty => default)
    Intent      Intent      // default Narrate
    Event       string      // label/intent for narrate/describe
    Data        any         // structured payload (map, struct, JSON value)
    Labels      []string    // for Classify
    Hint        string      // optional one-off steer
    Constraints Constraints // overrides persona defaults
    Seed        uint64      // determinism; 0 => derive from Task
}

type Result struct {
    Text  string
    Label string // set for Classify
}

type Client struct{ /* personas, grammar, fallback */ }

func New(opts ...Option) (*Client, error)
func (c *Client) Generate(ctx context.Context, t Task) (Result, error)

// Options
func WithPersona(p Persona) Option
func WithPersonasJSON(data []byte) Option
func WithFallback(b Backend) Option // for open-ended intents / unknown input

// Backend is the escape hatch: a real LLM for the tasks procedural logic can't do.
// Any Generate(ctx, prompt) (string, error) satisfies this.
type Backend interface {
    Generate(ctx context.Context, prompt string) (string, error)
}
```

Narrata's `native` text backend becomes a thin adapter over `nlg` (parse the
prompt → `Task` → `Generate`), so the engine keeps working unchanged.

## Generation pipeline (for `narrate` / `describe`)

```mermaid
flowchart LR
  T[Task: persona, intent, event, data, seed] --> Sal[Salience: pick 1-2 fields]
  Sal --> Real[Realize fields -> phrases]
  Real --> Gram[Grammar: choose + fill template by persona/tone]
  Gram --> Con[Constraint-aware expansion]
  Con --> Out[Result.Text]
```

1. Salience. Classify fields by kind (quantity, state, name, place, time,
   flag) and keep the top 1-2 by a narration-worthiness score, scaled by
   verbosity and the word budget. Avoids dumping every key/value.
2. Realize. Per-kind realizers turn fields into fragments: `health:8` becomes
   "health at 8", `room:"utility"` becomes "in the utility room".
3. Grammar. A small grammar expands
   `S → {opener}? {event_clause} {data_clause}? {closer}?`, each non-terminal
   choosing a production conditioned on the persona's tone/energy/humour, with
   event-shape templates (completion / threshold / arrival / state-change /
   generic) matched from the event name.
4. Constraint-aware expansion. Budget-driven: optional constituents are
   included only if they fit `MaxWords`/`MaxSentences`, with a final word and
   sentence cap as a backstop. `Humour=="none"` disables witty intensifiers.
5. Determinism. A seeded PRNG (from `Seed` or a hash of the Task) drives all
   weighted choices: stable per input, reproducible in tests, varied across
   inputs.

Other intents reuse the pieces. `summarize` is salience over many fields,
`classify` is keyword scoring against `Labels`, and `extract` is schema-guided
field realization returning structured data.

## What it can and can't do

Can (pure Go, persona-conditioned): narrate/describe structured data,
summarize known fields, classify into a fixed label set, extract against a
schema. Deterministic, fast, zero deps.

Can't (needs weights): open-ended Q&A, reasoning, understanding arbitrary
free-form natural language, novel fluent prose. For these, set `WithFallback`
to a real backend; `Generate` routes there (or returns a typed
`ErrNeedsModel`) instead of faking it.

Quality is bounded by the authored grammar and phrase pools. Breadth comes
from authored variety, not scale.

## Technique mapping

| LLM idea | Weight-free analogue |
|---|---|
| Constrained / structured decoding | Grammar productions guarantee shape and spec compliance |
| Sampling (temp/top-p) | Seeded choice among authored productions |
| System prompt / control tokens | `Persona` selects grammars, pools, connectives |
| Summarization focus | Salience scoring picks narration-worthy fields |
| n-gram smoothing (optional) | Low-order Markov over an authored per-tone connective set |

## Package layout

```
narrata/
  internal/nlg/         # implementation detail of the narrata module (stdlib only)
    nlg.go              # Client, Persona, Task, Result, Options
    grammar.go          # openers, joins, budgets, seeded rng
    shape.go            # event-shape templates + tone buckets
    field.go            # field kind inference, salience, realizers
    persona.go          # Persona + JSON loading (shared shape with narrata)
    nlg_test.go
  backend/text/native.go  # adapter + structured path: -> nlg.Task -> nlg.Generate
```

Grammar and phrases are authored Go literals (KB-scale content, not a trained
model, not embedded files), which keeps the library zero-asset. Hosts extend it
through `WithOpeners` (per-tone opener pools) and per-event `Task.Examples`
templates; the clause templates themselves are not currently replaceable.

## Decisions

- v1 intents: `narrate` and `describe`. `summarize`/`classify`/`extract` are in
  the API from day one but implemented later.
- Identity: lives at `internal/nlg` as an implementation detail of the narrata
  module, not a public package. `narrata` is the package; `nlg` is not imported
  by external users. It keeps no Narrata imports so the boundary stays clean.
- Fallback: open-ended intents require `WithFallback`; without it they return
  `ErrNeedsModel` rather than degraded output.

## Phased plan

Status: phases 1-6 are implemented on `exp/pure-go-backends` (package `nlg`,
with `narrate`/`describe`/`summarize`/`classify`/`extract`, event-shape
templates, salience and realizers, constraint-aware expansion,
`WithFallback`/`ErrNeedsModel`, `WithOpeners` extensibility, tests, benchmarks
and a godoc example). The `native` text backend routes through it.

1. `nlg` core: package, `Client`/`Persona`/`Task`/`Constraints`, seeded
   grammar expansion (port and generalise `native`), `narrate`. `native`
   backend adapts to `nlg`.
2. Salience and realizers: field kind inference and per-kind realization;
   verbosity/budget-aware selection. Adds `describe`.
3. Event-shape templates: completion/threshold/arrival/state-change/generic
   clause families; richer per-tone pools.
4. Constraint-aware expansion: budget-driven optional constituents; humour
   gating; sentence budgeting.
5. Fallback and more intents: `WithFallback`, `ErrNeedsModel`; `summarize`,
   then `classify`/`extract`.
6. Polish: docs, examples, opener overrides, benchmarks.

## Testing

- Golden tests per `(persona, intent, event, data)`.
- Property tests: output within constraints; never emits `{`/`}` or raw
  `key: value`; non-empty for valid input; deterministic per seed; varies
  across events/tones.
- Fuzz the realizers with arbitrary data (no panics, bounded output).
- Benchmarks on the hot path.

## Non-goals

No weights, no embedded trained model, no external files, no dependencies beyond
the standard library. No chat/memory/agent behaviour (Narrata guardrails hold).
