# Development CLI

`cmd/narrata` is a developer aid for validating personas, scaffolding new ones,
and rendering events from the terminal. Narrata itself is embedded as a library;
the CLI is not a runtime service.

```bash
go run ./cmd/narrata <command> [flags]
# or install it:
go install github.com/danielriddell21/narrata/cmd/narrata@latest
```

## Commands

### `validate <path>`

Validate a `personas.json` file or a `personas/` directory. Loads the file
through the engine (which validates every persona) and reports the result.

```bash
narrata validate personas.json
narrata validate personas/
```

### `gen <description>`

Scaffold a persona from a short description and print it as JSON. Use `-o` to
write it to a file instead of stdout.

```bash
narrata gen "calm ship computer for a space sim"
narrata gen "pirate captain for server alerts" -o personas/pirate.json
```

### `generate`

Render a single event to narration text using the deterministic backends (no
model file required).

| Flag | Description |
|------|-------------|
| `--persona` | Persona id (default: the configured default). |
| `--event` | Event name, e.g. `door_opened`. |
| `--data` | Event data as a JSON object. |
| `--personas` | Path to a `personas.json` file. |
| `--dir` | Path to a `personas/` directory. |
| `--backend` | Text backend: `mock` or `template` (default `template`). |
| `--max-words` | Override the persona word limit. |
| `--instruction` | Optional one-off steering note. |

```bash
narrata generate --persona funny_narrator --event service_degraded \
  --data '{"service":"payments-api","cpu":96}'
```

### `personas`

List the bundled default personas with their descriptions.

### `version`

Print the CLI version.
