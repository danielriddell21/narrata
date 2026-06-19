package tts

import (
	"context"
	"encoding/binary"
)

// defaultSampleRate is used when no sample rate is configured.
const defaultSampleRate = 24000

// Mock is a deterministic TTS backend that emits a valid mono 16-bit PCM WAV
// whose contents are derived from the input text. It lets tests and examples
// exercise the speech path without a real model or audio runtime.
type Mock struct {
	sampleRate int
}

// NewMock returns a Mock backend. A zero sampleRate uses the default.
func NewMock(sampleRate int) *Mock {
	if sampleRate <= 0 {
		sampleRate = defaultSampleRate
	}
	return &Mock{sampleRate: sampleRate}
}

// Speak returns a deterministic WAV buffer sized from the text length.
func (m *Mock) Speak(ctx context.Context, text string, opts SpeakOptions) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	rate := m.sampleRate
	if opts.SampleRate > 0 {
		rate = opts.SampleRate
	}

	// ~30ms of audio per character, bounded, so output is non-trivial but small.
	samples := len(text) * rate / 33
	if samples < rate/10 {
		samples = rate / 10 // minimum 100ms
	}
	if samples > rate*5 {
		samples = rate * 5 // cap at 5s
	}

	pcm := make([]int16, samples)
	// Deterministic low-amplitude waveform seeded by the text bytes.
	var seed uint32 = 2166136261
	for _, b := range []byte(text) {
		seed = (seed ^ uint32(b)) * 16777619
	}
	for i := range pcm {
		seed = seed*1664525 + 1013904223
		pcm[i] = int16(seed >> 18) // small amplitude
	}

	return Result{
		Audio:      encodeWAV(pcm, rate),
		Format:     "wav",
		SampleRate: rate,
	}, nil
}

// Close is a no-op.
func (m *Mock) Close() error { return nil }

// encodeWAV writes a canonical 16-bit mono PCM WAV file.
func encodeWAV(pcm []int16, sampleRate int) []byte {
	const (
		numChannels   = 1
		bitsPerSample = 16
	)
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	dataSize := len(pcm) * 2

	buf := make([]byte, 0, 44+dataSize)
	buf = append(buf, "RIFF"...)
	buf = le32(buf, uint32(36+dataSize))
	buf = append(buf, "WAVE"...)
	buf = append(buf, "fmt "...)
	buf = le32(buf, 16) // PCM fmt chunk size
	buf = le16(buf, 1)  // audio format: PCM
	buf = le16(buf, numChannels)
	buf = le32(buf, uint32(sampleRate))
	buf = le32(buf, uint32(byteRate))
	buf = le16(buf, uint16(blockAlign))
	buf = le16(buf, bitsPerSample)
	buf = append(buf, "data"...)
	buf = le32(buf, uint32(dataSize))
	for _, s := range pcm {
		buf = le16(buf, uint16(s))
	}
	return buf
}

func le16(b []byte, v uint16) []byte {
	var tmp [2]byte
	binary.LittleEndian.PutUint16(tmp[:], v)
	return append(b, tmp[:]...)
}

func le32(b []byte, v uint32) []byte {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], v)
	return append(b, tmp[:]...)
}
