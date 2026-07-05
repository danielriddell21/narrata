package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/narrata"
)

func Execute(version string) error {
	root := &cobra.Command{
		Use:   "narrata",
		Short: "Development CLI for the Narrata narration runtime",
		Long: `narrata is a development CLI for the Narrata narration runtime.

It validates persona files, scaffolds personas from a description, renders a
single event to narration text, and lists the bundled default personas. It is a
developer aid, not a runtime service: Narrata is embedded as a library in host
applications.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(validateCmd(), genCmd(), generateCmd(), personasCmd(), completionCmd())
	if err := root.Execute(); err != nil {
		return fmt.Errorf("narrata: %w", err)
	}
	return nil
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "validate <path>",
		Short:        "Validate a personas.json file or a personas/ directory",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
}

func genCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:          "gen <description>",
		Short:        "Generate a persona from a short description",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			desc := strings.TrimSpace(strings.Join(args, " "))
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

			if out == "" {
				if _, err := os.Stdout.Write(data); err != nil {
					return fmt.Errorf("gen: write: %w", err)
				}
				return nil
			}
			if err := os.WriteFile(out, data, 0o600); err != nil {
				return fmt.Errorf("gen: write %s: %w", out, err)
			}
			fmt.Fprintf(os.Stderr, "wrote %s (persona id %q)\n", out, persona.ID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&out, "out", "o", "", "write the persona JSON to this file instead of stdout")
	return cmd
}

func generateCmd() *cobra.Command {
	var (
		persona      string
		event        string
		dataJSON     string
		personasPath string
		personasDir  string
		backend      string
		maxWords     int
		instruction  string
	)
	cmd := &cobra.Command{
		Use:          "generate",
		Short:        "Render a single event to narration text",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(event) == "" && strings.TrimSpace(dataJSON) == "" {
				return fmt.Errorf("generate: --event or --data is required")
			}

			var data any
			if strings.TrimSpace(dataJSON) != "" {
				if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
					return fmt.Errorf("generate: parsing --data: %w", err)
				}
			}

			engine, err := narrata.New(narrata.Config{
				PersonasPath: personasPath,
				PersonasDir:  personasDir,
				Text:         narrata.TextConfig{Backend: backend},
			})
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}
			defer func() { _ = engine.Close() }()

			res, err := engine.Generate(context.Background(), narrata.Request{
				PersonaID:   persona,
				Event:       event,
				Data:        data,
				Instruction: instruction,
				Output:      narrata.OutputText,
				Constraints: narrata.Constraints{MaxWords: maxWords},
			})
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}
			fmt.Printf("[%s] %s\n", res.PersonaID, res.Text)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&persona, "persona", "", "persona id (default: the configured default)")
	f.StringVar(&event, "event", "", "event name, e.g. door_opened")
	f.StringVar(&dataJSON, "data", "", "event data as a JSON object")
	f.StringVar(&personasPath, "personas", "", "path to a personas.json file")
	f.StringVar(&personasDir, "dir", "", "path to a personas/ directory")
	f.StringVar(&backend, "backend", "native", "text backend: mock, template, or native")
	f.IntVar(&maxWords, "max-words", 0, "override the persona word limit")
	f.StringVar(&instruction, "instruction", "", "optional one-off steering note")
	return cmd
}

func personasCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "personas",
		Short:        "List the bundled default personas",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := narrata.New(narrata.Config{})
			if err != nil {
				return fmt.Errorf("personas: %w", err)
			}
			defer func() { _ = engine.Close() }()

			for _, p := range engine.Personas().List() {
				fmt.Printf("%-20s %s\n", p.ID, p.Description)
			}
			return nil
		},
	}
}
