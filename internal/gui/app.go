package gui

import (
	"context"

	"github.com/dannyeldridge/differ/git"
)

// App is bound to the frontend: every exported method becomes callable
// from JavaScript as window.go.gui.App.<MethodName>(...).
type App struct {
	repoPath string
	ctx      context.Context
}

func (a *App) CurrentBranch() (string, error) {
	return git.CurrentBranch(a.repoPath)
}

func (a *App) LoadCommits() ([]git.Commit, error) {
	return git.LoadCommits(a.repoPath)
}
