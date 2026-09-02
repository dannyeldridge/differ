import { state, setDiffContainer } from './state'

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function statusClass(status: string): string {
  switch (status) {
    case 'M':
      return 'status-m'
    case 'A':
      return 'status-a'
    case 'D':
      return 'status-d'
    case 'R':
      return 'status-r'
    case 'C':
      return 'status-c'
    default:
      return ''
  }
}

function renderCommitsPane(): string {
  const tabs = `${state.mode === 'changes' ? '[Changes]' : 'Changes'}   ${
    state.mode === 'history' ? '[History]' : 'History'
  } <span class="hint">c</span>`

  if (state.mode === 'changes') {
    const rows = state.changesEntries
      .map((e, i) => {
        if (e.kind === 'header') return `<div class="row header-row">${escapeHtml(e.label)}</div>`
        const f = e.file!
        const selected = i === state.changesIndex ? 'selected' : ''
        return `<div class="row ${selected}"><span class="status ${statusClass(f.Status)}">${f.Status}</span> ${escapeHtml(f.Path)}</div>`
      })
      .join('')
    return `<div class="pane-title">${tabs}</div><div class="list">${rows}</div>`
  }

  const rows = state.commits
    .map((c, i) => {
      const selected = i === state.commitIndex ? 'selected' : ''
      return `<div class="row ${selected}"><div class="commit-subject">${escapeHtml(c.ShortHash)} ${escapeHtml(c.Subject)}</div><div class="commit-meta">${escapeHtml(c.Author)} · ${escapeHtml(c.Date)}</div></div>`
    })
    .join('')
  const detail = state.commits[state.commitIndex]
  const detailHtml = detail
    ? `<div class="commit-detail">
         <div class="commit-detail-subject">${escapeHtml(detail.Subject)}</div>
         <div class="commit-detail-meta">${escapeHtml(detail.ShortHash)}  ${escapeHtml(detail.Author)}  ${escapeHtml(detail.Date)}</div>
         ${detail.Body ? `<div class="commit-detail-body">${escapeHtml(detail.Body)}</div>` : ''}
       </div>`
    : ''
  return `<div class="pane-title">${tabs}</div><div class="list">${rows}</div>${detailHtml}`
}

function renderFilesPane(): string {
  const rows = state.files
    .map((f, i) => {
      const selected = i === state.fileIndex ? 'selected' : ''
      return `<div class="row ${selected}"><span class="status ${statusClass(f.Status)}">${f.Status}</span> ${escapeHtml(f.Path)}</div>`
    })
    .join('')
  return `<div class="pane-title">Files</div><div class="list">${rows}</div>`
}

function renderDiffPane(): string {
  return state.diffLines
    .map((l) => {
      switch (l.Type) {
        case 'add':
          return `<div class="diff-line diff-add"><span class="gutter">+</span><span class="content">${escapeHtml(l.Content)}</span></div>`
        case 'del':
          return `<div class="diff-line diff-del"><span class="gutter">-</span><span class="content">${escapeHtml(l.Content)}</span></div>`
        case 'hunk':
          return `<div class="diff-line diff-hunk"><span class="content">${escapeHtml(l.Content)}</span></div>`
        case 'header':
          return `<div class="diff-line diff-header"><span class="content">${escapeHtml(l.Content)}</span></div>`
        default:
          return `<div class="diff-line diff-context"><span class="gutter"></span><span class="content">${escapeHtml(l.Content)}</span></div>`
      }
    })
    .join('')
}

function selectedPathLine(): string {
  if (state.mode === 'changes') {
    const entry = state.changesEntries[state.changesIndex]
    return entry && entry.kind === 'file' && entry.file ? entry.file.Path : ''
  }
  if (state.focused === 'commits') return state.commits[state.commitIndex]?.Subject ?? ''
  return state.files[state.fileIndex]?.Path ?? ''
}

function statusBarText(): string {
  const prefix = state.errorMessage ? `error: ${state.errorMessage}  ` : ''
  if (state.mode === 'changes') {
    return `${prefix}branch: ${state.branch}  staged: ${state.stagedCount}  unstaged: ${state.unstagedCount}`
  }
  const commit = state.commits[state.commitIndex]
  let text = `${prefix}branch: ${state.branch}  commit: ${commit?.ShortHash ?? ''}  author: ${commit?.Author ?? ''}  date: ${commit?.Date ?? ''}`
  if (state.focused !== 'commits' && state.files.length > 0) {
    text += `  [${state.fileIndex + 1}/${state.files.length} files]`
  }
  return text
}

export function render(): void {
  const app = document.getElementById('app')!
  app.innerHTML = `
    <div class="panes ${state.mode === 'changes' ? 'changes-mode' : ''}">
      <div id="pane-commits" class="pane ${state.focused === 'commits' ? 'focused' : ''}">${renderCommitsPane()}</div>
      ${
        state.mode === 'history'
          ? `<div id="pane-files" class="pane ${state.focused === 'files' ? 'focused' : ''}">${renderFilesPane()}</div>`
          : ''
      }
      <div id="pane-diff" class="pane diff-pane ${state.focused === 'diff' ? 'focused' : ''}">${renderDiffPane()}</div>
    </div>
    <div class="path-bar">${escapeHtml(selectedPathLine())}</div>
    <div class="status-bar">${escapeHtml(statusBarText())}</div>
  `
  setDiffContainer(document.getElementById('pane-diff'))
  document.querySelector(`#pane-${state.focused} .row.selected`)?.scrollIntoView({ block: 'nearest' })
}
