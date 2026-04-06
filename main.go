package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not get working directory")
		os.Exit(1)
	}

	if !gitIsRepo(cwd) {
		fmt.Fprintln(os.Stderr, "error: not inside a git repository")
		os.Exit(1)
	}

	root, err := gitRepoRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not find repo root: %v\n", err)
		os.Exit(1)
	}

	m := newModel(root)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
