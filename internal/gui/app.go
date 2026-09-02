package gui

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/dannyeldridge/differ/git"
	"github.com/dannyeldridge/differ/internal/diffparse"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

// startWatcher polls for HEAD/reflog/index changes every second and emits
// "repo-changed" to the frontend, mirroring the TUI's watchRepoCmd.
func (a *App) startWatcher() {
	go func() {
		var headHash string
		var reflogMtime, indexMtime int64
		first := true
		for {
			time.Sleep(time.Second)

			hash, _ := git.HeadHash(a.repoPath)
			var newReflogMtime int64
			if info, err := os.Stat(filepath.Join(a.repoPath, ".git", "logs", "HEAD")); err == nil {
				newReflogMtime = info.ModTime().UnixNano()
			}
			var newIndexMtime int64
			if info, err := os.Stat(filepath.Join(a.repoPath, ".git", "index")); err == nil {
				newIndexMtime = info.ModTime().UnixNano()
			}

			changed := hash != headHash || newReflogMtime != reflogMtime || newIndexMtime != indexMtime
			headHash, reflogMtime, indexMtime = hash, newReflogMtime, newIndexMtime

			if changed && !first {
				wailsruntime.EventsEmit(a.ctx, "repo-changed")
			}
			first = false
		}
	}()
}
