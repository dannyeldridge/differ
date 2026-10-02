import './style.css'
import {
  init,
  handleKey,
  setRenderer,
  refreshForRepoChange,
  state,
  switchMode,
  focusPane,
  selectCommit,
  selectFile,
  selectChangesEntry,
} from './state'
import { render } from './render'

setRenderer(render)

function safely(fn: () => Promise<void>): void {
  fn().catch((err) => {
    state.errorMessage = err instanceof Error ? err.message : String(err)
    render()
  })
}

document.getElementById('app')!.addEventListener('click', (e) => {
  // A drag-select ends in a click; re-rendering here would wipe the selection.
  if (!window.getSelection()?.isCollapsed) return

  const target = e.target as HTMLElement

  const copyBtn = target.closest<HTMLElement>('.copy-btn')
  if (copyBtn) {
    const path = copyBtn.dataset.path ?? ''
    navigator.clipboard.writeText(path).then(() => {
      copyBtn.classList.add('copied')
      setTimeout(() => copyBtn.classList.remove('copied'), 1000)
    })
    return
  }

  const tab = target.closest<HTMLElement>('[data-tab]')
  if (tab) {
    const wantsChanges = tab.dataset.tab === 'changes'
    if (wantsChanges !== (state.mode === 'changes')) safely(() => switchMode())
    return
  }

  const pane = target.closest<HTMLElement>('.pane')
  if (!pane) return
  const row = target.closest<HTMLElement>('.row[data-index]')
  const index = row ? Number(row.dataset.index) : -1

  if (pane.id === 'pane-commits') {
    if (state.mode === 'changes') {
      if (index >= 0) safely(() => selectChangesEntry(index))
      else focusPane('commits')
    } else if (index >= 0) {
      safely(() => selectCommit(index))
    } else {
      focusPane('commits')
    }
  } else if (pane.id === 'pane-files') {
    if (index >= 0) safely(() => selectFile(index))
    else focusPane('files')
  } else if (pane.id === 'pane-diff') {
    focusPane('diff')
  }
})

const HANDLED_KEYS = new Set([
  'j', 'k', 'h', 'l', 'g', 'G', 'c', 'q',
  'Tab', 'ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Shift+Tab',
])

window.addEventListener('keydown', (e) => {
  if (e.metaKey && (e.key === 'q' || e.key === 'w')) {
    e.preventDefault()
    window.runtime.Quit()
    return
  }
  if (e.metaKey || e.ctrlKey) return
  const key = e.shiftKey && e.key === 'Tab' ? 'Shift+Tab' : e.key
  if (HANDLED_KEYS.has(key)) {
    e.preventDefault()
    safely(() => handleKey(key))
  }
})

window.runtime.EventsOn('repo-changed', () => {
  safely(() => refreshForRepoChange())
})

window.addEventListener('unhandledrejection', (e) => {
  state.errorMessage = String(e.reason)
  render()
})

safely(() => init())
