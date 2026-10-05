import * as api from './api'
import type { Commit, FileChange, DiffLine } from './types'

export type Mode = 'history' | 'changes'
export type Pane = 'commits' | 'files' | 'diff'

export interface ChangesEntry {
  kind: 'header' | 'file'
  label: string
  file: FileChange | null
  staged: boolean
}

export interface State {
  repo: string
  branch: string
  mode: Mode
  focused: Pane

  commits: Commit[]
  commitIndex: number

  files: FileChange[]
  fileIndex: number

  changesEntries: ChangesEntry[]
  changesIndex: number
  stagedCount: number
  unstagedCount: number

  diffLines: DiffLine[]
  diffScrollTop: number
  diffScrollLeft: number

  errorMessage: string
}

export function createInitialState(repo = ''): State {
  return {
    repo,
    branch: '',
    mode: 'history',
    focused: 'commits',
    commits: [],
    commitIndex: 0,
    files: [],
    fileIndex: 0,
    changesEntries: [],
    changesIndex: 0,
    stagedCount: 0,
    unstagedCount: 0,
    diffLines: [],
    diffScrollTop: 0,
    diffScrollLeft: 0,
    errorMessage: '',
  }
}

// One State per open repo, in tab order. `state` is a live binding to the
// active tab's State, so importers always see the current tab. Async work must
// capture the State it started with rather than re-reading `state`, so a slow
// load can't land in whichever tab is active when it finishes.
export const tabs: State[] = []
export let state: State = createInitialState()

let render: () => void = () => {}
export function setRenderer(fn: () => void) {
  render = fn
}

// Re-render only if s is the tab being shown; background tabs update silently.
function rerender(s: State): void {
  if (s === state) render()
}

let diffContainer: HTMLElement | null = null
export function setDiffContainer(el: HTMLElement | null) {
  diffContainer = el
}

function firstChangesFileIndex(s: State): number {
  return s.changesEntries.findIndex((e) => e.kind === 'file')
}

function lastChangesFileIndex(s: State): number {
  for (let i = s.changesEntries.length - 1; i >= 0; i--) {
    if (s.changesEntries[i].kind === 'file') return i
  }
  return s.changesIndex
}

function nextChangesFileIndex(s: State, from: number, direction: 1 | -1): number {
  let i = from
  let candidate = i
  do {
    i += direction
    if (i < 0 || i >= s.changesEntries.length) return candidate
    if (s.changesEntries[i].kind === 'file') candidate = i
  } while (s.changesEntries[i].kind !== 'file')
  return candidate
}

async function loadDiffForSelectedFile(s: State): Promise<void> {
  const modeAtStart = s.mode
  if (modeAtStart === 'history') {
    const file = s.files[s.fileIndex]
    const commit = s.commits[s.commitIndex]
    const lines = file && commit ? await api.loadDiffLines(s.repo, commit.Hash, file.Path) : []
    if (s.mode !== modeAtStart) return
    s.diffLines = lines
  } else {
    const entry = s.changesEntries[s.changesIndex]
    const lines =
      entry && entry.kind === 'file' && entry.file
        ? await api.loadWorkingDiffLines(s.repo, entry.file.Path, entry.staged)
        : []
    if (s.mode !== modeAtStart) return
    s.diffLines = lines
  }
  if (s.mode !== modeAtStart) return
  s.diffScrollTop = 0
  s.diffScrollLeft = 0
  if (s === state) diffContainer?.scrollTo({ top: 0, left: 0 })
}

async function loadFilesForSelectedCommit(s: State): Promise<void> {
  const modeAtStart = s.mode
  const commit = s.commits[s.commitIndex]
  const files = commit ? await api.loadFiles(s.repo, commit.Hash) : []
  if (s.mode !== modeAtStart) return
  s.files = files
  s.fileIndex = 0
}

async function loadTab(s: State): Promise<void> {
  s.branch = await api.currentBranch(s.repo)
  s.commits = await api.loadCommits(s.repo)
  s.commitIndex = 0
  await loadFilesForSelectedCommit(s)
  await loadDiffForSelectedFile(s)
  rerender(s)
}

// Create a tab for repo if it isn't open yet, returning it. A new tab loads in
// the background; callers decide whether to activate it.
function ensureTab(repo: string): { tab: State; created: boolean } {
  const existing = tabs.find((t) => t.repo === repo)
  if (existing) return { tab: existing, created: false }
  const tab = createInitialState(repo)
  tabs.push(tab)
  return { tab, created: true }
}

export function activateTab(repo: string): void {
  const next = tabs.find((t) => t.repo === repo)
  if (!next || next === state) return
  state.diffScrollTop = diffContainer?.scrollTop ?? state.diffScrollTop
  state.diffScrollLeft = diffContainer?.scrollLeft ?? state.diffScrollLeft
  state = next
  render()
  diffContainer?.scrollTo({ top: state.diffScrollTop, left: state.diffScrollLeft })
}

export function activateTabAt(index: number): void {
  const tab = tabs[index]
  if (tab) activateTab(tab.repo)
}

// Called when the backend opens a repo (a later `differ` launch): add its tab if
// new and bring it to the front.
export async function openTab(repo: string): Promise<void> {
  const { tab, created } = ensureTab(repo)
  activateTab(repo)
  if (created) await loadTab(tab)
}

export async function closeTab(repo: string): Promise<void> {
  const index = tabs.findIndex((t) => t.repo === repo)
  if (index < 0) return
  const remaining = await api.closeRepo(repo)
  if (remaining === 0) {
    window.runtime.Quit()
    return
  }
  const wasActive = tabs[index] === state
  tabs.splice(index, 1)
  if (wasActive) {
    state = tabs[Math.min(index, tabs.length - 1)]
  }
  render()
}

export async function init(): Promise<void> {
  const list = await api.listRepos()
  const created = list.Repos.map((repo) => ensureTab(repo).tab)
  const active = tabs.find((t) => t.repo === list.Active) ?? tabs[0]
  if (!active) return
  state = active
  render()
  await Promise.all(created.map((t) => loadTab(t)))
}

export async function refreshForRepoChange(repo: string): Promise<void> {
  const s = tabs.find((t) => t.repo === repo)
  if (!s) return
  s.branch = await api.currentBranch(s.repo)
  s.commits = await api.loadCommits(s.repo)
  if (s.commitIndex >= s.commits.length) s.commitIndex = 0

  if (s.mode === 'history') {
    await loadFilesForSelectedCommit(s)
    await loadDiffForSelectedFile(s)
  } else {
    await loadChanges(s) // loadChanges() already calls rerender()
    return
  }
  rerender(s)
}

export async function loadChanges(s: State = state): Promise<void> {
  const [staged, unstaged] = await Promise.all([api.loadStagedFiles(s.repo), api.loadUnstagedFiles(s.repo)])
  s.stagedCount = staged.length
  s.unstagedCount = unstaged.length

  const entries: ChangesEntry[] = []
  if (staged.length > 0) {
    entries.push({ kind: 'header', label: `Staged Changes (${staged.length})`, file: null, staged: false })
    for (const f of staged) entries.push({ kind: 'file', label: '', file: f, staged: true })
  }
  if (unstaged.length > 0) {
    entries.push({ kind: 'header', label: `Unstaged Changes (${unstaged.length})`, file: null, staged: false })
    for (const f of unstaged) entries.push({ kind: 'file', label: '', file: f, staged: false })
  }
  s.changesEntries = entries

  const first = firstChangesFileIndex(s)
  s.changesIndex = first >= 0 ? first : 0
  if (first >= 0) {
    await loadDiffForSelectedFile(s)
  } else {
    s.diffLines = []
  }
  rerender(s)
}

export async function switchMode(): Promise<void> {
  const s = state
  if (s.mode === 'history') {
    s.mode = 'changes'
    s.focused = 'commits'
    s.diffLines = []
    s.changesEntries = []
    rerender(s)
    await loadChanges(s)
  } else {
    s.mode = 'history'
    s.focused = 'commits'
    rerender(s)
  }
}

export function focusPane(pane: Pane): void {
  if (pane === 'files' && state.mode === 'changes') return
  if (state.focused === pane) return
  state.focused = pane
  render()
}

export async function selectCommit(index: number): Promise<void> {
  const s = state
  if (s.mode !== 'history' || index < 0 || index >= s.commits.length) return
  s.focused = 'commits'
  const prev = s.commitIndex
  s.commitIndex = index
  if (index !== prev) {
    await loadFilesForSelectedCommit(s)
    await loadDiffForSelectedFile(s)
  }
  rerender(s)
}

export async function selectFile(index: number): Promise<void> {
  const s = state
  if (s.mode !== 'history' || index < 0 || index >= s.files.length) return
  s.focused = 'files'
  const prev = s.fileIndex
  s.fileIndex = index
  if (index !== prev) await loadDiffForSelectedFile(s)
  rerender(s)
}

export async function selectChangesEntry(index: number): Promise<void> {
  const s = state
  if (s.mode !== 'changes') return
  const entry = s.changesEntries[index]
  if (!entry || entry.kind !== 'file') return
  s.focused = 'commits'
  const prev = s.changesIndex
  s.changesIndex = index
  if (index !== prev) await loadDiffForSelectedFile(s)
  rerender(s)
}

export async function handleKey(key: string): Promise<void> {
  const s = state

  if (key === 'q') {
    await closeTab(s.repo)
    return
  }

  if (key === 'c') {
    await switchMode()
    return
  }

  if ((key === 'l' || key === 'ArrowRight' || key === 'Tab') && s.focused !== 'diff') {
    if (s.mode === 'changes') {
      if (s.focused === 'commits') s.focused = 'diff'
    } else if (s.focused === 'commits') {
      s.focused = 'files'
      s.fileIndex = 0
      if (s.files.length > 0) await loadDiffForSelectedFile(s)
    } else if (s.focused === 'files') {
      s.focused = 'diff'
    }
    rerender(s)
    return
  }
  if ((key === 'h' || key === 'ArrowLeft' || key === 'Shift+Tab') && s.focused !== 'diff') {
    // s.focused is already narrowed to exclude 'diff' here (guarded above); its own
    // left-navigation is handled in the `s.focused === 'diff'` branch further below.
    if (s.mode === 'changes') {
      // 'commits' is the leftmost pane in changes mode; nothing further left to focus.
    } else if (s.focused === 'files') {
      s.focused = 'commits'
    }
    rerender(s)
    return
  }

  if (s.focused === 'commits') {
    if (s.mode === 'changes') {
      const prev = s.changesIndex
      if (key === 'j' || key === 'ArrowDown') s.changesIndex = nextChangesFileIndex(s, prev, 1)
      else if (key === 'k' || key === 'ArrowUp') s.changesIndex = nextChangesFileIndex(s, prev, -1)
      else if (key === 'g') s.changesIndex = firstChangesFileIndex(s) >= 0 ? firstChangesFileIndex(s) : prev
      else if (key === 'G') s.changesIndex = lastChangesFileIndex(s)
      if (s.changesIndex !== prev) await loadDiffForSelectedFile(s)
      rerender(s)
      return
    }
    const prev = s.commitIndex
    if (key === 'g') s.commitIndex = 0
    else if (key === 'G') s.commitIndex = s.commits.length - 1
    else if (key === 'j' || key === 'ArrowDown') s.commitIndex = Math.min(prev + 1, s.commits.length - 1)
    else if (key === 'k' || key === 'ArrowUp') s.commitIndex = Math.max(prev - 1, 0)
    if (s.commitIndex !== prev) {
      await loadFilesForSelectedCommit(s)
      await loadDiffForSelectedFile(s)
    }
    rerender(s)
    return
  }

  if (s.focused === 'files') {
    const prev = s.fileIndex
    if (key === 'g') s.fileIndex = 0
    else if (key === 'G') s.fileIndex = s.files.length - 1
    else if (key === 'j' || key === 'ArrowDown') s.fileIndex = Math.min(prev + 1, s.files.length - 1)
    else if (key === 'k' || key === 'ArrowUp') s.fileIndex = Math.max(prev - 1, 0)
    if (s.fileIndex !== prev) await loadDiffForSelectedFile(s)
    rerender(s)
    return
  }

  if (s.focused === 'diff') {
    const el = diffContainer
    if (key === 'k' || key === 'ArrowUp') el?.scrollBy({ top: -60 })
    else if (key === 'j' || key === 'ArrowDown') el?.scrollBy({ top: 60 })
    else if (key === 'h' || key === 'ArrowLeft') {
      if (!el || el.scrollLeft <= 0) {
        s.focused = s.mode === 'changes' ? 'commits' : 'files'
        rerender(s)
      } else {
        el.scrollBy({ left: -32 })
      }
    } else if (key === 'l' || key === 'ArrowRight') {
      el?.scrollBy({ left: 32 })
    } else if (key === 'g') {
      el?.scrollTo({ top: 0 })
    } else if (key === 'G') {
      el?.scrollTo({ top: el.scrollHeight })
    } else if (key === 'Shift+Tab') {
      s.focused = s.mode === 'changes' ? 'commits' : 'files'
      rerender(s)
    }
    return
  }
}
