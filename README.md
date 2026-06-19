# Narrata

An embedded local narration runtime for Go applications. Transform structured data, events, and state changes into human-readable text and speech.

## What is Narrata?

Narrata turns machine-readable data into narration. It's a library you embed directly into your application—not a separate AI assistant service. Use it to:

- **Generate game narration** — Dynamic NPC dialogue, narrator commentary, character reactions
- **Announce smart home events** — Local spoken alerts without cloud dependency
- **Create monitoring summaries** — Turn dashboards and alerts into readable text
- **Build conversational interfaces** — Add character and voice to applications

Narrata is **not** an assistant framework, chatbot, agent, memory system, or tool-calling platform. It only converts your data into narration through configurable personas.

## Quick Start

### Installation

```bash
go get github.com/danielriddell21/narrata/pkg/narrata
```

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"github.com/danielriddell21/narrata/pkg/narrata"
)

func main() {
	ctx := context.Background()

	engine, err := narrata.New(narrata.Config{
		PersonasPath: "./personas.json",
		Text: narrata.TextConfig{
			Backend:   "llama.cpp",
			ModelPath: "./models/model.gguf",
		},
	})
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	result, err := engine.Generate(ctx, narrata.Request{
		PersonaID: "narrator",
		Event:     "door_opened",
		Data: map[string]any{
			"room": "garage",
			"time": "22:41",
		},
		Output: narrata.OutputText,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Text)
}
```

## Core Concepts

### Engine
The main runtime object. Loads personas, manages model backends, and processes requests.

### Persona
A configurable profile that controls tone, style, length, humour, and voice. Edit them in `personas.json` without code changes.

### Request
What you want narrated: an event type, data payload, and output preferences.

### Result
The output: text and/or audio, plus metadata about what was generated.

## Use Cases

### Game Narration

```go
result, _ := engine.Generate(ctx, narrata.Request{
	PersonaID: "dungeon_master",
	Event:     "player_low_health",
	Data: map[string]any{
		"player": "Ari",
		"health": 8,
		"enemy":  "Bone Dragon",
	},
})
// Output: "Ari staggers as the Bone Dragon closes in. This is not the time for optimism."
```

### Home Automation

```go
result, _ := engine.Generate(ctx, narrata.Request{
	PersonaID: "home_announcer",
	Event:     "washing_machine_done",
	Data: map[string]any{
		"room":  "utility room",
		"cycle": "cottons",
	},
	Output: narrata.OutputTextSpeech, // Text + audio
})
speaker.Play(result.Audio)
```

### Monitoring Alerts

```go
result, _ := engine.Generate(ctx, narrata.Request{
	PersonaID: "funny_narrator",
	Event:     "service_degraded",
	Data: alert,
	Output: narrata.OutputText,
})
dashboard.ShowAlert(result.Text)
```

## Default Personas

Narrata comes with built-in personas. Create custom ones by editing `personas.json`:

- **narrator** — Clear, calm, neutral narration
- **home_announcer** — Simple home automation announcements
- **funny_narrator** — Witty commentary with useful information
- **dungeon_master** — Dramatic RPG-style narration
- **executive_briefing** — Short, decision-focused business summaries
- **newsreader** — Neutral update style
- **sports_commentator** — High-energy live reactions
- **sci_fi_computer** — Calm ship computer style
- **robot_butler** — Light sci-fi servant style

## Architecture

Narrata is **library-first and embedded**. The host application controls:

- When to call Narrata
- What data to pass in
- Where output is routed
- Lifecycle and configuration

Narrata only transforms your data through personas and backend models.

```
Host Application
  ├── Game loop / automation event / dashboard alert
  ├── Host data and business rules
  ├── Host decision: should Narrata be called?
  └── Narrata embedded runtime
        ├── Persona registry
        ├── Prompt builder
        ├── Text model backend (llama.cpp/GGUF)
        ├── Output policy
        └── Optional TTS backend (Kokoro)
```

## Configuration

### Config Structure

```go
type Config struct {
	PersonasPath  string
	PersonasDir   string
	DefaultPersona string
	Text          TextConfig
	TTS           TTSConfig
	MaxConcurrent int
	Timeout       time.Duration
}
```

### Text Backend

MVP supports **llama.cpp/GGUF** for local inference. Backends are pluggable through an interface.

```go
Text: narrata.TextConfig{
	Backend:        "llama.cpp",
	ModelPath:      "./models/model.gguf",
	ContextTokens:  512,
	Temperature:    0.7,
	MaxTokens:      100,
}
```

### TTS Backend (Optional)

Speech synthesis is optional. Supported:
- **Kokoro** — 82M parameter TTS model (ONNX)
- **Mock** — For testing

```go
TTS: narrata.TTSConfig{
	Enabled:      true,
	Backend:      "kokoro",
	ModelPath:    "./models/kokoro.onnx",
	DefaultVoice: "neutral_british",
	SampleRate:   24000,
}
```

## Personas

Create `personas.json` with custom personas:

```json
{
  "version": "0.1",
  "default": "narrator",
  "personas": [
    {
      "id": "my_narrator",
      "name": "My Custom Narrator",
      "description": "A persona for my app",
      "style": {
        "tone": "dramatic",
        "energy": "high",
        "humour": "light",
        "verbosity": "short"
      },
      "voice": {
        "id": "deep_storyteller",
        "speed": 0.95,
        "pitch": 0.9
      },
      "rules": [
        "Be dramatic but clear.",
        "Mention key facts.",
        "Keep it under 30 words."
      ],
      "constraints": {
        "max_words": 30,
        "max_sentences": 2,
        "allow_profanity": false,
        "allow_markdown": false
      }
    }
  ]
}
```

## Output Modes

- **`OutputText`** — Text only
- **`OutputSpeech`** — Audio only (requires TTS)
- **`OutputTextSpeech`** — Both text and audio (requires TTS)
- **`OutputStreamText`** — Token-by-token streaming (future)

## What Narrata Is NOT

Narrata explicitly avoids:

- ❌ **Assistant frameworks** — No chat interface or assistant behaviour
- ❌ **Memory systems** — No conversation history or recall
- ❌ **Tool calling** — No function execution or external APIs
- ❌ **Planning engines** — No task generation or workflows
- ❌ **RAG platforms** — No document indexing or retrieval
- ❌ **Autonomous agents** — No background loops or self-direction
- ❌ **Chatbots** — No multi-turn conversation

If your app needs these, build them outside Narrata and call `Generate()` only for final narration.

## Development Roadmap

### Phase 0 (Current)
- Core Go SDK and API shape
- Persona loading and validation
- Mock text backend
- Example apps

### Phase 1
- llama.cpp/GGUF backend
- Real text generation
- Timeout and cancellation
- Game and monitoring examples

### Phase 2
- TTS interface and Kokoro implementation
- Optional audio output
- Home automation example

### Phase 3
- Developer documentation
- Persona authoring guide
- Benchmarks and performance tuning

### Phase 4
- Thread-safety and memory optimization
- Prompt injection mitigation
- Structured logging hooks
- Event policy groundwork

### Phase 5 (Post-MVP)
- Event awareness (importance, cooldown, silence)
- Response intensity control
- Persona hot reload

## Project Structure

```
narrata/
├── go.mod
├── README.md
├── personas.default.json
├── examples/
│   ├── game/
│   ├── home_automation/
│   └── monitoring/
├── internal/
│   ├── prompt/
│   ├── policy/
│   └── validate/
└── pkg/
    ├── narrata/
    │   ├── engine.go
    │   ├── config.go
    │   ├── request.go
    │   ├── result.go
    │   ├── persona.go
    │   └── errors.go
    └── backend/
        ├── text/
        │   ├── backend.go
        │   ├── llama_cpp.go
        │   ├── mock.go
        │   └── template.go
        └── tts/
            ├── backend.go
            ├── kokoro.go
            └── mock.go
```

## Examples

See the `examples/` directory:

- **Game** — Dynamic narrator and NPC voices
- **Home Automation** — Local spoken announcements
- **Monitoring** — Alert narration for dashboards

## Performance Considerations

- Narrata is thread-safe for concurrent requests
- Configure `MaxConcurrent` to limit parallel generation
- Support `context.Context` for cancellation and timeouts
- Keep data payloads small; Narrata compacts input before generation
- Model backends are long-lived; reuse the Engine instance

## Privacy and Security

- **Local-first** — No cloud dependency in core
- **Host-controlled** — You decide what data Narrata sees
- **No telemetry** — Narrata does not phone home
- **Embedded** — Models and personas live in your app

## Error Handling

Narrata provides typed errors for common cases:

```go
err := result.Err
switch err {
case narrata.ErrPersonaNotFound:
	// Persona ID is invalid
case narrata.ErrModelNotLoaded:
	// Text or TTS backend failed to initialize
case narrata.ErrGenerationTimeout:
	// Generation exceeded deadline
case narrata.ErrTTSUnavailable:
	// TTS was requested but not enabled
}
```

## Contributing

Contributions welcome! Please review:

- [Architecture Design](./architecture.md) — System boundaries and components
- [Specification](./spec.md) — Goals, non-goals, and requirements
- [Roadmap](./roadmap.md) — Development phases and milestones
- [Research Notes](./research.md) — Backend choices and rationale

**Important:** Narrata maintains strict non-goals to avoid scope creep into agent frameworks. Any PR that adds `Assistant`, `ChatSession`, `MemoryStore`, `ToolRegistry`, `Planner`, or similar concepts will be rejected.

## License

[LICENSE](./LICENSE)

## Resources

- [llama.cpp](https://github.com/ggml-org/llama.cpp) — Local LLM inference
- [Kokoro TTS](https://huggingface.co/hexgrad/Kokoro-82M) — Lightweight speech synthesis
- [GGUF Format](https://huggingface.co/docs/hub/en/gguf) — Model format for efficient inference

## Questions?

See the documentation files:
- [architecture.md](./architecture.md) — Technical design
- [config.md](./config.md) — Persona schema and configuration
- [sdk.md](./sdk.md) — Go SDK API reference
- [spec.md](./spec.md) — Full specification and goals
