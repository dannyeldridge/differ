// Drag-to-resize for the pane columns. Widths live in CSS variables on <html>
// as `fr` weights, because render() rebuilds .panes on every state change and
// would otherwise wipe them. Weights (rather than px) keep the columns
// proportional when the window itself is resized.

const STORAGE_KEY = 'differ.paneWeights'
const MIN_PANE = 120
const MIN_DIFF = 240
const NAMES = ['commits', 'files', 'diff'] as const
type Name = (typeof NAMES)[number]

const root = document.documentElement

function minWidth(name: Name): number {
  return name === 'diff' ? MIN_DIFF : MIN_PANE
}

function setWeight(name: Name, px: number): void {
  root.style.setProperty(`--w-${name}`, `${px}fr`)
}

function paneName(el: Element): Name {
  return el.id.replace('pane-', '') as Name
}

interface Drag {
  left: Name
  right: Name
  startX: number
  leftWidth: number
  total: number
}

let drag: Drag | null = null

function save(): void {
  try {
    const weights = NAMES.map((n) => parseFloat(root.style.getPropertyValue(`--w-${n}`)))
    localStorage.setItem(STORAGE_KEY, JSON.stringify(weights))
  } catch {
    // Storage can be unavailable; the layout just won't persist.
  }
}

function restore(): void {
  try {
    const weights = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? 'null')
    if (!Array.isArray(weights)) return
    NAMES.forEach((n, i) => {
      if (Number.isFinite(weights[i]) && weights[i] > 0) setWeight(n, weights[i])
    })
  } catch {
    // Corrupt or inaccessible storage: keep the default layout.
  }
}

export function initPaneResize(): void {
  restore()

  window.addEventListener('mousedown', (e) => {
    const handle = (e.target as HTMLElement).closest<HTMLElement>('.resizer')
    const leftEl = handle?.previousElementSibling
    const rightEl = handle?.nextElementSibling
    if (!leftEl || !rightEl) return
    e.preventDefault()

    // Freeze every pane at its current width first, so only the two panes on
    // either side of the handle change while dragging. Measure everything
    // before writing anything: each write re-lays out the grid, so reading in
    // between would see a half-updated layout.
    const widths = Array.from(document.querySelectorAll('.panes > .pane')).map(
      (el) => [paneName(el), el.getBoundingClientRect().width] as const,
    )
    widths.forEach(([name, width]) => setWeight(name, width))
    const left = paneName(leftEl)
    const right = paneName(rightEl)
    const leftWidth = widths.find(([n]) => n === left)![1]
    drag = {
      left,
      right,
      startX: e.clientX,
      leftWidth,
      total: leftWidth + widths.find(([n]) => n === right)![1],
    }
    document.body.classList.add('resizing')
  })

  window.addEventListener('mousemove', (e) => {
    if (!drag) return
    const { left, right, startX, leftWidth, total } = drag
    const width = Math.min(Math.max(leftWidth + e.clientX - startX, minWidth(left)), total - minWidth(right))
    setWeight(left, width)
    setWeight(right, total - width)
  })

  window.addEventListener('mouseup', () => {
    if (!drag) return
    drag = null
    document.body.classList.remove('resizing')
    save()
  })
}
