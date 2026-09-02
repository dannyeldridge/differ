import './style.css'
import { init, handleKey, setRenderer } from './state'
import { render } from './render'

setRenderer(render)

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
    void handleKey(key)
  }
})

void init()
