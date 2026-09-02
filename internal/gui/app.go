package gui

import (
	"context"

	"github.com/dannyeldridge/differ/git"
	"github.com/dannyeldridge/differ/internal/diffparse"
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

func (a *App) LoadFiles(hash string) ([]git.FileChange, error) {
	return git.LoadFiles(a.repoPath, hash)
}

func (a *App) LoadDiffLines(hash, file string) ([]diffparse.Line, error) {
	raw, err := git.LoadDiff(a.repoPath, hash, file)
	if err != nil {
		return nil, err
	}
	return diffparse.Parse(raw), nil
}

func (a *App) LoadStagedFiles() ([]git.FileChange, error) {
	return git.LoadStagedFiles(a.repoPath)
}

func (a *App) LoadUnstagedFiles() ([]git.FileChange, error) {
	return git.LoadUnstagedFiles(a.repoPath)
}

func (a *App) LoadWorkingDiffLines(file string, staged bool) ([]diffparse.Line, error) {
	raw, err := git.LoadWorkingDiff(a.repoPath, file, staged)
	if err != nil {
		return nil, err
	}
	return diffparse.Parse(raw), nil
}
