package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dannyeldridge/differ/git"
	"github.com/dannyeldridge/differ/internal/gui"
	"github.com/dannyeldridge/differ/internal/tui"
)

func main() {
	tuiMode := flag.Bool("tui", false, "run the terminal UI instead of the GUI")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not get working directory")
		os.Exit(1)
	}

	if !git.IsGitRepo(cwd) {
		fmt.Fprintln(os.Stderr, "error: not inside a git repository")
		os.Exit(1)
	}

	root, err := git.RepoRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not find repo root: %v\n", err)
		os.Exit(1)
	}

	if *tuiMode {
		err = tui.Run(root)
	} else {
		err = gui.Run(root)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
