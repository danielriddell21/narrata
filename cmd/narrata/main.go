package main

import (
	"fmt"
	"os"

	"github.com/danielriddell21/narrata/internal/cli"
)

var version = "0.1.0-dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
