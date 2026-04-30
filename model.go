package main

import (
	"github.com/dannyeldridge/differ/git"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Pane constants ───────────────────────────────────────────────────────────

const (
	paneCommits = 0
	paneFiles   = 1
	paneDiff    = 2
)

// ── Mode constants ───────────────────────────────────────────────────────────

const (
	modeHistory = 0
	modeChanges = 1
)

// ── Styles ───────────────────────────────────────────────────────────────────

var (
	focusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39"))

	unfocusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("237"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	hintRendered = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true).Render("c")

	filePathBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)
)

// ── List item types ──────────────────────────────────────────────────────────

type commitItem struct{ commit git.Commit }

func (i commitItem) FilterValue() string { return i.commit.Subject }
func (i commitItem) Title() string       { return i.commit.ShortHash + " " + i.commit.Subject }
func (i commitItem) Description() string { return i.commit.Author + " · " + i.commit.Date }

type fileItem struct{ file git.FileChange }

func (i fileItem) FilterValue() string { return i.file.Path }
func (i fileItem) Title() string       { return i.file.Status + " " + i.file.Path }
func (i fileItem) Description() string { return "" }

// fileItemDelegate renders file paths with middle truncation so the filename is always visible.
type fileItemDelegate struct {
	width  int
	styles list.DefaultItemStyles
}

func newFileItemDelegate() fileItemDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		BorderForeground(lipgloss.Color("39")).
		Foreground(lipgloss.Color("252"))
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(lipgloss.Color("245"))
	return fileItemDelegate{styles: d.Styles}
}

func (d fileItemDelegate) Height() int                             { return 1 }
func (d fileItemDelegate) Spacing() int                            { return 0 }
func (d fileItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d fileItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var fc git.FileChange
	switch v := item.(type) {
	case fileItem:
		fc = v.file
	case changesFileItem:
		fc = v.file
	case changesHeaderItem:
		fmt.Fprint(w, d.styles.NormalTitle.Render(v.label))
		return
	default:
		return
	}
	// 4 = left padding (2) + status char (1) + space (1)
	maxPath := d.width - 4
	if maxPath < 3 {
		maxPath = 3
	}
	label := fc.Status + " " + truncateMiddle(fc.Path, maxPath)
	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render(label))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render(label))
	}
}

// truncateMiddle shortens s to max runes by replacing the middle with "…",
// keeping the start and the filename suffix visible.
func truncateMiddle(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 3 {
		return string(runes[:max])
	}
	left := (max - 1) / 2
	right := max - left - 1
	return string(runes[:left]) + "…" + string(runes[len(runes)-right:])
}

type changesFileItem struct {
	file   git.FileChange
	staged bool
}

func (i changesFileItem) FilterValue() string { return i.file.Path }
func (i changesFileItem) Title() string       { return i.file.Status + " " + i.file.Path }
func (i changesFileItem) Description() string { return "" }

type changesHeaderItem struct{ label string }

func (i changesHeaderItem) FilterValue() string { return "" }
func (i changesHeaderItem) Title() string       { return i.label }
func (i changesHeaderItem) Description() string { return "" }

// ── Message types ────────────────────────────────────────────────────────────

type commitsLoadedMsg struct {
	commits     []git.Commit
	preserveIdx int
}
type repoWatchMsg struct {
	hash        string
	reflogMtime int64
	indexMtime  int64
}
type filesLoadedMsg struct {
	hash  string
	files []git.FileChange
}
type diffLoadedMsg struct {
	content string
	mode    int
}
type branchLoadedMsg struct{ branch string }
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type changesLoadedMsg struct {
	staged   []git.FileChange
	unstaged []git.FileChange
}

// ── Model ────────────────────────────────────────────────────────────────────

type Model struct {
	repoPath string
	branch   string
	focused  int

	commitList list.Model
	fileList   list.Model
	diffView   viewport.Model

	commits []git.Commit
	files   []git.FileChange

	fileErr  string
	diffErr  string
	headHash string

	width  int
	height int
	ready  bool

	watchInitialized bool
	reflogMtime      int64

	mode          int
	changesList   list.Model
	stagedFiles   []git.FileChange
	unstagedFiles []git.FileChange
	indexMtime    int64

	xOffset   int
	statusMsg string
}

func newModel(repoPath string) Model {
	commitDelegate := list.NewDefaultDelegate()
	commitDelegate.Styles.SelectedTitle = commitDelegate.Styles.SelectedTitle.
		BorderForeground(lipgloss.Color("39")).
		Foreground(lipgloss.Color("252"))
	commitDelegate.Styles.SelectedDesc = commitDelegate.Styles.SelectedDesc.
		BorderForeground(lipgloss.Color("39")).
		Foreground(lipgloss.Color("246"))
	commitDelegate.Styles.NormalTitle = commitDelegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("245"))
	commitDelegate.Styles.NormalDesc = commitDelegate.Styles.NormalDesc.
		Foreground(lipgloss.Color("240"))

	listTitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true).
		Padding(0, 1)

	commitList := list.New(nil, commitDelegate, 0, 0)
	commitList.Title = "Commits"
	commitList.Styles.Title = listTitleStyle
	commitList.SetShowStatusBar(false)
	commitList.SetFilteringEnabled(false)
	commitList.KeyMap.NextPage.SetKeys("pgdown")
	commitList.KeyMap.PrevPage.SetKeys("pgup")
	commitList.KeyMap.GoToStart.SetKeys("home")
	commitList.KeyMap.GoToEnd.SetKeys("end")

	fileList := list.New(nil, newFileItemDelegate(), 0, 0)
	fileList.Title = "Files"
	fileList.Styles.Title = listTitleStyle
	fileList.SetShowStatusBar(false)
	fileList.SetFilteringEnabled(false)
	fileList.KeyMap.NextPage.SetKeys("pgdown")
	fileList.KeyMap.PrevPage.SetKeys("pgup")
	fileList.KeyMap.GoToStart.SetKeys("home")
	fileList.KeyMap.GoToEnd.SetKeys("end")

	changesList := list.New(nil, newFileItemDelegate(), 0, 0)
	changesList.Title = "Changes"
	changesList.Styles.Title = listTitleStyle
	changesList.SetShowStatusBar(false)
	changesList.SetFilteringEnabled(false)
	changesList.KeyMap.CursorUp.SetKeys("up")
	changesList.KeyMap.CursorDown.SetKeys("down")
	changesList.KeyMap.NextPage.SetKeys("pgdown")
	changesList.KeyMap.PrevPage.SetKeys("pgup")
	changesList.KeyMap.GoToStart.SetKeys("home")
	changesList.KeyMap.GoToEnd.SetKeys("end")

	return Model{
		repoPath:    repoPath,
		focused:     paneCommits,
		commitList:  commitList,
		fileList:    fileList,
		changesList: changesList,
	}
}

// ── Init ─────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadBranchCmd(m.repoPath), loadCommitsCmd(m.repoPath, 0), watchRepoCmd(m.repoPath))
}

// ── Commands ─────────────────────────────────────────────────────────────────

func loadBranchCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		branch, err := git.CurrentBranch(repoPath)
		if err != nil {
			return errMsg{err}
		}
		return branchLoadedMsg{branch}
	}
}

func loadCommitsCmd(repoPath string, preserveIdx int) tea.Cmd {
	return func() tea.Msg {
		commits, err := git.LoadCommits(repoPath)
		if err != nil {
			return errMsg{err}
		}
		return commitsLoadedMsg{commits: commits, preserveIdx: preserveIdx}
	}
}

func watchRepoCmd(repoPath string) tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		hash, _ := git.HeadHash(repoPath)
		var reflogMtime int64
		if info, err := os.Stat(filepath.Join(repoPath, ".git", "logs", "HEAD")); err == nil {
			reflogMtime = info.ModTime().UnixNano()
		}
		var indexMtime int64
		if info, err := os.Stat(filepath.Join(repoPath, ".git", "index")); err == nil {
			indexMtime = info.ModTime().UnixNano()
		}
		return repoWatchMsg{hash: hash, reflogMtime: reflogMtime, indexMtime: indexMtime}
	})
}

func loadFilesCmd(repoPath, hash string) tea.Cmd {
	return func() tea.Msg {
		files, err := git.LoadFiles(repoPath, hash)
		if err != nil {
			return errMsg{err}
		}
		return filesLoadedMsg{hash: hash, files: files}
	}
}

func loadDiffCmd(repoPath, hash, file string) tea.Cmd {
	return func() tea.Msg {
		diff, err := git.LoadDiff(repoPath, hash, file)
		if err != nil {
			return errMsg{err}
		}
		return diffLoadedMsg{content: diff, mode: modeHistory}
	}
}

func loadChangesCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		staged, err := git.LoadStagedFiles(repoPath)
		if err != nil {
			return errMsg{err}
		}
		unstaged, err := git.LoadUnstagedFiles(repoPath)
		if err != nil {
			return errMsg{err}
		}
		return changesLoadedMsg{staged: staged, unstaged: unstaged}
	}
}

func loadWorkingDiffCmd(repoPath, file string, staged bool) tea.Cmd {
	return func() tea.Msg {
		diff, err := git.LoadWorkingDiff(repoPath, file, staged)
		if err != nil {
			return errMsg{err}
		}
		return diffLoadedMsg{content: diff, mode: modeChanges}
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func gitIsRepo(path string) bool              { return git.IsGitRepo(path) }
func gitRepoRoot(path string) (string, error) { return git.RepoRoot(path) }

func (m Model) selectedCommitHash() string {
	if item, ok := m.commitList.SelectedItem().(commitItem); ok {
		return item.commit.Hash
	}
	return ""
}

func (m Model) selectedFile() string {
	if item, ok := m.fileList.SelectedItem().(fileItem); ok {
		return item.file.Path
	}
	return ""
}

func (m *Model) setPaneSizes() {
	if m.width == 0 || m.height == 0 {
		return
	}
	const borderW, borderH = 2, 2
	const statusBarH = 1

	paneH := m.height - statusBarH - borderH

	commitW := m.width/4 - borderW
	if commitW < 1 {
		commitW = 1
	}
	if paneH < 1 {
		paneH = 1
	}

	m.commitList.SetSize(commitW, paneH)
	m.changesList.SetSize(commitW, paneH)

	if m.mode == modeChanges {
		diffW := m.width - m.width/4 - borderW*2
		if diffW < 1 {
			diffW = 1
		}
		m.diffView.Width = diffW
		m.diffView.Height = paneH
	} else {
		fileW := m.width/4 - borderW
		diffW := m.width - m.width/4 - m.width/4 - borderW*3
		if fileW < 1 {
			fileW = 1
		}
		if diffW < 1 {
			diffW = 1
		}
		m.fileList.SetSize(fileW, paneH)
		m.fileList.SetDelegate(fileItemDelegate{width: fileW, styles: newFileItemDelegate().styles})
		m.diffView.Width = diffW
		m.diffView.Height = paneH
	}
}

// ── Update ───────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.diffView = viewport.New(0, 0)
			m.ready = true
		}
		m.setPaneSizes()
		return m, nil

	case branchLoadedMsg:
		m.branch = msg.branch
		return m, nil

	case commitsLoadedMsg:
		m.commits = msg.commits
		items := make([]list.Item, len(msg.commits))
		for i, c := range msg.commits {
			items[i] = commitItem{c}
		}
		m.commitList.SetItems(items)
		idx := msg.preserveIdx
		if idx >= len(msg.commits) {
			idx = 0
		}
		m.commitList.Select(idx)
		if len(msg.commits) > 0 && m.mode == modeHistory {
			return m, loadFilesCmd(m.repoPath, msg.commits[idx].Hash)
		}
		return m, nil

	case repoWatchMsg:
		if !m.watchInitialized {
			m.watchInitialized = true
			m.headHash = msg.hash
			m.reflogMtime = msg.reflogMtime
			m.indexMtime = msg.indexMtime
			return m, watchRepoCmd(m.repoPath)
		}
		hashChanged := msg.hash != "" && msg.hash != m.headHash
		reflogChanged := msg.reflogMtime != 0 && msg.reflogMtime != m.reflogMtime
		indexChanged := msg.indexMtime != 0 && msg.indexMtime != m.indexMtime
		if msg.hash != "" {
			m.headHash = msg.hash
		}
		if msg.reflogMtime != 0 {
			m.reflogMtime = msg.reflogMtime
		}
		if msg.indexMtime != 0 {
			m.indexMtime = msg.indexMtime
		}
		var cmds []tea.Cmd
		cmds = append(cmds, watchRepoCmd(m.repoPath))
		if hashChanged || reflogChanged {
			idx := m.commitList.Index()
			cmds = append(cmds, loadBranchCmd(m.repoPath), loadCommitsCmd(m.repoPath, idx))
		}
		if indexChanged && m.mode == modeChanges {
			cmds = append(cmds, loadChangesCmd(m.repoPath))
		}
		return m, tea.Batch(cmds...)

	case filesLoadedMsg:
		if m.mode != modeHistory {
			return m, nil
		}
		m.files = msg.files
		m.fileErr = ""
		items := make([]list.Item, len(msg.files))
		for i, f := range msg.files {
			items[i] = fileItem{f}
		}
		m.fileList.SetItems(items)
		m.fileList.Select(0)
		if len(msg.files) > 0 {
			return m, loadDiffCmd(m.repoPath, msg.hash, msg.files[0].Path)
		}
		m.diffView.SetContent("")
		return m, nil

	case changesLoadedMsg:
		if m.mode != modeChanges {
			return m, nil
		}
		m.stagedFiles = msg.staged
		m.unstagedFiles = msg.unstaged
		m.fileErr = ""

		var items []list.Item
		if len(msg.staged) > 0 {
			items = append(items, changesHeaderItem{label: fmt.Sprintf("Staged Changes (%d)", len(msg.staged))})
			for _, f := range msg.staged {
				items = append(items, changesFileItem{file: f, staged: true})
			}
		}
		if len(msg.unstaged) > 0 {
			items = append(items, changesHeaderItem{label: fmt.Sprintf("Unstaged Changes (%d)", len(msg.unstaged))})
			for _, f := range msg.unstaged {
				items = append(items, changesFileItem{file: f, staged: false})
			}
		}
		m.changesList.SetItems(items)

		firstFile := firstChangesFile(items)
		if firstFile >= 0 {
			m.changesList.Select(firstFile)
			item := items[firstFile].(changesFileItem)
			return m, loadWorkingDiffCmd(m.repoPath, item.file.Path, item.staged)
		}
		m.diffView.SetContent("")
		return m, nil

	case diffLoadedMsg:
		if msg.mode != m.mode {
			return m, nil
		}
		m.diffErr = ""
		m.diffView.SetContent(ColorizeDiff(msg.content))
		m.diffView.GotoTop()
		return m, nil

	case errMsg:
		m.fileErr = fmt.Sprintf("error: %v", msg.err)
		m.diffErr = m.fileErr
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, keys.ToggleMode) {
			if m.mode == modeHistory {
				m.mode = modeChanges
				m.focused = paneCommits
				m.diffView.SetContent("")
				m.changesList.SetItems([]list.Item{})
				m.setPaneSizes()
				return m, loadChangesCmd(m.repoPath)
			}
			m.mode = modeHistory
			m.focused = paneCommits
			m.setPaneSizes()
			return m, nil
		}
		if key.Matches(msg, keys.FocusNext) {
			if m.focused < paneDiff {
				m.focused++
				if m.focused == paneFiles {
					m.fileList.Select(0)
					if len(m.files) > 0 {
						return m, loadDiffCmd(m.repoPath, m.selectedCommitHash(), m.files[0].Path)
					}
				}
			}
			return m, nil
		}
		if key.Matches(msg, keys.FocusPrev) {
			if m.focused > paneCommits {
				m.focused--
			}
			return m, nil
		}

		var cmd tea.Cmd
		switch m.focused {
		case paneCommits:
			if m.mode == modeChanges {
				prevIdx := m.changesList.Index()
				m.changesList, cmd = m.changesList.Update(msg)
				newIdx := m.changesList.Index()
				// Skip header items
				items := m.changesList.Items()
				if len(items) > 0 {
					if _, isHeader := items[newIdx].(changesHeaderItem); isHeader {
						if newIdx > prevIdx {
							if newIdx+1 < len(items) {
								m.changesList.Select(newIdx + 1)
							}
						} else if newIdx > 0 {
							m.changesList.Select(newIdx - 1)
						}
						newIdx = m.changesList.Index()
					}
				}
				if newIdx != prevIdx {
					if item, ok := m.changesList.Items()[newIdx].(changesFileItem); ok {
						return m, tea.Batch(cmd, loadWorkingDiffCmd(m.repoPath, item.file.Path, item.staged))
					}
				}
				return m, cmd
			}
			// History mode — existing behavior unchanged
			if key.Matches(msg, keys.GotoTop) {
				m.commitList.Select(0)
				return m, loadFilesCmd(m.repoPath, m.selectedCommitHash())
			}
			if key.Matches(msg, keys.GotoBottom) {
				m.commitList.Select(len(m.commits) - 1)
				return m, loadFilesCmd(m.repoPath, m.selectedCommitHash())
			}
			prevIdx := m.commitList.Index()
			m.commitList, cmd = m.commitList.Update(msg)
			if m.commitList.Index() != prevIdx {
				hash := m.selectedCommitHash()
				return m, tea.Batch(cmd, loadFilesCmd(m.repoPath, hash))
			}
		case paneFiles:
			if key.Matches(msg, keys.GotoTop) {
				m.fileList.Select(0)
				if file := m.selectedFile(); file != "" {
					return m, loadDiffCmd(m.repoPath, m.selectedCommitHash(), file)
				}
				return m, nil
			}
			if key.Matches(msg, keys.GotoBottom) {
				m.fileList.Select(len(m.files) - 1)
				if file := m.selectedFile(); file != "" {
					return m, loadDiffCmd(m.repoPath, m.selectedCommitHash(), file)
				}
				return m, nil
			}
			prevIdx := m.fileList.Index()
			m.fileList, cmd = m.fileList.Update(msg)
			if m.fileList.Index() != prevIdx {
				file := m.selectedFile()
				if file != "" {
					return m, tea.Batch(cmd, loadDiffCmd(m.repoPath, m.selectedCommitHash(), file))
				}
			}
		case paneDiff:
			switch {
			case key.Matches(msg, keys.Up):
				m.diffView.LineUp(3)
			case key.Matches(msg, keys.Down):
				m.diffView.LineDown(3)
			case key.Matches(msg, keys.GotoTop):
				m.diffView.GotoTop()
			case key.Matches(msg, keys.GotoBottom):
				m.diffView.GotoBottom()
			default:
				m.diffView, cmd = m.diffView.Update(msg)
			}
			return m, cmd
		}
		return m, cmd
	}

	// Pass non-key messages to sub-models
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.commitList, cmd = m.commitList.Update(msg)
	cmds = append(cmds, cmd)
	m.fileList, cmd = m.fileList.Update(msg)
	cmds = append(cmds, cmd)
	m.diffView, cmd = m.diffView.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// ── View ─────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.renderCommits(),
			m.renderFiles(),
			m.renderDiff(),
		),
		m.renderStatusBar(),
	)
}

func (m Model) renderCommits() string {
	var tabBar string
	if m.mode == modeHistory {
		tabBar = " Changes   [History]  " + hintRendered
	} else {
		tabBar = "[Changes]   History   " + hintRendered
	}

	if m.mode == modeChanges {
		m.changesList.Title = tabBar
		return borderStyle(m.focused == paneCommits).Render(m.changesList.View())
	}
	m.commitList.Title = tabBar
	return borderStyle(m.focused == paneCommits).Render(m.commitList.View())
}

func (m Model) renderFiles() string {
	if m.mode == modeChanges {
		return ""
	}
	content := m.fileList.View()
	if m.fileErr != "" {
		content = errorStyle.Render(m.fileErr)
	}
	return borderStyle(m.focused == paneFiles).Render(content)
}

func (m Model) renderDiff() string {
	content := m.diffView.View()
	if m.diffErr != "" {
		content = errorStyle.Render(m.diffErr)
	}
	return borderStyle(m.focused == paneDiff).Render(content)
}

func (m Model) renderStatusBar() string {
	var status string
	if m.mode == modeChanges {
		status = fmt.Sprintf("branch: %s  staged: %d  unstaged: %d",
			m.branch,
			len(m.stagedFiles),
			len(m.unstagedFiles),
		)
	} else {
		if len(m.commits) == 0 {
			return statusBarStyle.Render("branch: " + m.branch + "  No commits on this branch")
		}
		var selectedCommit git.Commit
		if item, ok := m.commitList.SelectedItem().(commitItem); ok {
			selectedCommit = item.commit
		}
		status = fmt.Sprintf("branch: %s  commit: %s  author: %s  date: %s",
			m.branch,
			selectedCommit.ShortHash,
			selectedCommit.Author,
			selectedCommit.Date,
		)
		if m.focused >= paneFiles && len(m.files) > 0 {
			status += fmt.Sprintf("  [%d/%d files]", m.fileList.Index()+1, len(m.files))
		}
		if m.focused == paneDiff && m.xOffset > 0 {
			status += fmt.Sprintf("  [→%d]", m.xOffset)
		}
	}

	if m.fileErr != "" {
		status = errorStyle.Render(m.fileErr)
	}

	statsLine := statusBarStyle.Width(m.width).Render(status)
	pathContent := m.selectedFile()
	if m.statusMsg != "" {
		pathContent = m.statusMsg
	}
	pathLine := filePathBarStyle.Width(m.width).Render(pathContent)
	return lipgloss.JoinVertical(lipgloss.Left, pathLine, statsLine)
}

func borderStyle(focused bool) lipgloss.Style {
	if focused {
		return focusedBorder
	}
	return unfocusedBorder
}

func firstChangesFile(items []list.Item) int {
	for i, item := range items {
		if _, ok := item.(changesFileItem); ok {
			return i
		}
	}
	return -1
}
