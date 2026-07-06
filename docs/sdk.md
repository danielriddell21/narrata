# Narrata Go SDK Design

## 1. SDK Philosophy

The Go SDK should feel like a normal Go library, not a chatbot wrapper. Developers should be able to instantiate an engine, pass an event, and receive text or audio.

Narrata's public API should communicate the core model:

```text
Input data -> persona -> generate narration -> text/audio result
```

The SDK should not expose assistant, chat, memory, planning, or tool-calling primitives.

## 2. Basic Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/danielriddell21/narrata"
)

func main() {
    ctx := context.Background()

    engine, err := narrata.New(narrata.Config{
        PersonasPath: "./personas.json",
        Text:         narrata.TextConfig{Backend: "native"}, // pure Go, no model
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

## 3. Public Types

### Config

```go
type Config struct {
    PersonasPath string
    PersonasDir string
    DefaultPersona string
    Text TextConfig
    TTS TTSConfig
    MaxConcurrent int
    Timeout time.Duration
}
```

### TextConfig

```go
type TextConfig struct {
    Backend string
    ContextTokens int
    Temperature float32
    TopP float32
    MaxTokens int
}
```

### TTSConfig

```go
type TTSConfig struct {
    Enabled bool
    Backend string
    DefaultVoice string
    SampleRate int
}
```

### Request

```go
type Request struct {
    PersonaID string
    Event string
    Data any
    Instruction string
    Output OutputMode
    Constraints Constraints
    EventPolicy EventPolicy
}
```

### Constraints

```go
type Constraints struct {
    MaxWords int
    MaxSentences int
    AllowProfanity bool
    Format Format
}
```

### EventPolicy

```go
type EventPolicy struct {
    Importance string
    Urgency string
    Cooldown time.Duration
    AllowSilence bool
}
```

For MVP, `EventPolicy` may be parsed but not fully acted on. It exists to reserve the correct shape without adding agent behaviour.

### Result

```go
type Result struct {
    Text string
    Audio []byte
    AudioFormat string
    PersonaID string
    Spoken bool
    Silent bool
    Duration time.Duration
    Tokens int
}
```

## 4. Interfaces

### TextBackend

```go
type TextBackend interface {
    Generate(ctx context.Context, prompt string, opts GenerateOptions) (TextResult, error)
    Close() error
}
```

### TTSBackend

```go
type TTSBackend interface {
    Speak(ctx context.Context, text string, opts SpeakOptions) (AudioResult, error)
    Close() error
}
```

### PersonaStore

```go
type PersonaStore interface {
    Get(id string) (Persona, bool)
    List() []Persona
    Save(persona Persona) error
}
```

## 5. Forbidden API Shapes

Do not add these to the SDK core:

```go
type Assistant struct{}
type ChatSession struct{}
type MemoryStore interface{}
type Tool interface{}
type Agent struct{}
type Planner interface{}
type Retriever interface{}
```

If a host app needs these, it should own them outside Narrata and call `Generate` only for final narration.

## 6. Game Integration Example

```go
func onPlayerLowHealth(engine *narrata.Engine, player Player) {
    result, _ := engine.Generate(context.Background(), narrata.Request{
        PersonaID: "dungeon_master",
        Event: "player_low_health",
        Data: map[string]any{
            "player": player.Name,
            "health": player.Health,
            "enemy": player.CurrentEnemy,
        },
        Output: narrata.OutputText,
        Constraints: narrata.Constraints{
            MaxWords: 20,
        },
    })

    game.ShowSubtitle(result.Text)
}
```

## 7. Home Automation Example

```go
func announceWashingDone(engine *narrata.Engine, speaker Speaker) error {
    result, err := engine.Generate(context.Background(), narrata.Request{
        PersonaID: "home_announcer",
        Event: "washing_machine_done",
        Data: map[string]any{
            "room": "utility room",
            "cycle": "cottons",
        },
        Output: narrata.OutputTextSpeech,
    })
    if err != nil {
        return err
    }

    speaker.Play(result.Audio)
    return nil
}
```

## 8. Monitoring Example

```go
func onAlert(engine *narrata.Engine, alert Alert) string {
    result, err := engine.Generate(context.Background(), narrata.Request{
        PersonaID: "executive_briefing",
        Event: "service_alert",
        Data: alert,
        Output: narrata.OutputText,
        Constraints: narrata.Constraints{
            MaxSentences: 2,
        },
    })
    if err != nil {
        return alert.Summary
    }
    return result.Text
}
```

## 9. Pluggable Backends

The default `native` backend generates narration in pure Go at runtime — no model
and no cgo. The backend stays behind an interface, so `native`, `template`, and
`mock` are interchangeable, and a host can supply its own generator (e.g. wrapping
a real model) without changing the integration:

```go
type Backend interface {
    Generate(ctx context.Context, prompt string) (string, error)
}
```

## 10. Error Handling

`Generate` and `New` return typed, wrapped errors. Compare them with
`errors.Is`:

```go
res, err := engine.Generate(ctx, req)
switch {
case errors.Is(err, narrata.ErrPersonaNotFound):
    // Unknown persona ID.
case errors.Is(err, narrata.ErrInvalidRequest):
    // Request had nothing to narrate.
case errors.Is(err, narrata.ErrBackendUnavailable):
    // The configured backend could not be initialised.
case errors.Is(err, narrata.ErrGenerationTimeout):
    // Generation exceeded the deadline.
case errors.Is(err, narrata.ErrTTSUnavailable):
    // Speech was requested but TTS is disabled or failed to initialise.
case err != nil:
    // Other error (e.g. context cancellation).
default:
    use(res.Text)
}
```

| Error | Meaning |
|-------|---------|
| `ErrPersonaNotFound` | The requested (or default) persona ID is not registered. |
| `ErrInvalidRequest` | The request failed validation. |
| `ErrBackendUnavailable` | The configured backend could not be initialised (e.g. an unknown backend name). |
| `ErrGenerationTimeout` | Generation exceeded the deadline. |
| `ErrTTSUnavailable` | Speech requested but TTS is disabled or failed to initialise. |
| `ErrScopeViolation` | An operation outside Narrata's narration scope was attempted. |
