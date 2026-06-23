# Narrata MVP Roadmap

## Phase 0: Prototype Shape

**Status: implemented.** Engine API, persona loader, prompt builder, mock
backend, default personas, the monitoring example, and scope-guardrail tests
are all in place.

Goal: prove the embedded API shape without committing to model backend complexity.

Deliverables:

- Go module.
- `Engine` API.
- `Request` and `Result` types.
- Persona loader.
- Prompt builder.
- Mock text backend.
- Example `personas.json`.
- Example app for monitoring alerts.
- Scope guardrail tests/checklist: no assistant, chat, memory, tools, planner, or RAG primitives.

Exit criteria:

- Host Go app can call Narrata and get deterministic fake output.
- Personas validate correctly.
- Prompt generation has golden tests.
- Public API only reflects embedded narration concepts.

## Phase 1: Local Text Generation

**Status: implemented.** A pure-Go `template` backend generates real text
locally with no external process; timeout, cancellation, and bounded
concurrency are handled by the engine; the game example is included. An
experimental llama.cpp/GGUF backend is available behind the `llama` build tag.

Goal: generate real text locally.

Deliverables:

- llama.cpp/GGUF backend experiment.
- Model loading config.
- Timeout and cancellation.
- Basic concurrency handling.
- Response post-processing.
- Example game event integration.
- Template-only fallback backend for constrained systems.

Exit criteria:

- Event input produces useful persona-shaped text locally.
- No required external process.
- Works on MacBook/desktop.
- Backend is replaceable behind an interface.

## Phase 2: Speech Output

**Status: implemented.** The TTS interface, a deterministic mock backend
(valid WAV), optional audio output, and the home automation example are in
place; text-only usage is unaffected when TTS is disabled or its model is
missing. An experimental Kokoro/ONNX backend is available behind the `kokoro`
build tag.

Goal: optional text-to-speech output.

Deliverables:

- TTS interface.
- Mock TTS backend.
- Kokoro/ONNX feasibility implementation.
- Audio result as bytes.
- Example home automation announcement.

Exit criteria:

- Host app can request text only or text+speech.
- Speech layer can be disabled entirely.
- Missing TTS model does not break text-only usage.

## Phase 3: Developer Experience

**Status: implemented.** README quickstart and config examples, a persona
authoring guide, the `narrata` development CLI (`validate`, `gen`, `generate`,
`personas`), benchmarks for the prompt/policy/generate paths, example
integrations, and the "What Narrata Is Not" documentation are all in place.

Goal: make it usable by other developers.

Deliverables:

- README quickstart.
- Config examples.
- Persona authoring guide.
- Development-only CLI for validation/testing.
- Benchmarks.
- Example integrations.
- Documentation section called "What Narrata Is Not".

Exit criteria:

- A developer can embed Narrata from documentation alone.
- Default personas are useful.
- Common errors are understandable.
- Developers understand that Narrata is an embedded narration engine, not an assistant framework.

## Phase 4: Hardening

**Status: implemented on branch `feat/phase-4-hardening` (under review).**
Delivered: prompt-injection sanitization of untrusted event data, host-controlled
`slog` logging hooks (metadata only), versioned persona schema validation, a
regex-precompilation memory fix, and concurrency coverage under the race
detector. Event-policy behaviour is deferred to Phase 5.

Goal: make Narrata stable enough for real host systems.

Deliverables:

- Thread-safety review.
- Memory usage review.
- Backend abstraction cleanup.
- Prompt injection mitigation for untrusted event data.
- Structured logging hooks controlled by host.
- Versioned persona schema.
- Event policy groundwork for cooldown/importance/silence decisions.

Exit criteria:

- Narrata can run in long-lived host processes.
- Host systems can control logging and privacy.
- Persona files can evolve safely.
- Event policy remains rendering logic, not autonomous agent behaviour.

## Phase 5: Post-MVP Event Awareness

Goal: allow Narrata to decide how to render an event within host-defined boundaries.

Deliverables:

- `EventPolicy` support.
- Importance and urgency handling.
- Cooldown handling.
- Optional silence result.
- Per-persona event response intensity.

Exit criteria:

- Narrata can decide whether to speak, stay silent, shorten output, or increase drama.
- The host remains in charge of when Narrata is called.
- No background monitoring, scheduling, or autonomous task execution is added.

## Recommended First Build Order

1. `pkg/narrata` public API.
2. Scope guardrail checklist in README/spec.
3. Persona schema and loader.
4. Prompt builder.
5. Mock backend.
6. Default personas.
7. Monitoring alert example.
8. Game example.
9. llama.cpp/GGUF backend.
10. TTS interface.
11. Kokoro experiment.
12. Event policy groundwork.

## Suggested Repo Milestones

### v0.1.0

- Text-only embedded runtime.
- Mock backend and one real local backend.
- Default personas.
- Go examples.
- Explicit non-goals documented.

### v0.2.0

- Optional TTS backend.
- Home automation example.
- Better config and validation.

### v0.3.0

- Persona generation helper.
- Hot reload.
- Benchmarks.

### v0.4.0

- Multiple model backends.
- Streaming text.
- Improved concurrency.

### v0.5.0

- Event awareness.
- Cooldowns.
- Silence handling.
- Response intensity control.
