package nlg

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Kind classifies a data field so it can be realised and scored appropriately.
type Kind int

// Field kinds.
const (
	KindOther    Kind = iota // fallback
	KindQuantity             // numbers / measurements
	KindName                 // named entities (people, enemies, services)
	KindPlace                // rooms, locations
	KindTime                 // times, dates
	KindFlag                 // booleans / states
	KindState                // descriptive state strings
)

// Field is one typed datum that may be narrated.
type Field struct {
	Key   string
	Value any
	Kind  Kind // inferred when zero (KindOther) and Value/Key give a better hint
}

// flattenData turns an arbitrary JSON-serialisable value into typed Fields with
// stable, sorted keys. Nested objects are dotted (e.g. "player.health").
func flattenData(data any) []Field {
	if data == nil {
		return nil
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return []Field{{Key: "value", Value: fmt.Sprintf("%v", data)}}
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return []Field{{Key: "value", Value: string(raw)}}
	}
	var fields []Field
	flatten("", decoded, &fields)
	sort.Slice(fields, func(i, j int) bool { return fields[i].Key < fields[j].Key })
	return fields
}

func flatten(prefix string, v any, out *[]Field) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flatten(key, val, out)
		}
	case []any:
		for i, val := range t {
			key := strconv.Itoa(i)
			if prefix != "" {
				key = prefix + "." + key
			}
			flatten(key, val, out)
		}
	default:
		key := prefix
		if key == "" {
			key = "value"
		}
		*out = append(*out, Field{Key: key, Value: t})
	}
}

// withKinds returns a copy of fields with Kind inferred where unset.
func withKinds(fields []Field) []Field {
	out := make([]Field, len(fields))
	for i, f := range fields {
		if f.Kind == KindOther {
			f.Kind = inferKind(f.Key, f.Value)
		}
		out[i] = f
	}
	return out
}

func inferKind(key string, v any) Kind {
	switch t := v.(type) {
	case bool:
		return KindFlag
	case float64, float32, int, int64, int32:
		return KindQuantity
	case string:
		if _, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil && t != "" {
			return KindQuantity
		}
	}
	lk := strings.ToLower(key)
	switch {
	case containsAny(lk, "room", "place", "location", "zone", "area"):
		return KindPlace
	case containsAny(lk, "time", "date", "when", "clock"):
		return KindTime
	case containsAny(lk, "name", "player", "user", "enemy", "who", "actor", "service"):
		return KindName
	default:
		return KindOther
	}
}

// score ranks a field's narration-worthiness.
func score(k Kind) int {
	switch k {
	case KindQuantity, KindName:
		return 3
	case KindPlace, KindState, KindFlag:
		return 2
	case KindTime:
		return 1
	default:
		return 1
	}
}

// salient returns up to max fields, highest score first, stable within a score.
func salient(fields []Field, max int) []Field {
	if max <= 0 || len(fields) <= max {
		return fields
	}
	idx := make([]int, len(fields))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		return score(fields[idx[a]].Kind) > score(fields[idx[b]].Kind)
	})
	kept := make([]Field, 0, max)
	for _, i := range idx[:max] {
		kept = append(kept, fields[i])
	}
	// Restore original (sorted-by-key) order among the kept fields.
	sort.SliceStable(kept, func(a, b int) bool { return kept[a].Key < kept[b].Key })
	return kept
}

// realize turns a field into a short natural fragment.
func realize(f Field) string {
	switch f.Kind {
	case KindName:
		return valueString(f.Value)
	case KindPlace:
		return "in the " + valueString(f.Value)
	case KindTime:
		return "at " + valueString(f.Value)
	case KindFlag:
		if b, ok := f.Value.(bool); ok && !b {
			return "no " + humanizeKey(f.Key)
		}
		return humanizeKey(f.Key)
	case KindQuantity:
		return humanizeKey(f.Key) + " " + valueString(f.Value)
	default:
		return humanizeKey(f.Key) + " " + valueString(f.Value)
	}
}

func valueString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func humanizeKey(key string) string {
	key = strings.ReplaceAll(key, "_", " ")
	key = strings.ReplaceAll(key, ".", " ")
	return strings.TrimSpace(key)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
