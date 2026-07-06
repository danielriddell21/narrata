package narrata

import "errors"

// Typed errors returned by the [Engine]. Callers may compare with [errors.Is].
var (
	// ErrPersonaNotFound means the requested persona ID is not registered.
	ErrPersonaNotFound = errors.New("narrata: persona not found")
	// ErrInvalidRequest means the request failed validation.
	ErrInvalidRequest = errors.New("narrata: invalid request")
	// ErrBackendUnavailable means the configured backend could not be
	// initialised, such as an unknown backend name.
	ErrBackendUnavailable = errors.New("narrata: backend unavailable")
	// ErrGenerationTimeout means generation exceeded the deadline.
	ErrGenerationTimeout = errors.New("narrata: generation timed out")
	// ErrGeneration means generation failed for a reason other than a timeout.
	ErrGeneration = errors.New("narrata: generation failed")
	// ErrTTSUnavailable means speech was requested but TTS is disabled or
	// failed to initialise.
	ErrTTSUnavailable = errors.New("narrata: text-to-speech unavailable")
	// ErrScopeViolation means an operation outside Narrata's narration scope
	// was attempted (assistant/agent/tool/memory behaviour).
	ErrScopeViolation = errors.New("narrata: scope violation")
)
