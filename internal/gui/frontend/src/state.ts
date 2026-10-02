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

  errorMessage: string
}

export function createInitialState(): State {
  return {
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
    errorMessage: '',
  }
}

export const state = createInitialState()

let render: () => void = () => {}
export function setRenderer(fn: () => void) {
  render = fn
}

let diffContainer: HTMLElement | null = null
export function setDiffContainer(el: HTMLElement | null) {
  diffContainer = el
}

function firstChangesFileIndex(): number {
  return state.changesEntries.findIndex((e) => e.kind === 'file')
}

function lastChangesFileIndex(): number {
  for (let i = state.changesEntries.length - 1; i >= 0; i--) {
    if (state.changesEntries[i].kind === 'file') return i
  }
  return state.changesIndex
}

function nextChangesFileIndex(from: number, direction: 1 | -1): number {
  let i = from
  let candidate = i
  do {
    i += direction
    if (i < 0 || i >= state.changesEntries.length) return candidate
    if (state.changesEntries[i].kind === 'file') candidate = i
  } while (state.changesEntries[i].kind !== 'file')
  return candidate
}

async function loadDiffForSelectedFile(): Promise<void> {
  const modeAtStart = state.mode
  if (modeAtStart === 'history') {
    const file = state.files[state.fileIndex]
    const commit = state.commits[state.commitIndex]
    const lines = file && commit ? await api.loadDiffLines(commit.Hash, file.Path) : []
    if (state.mode !== modeAtStart) return
    state.diffLines = lines
  } else {
    const entry = state.changesEntries[state.changesIndex]
    const lines =
      entry && entry.kind === 'file' && entry.file
        ? await api.loadWorkingDiffLines(entry.file.Path, entry.staged)
        : []
    if (state.mode !== modeAtStart) return
    state.diffLines = lines
  }
  if (state.mode !== modeAtStart) return
  diffContainer?.scrollTo({ top: 0, left: 0 })
}

async function loadFilesForSelectedCommit(): Promise<void> {
  const modeAtStart = state.mode
  const commit = state.commits[state.commitIndex]
  const files = commit ? await api.loadFiles(commit.Hash) : []
  if (state.mode !== modeAtStart) return
  state.files = files
  state.fileIndex = 0
}

export async function init(): Promise<void> {
  state.branch = await api.currentBranch()
  state.commits = await api.loadCommits()
  state.commitIndex = 0
  await loadFilesForSelectedCommit()
  await loadDiffForSelectedFile()
  render()
}

export async function refreshForRepoChange(): Promise<void> {
  state.branch = await api.currentBranch()
  state.commits = await api.loadCommits()
  if (state.commitIndex >= state.commits.length) state.commitIndex = 0

  if (state.mode === 'history') {
    await loadFilesForSelectedCommit()
    await loadDiffForSelectedFile()
  } else {
    await loadChanges() // loadChanges() already calls render()
    return
  }
  render()
}

export async function loadChanges(): Promise<void> {
  const [staged, unstaged] = await Promise.all([api.loadStagedFiles(), api.loadUnstagedFiles()])
  state.stagedCount = staged.length
  state.unstagedCount = unstaged.length

  const entries: ChangesEntry[] = []
  if (staged.length > 0) {
    entries.push({ kind: 'header', label: `Staged Changes (${staged.length})`, file: null, staged: false })
    for (const f of staged) entries.push({ kind: 'file', label: '', file: f, staged: true })
  }
  if (unstaged.length > 0) {
    entries.push({ kind: 'header', label: `Unstaged Changes (${unstaged.length})`, file: null, staged: false })
    for (const f of unstaged) entries.push({ kind: 'file', label: '', file: f, staged: false })
  }
  state.changesEntries = entries

  const first = firstChangesFileIndex()
  state.changesIndex = first >= 0 ? first : 0
  if (first >= 0) {
    await loadDiffForSelectedFile()
  } else {
    state.diffLines = []
  }
  render()
}

export async function switchMode(): Promise<void> {
  if (state.mode === 'history') {
    state.mode = 'changes'
    state.focused = 'commits'
    state.diffLines = []
    state.changesEntries = []
    render()
    await loadChanges()
  } else {
    state.mode = 'history'
    state.focused = 'commits'
    render()
  }
}

export function focusPane(pane: Pane): void {
  if (pane === 'files' && state.mode === 'changes') return
  if (state.focused === pane) return
  state.focused = pane
  render()
}

export async function selectCommit(index: number): Promise<void> {
  if (state.mode !== 'history' || index < 0 || index >= state.commits.length) return
  state.focused = 'commits'
  const prev = state.commitIndex
  state.commitIndex = index
  if (index !== prev) {
    await loadFilesForSelectedCommit()
    await loadDiffForSelectedFile()
  }
  render()
}

export async function selectFile(index: number): Promise<void> {
  if (state.mode !== 'history' || index < 0 || index >= state.files.length) return
  state.focused = 'files'
  const prev = state.fileIndex
  state.fileIndex = index
  if (index !== prev) await loadDiffForSelectedFile()
  render()
}

export async function selectChangesEntry(index: number): Promise<void> {
  if (state.mode !== 'changes') return
  const entry = state.changesEntries[index]
  if (!entry || entry.kind !== 'file') return
  state.focused = 'commits'
  const prev = state.changesIndex
  state.changesIndex = index
  if (index !== prev) await loadDiffForSelectedFile()
  render()
}

export async function handleKey(key: string): Promise<void> {
  if (key === 'q') {
    window.runtime.Quit()
    return
  }

  if (key === 'c') {
    await switchMode()
    return
  }

  if ((key === 'l' || key === 'ArrowRight' || key === 'Tab') && state.focused !== 'diff') {
    if (state.mode === 'changes') {
      if (state.focused === 'commits') state.focused = 'diff'
    } else if (state.focused === 'commits') {
      state.focused = 'files'
      state.fileIndex = 0
      if (state.files.length > 0) await loadDiffForSelectedFile()
    } else if (state.focused === 'files') {
      state.focused = 'diff'
    }
    render()
    return
  }
  if ((key === 'h' || key === 'ArrowLeft' || key === 'Shift+Tab') && state.focused !== 'diff') {
    // state.focused is already narrowed to exclude 'diff' here (guarded above); its own
    // left-navigation is handled in the `state.focused === 'diff'` branch further below.
    if (state.mode === 'changes') {
      // 'commits' is the leftmost pane in changes mode; nothing further left to focus.
    } else if (state.focused === 'files') {
      state.focused = 'commits'
    }
    render()
    return
  }

  if (state.focused === 'commits') {
    if (state.mode === 'changes') {
      const prev = state.changesIndex
      if (key === 'j' || key === 'ArrowDown') state.changesIndex = nextChangesFileIndex(prev, 1)
      else if (key === 'k' || key === 'ArrowUp') state.changesIndex = nextChangesFileIndex(prev, -1)
      else if (key === 'g') state.changesIndex = firstChangesFileIndex() >= 0 ? firstChangesFileIndex() : prev
      else if (key === 'G') state.changesIndex = lastChangesFileIndex()
      if (state.changesIndex !== prev) await loadDiffForSelectedFile()
      render()
      return
    }
    const prev = state.commitIndex
    if (key === 'g') state.commitIndex = 0
    else if (key === 'G') state.commitIndex = state.commits.length - 1
    else if (key === 'j' || key === 'ArrowDown') state.commitIndex = Math.min(prev + 1, state.commits.length - 1)
    else if (key === 'k' || key === 'ArrowUp') state.commitIndex = Math.max(prev - 1, 0)
    if (state.commitIndex !== prev) {
      await loadFilesForSelectedCommit()
      await loadDiffForSelectedFile()
    }
    render()
    return
  }

  if (state.focused === 'files') {
    const prev = state.fileIndex
    if (key === 'g') state.fileIndex = 0
    else if (key === 'G') state.fileIndex = state.files.length - 1
    else if (key === 'j' || key === 'ArrowDown') state.fileIndex = Math.min(prev + 1, state.files.length - 1)
    else if (key === 'k' || key === 'ArrowUp') state.fileIndex = Math.max(prev - 1, 0)
    if (state.fileIndex !== prev) await loadDiffForSelectedFile()
    render()
    return
  }

  if (state.focused === 'diff') {
    const el = diffContainer
    if (key === 'k' || key === 'ArrowUp') el?.scrollBy({ top: -60 })
    else if (key === 'j' || key === 'ArrowDown') el?.scrollBy({ top: 60 })
    else if (key === 'h' || key === 'ArrowLeft') {
      if (!el || el.scrollLeft <= 0) {
        state.focused = state.mode === 'changes' ? 'commits' : 'files'
        render()
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
      state.focused = state.mode === 'changes' ? 'commits' : 'files'
      render()
    }
    return
  }
}
