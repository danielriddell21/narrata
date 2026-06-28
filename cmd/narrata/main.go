// Command narrata is a development CLI for the Narrata narration runtime. It
// validates persona files, scaffolds personas from a description, renders a
// single event to text, and lists the bundled default personas.
//
// It is a developer aid, not a runtime service: Narrata is embedded as a
// library in host applications.
//
// Usage:
//
//	narrata <command> [flags]
//
//	validate <path>     Validate a personas.json file or a personas/ directory.
//	gen <description>   Generate a persona from a short description.
//	generate            Render a single event to narration text.
//	personas            List the bundled default personas.
//	version             Print the CLI version.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/danielriddell21/narrata/pkg/narrata"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "0.1.0-dev"

const usage = `narrata is a development CLI for the Narrata narration runtime.

Usage:
  narrata <command> [flags]

Commands:
  validate <path>     Validate a personas.json file or a personas/ directory.
  gen <description>   Generate a persona from a short description.
  generate            Render a single event to narration text.
  personas            List the bundled default personas.
  version             Print the CLI version.

Run "narrata <command> -h" for command flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "validate":
		err = runValidate(args)
	case "gen":
		err = runGen(args)
	case "generate":
		err = runGenerate(args)
	case "personas":
		err = runPersonas(args)
	case "version", "-v", "--version":
		fmt.Println(version)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "narrata: unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
	if err != nil {
		// Typed package errors and subcommand errors are already prefixed.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runValidate loads a personas file or directory through the engine, which
// validates every persona, and reports the result.
func runValidate(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("validate: expected exactly one path argument")
	}
	path := args[0]
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	var cfg narrata.Config
	if info.IsDir() {
		cfg.PersonasDir = path
	} else {
		cfg.PersonasPath = path
	}
	engine, err := narrata.New(cfg)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	defer func() { _ = engine.Close() }()

	// Count personas beyond the bundled defaults to report what the file added.
	total := len(engine.Personas().List())
	fmt.Printf("OK: %d personas available (including bundled defaults)\n", total)
	return nil
}

// runGen scaffolds a persona from a free-text description.
func runGen(args []string) error {
	fs := flag.NewFlagSet("gen", flag.ContinueOnError)
	out := fs.String("o", "", "write the persona JSON to this file instead of stdout")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	desc := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if desc == "" {
		return fmt.Errorf("gen: a description is required, e.g. narrata gen \"pirate ship captain for server alerts\"")
	}

	persona, err := narrata.GeneratePersona(desc)
	if err != nil {
		return fmt.Errorf("gen: %w", err)
	}
	data, err := json.MarshalIndent(persona, "", "  ")
	if err != nil {
		return fmt.Errorf("gen: marshal persona: %w", err)
	}
	data = append(data, '\n')

	if *out == "" {
		if _, err := os.Stdout.Write(data); err != nil {
			return fmt.Errorf("gen: write: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(*out, data, 0o600); err != nil {
		return fmt.Errorf("gen: write %s: %w", *out, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (persona id %q)\n", *out, persona.ID)
	return nil
}

// runGenerate renders a single event using the deterministic backends.
func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	persona := fs.String("persona", "", "persona id (default: the configured default)")
	event := fs.String("event", "", "event name, e.g. door_opened")
	dataJSON := fs.String("data", "", "event data as a JSON object")
	personasPath := fs.String("personas", "", "path to a personas.json file")
	personasDir := fs.String("dir", "", "path to a personas/ directory")
	backend := fs.String("backend", "template", "text backend: mock or template")
	maxWords := fs.Int("max-words", 0, "override the persona word limit")
	instruction := fs.String("instruction", "", "optional one-off steering note")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if strings.TrimSpace(*event) == "" && strings.TrimSpace(*dataJSON) == "" {
		return fmt.Errorf("generate: -event or -data is required")
	}

	var data any
	if strings.TrimSpace(*dataJSON) != "" {
		if err := json.Unmarshal([]byte(*dataJSON), &data); err != nil {
			return fmt.Errorf("generate: parsing -data: %w", err)
		}
	}

	engine, err := narrata.New(narrata.Config{
		PersonasPath: *personasPath,
		PersonasDir:  *personasDir,
		Text:         narrata.TextConfig{Backend: *backend},
	})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	defer func() { _ = engine.Close() }()

	res, err := engine.Generate(context.Background(), narrata.Request{
		PersonaID:   *persona,
		Event:       *event,
		Data:        data,
		Instruction: *instruction,
		Output:      narrata.OutputText,
		Constraints: narrata.Constraints{MaxWords: *maxWords},
	})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	fmt.Printf("[%s] %s\n", res.PersonaID, res.Text)
	return nil
}

// runPersonas lists the bundled default personas.
func runPersonas(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("personas: takes no arguments")
	}
	engine, err := narrata.New(narrata.Config{})
	if err != nil {
		return fmt.Errorf("personas: %w", err)
	}
	defer func() { _ = engine.Close() }()

	for _, p := range engine.Personas().List() {
		fmt.Printf("%-20s %s\n", p.ID, p.Description)
	}
	return nil
}
