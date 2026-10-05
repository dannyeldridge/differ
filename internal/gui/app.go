package gui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dannyeldridge/differ/git"
	"github.com/dannyeldridge/differ/internal/diffparse"
	"github.com/wailsapp/wails/v2/pkg/options"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is bound to the frontend: every exported method becomes callable
// from JavaScript as window.go.gui.App.<MethodName>(...).
type App struct {
	ctx context.Context
	reg *registry

	mu     sync.Mutex
	active string // repo the most recent launch asked to show
}

func newApp() *App {
	return &App{reg: newRegistry()}
}

// RepoList is the set of open repos in tab order, plus the one to show first.
type RepoList struct {
	Repos  []string
	Active string
}

func (a *App) ListRepos() RepoList {
	a.mu.Lock()
	defer a.mu.Unlock()
	return RepoList{Repos: a.reg.list(), Active: a.active}
}

// CloseRepo stops watching repo and reports how many repos remain open.
func (a *App) CloseRepo(repo string) int {
	a.reg.remove(repo)
	return len(a.reg.list())
}

// openRepo registers repo (starting its watcher if new), records it as the
// active tab, and tells the frontend. Safe to call before the frontend loads:
// the frontend also pulls ListRepos on startup.
func (a *App) openRepo(repo string) {
	if stop, added := a.reg.add(repo); added {
		a.startWatcher(repo, stop)
	}
	a.mu.Lock()
	a.active = repo
	a.mu.Unlock()
	wailsruntime.EventsEmit(a.ctx, "repo-opened", repo)
}

// repoFlag is how a launching process tells the running instance which repo it
// was started in. Wails' own SecondInstanceData.WorkingDirectory is useless for
// this: on macOS it is the executable's directory, not the caller's cwd.
const repoFlag = "--repo"

// repoFromArgs extracts the --repo value from a launch's args, or "" if absent.
func repoFromArgs(args []string) string {
	for i, arg := range args {
		arg = strings.TrimPrefix(arg, "-")
		arg = strings.TrimPrefix(arg, "-")
		if arg == "repo" {
			if i+1 < len(args) {
				return args[i+1]
			}
			return ""
		}
		if v, ok := strings.CutPrefix(arg, "repo="); ok {
			return v
		}
	}
	return ""
}

// handleSecondInstance runs in the already-running app when another `differ`
// is launched. data.Args carries that launch's --repo (see Run).
func (a *App) handleSecondInstance(data options.SecondInstanceData) {
	if dir := repoFromArgs(data.Args); dir != "" && git.IsGitRepo(dir) {
		if root, err := git.RepoRoot(dir); err == nil {
			a.openRepo(root)
		}
	}
	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.WindowShow(a.ctx)
}

func (a *App) CurrentBranch(repo string) (string, error) {
	return git.CurrentBranch(repo)
}

func (a *App) LoadCommits(repo string) ([]git.Commit, error) {
	return git.LoadCommits(repo)
}

func (a *App) LoadFiles(repo, hash string) ([]git.FileChange, error) {
	files, err := git.LoadFiles(repo, hash)
	if err != nil {
		return nil, err
	}
	if files == nil {
		files = []git.FileChange{}
	}
	return files, nil
}

func (a *App) LoadDiffLines(repo, hash, file string) ([]diffparse.Line, error) {
	raw, err := git.LoadDiff(repo, hash, file)
	if err != nil {
		return nil, err
	}
	lines := diffparse.Parse(raw)
	if lines == nil {
		lines = []diffparse.Line{}
	}
	return lines, nil
}

func (a *App) LoadStagedFiles(repo string) ([]git.FileChange, error) {
	files, err := git.LoadStagedFiles(repo)
	if err != nil {
		return nil, err
	}
	if files == nil {
		files = []git.FileChange{}
	}
	return files, nil
}

func (a *App) LoadUnstagedFiles(repo string) ([]git.FileChange, error) {
	files, err := git.LoadUnstagedFiles(repo)
	if err != nil {
		return nil, err
	}
	if files == nil {
		files = []git.FileChange{}
	}
	return files, nil
}

func (a *App) LoadWorkingDiffLines(repo, file string, staged bool) ([]diffparse.Line, error) {
	raw, err := git.LoadWorkingDiff(repo, file, staged)
	if err != nil {
		return nil, err
	}
	lines := diffparse.Parse(raw)
	if lines == nil {
		lines = []diffparse.Line{}
	}
	return lines, nil
}

// startWatcher polls repo for HEAD/reflog/index changes every second and emits
// "repo-changed" (with the repo path) to the frontend, mirroring the TUI's
// watchRepoCmd. It exits when stop is closed.
func (a *App) startWatcher(repo string, stop <-chan struct{}) {
	go func() {
		var headHash string
		var reflogMtime, indexMtime int64
		first := true
		for {
			select {
			case <-stop:
				return
			case <-time.After(time.Second):
			}

			hash, _ := git.HeadHash(repo)
			var newReflogMtime int64
			if info, err := os.Stat(filepath.Join(repo, ".git", "logs", "HEAD")); err == nil {
				newReflogMtime = info.ModTime().UnixNano()
			}
			var newIndexMtime int64
			if info, err := os.Stat(filepath.Join(repo, ".git", "index")); err == nil {
				newIndexMtime = info.ModTime().UnixNano()
			}

			changed := hash != headHash || newReflogMtime != reflogMtime || newIndexMtime != indexMtime
			headHash, reflogMtime, indexMtime = hash, newReflogMtime, newIndexMtime

			if changed && !first {
				wailsruntime.EventsEmit(a.ctx, "repo-changed", repo)
			}
			first = false
		}
	}()
}
