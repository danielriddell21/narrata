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
package main

import (
	"fmt"
	"os"

	"github.com/danielriddell21/narrata/internal/cli"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "0.1.0-dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
