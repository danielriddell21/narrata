# Event Policy

Event policy lets Narrata decide *how* to render an event within host-defined
boundaries. It is rendering policy, not autonomy: Narrata never schedules work,
polls, or decides tasks. The host still decides **when** to call `Generate`.

Policy comes from two places and is resolved per request:

- `Request.EventPolicy` — per-call overrides.
- `Persona.EventPolicy` (the `event_policy` JSON field) — persona defaults.

## Cooldown and silence

When an event recurs within its cooldown window **and** silence is allowed,
Narrata returns a silent result (`Result.Silent == true`, empty text, no audio)
instead of repeating itself.

```go
res, _ := engine.Generate(ctx, narrata.Request{
    PersonaID: "home_announcer",
    Event:     "motion_detected",
    Data:      map[string]any{"room": "hallway"},
    EventPolicy: narrata.EventPolicy{
        Cooldown:     30 * time.Second,
        AllowSilence: true,
    },
})
if res.Silent {
    return // within cooldown; nothing to announce
}
```

Cooldown is tracked per persona + event. The request `Cooldown` takes precedence
over the persona's `cooldown_seconds`. Without `AllowSilence`, output is never
suppressed.

## Importance, urgency, and intensity

| Field | Effect |
|-------|--------|
| `Importance` | `low` tightens the word limit (~60%); `high` relaxes it (~130%). |
| `Urgency` | `immediate`/`urgent`/`high` forces a single, terse sentence. |
| `Intensity` (persona) | `high` adds a "more drama, keep the facts clear" directive; `low` keeps it understated. |

With all fields unset, rendering is unchanged — default requests behave exactly
as before. These adjustments shape the prompt and the output policy only; they
never change *whether* the host calls Narrata.
