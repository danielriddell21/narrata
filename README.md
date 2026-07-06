# narrata

*n.* the telling of what happened.

[![Go Reference](https://pkg.go.dev/badge/github.com/danielriddell21/narrata.svg)](https://pkg.go.dev/github.com/danielriddell21/narrata)
[![CI](https://github.com/danielriddell21/narrata/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/narrata/actions/workflows/ci.yaml)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

An embedded narration runtime for Go. Narrata turns structured data and events
into short human-readable text (and optional speech), shaped by configurable
personas.

```text
input data -> persona -> generate narration -> text/audio result
```

Everything happens at runtime in pure Go. There are no models, no cgo and no
external files, and the only dependency is the standard library. Narration
comes from a small procedural engine conditioned on the persona, so the
resulting binary is self-contained and starts immediately.

It is a library you embed, **not** an assistant, agent, chatbot, memory system,
or tool-calling framework. See [what Narrata is not](#what-narrata-is-not).

## Install

```bash
go get github.com/danielriddell21/narrata
```

Requires Go 1.26+.

## Quick start

This runs as-is, no setup needed:

```go
package main

import (
	"context"
	"fmt"

	"github.com/danielriddell21/narrata"
)

func main() {
	engine, err := narrata.New(narrata.Config{})
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	res, err := engine.Generate(context.Background(), narrata.Request{
		PersonaID: "narrator",
		Event:     "door_opened",
		Data:      map[string]any{"room": "garage", "time": "22:41"},
		Output:    narrata.OutputText,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Text)
}
```

The default `native` backend builds sentences from the event shape and the
persona's tone. See [docs/backends.md](docs/backends.md).

Runnable programs live in [`examples/`](examples): `monitoring`, `game`, and
`home_automation`. There is also a development CLI (`go run ./cmd/narrata`)
that validates personas, scaffolds new ones, and renders events from the
terminal. See [docs/cli.md](docs/cli.md).

## Architecture

The host owns execution, data, and timing, and calls Narrata when it wants
narration. Narrata only transforms host-provided data.

```mermaid
flowchart LR
    H[Host application] -->|Request| E
    subgraph E["Engine"]
        direction TB
        PR[Persona registry] --> GE[Narration engine]
        GE --> OP[Output policy]
        OP --> TT[Optional TTS]
    end
    E -->|"Result: text + audio"| H
```

Details in [docs/architecture.md](docs/architecture.md).

## Default personas

`narrator`, `home_announcer`, `funny_narrator`, `dungeon_master`,
`executive_briefing`, `newsreader`, `sports_commentator`, `sci_fi_computer`,
`robot_butler`. Customise or add your own via `personas.json`, described in
[docs/personas.md](docs/personas.md).

## What Narrata is NOT

Narrata deliberately excludes assistant frameworks, chat sessions, memory,
tool calling, planners, RAG, and autonomous agents. If your app needs those,
build them outside Narrata and call `Generate()` only for final narration. The
public API exposes none of those concepts, and a test enforces it.

## Documentation

Full docs are in [`docs/`](docs/README.md):

- [spec.md](docs/spec.md): mission, goals, non-goals, use cases
- [architecture.md](docs/architecture.md): boundaries, flow, components
- [sdk.md](docs/sdk.md): Go API reference and error handling
- [personas.md](docs/personas.md): persona schema and defaults
- [persona-authoring.md](docs/persona-authoring.md): writing and scaffolding personas
- [backends.md](docs/backends.md): the pure-Go backends
- [cli.md](docs/cli.md): the `narrata` development CLI
- [roadmap.md](docs/roadmap.md): phased delivery plan

## License

[LICENSE](./LICENSE)
