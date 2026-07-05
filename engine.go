package narrata

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/danielriddell21/narrata/backend/text"
	"github.com/danielriddell21/narrata/backend/tts"
	"github.com/danielriddell21/narrata/internal/policy"
	"github.com/danielriddell21/narrata/internal/prompt"
	"github.com/danielriddell21/narrata/internal/validate"
)

// Engine is the embedded narration runtime. It is safe for concurrent use.
type Engine struct {
	cfg      Config
	personas *registry
	text     text.Backend
	tts      tts.Backend
	ttsErr   error // why TTS is unavailable, if it failed to initialise
	sem      chan struct{}
	defVoice string
	sampleRt int
}

// New constructs an Engine from cfg. It always loads the bundled default
// personas first, then merges PersonasPath and PersonasDir if provided. With no
// text backend configured it uses the deterministic mock backend.
func New(cfg Config) (*Engine, error) {
	reg := newRegistry()
	if err := reg.loadDefaults(); err != nil {
		return nil, err
	}
	if cfg.PersonasPath != "" {
		if err := reg.loadFile(cfg.PersonasPath); err != nil {
			return nil, err
		}
	}
	if cfg.PersonasDir != "" {
		if err := reg.loadDir(cfg.PersonasDir); err != nil {
			return nil, err
		}
	}
	if cfg.DefaultPersona != "" {
		if _, ok := reg.Get(cfg.DefaultPersona); !ok {
			return nil, fmt.Errorf("%w: default persona %q", ErrPersonaNotFound, cfg.DefaultPersona)
		}
		reg.defaultPersona = cfg.DefaultPersona
	}

	tb, err := text.New(text.Options{
		Backend:       cfg.Text.Backend,
		ModelPath:     cfg.Text.ModelPath,
		ContextTokens: cfg.Text.ContextTokens,
		Temperature:   cfg.Text.Temperature,
		TopP:          cfg.Text.TopP,
		MaxTokens:     cfg.Text.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrModelNotLoaded, err)
	}

	e := &Engine{
		cfg:      cfg,
		personas: reg,
		text:     tb,
		defVoice: cfg.TTS.DefaultVoice,
		sampleRt: cfg.TTS.SampleRate,
	}
	if cfg.MaxConcurrent > 0 {
		e.sem = make(chan struct{}, cfg.MaxConcurrent)
	}

	// TTS is optional. If it fails to initialise, keep the engine usable for
	// text and surface ErrTTSUnavailable only when speech is requested.
	if cfg.TTS.Enabled {
		sb, err := tts.New(tts.Options{
			Backend:      cfg.TTS.Backend,
			ModelPath:    cfg.TTS.ModelPath,
			DefaultVoice: cfg.TTS.DefaultVoice,
			SampleRate:   cfg.TTS.SampleRate,
		})
		if err != nil {
			e.ttsErr = err
		} else {
			e.tts = sb
		}
	}

	return e, nil
}

// Personas exposes the persona store (read/save).
func (e *Engine) Personas() PersonaStore { return e.personas }

// Close releases backend resources.
func (e *Engine) Close() error {
	var errs []error
	if e.text != nil {
		if err := e.text.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if e.tts != nil {
		if err := e.tts.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Generate renders a single request into text and/or speech.
func (e *Engine) Generate(ctx context.Context, req Request) (Result, error) {
	if err := validate.Request(validate.RequestSpec{
		Event:          req.Event,
		HasData:        req.Data != nil,
		HasInstruction: req.Instruction != "",
	}); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	personaID := req.PersonaID
	if personaID == "" {
		personaID = e.personas.defaultID()
	}
	persona, ok := e.personas.Get(personaID)
	if !ok {
		return Result{}, fmt.Errorf("%w: %q", ErrPersonaNotFound, personaID)
	}

	wantSpeech := req.Output.wantsSpeech()
	wantText := req.Output.wantsText()
	if wantSpeech && e.tts == nil {
		if e.ttsErr != nil {
			return Result{}, fmt.Errorf("%w: %w", ErrTTSUnavailable, e.ttsErr)
		}
		return Result{}, ErrTTSUnavailable
	}

	// Apply the engine's default timeout if the caller set no earlier deadline.
	if e.cfg.Timeout > 0 {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, e.cfg.Timeout)
			defer cancel()
		}
	}

	if err := e.acquire(ctx); err != nil {
		return Result{}, err
	}
	defer e.release()

	start := time.Now()
	eff := effectiveConstraints(persona, req)

	opts := text.GenerateOptions{
		Temperature:   e.cfg.Text.Temperature,
		TopP:          e.cfg.Text.TopP,
		MaxTokens:     e.cfg.Text.MaxTokens,
		ContextTokens: e.cfg.Text.ContextTokens,
	}

	var gen text.Result
	var err error
	if sb, ok := e.text.(text.Structured); ok {
		// Typed path: hand the backend structured inputs (no prompt round-trip).
		gen, err = sb.GenerateStructured(ctx, text.StructuredInput{
			Event:       req.Event,
			Data:        req.Data,
			Instruction: req.Instruction,
			Style: text.Style{
				Tone:      persona.Style.Tone,
				Energy:    persona.Style.Energy,
				Humour:    persona.Style.Humour,
				Verbosity: persona.Style.Verbosity,
			},
			Rules:         persona.Rules,
			MaxWords:      eff.maxWords,
			MaxSentences:  eff.maxSentences,
			AllowMarkdown: eff.allowMarkdown,
		}, opts)
	} else {
		p := prompt.Build(prompt.Input{
			PersonaName:   persona.Name,
			Description:   persona.Description,
			Tone:          persona.Style.Tone,
			Energy:        persona.Style.Energy,
			Humour:        persona.Style.Humour,
			Verbosity:     persona.Style.Verbosity,
			Rules:         persona.Rules,
			MaxWords:      eff.maxWords,
			MaxSentences:  eff.maxSentences,
			AllowMarkdown: eff.allowMarkdown,
			Event:         req.Event,
			DataLines:     prompt.CompactData(req.Data),
			Instruction:   req.Instruction,
		})
		gen, err = e.text.Generate(ctx, p, opts)
	}
	if err != nil {
		return Result{}, mapGenErr(err)
	}

	processed := policy.Apply(gen.Text, policy.Options{
		MaxWords:       eff.maxWords,
		MaxSentences:   eff.maxSentences,
		AllowMarkdown:  eff.allowMarkdown,
		AllowProfanity: eff.allowProfanity,
		AllowSilence:   eff.allowSilence,
	})

	res := Result{
		PersonaID: persona.ID,
		Tokens:    gen.Tokens,
		Silent:    processed.Silent,
	}
	if wantText {
		res.Text = processed.Text
	}

	if wantSpeech && !processed.Silent {
		audio, err := e.tts.Speak(ctx, processed.Text, tts.SpeakOptions{
			VoiceID:    e.voiceFor(persona),
			Speed:      persona.Voice.Speed,
			Pitch:      persona.Voice.Pitch,
			SampleRate: e.sampleRt,
		})
		if err != nil {
			return Result{}, fmt.Errorf("%w: %w", ErrTTSUnavailable, err)
		}
		res.Audio = audio.Audio
		res.AudioFormat = audio.Format
		res.Spoken = len(audio.Audio) > 0
	}

	res.Duration = time.Since(start)
	return res, nil
}

func (e *Engine) voiceFor(p Persona) string {
	if p.Voice.ID != "" {
		return p.Voice.ID
	}
	return e.defVoice
}

func (e *Engine) acquire(ctx context.Context) error {
	if e.sem == nil {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("acquire: %w", err)
		}
		return nil
	}
	select {
	case e.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("acquire: %w", ctx.Err())
	}
}

func (e *Engine) release() {
	if e.sem != nil {
		<-e.sem
	}
}

// effective holds the merged request+persona constraints.
type effective struct {
	maxWords       int
	maxSentences   int
	allowMarkdown  bool
	allowProfanity bool
	allowSilence   bool
}

func effectiveConstraints(p Persona, req Request) effective {
	e := effective{
		maxWords:       p.Constraints.MaxWords,
		maxSentences:   p.Constraints.MaxSentences,
		allowMarkdown:  p.Constraints.AllowMarkdown,
		allowProfanity: p.Constraints.AllowProfanity,
		allowSilence:   p.EventPolicy.AllowSilence || req.EventPolicy.AllowSilence,
	}
	if req.Constraints.MaxWords > 0 {
		e.maxWords = req.Constraints.MaxWords
	}
	if req.Constraints.MaxSentences > 0 {
		e.maxSentences = req.Constraints.MaxSentences
	}
	if req.Constraints.Format == FormatMarkdown {
		e.allowMarkdown = true
	}
	if req.Constraints.AllowProfanity {
		e.allowProfanity = true
	}
	return e
}

func mapGenErr(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%w: %w", ErrGenerationTimeout, err)
	case errors.Is(err, context.Canceled):
		return err
	default:
		return fmt.Errorf("%w: %w", ErrGeneration, err)
	}
}
