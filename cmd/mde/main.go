package main

import (
	"flag"
	"fmt"
	"os"

	"mdt/internal/editor"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <markdown-file>\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Edit a Markdown file in a Nano-style terminal interface.")
		fmt.Fprintln(flag.CommandLine.Output(), "The path may be absolute or relative to the current directory.")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if err := editor.Run(flag.Arg(0)); err != nil {
		fmt.Fprintf(os.Stderr, "mde: %v\n", err)
		os.Exit(1)
	}
}
