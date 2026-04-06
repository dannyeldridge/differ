package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Commit represents a single git commit.
type Commit struct {
	Hash      string
	ShortHash string
	Subject   string
	Author    string
	Date      string
}

// FileChange represents a file affected by a commit.
type FileChange struct {
	Status string // M, A, D, R, C
	Path   string
}

// run executes a git command in repoPath and returns trimmed stdout.
func run(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// IsGitRepo returns true if path is inside a git repository.
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

// RepoRoot returns the absolute path to the repository root.
func RepoRoot(path string) (string, error) {
	return run(path, "rev-parse", "--show-toplevel")
}

// CurrentBranch returns the name of the current branch (e.g. "main").
func CurrentBranch(repoPath string) (string, error) {
	return run(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}

// HeadHash returns the full hash of the current HEAD commit.
func HeadHash(repoPath string) (string, error) {
	return run(repoPath, "rev-parse", "HEAD")
}

// LoadCommits returns the last 100 commits on the current branch.
func LoadCommits(repoPath string) ([]Commit, error) {
	out, err := run(repoPath, "log",
		"--format=%H\x1f%h\x1f%s\x1f%an\x1f%ad",
		"--date=short",
		"-100",
	)
	if err != nil {
		return nil, err
	}
	return parseCommits(out), nil
}

// parseCommits parses the output of `git log --format=...` into Commit slices.
// Unexported; accessible to tests in this package. Callers should use LoadCommits.
func parseCommits(output string) []Commit {
	if output == "" {
		return nil
	}
	lines := strings.Split(output, "\n")
	commits := make([]Commit, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\x1f", 5)
		if len(parts) != 5 {
			continue
		}
		commits = append(commits, Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Subject:   parts[2],
			Author:    parts[3],
			Date:      parts[4],
		})
	}
	return commits
}

// isInitialCommit returns true if hash has no parent commit.
func isInitialCommit(repoPath, hash string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", hash+"^")
	cmd.Dir = repoPath
	err := cmd.Run()
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode() == 128
	}
	return false
}

// LoadFiles returns the list of files changed in a given commit.
func LoadFiles(repoPath, hash string) ([]FileChange, error) {
	var out string
	var err error
	if isInitialCommit(repoPath, hash) {
		out, err = run(repoPath, "show", "--name-status", "--format=", hash)
	} else {
		out, err = run(repoPath, "diff", "--name-status", hash+"^", hash)
	}
	if err != nil {
		return nil, err
	}
	return parseFiles(out), nil
}

// parseFiles parses the output of `git diff --name-status` or `git show --name-status`.
func parseFiles(output string) []FileChange {
	if output == "" {
		return nil
	}
	lines := strings.Split(output, "\n")
	files := make([]FileChange, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		status := string(parts[0][0]) // take first char: M, A, D, R, C
		files = append(files, FileChange{Status: status, Path: parts[1]})
	}
	return files
}

// LoadDiff returns the raw unified diff for a single file in a commit.
func LoadDiff(repoPath, hash, file string) (string, error) {
	if isInitialCommit(repoPath, hash) {
		return run(repoPath, "show", hash, "--", file)
	}
	return run(repoPath, "diff", hash+"^", hash, "--", file)
}
