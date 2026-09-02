package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the terminal UI rooted at repoPath. It blocks until the user quits.
func Run(repoPath string) error {
	m := newModel(repoPath)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
