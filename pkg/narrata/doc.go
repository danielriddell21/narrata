// Package narrata is an embedded, local-first narration runtime for Go
// applications. It transforms host-provided structured data and events into
// short human-readable text and, optionally, speech, shaped by configurable
// personas.
//
// The core model is intentionally small:
//
//	input data -> persona -> generate narration -> text/audio result
//
// # Non-goals (enforced)
//
// Narrata is NOT an assistant, agent, chatbot, workflow engine, RAG platform,
// memory system, or tool-calling framework. It must never gain those concepts.
// The public API deliberately exposes none of the following, and contributions
// adding them will be rejected:
//
//   - Assistant, ChatSession
//   - MemoryStore
//   - Tool, ToolRegistry
//   - Agent, Planner
//   - Retriever, VectorStore
//   - Browser, Scheduler
//
// Allowed concepts are: Engine, Request, Result, Persona, Renderer,
// TextBackend, TTSBackend, OutputPolicy, and EventPolicy. The host application
// owns execution, lifecycle, data, permissions, timing, and output routing.
// Narrata only converts host-provided data into narration; it does not poll,
// remember, plan, browse, call tools, or hold conversations.
//
// Narrata is local-first and emits no telemetry. The host decides what data is
// passed in, when generation happens, and where output goes.
package narrata
