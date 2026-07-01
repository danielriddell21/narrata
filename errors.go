package narrata

import "errors"

// Typed errors returned by the [Engine]. Callers may compare with [errors.Is].
var (
	// ErrPersonaNotFound means the requested persona ID is not registered.
	ErrPersonaNotFound = errors.New("narrata: persona not found")
	// ErrInvalidRequest means the request failed validation.
	ErrInvalidRequest = errors.New("narrata: invalid request")
	// ErrModelNotLoaded means a backend failed to initialise or is missing.
	ErrModelNotLoaded = errors.New("narrata: model not loaded")
	// ErrGenerationTimeout means generation exceeded the deadline.
	ErrGenerationTimeout = errors.New("narrata: generation timed out")
	// ErrGeneration means text generation failed for a reason other than a
	// timeout or a missing model, such as a mid-stream decode failure.
	ErrGeneration = errors.New("narrata: generation failed")
	// ErrTTSUnavailable means speech was requested but TTS is disabled or
	// failed to initialise.
	ErrTTSUnavailable = errors.New("narrata: text-to-speech unavailable")
	// ErrScopeViolation means an operation outside Narrata's narration scope
	// was attempted (assistant/agent/tool/memory behaviour).
	ErrScopeViolation = errors.New("narrata: scope violation")
)
