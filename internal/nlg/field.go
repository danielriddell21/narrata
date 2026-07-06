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
	if _, ok := v.(bool); ok {
		return KindFlag
	}
	lk := strings.ToLower(key)
	// An integer keyed by an index word (minute, lap, round) is an ordinal
	// position in time, not a measurement: it realises as "in the 89th minute".
	if ordinalIndexKeys[lk] {
		if _, ok := toInt(v); ok {
			return KindTime
		}
	}
	switch t := v.(type) {
	case float64, float32, int, int64, int32:
		return KindQuantity
	case string:
		if _, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil && t != "" {
			return KindQuantity
		}
	}
	switch {
	case containsAny(lk, "room", "place", "location", "zone", "area",
		"region", "datacenter", "datacentre", "cluster", "site"):
		return KindPlace
	case containsAny(lk, "time", "date", "when", "clock"):
		return KindTime
	case containsAny(lk, "name", "player", "user", "enemy", "who", "actor", "service",
		"team", "host", "node", "device", "app", "job", "channel", "sensor",
		"bot", "npc", "unit", "character", "agent", "worker",
		"opponent", "foe", "boss", "villain", "attacker", "adversary"):
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

// fieldScore ranks a field's narration-worthiness, boosting notable quantities:
// values with units, implied percentages, and percentages near their limit read
// as the interesting fact.
func fieldScore(f Field) int {
	s := score(f.Kind)
	if f.Kind != KindQuantity {
		return s
	}
	if _, unit := unitFor(f.Key); unit != "" {
		s++
	}
	if impliedPercent(f.Key, f.Value) {
		s++
		if v, ok := toFloat(f.Value); ok && (v >= 90 || v <= 10) {
			s++ // near a limit — likely why the event fired
		}
	}
	return s
}

// partitionAdjuncts splits fields into adjuncts, which attach to a verb clause
// ("... in the utility room", "... facing the Bone Dragon"), and coordinate
// details, which read as a list ("cpu 96% and latency 950ms"). Adjuncts are
// ordered place, then time, then relation, so the clause reads naturally
// ("done in the lab at noon facing the boss").
func partitionAdjuncts(fields []Field) (adjuncts, details []Field) {
	var places, times, relations []Field
	for _, f := range fields {
		switch {
		case f.Kind == KindPlace:
			places = append(places, f)
		case f.Kind == KindTime:
			times = append(times, f)
		case f.Kind == KindName && nameRelation(f.Key) != "":
			relations = append(relations, f)
		default:
			details = append(details, f)
		}
	}
	adjuncts = append(places, times...)
	adjuncts = append(adjuncts, relations...)
	return adjuncts, details
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
		return fieldScore(fields[idx[a]]) > fieldScore(fields[idx[b]])
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
		return realizeName(f)
	case KindPlace:
		return realizePlace(f)
	case KindTime:
		return realizeTime(f)
	case KindFlag:
		if b, ok := f.Value.(bool); ok && !b {
			return "no " + humanizeKey(f.Key)
		}
		return humanizeKey(f.Key)
	case KindQuantity:
		return realizeQuantity(f)
	default:
		return humanizeKey(f.Key) + " " + valueString(f.Value)
	}
}

// realizeStatement turns a field into a full copular clause ("cpu is at 96%")
// for the second sentence of verbose narration, where a fragment would read as
// telegraphic.
func realizeStatement(f Field) string {
	switch f.Kind {
	case KindQuantity:
		return quantityStatement(f)
	case KindFlag:
		if b, ok := f.Value.(bool); ok && !b {
			return humanizeKey(f.Key) + " is off"
		}
		return humanizeKey(f.Key) + " is on"
	case KindPlace, KindTime:
		return realize(f)
	default:
		return humanizeKey(f.Key) + " is " + valueString(f.Value)
	}
}

// quantityStatement renders a numeric field as a copular clause: "latency is
// 950ms", "cpu is at 96%", "queue is 3".
func quantityStatement(f Field) string {
	base, unit := unitFor(f.Key)
	v := formatNumber(f.Value)
	if unit != "" {
		return humanizeKey(base) + " is " + v + unit
	}
	if impliedPercent(f.Key, f.Value) {
		return humanizeKey(f.Key) + " is at " + v + "%"
	}
	return humanizeKey(f.Key) + " is " + v
}

// ordinalIndexKeys name an integer's position in time ("minute 89" -> "in the
// 89th minute") rather than a measured quantity.
var ordinalIndexKeys = map[string]bool{
	"minute": true, "lap": true, "round": true, "wave": true,
}

// realizeTime renders a time field: an ordinal index ("in the 89th minute") for
// an index key, else a timestamp ("at 14:30").
func realizeTime(f Field) string {
	lk := strings.ToLower(f.Key)
	if ordinalIndexKeys[lk] {
		if n, ok := toInt(f.Value); ok {
			return "in the " + ordinal(n) + " " + lk
		}
	}
	return "at " + valueString(f.Value)
}

// ordinal renders an integer as an English ordinal ("1st", "2nd", "89th").
func ordinal(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return strconv.Itoa(n) + suffix
}

// realizePlace renders a location. A named location key (region, datacenter,
// cluster, site) takes no article ("in eu-west"); a common location does
// ("in the utility room").
func realizePlace(f Field) string {
	v := valueString(f.Value)
	if containsAny(strings.ToLower(f.Key),
		"region", "datacenter", "datacentre", "cluster", "site") {
		return "in " + v
	}
	return "in the " + v
}

// realizeName renders a named entity. A combatant key ("enemy", "boss") frames
// the name as a relation ("facing the Bone Dragon"); other names render bare.
// The definite article is added only for multi-word names ("the Bone Dragon"),
// not proper names ("facing Ari").
func realizeName(f Field) string {
	name := valueString(f.Value)
	rel := nameRelation(f.Key)
	if rel == "" {
		return name
	}
	if strings.Contains(strings.TrimSpace(name), " ") {
		return rel + " the " + name
	}
	return rel + " " + name
}

// nameRelation returns the preposition that frames a named non-subject entity,
// or "" when the key implies no particular relation.
func nameRelation(key string) string {
	if containsAny(strings.ToLower(key),
		"enemy", "opponent", "foe", "boss", "villain", "attacker", "adversary") {
		return "facing"
	}
	return ""
}

// realizeQuantity formats a numeric field with a unit inferred from the key
// ("latency_ms" -> "latency 950ms") or an implied percent ("cpu" 96 -> "cpu 96%").
func realizeQuantity(f Field) string {
	base, unit := unitFor(f.Key)
	v := formatNumber(f.Value)
	if unit != "" {
		// The unit already names the dimension, so a generic key ("size") is
		// redundant: render "512MB", not "size 512MB".
		if redundantDims[strings.ToLower(humanizeKey(base))] {
			return v + unit
		}
		return humanizeKey(base) + " " + v + unit
	}
	if impliedPercent(f.Key, f.Value) {
		return humanizeKey(f.Key) + " " + v + "%"
	}
	return humanizeKey(f.Key) + " " + v
}

// redundantDims are generic dimension keys made redundant by a unit: a value
// keyed "size" with a unit needs no "size" prefix ("512MB", not "size 512MB").
var redundantDims = map[string]bool{
	"size": true, "length": true, "amount": true,
	"value": true, "duration": true, "total": true,
}

// unitSuffixes maps a key suffix to the unit that replaces it.
var unitSuffixes = []struct{ suf, unit string }{
	{"_ms", "ms"}, {"_msec", "ms"},
	{"_seconds", "s"}, {"_secs", "s"}, {"_sec", "s"},
	{"_kb", "KB"}, {"_mb", "MB"}, {"_gb", "GB"}, {"_tb", "TB"},
	{"_hz", "Hz"}, {"_khz", "kHz"}, {"_mhz", "MHz"},
	{"_bps", "bps"}, {"_percent", "%"}, {"_pct", "%"},
}

func unitFor(key string) (base, unit string) {
	lk := strings.ToLower(key)
	for _, u := range unitSuffixes {
		if strings.HasSuffix(lk, u.suf) && len(key) > len(u.suf) {
			return key[:len(key)-len(u.suf)], u.unit
		}
	}
	return key, ""
}

// impliedPercent reports keys that read as a percentage in the 0–100 range.
func impliedPercent(key string, v any) bool {
	lk := strings.ToLower(key)
	if !containsAny(lk, "cpu", "mem", "disk", "battery", "usage", "util") {
		return false
	}
	f, ok := toFloat(v)
	return ok && f >= 0 && f <= 100
}

// toInt reports an integer-valued number, rejecting fractions and non-numbers.
func toInt(v any) (int, bool) {
	f, ok := toFloat(v)
	if !ok || f != float64(int64(f)) {
		return 0, false
	}
	return int(f), true
}

func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// formatNumber renders a value, adding thousands separators to large integers.
func formatNumber(v any) string {
	s := valueString(v)
	f, ok := toFloat(v)
	if !ok || f != float64(int64(f)) || f < 10000 && f > -10000 {
		return s
	}
	neg := strings.HasPrefix(s, "-")
	digits := strings.TrimPrefix(s, "-")
	var out []byte
	for i, c := range []byte(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
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

// renderExample fills {placeholder} tokens in a template from the fields
// (matched by key or humanized key, case-insensitively). It reports whether
// every placeholder was filled; an unfilled template should be discarded.
func renderExample(tmpl string, fields []Field) (string, bool) {
	lookup := make(map[string]string, len(fields)*2)
	for _, f := range fields {
		v := valueString(f.Value)
		lookup[strings.ToLower(f.Key)] = v
		lookup[strings.ToLower(humanizeKey(f.Key))] = v
	}
	var b strings.Builder
	filled := true
	for i := 0; i < len(tmpl); {
		if tmpl[i] != '{' {
			b.WriteByte(tmpl[i])
			i++
			continue
		}
		j := strings.IndexByte(tmpl[i:], '}')
		if j < 0 {
			b.WriteByte(tmpl[i])
			i++
			continue
		}
		name := strings.ToLower(strings.TrimSpace(tmpl[i+1 : i+j]))
		if v, ok := lookup[name]; ok {
			b.WriteString(v)
		} else {
			filled = false
		}
		i += j + 1
	}
	return strings.TrimSpace(b.String()), filled
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
