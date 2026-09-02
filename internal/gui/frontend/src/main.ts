import './style.css'
import { init, handleKey, setRenderer, refreshForRepoChange, state } from './state'
import { render } from './render'

setRenderer(render)

function safely(fn: () => Promise<void>): void {
  fn().catch((err) => {
    state.errorMessage = err instanceof Error ? err.message : String(err)
    render()
  })
}

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
