package nlg

import (
	"fmt"
	"strings"
)

// shape is the narrative family an event falls into, inferred from its name.
type shape int

const (
	shapeGeneric shape = iota
	shapeCompletion
	shapeThreshold
	shapeArrival
	shapeDeparture
	shapeChange
)

// shapeKeywords maps an event's action word to its shape.
var shapeKeywords = map[string]shape{
	"done": shapeCompletion, "complete": shapeCompletion, "completed": shapeCompletion,
	"finished": shapeCompletion, "finish": shapeCompletion, "ready": shapeCompletion,

	"low": shapeThreshold, "critical": shapeThreshold, "high": shapeThreshold,
	"over": shapeThreshold, "under": shapeThreshold, "empty": shapeThreshold,
	"full": shapeThreshold, "breach": shapeThreshold, "exceeded": shapeThreshold,
	"overheat": shapeThreshold, "overloaded": shapeThreshold, "drained": shapeThreshold,

	"opened": shapeArrival, "open": shapeArrival, "started": shapeArrival,
	"start": shapeArrival, "arrived": shapeArrival, "connected": shapeArrival,
	"detected": shapeArrival, "spawned": shapeArrival, "joined": shapeArrival,
	"entered": shapeArrival, "online": shapeArrival,

	"closed": shapeDeparture, "left": shapeDeparture, "exited": shapeDeparture,
	"stopped": shapeDeparture, "disconnected": shapeDeparture, "killed": shapeDeparture,
	"died": shapeDeparture, "destroyed": shapeDeparture, "lost": shapeDeparture,
	"offline": shapeDeparture, "down": shapeDeparture, "failed": shapeDeparture,

	"changed": shapeChange, "updated": shapeChange, "degraded": shapeChange,
	"increased": shapeChange, "decreased": shapeChange, "moved": shapeChange,
	"switched": shapeChange, "renamed": shapeChange,
}

// classifyEvent splits an event name into a shape, the subject nouns (words
// before the action), and the action word. It returns shapeGeneric when no
// action keyword is found or there is no subject to attach it to.
func classifyEvent(event string) (shape, string, string) {
	words := strings.FieldsFunc(strings.ToLower(event), func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i := len(words) - 1; i >= 0; i-- {
		if sh, ok := shapeKeywords[words[i]]; ok {
			subject := strings.Join(words[:i], " ")
			if subject == "" {
				return shapeGeneric, "", ""
			}
			return sh, subject, words[i]
		}
	}
	return shapeGeneric, "", ""
}

// subjectPhrase resolves the subject: if a data field's key matches the subject
// noun, its value becomes a proper subject (no article) and is removed from the
// remaining fields; otherwise the noun is used with a definite article.
func subjectPhrase(noun string, fields []Field) (string, []Field) {
	if noun == "" {
		return "", fields
	}
	for i, f := range fields {
		if strings.EqualFold(humanizeKey(f.Key), noun) {
			rest := make([]Field, 0, len(fields)-1)
			rest = append(rest, fields[:i]...)
			rest = append(rest, fields[i+1:]...)
			return valueString(f.Value), rest
		}
	}
	return "the " + noun, fields
}

// clauseFor builds the verb clause for a shape/action, filled with subject.
func clauseFor(sh shape, action, subject string, r *rng) string {
	return fmt.Sprintf(r.pick(templatesFor(sh, action)), subject)
}

func templatesFor(sh shape, action string) []string {
	switch sh {
	case shapeCompletion:
		return []string{"%s has finished", "%s is done", "%s wrapped up"}
	case shapeThreshold:
		switch action {
		case "low", "under", "empty", "drained":
			return []string{"%s is running low", "%s is nearly out"}
		case "critical", "breach":
			return []string{"%s has hit critical", "%s needs attention now"}
		default: // high, over, full, exceeded, overheat, overloaded
			return []string{"%s is over the line", "%s is spiking"}
		}
	case shapeArrival:
		switch action {
		case "opened", "open":
			return []string{"%s opened", "%s is open"}
		case "started", "start", "spawned":
			return []string{"%s has started", "%s is up"}
		default: // arrived, connected, detected, joined, entered, online
			return []string{"%s has arrived", "%s is online"}
		}
	case shapeDeparture:
		switch action {
		case "closed":
			return []string{"%s closed", "%s is shut"}
		case "died", "killed", "destroyed", "lost":
			return []string{"%s is down", "%s has fallen"}
		default: // stopped, offline, disconnected, failed, down, left, exited
			return []string{"%s has stopped", "%s went dark"}
		}
	case shapeChange:
		if action == "degraded" {
			return []string{"%s has degraded", "%s is struggling"}
		}
		return []string{"%s has changed", "%s updated"}
	default:
		return []string{"%s"}
	}
}
