# Narrata Specification

## 1. Summary

Narrata is an embedded narration runtime for Go applications. It lets a host system turn structured data, events, state changes, metrics, or text into short natural-language output, with optional speech. Everything runs at runtime in **pure Go** — no models, no cgo, no external files, no dependencies beyond the standard library.

Narrata is not a separate assistant application. It is a library/runtime baked directly into the systems that use it, such as games, home automation systems, dashboards, monitoring tools, robotics projects, local developer tools, and simulations.

## 2. Core Mission

**Transform machine-readable events into human-readable narration.**

Narrata accepts:

- JSON
- Events
- Metrics
- State
- Short text context

Narrata returns:

- Text
- Optional speech audio

Nothing else belongs in the core mission.

## 3. Product Positioning

**Narrata: an embedded local narration engine for Go systems.**

Applications use Narrata to explain their state, events, and actions to humans through text and speech. The distinctive value is turning application data into in-character narration in real time — generated procedurally in Go, with no model and no inference service.

Narrata allows developers to add character, narration, spoken alerts, contextual commentary, and data-driven dialogue directly inside their software, with nothing to download and no data leaving the process.

## 4. Goals

- Provide a Go-first embedded SDK.
- Stay **pure Go**: no models, no cgo, no external files, stdlib-only.
- Generate narration procedurally at runtime — instant startup, self-contained binary, runs anywhere Go runs.
- Accept data-first inputs, especially JSON/event payloads.
- Generate short text output quickly.
- Optionally generate speech output.
- Support editable and generated personas through `personas.json` or a personas directory.
- Allow host systems to control when Narrata speaks and what data it sees.
- Keep backends behind an interface (for tests and optional extensions).
- Keep the API small: input -> persona -> generate -> output.

## 5. Non-Goals

Narrata is not:

- An AI assistant.
- An agent framework.
- A workflow engine.
- A RAG platform.
- A memory system.
- A chatbot.
- A tool-calling framework.
- A general-purpose ChatGPT clone.
- A home assistant platform by itself.
- A game engine.
- A remote API service by default.

Narrata should not plan, browse, remember, call tools, hold conversations, or autonomously decide tasks. Narrata only converts host-provided application data into text and/or speech through configurable personas.

## 6. Design Principles

1. **Embedded-first** — Narrata is a library inside a host system, not an app beside it.
2. **Local-first** — No cloud dependency in the core path.
3. **Persona-driven** — Tone, style, voice, and rules come from personas.
4. **Event-driven** — The host provides events and state; Narrata narrates them.
5. **In-house & pure Go** — narration is generated procedurally at runtime; no models, no cgo, no dependencies.
6. **Low-latency** — Outputs should be short and suitable for real-time use.
7. **No agent behaviour** — No memory, tools, planning, autonomous loops, or chatbot scope creep.
8. **Host-controlled** — The host decides what is passed in, when generation happens, and where output goes.

## 7. Target Users

### Primary

- Go developers embedding local narration into applications.
- Indie game developers wanting dynamic NPC or narrator lines.
- Home automation developers wanting local spoken alerts.
- Dashboard and monitoring tool builders wanting contextual summaries.

### Secondary

- Hobbyists building local narration into their tools.
- Robotics/simulation developers.
- Privacy-conscious users who do not want event data sent to cloud AI APIs.

## 8. Core Use Cases

### 8.1 Game Narration

Input:

```json
{
  "persona": "dungeon_master",
  "event": "player_low_health",
  "data": {
    "player": "Ari",
    "health": 8,
    "enemy": "Bone Dragon"
  }
}
```

Output:

```json
{
  "text": "Ari is fading fast — Bone Dragon!",
  "audio": null
}
```

### 8.2 Home Assistant Style Announcement

Input:

```json
{
  "persona": "home_announcer",
  "event": "washing_machine_done",
  "data": {
    "room": "utility room",
    "cycle": "cottons"
  },
  "output": ["text", "speech"]
}
```

Output:

```json
{
  "text": "The washing machine has finished — in the utility room.",
  "audio_format": "wav",
  "audio": "<bytes>"
}
```

### 8.3 Monitoring Alert

Input:

```json
{
  "persona": "funny_narrator",
  "event": "service_degraded",
  "data": {
    "service": "payments-api",
    "latency_ms": 950,
    "cpu": 96
  }
}
```

Output:

```json
{
  "text": "Payments-api has degraded — cpu 96, latency ms 950."
}
```

## 9. Core Concepts

### 9.1 Host System

The application embedding Narrata. Examples include a game, Home Assistant integration, monitoring dashboard, CLI tool, simulator, robot, or desktop app.

### 9.2 Event

A host-provided input describing something that happened or needs to be narrated.

### 9.3 Persona

A configurable behavioural and stylistic profile that controls tone, format, length, humour, and voice.

### 9.4 Renderer

The component that turns event data + persona + rules into text.

### 9.5 Speaker

The optional text-to-speech layer that converts generated text into an audio buffer or file.

### 9.6 Output Policy

The constraint layer that keeps Narrata within the host's rules: max words, format, profanity, safety, cooldown, and output mode.

## 10. High-Level API Philosophy

Preferred:

```go
result, err := engine.Generate(ctx, narrata.Request{
    PersonaID: "funny_narrator",
    Event:     "service_degraded",
    Data:      event,
})
```

Avoid:

```go
assistant.Chat(...)
assistant.Remember(...)
assistant.Plan(...)
assistant.CallTool(...)
```

Narrata should feel like an embedded renderer for narration, not a chatbot wrapper.

## 11. High-Level API Example

```go
engine, err := narrata.New(narrata.Config{
    PersonasPath: "./personas.json",
    Text:         narrata.TextConfig{Backend: "native"}, // pure Go, no model
    TTS:          narrata.TTSConfig{Enabled: true, Backend: "mock"},
})

result, err := engine.Generate(ctx, narrata.Request{
    PersonaID: "funny_narrator",
    Event:     "service_degraded",
    Data: map[string]any{
        "service": "payments-api",
        "latency_ms": 950,
        "cpu": 96,
    },
    Output: narrata.OutputText,
})

fmt.Println(result.Text)
```

## 12. Output Modes

- `text`: return generated text only.
- `speech`: return audio only.
- `text_speech`: return generated text and speech audio.
- `stream_text`: stream tokens back to the host.
- `stream_speech`: later feature; stream audio chunks as they are generated.

## 13. MVP Requirements

### Must Have

- Go SDK.
- Embedded runtime; no required external daemon.
- Persona loading from `personas.json` and/or `personas/*.json`.
- Default personas bundled with the package.
- JSON/event input.
- Text generation through the in-house, pure-Go narration engine.
- Persona-conditioned generation (event shapes, tone, salience).
- Pluggable backend interface (for tests and optional extensions).
- Deterministic response constraints: max words, max sentences, style, safety rules.
- Optional TTS interface.
- Clear errors for invalid personas and malformed input.
- Strict non-goals documented in code comments and README.

### Should Have

- Persona creation helper.
- Response caching for repeated event patterns.
- Richer authored grammar/phrase variety.
- Basic benchmark command for development only.
- Examples for games, home automation, and monitoring.

### Could Have

- Audio streaming.
- Per-persona voice mapping.
- Host-supplied custom grammar/phrase packs.
- Hot reload for `personas.json`.
- WASM or mobile support later.

## 14. Constraints

- Runs anywhere Go runs; no platform-specific requirements.
- Local-only by default.
- No cloud dependency, no cgo, and no model files anywhere in the path.
- Narrata should avoid long-running global state where possible.
- Thread safety matters for host systems with concurrent event streams.
- No conversation state in MVP.
- No memory store in MVP.
- No tool calling in MVP.

## 15. Post-MVP: Event Awareness

After the basic generation path works, Narrata may support event awareness. This should still remain host-controlled and should not become agent behaviour.

Example:

```json
{
  "event_type": "goal_scored",
  "importance": "high",
  "cooldown_seconds": 30,
  "urgency": "immediate",
  "allow_silence": true
}
```

This would allow Narrata to decide, within host-defined boundaries:

- Speak.
- Stay silent.
- Use a short response.
- Use a more dramatic response.
- Use a neutral response.

This is different from autonomy. Narrata is not deciding tasks. It is deciding how to render an event.

## 16. Success Criteria

- A Go app can embed Narrata in under 20 lines of setup code.
- A JSON event can produce a useful line of text in near real time.
- The host can select a persona per request.
- The host can disable speech entirely.
- The runtime can be used in games, dashboards, and automation examples without architectural changes.
- Default personas feel distinct without fine-tuning.
- The API does not expose assistant, memory, planning, tool-calling, or chatbot concepts.
