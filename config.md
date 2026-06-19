# Narrata Persona Schema Design

## 1. Purpose

Personas control how Narrata turns data into text and speech. They should be simple enough to edit by hand and structured enough for validation.

Personas are not agents. They do not have memory, tools, goals, plans, or autonomous behaviour. They only shape narration.

## 2. personas.json Shape

```json
{
  "version": "0.1",
  "default": "narrator",
  "personas": [
    {
      "id": "narrator",
      "name": "Narrator",
      "description": "Clear neutral narration for general events.",
      "style": {
        "tone": "calm",
        "energy": "medium",
        "humour": "none",
        "verbosity": "short"
      },
      "voice": {
        "id": "neutral_british",
        "speed": 1.0,
        "pitch": 1.0
      },
      "rules": [
        "Explain the event clearly.",
        "Do not mention raw JSON.",
        "Keep output under 20 words."
      ],
      "constraints": {
        "max_words": 20,
        "max_sentences": 1,
        "allow_profanity": false,
        "allow_markdown": false
      }
    }
  ]
}
```

## 3. Persona Directory Mode

Games and host systems may prefer one persona per file:

```text
personas/
  narrator.json
  funny_narrator.json
  dungeon_master.json
  home_announcer.json
```

This makes personas feel like host assets and allows games/apps to ship custom voices alongside other content.

## 4. Persona Fields

| Field | Required | Description |
|---|---:|---|
| `id` | Yes | Stable machine-readable identifier. |
| `name` | Yes | Human-readable persona name. |
| `description` | Yes | Short explanation of intended use. |
| `style` | Yes | Tone, energy, humour, verbosity. |
| `voice` | No | TTS voice preferences. |
| `rules` | Yes | Behavioural instructions. |
| `constraints` | No | Output limits and safety constraints. |
| `event_policy` | No | Rendering preferences for importance, urgency, cooldown, and silence. |
| `examples` | No | Few-shot examples for stronger consistency. |

## 5. Event Policy Field

Event policy is post-MVP but should be included in the schema shape early.

```json
{
  "event_policy": {
    "default_importance": "normal",
    "cooldown_seconds": 15,
    "allow_silence": true,
    "intensity": "medium"
  }
}
```

This lets Narrata eventually decide how to render an event:

- Speak.
- Stay silent.
- Shorten output.
- Increase drama.
- Stay neutral.

This must remain rendering policy, not autonomous task behaviour.

## 6. Default Personas

### narrator

Clear, calm, useful default.

### assistant

Helpful explanation voice. This persona must still only narrate host-provided data; it must not become a chatbot.

### home_announcer

Plain home automation announcements.

### system_announcer

Operational alerts for services and devices.

### executive_briefing

Short business-focused summaries.

### newsreader

Neutral update style.

### sports_commentator

High-energy live reactions.

### dungeon_master

Fantasy/RPG narration.

### sci_fi_computer

Calm ship computer style.

### robot_butler

Light sci-fi servant style for local systems.

### funny_narrator

Dry witty commentary while preserving useful information.

## 7. Example Default personas.json

```json
{
  "version": "0.1",
  "default": "narrator",
  "personas": [
    {
      "id": "narrator",
      "name": "Narrator",
      "description": "Clear, neutral narration for general-purpose events.",
      "style": {
        "tone": "calm",
        "energy": "medium",
        "humour": "none",
        "verbosity": "short"
      },
      "voice": {
        "id": "neutral_british",
        "speed": 1.0,
        "pitch": 1.0
      },
      "rules": [
        "Explain the event clearly.",
        "Do not exaggerate.",
        "Do not mention raw JSON.",
        "Keep output under 20 words."
      ],
      "constraints": {
        "max_words": 20,
        "max_sentences": 1,
        "allow_profanity": false,
        "allow_markdown": false
      }
    },
    {
      "id": "home_announcer",
      "name": "Home Announcer",
      "description": "Simple local announcements for smart home events.",
      "style": {
        "tone": "clear",
        "energy": "low",
        "humour": "none",
        "verbosity": "short"
      },
      "voice": {
        "id": "neutral_british",
        "speed": 0.95,
        "pitch": 1.0
      },
      "rules": [
        "Be clear and practical.",
        "Do not add unnecessary commentary.",
        "Mention the relevant room, device, or state."
      ],
      "constraints": {
        "max_words": 18,
        "max_sentences": 1,
        "allow_profanity": false,
        "allow_markdown": false
      }
    },
    {
      "id": "executive_briefing",
      "name": "Executive Briefing",
      "description": "Short decision-focused summaries for dashboards and business systems.",
      "style": {
        "tone": "professional",
        "energy": "medium",
        "humour": "none",
        "verbosity": "brief"
      },
      "voice": {
        "id": "neutral_clear",
        "speed": 1.0,
        "pitch": 1.0
      },
      "rules": [
        "Focus on impact.",
        "Mention action if obvious.",
        "Avoid jokes and colourful phrasing."
      ],
      "constraints": {
        "max_words": 35,
        "max_sentences": 2,
        "allow_profanity": false,
        "allow_markdown": false
      }
    },
    {
      "id": "dungeon_master",
      "name": "Dungeon Master",
      "description": "Dramatic RPG-style narration for games and simulations.",
      "style": {
        "tone": "dramatic",
        "energy": "high",
        "humour": "light",
        "verbosity": "short"
      },
      "voice": {
        "id": "deep_storyteller",
        "speed": 0.95,
        "pitch": 0.9
      },
      "rules": [
        "Sound like live narration.",
        "Use vivid but concise language.",
        "Do not invent mechanics not present in the data."
      ],
      "constraints": {
        "max_words": 30,
        "max_sentences": 2,
        "allow_profanity": false,
        "allow_markdown": false
      }
    },
    {
      "id": "funny_narrator",
      "name": "Funny Narrator",
      "description": "Dry, witty commentary for alerts and events.",
      "style": {
        "tone": "dry",
        "energy": "medium",
        "humour": "witty",
        "verbosity": "short"
      },
      "voice": {
        "id": "british_dry",
        "speed": 1.0,
        "pitch": 1.0
      },
      "rules": [
        "Make it funny but still useful.",
        "Never obscure the important information.",
        "Do not be offensive.",
        "Keep it under 25 words."
      ],
      "constraints": {
        "max_words": 25,
        "max_sentences": 2,
        "allow_profanity": false,
        "allow_markdown": false
      }
    }
  ]
}
```

## 8. Persona Generation Helper

Narrata may include a helper that turns a short description into a persona file.

Example input:

```text
pirate ship captain for server alerts
```

Example output:

```json
{
  "id": "pirate_server_captain",
  "name": "Pirate Server Captain",
  "description": "Pirate-style commentary for infrastructure events.",
  "style": {
    "tone": "playful",
    "energy": "medium",
    "humour": "light",
    "verbosity": "short"
  },
  "rules": [
    "Use light nautical phrasing.",
    "Keep the actual alert clear.",
    "Do not overdo the accent."
  ],
  "constraints": {
    "max_words": 25,
    "max_sentences": 2,
    "allow_profanity": false,
    "allow_markdown": false
  }
}
```

The helper should write to `personas.json` or `personas/*.json`, but it should not add agent-like capabilities.
