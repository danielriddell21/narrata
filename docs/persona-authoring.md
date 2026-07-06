# Persona Authoring Guide

Personas shape how Narrata turns data into text and speech. They are plain JSON,
editable by hand, and validated on load. This guide covers writing good ones.
For the full field reference see [personas.md](./personas.md).

## Where personas live

- A single document: `personas.json` with a `personas` array (set via
  `Config.PersonasPath`).
- One file per persona: a `personas/` directory of `*.json` files, each holding
  either a bare persona object or a full document (set via `Config.PersonasDir`).

Bundled defaults always load first; your files merge on top and may override a
default by reusing its `id`.

## Anatomy of a good persona

```json
{
  "id": "incident_reporter",
  "name": "Incident Reporter",
  "description": "Neutral, factual narration for service incidents.",
  "style": { "tone": "neutral", "energy": "medium", "humour": "none", "verbosity": "brief" },
  "voice": { "id": "neutral_clear", "speed": 1.0, "pitch": 1.0 },
  "rules": [
    "Lead with the most important fact.",
    "State impact and any required action.",
    "Do not speculate or invent details."
  ],
  "constraints": { "max_words": 30, "max_sentences": 2, "allow_profanity": false, "allow_markdown": false }
}
```

Tips:

- **Pick a clear `id`.** It is how hosts select the persona per request; keep it
  stable and lowercase with underscores.
- **Write rules as instructions, not prose.** Short imperative lines work best:
  "Lead with the key fact", "Keep it under 20 words".
- **Set constraints to match the surface.** A spoken home announcement wants
  `max_sentences: 1`; a dashboard summary can use two.
- **Keep humour honest.** When `humour` is on, add a rule like "never obscure the
  key information" so the joke never costs clarity.
- **Don't leak structure.** Add "Do not mention raw JSON" so field names and
  values are narrated, not dumped.
- **Use `event_policy` to pace and shape output.** Its `default_importance`,
  `intensity`, `cooldown_seconds`, and `allow_silence` fields adjust length,
  delivery energy, and cooldown silence; a request may override them.

## Scaffold one from a description

The `narrata gen` command infers a starting persona from a short description.
It is deterministic and adds no agent behaviour; treat the output as a draft to
review and tune.

```bash
narrata gen "dry witty narrator for build failures" -o personas/build_narrator.json
```

The same helper is available in code as `narrata.GeneratePersona`.

## Validate before shipping

```bash
narrata validate personas.json
narrata validate personas/
```

Validation checks that every persona has an `id`, `name`, `description`, and at
least one rule, and that numeric limits are non-negative. Invalid files fail
fast with a message naming the offending field.

## Try it without a model

The deterministic backends render personas without any model file, so you can
iterate on tone and constraints immediately:

```bash
narrata generate --persona incident_reporter --event service_degraded \
  --data '{"service":"payments-api","latency_ms":950}' --personas personas.json
```
