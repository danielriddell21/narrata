# Narrata Architecture Design

## 1. Architectural Principle

Narrata is a library-first embedded narration runtime. The host application owns execution, lifecycle, data, permissions, timing, and output routing. Narrata only transforms host-provided context into text and optionally speech.

Narrata must not become an assistant framework. The architecture should actively prevent scope creep into memory, RAG, tool calling, workflows, planning, or chat sessions.

## 2. System Boundary

```text
Host Application
  ├── Game loop / automation event / dashboard alert
  ├── Host data and business rules
  ├── Host decision: should Narrata be called?
  └── Narrata embedded runtime
        ├── Persona registry
        ├── Prompt builder
        ├── Text model backend
        ├── Output policy
        └── Optional TTS backend
```

Narrata does not poll the host system. The host calls Narrata when it wants narration.

## 3. Runtime Flow

```text
Request
  ↓
Validate input
  ↓
Load persona
  ↓
Build compact context
  ↓
Generate text
  ↓
Apply output policy
  ↓
Optional TTS
  ↓
Return result to host
```

## 4. Explicitly Excluded Flows

These flows should not exist in the core runtime:

```text
User chat → conversation memory → planning → tool calls → action
```

```text
Background poller → autonomous decision → external call → generated task
```

```text
Document ingestion → vector database → retrieval agent → answer
```

If a host system wants any of that, it should implement it outside Narrata and call Narrata only for final narration.

## 5. Package Layout

```text
narrata/
  go.mod
  README.md
  personas.default.json
  examples/
    game/
    home_automation/
    monitoring/
  internal/
    prompt/
    policy/
    validate/
  pkg/
    narrata/
      engine.go
      config.go
      request.go
      result.go
      persona.go
      errors.go
    backend/
      text/
        backend.go
        llama_cpp.go
        mock.go
        template.go
      tts/
        backend.go
        kokoro.go
        mock.go
```

Recommended public import:

```go
import "github.com/danielriddell21/narrata/pkg/narrata"
```

## 6. Core Components

### 6.1 Engine

The main embedded runtime object.

Responsibilities:

- Load config.
- Load personas.
- Manage model backend lifecycle.
- Accept generation requests.
- Return text/audio results.
- Expose close/shutdown methods.

Non-responsibilities:

- Conversation memory.
- Tool execution.
- Scheduling.
- Polling host systems.
- External data lookup.

### 6.2 Persona Registry

Stores personas loaded from default and user-provided persona files.

Responsibilities:

- Validate persona IDs.
- Resolve default persona.
- Merge built-in and user personas.
- Support optional hot reload later.

### 6.3 Prompt Builder

Converts request + persona + data into a compact model prompt.

Design goals:

- Keep prompts small.
- Prefer structured data summaries.
- Avoid leaking irrelevant data.
- Make output constraints explicit.
- Keep persona instructions separate from host data.
- Repeatedly reinforce that output should be narration, not chat.

### 6.4 Text Backend

Abstract interface for local text generation.

Initial target:

- llama.cpp/GGUF based backend.

Alternative backends later:

- MLX backend for Apple Silicon.
- Mock backend for tests.
- Template backend for very small devices.
- Optional cloud backend only as a separate adapter, not core default.

### 6.5 Output Policy

Post-processing stage that enforces host and persona constraints.

Examples:

- Maximum words.
- Maximum sentences.
- No markdown.
- No raw JSON unless requested.
- No profanity.
- No unsafe persona escalation.
- Strip surrounding quotes.
- Enforce silence if host/event policy says not to speak.

### 6.6 TTS Backend

Optional speech synthesis interface.

Initial target:

- Kokoro or Kokoro ONNX style backend.

TTS should be optional and pluggable. The text-generation engine must work without it.

## 7. Backend Strategy

### 7.1 Text Model

Preferred MVP route:

- GGUF model file.
- llama.cpp-backed inference.
- Go wrapper around llama.cpp or cgo integration.

Rationale:

- GGUF is a model format designed for efficient inference with GGML-style executors.
- llama.cpp is built for local inference across a wide range of hardware.
- Go can embed the backend directly through bindings or cgo.

### 7.2 TTS Model

Preferred MVP route:

- TTS interface first.
- Kokoro implementation second.
- Mock TTS for tests.

Rationale:

- Speech is useful but should not block the core runtime.
- Text output proves the engine first.
- Audio implementation may require ONNX/runtime decisions.

## 8. Concurrency Model

Narrata should support concurrent requests, but model backends may have their own limitations.

Recommended MVP:

- Engine is safe to call from multiple goroutines.
- Requests are queued or guarded if backend is single-session.
- Configurable max concurrency.
- Context cancellation is respected.

```go
type Config struct {
    MaxConcurrent int
}
```

## 9. Error Handling

Errors should be typed and actionable.

Examples:

- `ErrPersonaNotFound`
- `ErrInvalidRequest`
- `ErrModelNotLoaded`
- `ErrGenerationTimeout`
- `ErrTTSUnavailable`
- `ErrScopeViolation`

## 10. Privacy Model

Narrata is local-first. The host system is responsible for deciding what data to pass into Narrata.

No telemetry should be emitted by default.

## 11. Scope Guardrails for Implementation

Claude Code or any implementation agent should reject attempts to add the following to core:

- `ChatSession`
- `MemoryStore`
- `ToolRegistry`
- `Agent`
- `Planner`
- `Retriever`
- `VectorStore`
- `Browser`
- `Scheduler`

Allowed concepts:

- `Engine`
- `Request`
- `Result`
- `Persona`
- `Renderer`
- `TextBackend`
- `TTSBackend`
- `OutputPolicy`
- `EventPolicy`
