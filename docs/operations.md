# Operations & Hardening

Notes for running Narrata inside long-lived host processes.

## Concurrency

The `Engine` is safe for concurrent use from multiple goroutines. Persona reads
and writes are guarded by a read/write mutex, and in-flight generations are
bounded by `Config.MaxConcurrent` (zero means unbounded). Context cancellation
and the optional `Config.Timeout` are honoured per request. The test suite
exercises concurrent `Generate` and `Personas().Save` calls under the race
detector.

## Logging

Pass a `*slog.Logger` as `Config.Logger` to receive structured logs; leave it
nil to disable logging entirely. Narrata logs **metadata only** — persona id,
event name, token counts, and timings — and never the event data, preserving
the local-first privacy model.

```go
engine, _ := narrata.New(narrata.Config{
    Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})),
})
```

## Untrusted event data

Event data, keys, the event name, and any instruction are treated as untrusted:
before they enter a prompt, newlines and tabs are collapsed to spaces and other
control characters are dropped. This prevents injected content from forging
prompt lines (for example a fake `Narration:` directive) or breaking the offline
backends' Event/Data parsing. Persona rules remain host-authored and trusted.

## Persona schema versioning

A `personas.json` document may declare a `version`. Files whose **major** version
differs from `narrata.SchemaVersion` are rejected on load, so the schema can
evolve without silently mis-reading old files. An empty version is treated as
compatible.

## Privacy

Narrata is local-first and emits no telemetry. The host decides what data is
passed in, when generation happens, and where output goes. Models and personas
live inside the host application.
