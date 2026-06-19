# Narrata Research Notes

## Local Text Inference

Narrata should target local inference through a model format and runtime that can be embedded into host systems.

Recommended direction:

- GGUF model files.
- llama.cpp-backed runtime.
- Go binding or cgo integration.

Notes:

- llama.cpp is designed for local LLM inference with minimal setup and strong performance across hardware.
- GGUF is designed for GGML-style executors and efficient model loading/saving for inference.
- There are Go binding projects that expose llama.cpp-style local inference to Go applications, but backend choice should be tested before committing.

## Text-to-Speech

Recommended direction:

- Keep TTS optional.
- Define a stable `TTSBackend` interface early.
- Use mock TTS first.
- Evaluate Kokoro/Kokoro ONNX for a lightweight local implementation.

Notes:

- Kokoro is an open-weight 82M parameter TTS model.
- ONNX packaging may be attractive for embedding, depending on Go runtime support and deployment constraints.

## Why Not Ollama First?

Ollama is useful for prototyping, but Narrata is intended to be embedded directly into host systems. Depending on an Ollama sidecar would make Narrata feel like a separate service rather than baked-in runtime.

Possible compromise:

- Allow an Ollama backend for development only.
- Keep the production embedded path focused on direct local inference.

## Product Differentiation

Narrata should not position itself as a generic local AI runtime. That market is crowded.

The stronger position is:

> An embedded narration engine that turns structured application data into human text and speech.

Most local AI tooling focuses on chat, agents, or model hosting. Narrata should focus on making software explain what is happening in real time.

## Scope Guardrails

The core package should avoid these concepts:

- Assistant.
- Chat session.
- Memory.
- Tool calling.
- Planner.
- Agent.
- RAG.
- Vector database.
- Workflow engine.
- Scheduler.

Allowed concepts:

- Event.
- Persona.
- Narration.
- Text generation.
- Speech generation.
- Output policy.
- Event policy.

## Key Technical Risks

| Risk | Mitigation |
|---|---|
| Go LLM bindings are immature or unstable | Keep backend interface abstract; prototype more than one option. |
| TTS adds deployment complexity | Make TTS optional and second-phase. |
| Model startup time is too slow | Long-lived engine instance; preload models. |
| Output is too verbose | Strong persona constraints and post-processing. |
| Host passes too much data | Context builder should compact and filter input. |
| Embedded binary gets too large | Let host package models separately. |
| Persona consistency is weak | Add examples/few-shot support before fine-tuning. |
| Scope creeps into assistant framework | Keep non-goals in spec, README, and implementation review checklist. |

## Reference Links

- llama.cpp GitHub: https://github.com/ggml-org/llama.cpp
- Hugging Face GGUF docs: https://huggingface.co/docs/hub/en/gguf
- Kokoro-82M Hugging Face: https://huggingface.co/hexgrad/Kokoro-82M
- Kokoro ONNX: https://huggingface.co/onnx-community/Kokoro-82M-v1.0-ONNX
- seed-hypermedia/llama-go: https://github.com/seed-hypermedia/llama-go
- hybridgroup/yzma: https://github.com/hybridgroup/yzma
