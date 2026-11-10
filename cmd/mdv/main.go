package main

import (
	"flag"
	"fmt"
	"os"

	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"

	"mdt/internal/viewer"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <markdown-file>\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Render a Markdown file in the terminal.")
		fmt.Fprintln(flag.CommandLine.Output(), "The path may be absolute or relative to the current directory.")
		fmt.Fprintln(flag.CommandLine.Output(), "Set GLAMOUR_STYLE to a Glamour style name or JSON style file.")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	path := flag.Arg(0)
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdv: read %s: %v\n", path, err)
		os.Exit(1)
	}

	width := 80
	isTerminal := term.IsTerminal(os.Stdout.Fd())
	if isTerminal {
		if terminalWidth, _, err := term.GetSize(os.Stdout.Fd()); err == nil && terminalWidth > 0 {
			width = terminalWidth
		}
	}

	style := os.Getenv("GLAMOUR_STYLE")
	if style == "" {
		style = styles.NoTTYStyle
		if isTerminal {
			style = styles.LightStyle
			if lipgloss.HasDarkBackground(os.Stdin, os.Stdout) {
				style = styles.DarkStyle
			}
		}
	}

	output, err := viewer.Render(source, width, style)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdv: %v\n", err)
		os.Exit(1)
	}
	if _, err := lipgloss.Fprint(os.Stdout, output); err != nil {
		fmt.Fprintf(os.Stderr, "mdv: write output: %v\n", err)
		os.Exit(1)
	}
}
