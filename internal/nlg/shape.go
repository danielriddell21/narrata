package nlg

import (
	"fmt"
	"strings"
)

type shape int

const (
	shapeGeneric shape = iota
	shapeCompletion
	shapeThreshold
	shapeArrival
	shapeDeparture
	shapeChange
	shapeAction // unclassified but verb-like last word (e.g. "player_jumped")
)

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
	// No known keyword: if the last word is verb-like and has a subject before
	// it, treat it as a generic action ("player_jumped" -> "The player jumped").
	if n := len(words); n >= 2 && isVerbLike(words[n-1]) {
		return shapeAction, strings.Join(words[:n-1], " "), words[n-1]
	}
	return shapeGeneric, "", ""
}

func isVerbLike(w string) bool {
	return len(w) >= 4 && (strings.HasSuffix(w, "ed") || strings.HasSuffix(w, "ing"))
}

func subjectPhrase(noun string, fields []Field) (subject string, rest []Field, matched bool) {
	if noun == "" {
		return "", fields, false
	}
	for i, f := range fields {
		if strings.EqualFold(humanizeKey(f.Key), noun) {
			return valueString(f.Value), removeAt(fields, i), true
		}
	}
	return "the " + noun, fields, false
}

func takeFirstName(fields []Field) (subject string, rest []Field, ok bool) {
	for i, f := range fields {
		if f.Kind == KindName {
			return valueString(f.Value), removeAt(fields, i), true
		}
	}
	return "", fields, false
}

func removeAt(fields []Field, i int) []Field {
	rest := make([]Field, 0, len(fields)-1)
	rest = append(rest, fields[:i]...)
	rest = append(rest, fields[i+1:]...)
	return rest
}

type bucket int

const (
	bucketPlain bucket = iota // calm, professional, neutral, clear
	bucketDrama               // dramatic, excited, aggressive
	bucketWry                 // dry, witty, playful, sarcastic
)

func toneBucket(tone string) bucket {
	switch tone {
	case "dramatic", "excited", "aggressive", "intense", "high":
		return bucketDrama
	case "dry", "witty", "playful", "sarcastic", "cheeky":
		return bucketWry
	default:
		return bucketPlain
	}
}

func byBucket(b bucket, plain, drama, wry []string) []string {
	switch b {
	case bucketDrama:
		if len(drama) > 0 {
			return drama
		}
	case bucketWry:
		if len(wry) > 0 {
			return wry
		}
	}
	return plain
}

func clauseFor(sh shape, action, subject string, b bucket, r *rng) string {
	return fmt.Sprintf(r.pick(templatesFor(sh, action, b)), subject)
}

func templatesFor(sh shape, action string, b bucket) []string {
	switch sh {
	case shapeCompletion:
		return byBucket(b,
			[]string{"%s has finished", "%s is done"},
			[]string{"%s is complete at last", "%s is finally done"},
			[]string{"%s wrapped up", "%s is done, somehow"})
	case shapeThreshold:
		return thresholdTemplates(action, b)
	case shapeArrival:
		return arrivalTemplates(action, b)
	case shapeDeparture:
		return departureTemplates(action, b)
	case shapeChange:
		return changeTemplates(action, b)
	default:
		return []string{"%s"}
	}
}

func thresholdTemplates(action string, b bucket) []string {
	switch action {
	case "low", "under", "empty", "drained":
		return byBucket(b,
			[]string{"%s is running low", "%s is nearly out"},
			[]string{"%s is fading fast", "%s is on the brink"},
			[]string{"%s is running low, naturally", "%s is almost gone"})
	case "critical", "breach":
		return byBucket(b,
			[]string{"%s has hit critical", "%s needs attention now"},
			[]string{"%s has gone critical", "%s is at the edge"},
			[]string{"%s is, of course, critical", "%s needs a look"})
	default: // high, over, full, exceeded, overheat, overloaded
		return byBucket(b,
			[]string{"%s is over the line", "%s is spiking"},
			[]string{"%s is surging", "%s is off the charts"},
			[]string{"%s is over the top, naturally", "%s is high"})
	}
}

func arrivalTemplates(action string, b bucket) []string {
	switch action {
	case "opened", "open":
		return byBucket(b,
			[]string{"%s opened", "%s is open"},
			[]string{"%s swings open", "%s stands open"},
			[]string{"%s is open, at last", "%s opened"})
	case "started", "start", "spawned":
		return byBucket(b,
			[]string{"%s has started", "%s is up"},
			[]string{"%s roars to life", "%s is underway"},
			[]string{"%s finally started", "%s is up"})
	default: // arrived, connected, detected, joined, entered, online
		return byBucket(b,
			[]string{"%s has arrived", "%s is online"},
			[]string{"%s has arrived at last", "%s bursts online"},
			[]string{"%s turned up", "%s is online"})
	}
}

func departureTemplates(action string, b bucket) []string {
	switch action {
	case "closed":
		return byBucket(b,
			[]string{"%s closed", "%s is shut"},
			[]string{"%s slams shut", "%s is sealed"},
			[]string{"%s closed, finally", "%s is shut"})
	case "died", "killed", "destroyed", "lost":
		return byBucket(b,
			[]string{"%s is down", "%s has fallen"},
			[]string{"%s has fallen", "%s is no more"},
			[]string{"%s is down, alas", "%s is gone"})
	default: // stopped, offline, disconnected, failed, down, left, exited
		return byBucket(b,
			[]string{"%s has stopped", "%s went dark"},
			[]string{"%s has gone dark", "%s is silenced"},
			[]string{"%s stopped, naturally", "%s went dark"})
	}
}

func changeTemplates(action string, b bucket) []string {
	if action == "degraded" {
		return byBucket(b,
			[]string{"%s has degraded", "%s is struggling"},
			[]string{"%s is buckling", "%s is in trouble"},
			[]string{"%s is having a moment", "%s is struggling"})
	}
	return byBucket(b,
		[]string{"%s has changed", "%s updated"},
		[]string{"%s has shifted", "%s transformed"},
		[]string{"%s changed, naturally", "%s updated"})
}

func terminate(line string, style Style) string {
	if (style.Energy == "high" || toneBucket(style.Tone) == bucketDrama) &&
		strings.HasSuffix(line, ".") {
		return line[:len(line)-1] + "!"
	}
	return line
}
