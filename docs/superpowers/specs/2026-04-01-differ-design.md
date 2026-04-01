# differ — Terminal Git Diff Viewer

**Date:** 2026-04-01
**Status:** Approved

## Overview

`differ` is a read-only terminal UI for browsing git commit history, inspecting changed files, and viewing diffs — keyboard-driven, with no mouse required. It is invoked from within a git repository and operates on the current branch.

## Tech Stack

- **Language:** Go
- **TUI framework:** Bubbletea (Charmbracelet)
- **Styling:** Lipgloss
- **List/viewport components:** Bubbles (`list`, `viewport`)
- **Git interaction:** Shell out to `git` CLI — no libgit2 or Go git bindings

## Layout

Three-column layout. All three panes are always visible.

```
┌─────────────────┬─────────────────┬──────────────────────────────┐
│   Commits       │   Files         │   Diff                       │
│   (~25%)        │   (~25%)        │   (~50%)                     │
│                 │                 │                              │
│ ● abc1234 ...   │ M src/auth.go   │ @@ -12,6 +12,8 @@           │
│   def5678 ...   │ A src/token.go  │ + func validateToken(…       │
│   ghi9012 ...   │ D src/old.go    │ - func checkAuth(…           │
│                 │                 │   if t == "" { return }      │
└─────────────────┴─────────────────┴──────────────────────────────┘
│ branch: main  commit: abc1234  author: Jane  date: 2024-01-15  [1/3 files] │
```

The focused pane has a colored border. Unfocused panes are dimmed.

## Panes

### Commits Pane
- Populated at startup via `git log --oneline -100`
- Displays short hash + subject line per entry
- Selecting a commit loads its changed files into the Files pane

### Files Pane
- Populated via `git diff <hash>^..<hash> --name-only --diff-filter=ACDMR`
- Each entry is prefixed with its status: `M` (modified), `A` (added), `D` (deleted)
- Initial commit (no parent) handled via `git show <hash> --name-only`
- Selecting a file loads its diff into the Diff pane
- Auto-selects first file when focus enters this pane

### Diff Pane
- Populated via `git diff <hash>^..<hash> -- <file>`
- Initial commit handled via `git show <hash> -- <file>`
- Syntax: green for additions (`+`), red for deletions (`-`), dim for context
- Scrollable via Bubbletea `viewport`

## Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down in focused pane |
| `k` / `↑` | Move up in focused pane |
| `l` / `→` / `Tab` | Focus next pane (commits → files → diff) |
| `h` / `←` / `Shift+Tab` | Focus previous pane |
| `g` | Jump to top of focused pane |
| `G` | Jump to bottom of focused pane |
| `q` / `Ctrl+C` | Quit |

Focus starts on the Commits pane on launch.

## Status Bar

Single line at the bottom of the terminal:

```
branch: main  commit: abc1234  author: Jane Doe  date: 2024-01-15  [1/3 files]
```

- **branch**: current git branch (read once at startup)
- **commit**: short hash of selected commit
- **author**: author of selected commit
- **date**: author date of selected commit (YYYY-MM-DD)
- **[n/total files]**: position in files pane (only shown when files pane or diff pane is focused)

## App Model Structure

```
Model
├── commitList  bubbles/list.Model   — commits pane
├── fileList    bubbles/list.Model   — files pane
├── diffView    bubbles/viewport.Model — diff pane
├── focused     int (0=commits, 1=files, 2=diff)
├── branch      string
└── status      StatusBar
```

Bubbletea `Update` dispatches key messages to the focused sub-model. Selection changes in commits/files trigger `tea.Cmd` functions that run `git` and return messages to update downstream panes.

## Git Commands

| Purpose | Command |
|---------|---------|
| Load commits | `git log --oneline -100` |
| Files in commit | `git diff <hash>^..<hash> --name-only --diff-filter=ACDMR` |
| Files in initial commit | `git show <hash> --name-only --format=` |
| Diff for file | `git diff <hash>^..<hash> -- <file>` |
| Diff for initial commit | `git show <hash> -- <file>` |
| Current branch | `git rev-parse --abbrev-ref HEAD` |
| Commit metadata | `git log -1 --format=%H%n%an%n%ad --date=short <hash>` |

## Error Handling

- If CWD is not inside a git repo: print error message and exit with code 1
- If a `git` command fails: show error inline in the relevant pane (e.g., "could not load diff")
- Empty repo (no commits): show "No commits on this branch" in commits pane

## Out of Scope

- Staging, committing, or any write operations
- Branch switching
- Search/filter within commit list
- Side-by-side diff view
- Support for merge commits (treated as single-parent for now)
