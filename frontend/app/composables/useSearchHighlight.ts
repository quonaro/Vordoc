const SEARCH_MARK_CLASS = 'search-mark'

function unwrapMarks(container: HTMLElement) {
  const marks = container.querySelectorAll(`.${SEARCH_MARK_CLASS}`)
  marks.forEach((mark) => {
    const parent = mark.parentNode
    if (!parent) return
    while (mark.firstChild) {
      parent.insertBefore(mark.firstChild, mark)
    }
    parent.removeChild(mark)
    parent.normalize()
  })
}

function wrapTextNode(node: Text, pattern: RegExp) {
  const text = node.textContent ?? ''
  const matches = Array.from(text.matchAll(pattern))
  if (!matches.length) return

  const fragment = document.createDocumentFragment()
  let lastIndex = 0
  for (const match of matches) {
    const start = match.index ?? 0
    const end = start + match[0].length

    if (start > lastIndex) {
      fragment.appendChild(document.createTextNode(text.slice(lastIndex, start)))
    }

    const span = document.createElement('span')
    span.className = SEARCH_MARK_CLASS
    span.textContent = match[0]
    fragment.appendChild(span)
    lastIndex = end
  }

  if (lastIndex < text.length) {
    fragment.appendChild(document.createTextNode(text.slice(lastIndex)))
  }

  const parent = node.parentNode
  if (parent) {
    parent.replaceChild(fragment, node)
  }
}

function highlightTerm(container: HTMLElement, term: string) {
  const trimmed = term.trim()
  if (!trimmed) return

  const terms = trimmed
    .toLowerCase()
    .split(/\s+/)
    .filter((t) => t.length >= 2)
  if (!terms.length) return

  const pattern = new RegExp(
    `(${terms.map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`,
    'gi',
  )

  const walker = document.createTreeWalker(
    container,
    NodeFilter.SHOW_TEXT,
    {
      acceptNode: (node) => {
        const parent = node.parentElement
        if (!parent) return NodeFilter.FILTER_REJECT
        if (parent.classList.contains(SEARCH_MARK_CLASS)) {
          return NodeFilter.FILTER_REJECT
        }
        if (parent.closest('pre, code, script, style')) {
          return NodeFilter.FILTER_REJECT
        }
        return NodeFilter.FILTER_ACCEPT
      },
    },
  )

  const nodes: Text[] = []
  let node: Node | null
  while ((node = walker.nextNode())) {
    nodes.push(node as Text)
  }

  nodes.forEach((n) => wrapTextNode(n, pattern))
}

export function useSearchHighlight(
  containerRef: MaybeRefOrGetter<HTMLElement | null>,
  contentRef: MaybeRefOrGetter<string | undefined>,
  duration = 2500,
) {
  const route = useRoute()
  const term = computed(() => {
    const q = route.query.q
    return typeof q === 'string' ? q : ''
  })

  let timer: ReturnType<typeof setTimeout> | null = null

  function clear() {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
    const el = toValue(containerRef)
    if (el) unwrapMarks(el)
  }

  async function apply() {
    await nextTick()
    if (!import.meta.client) return
    clear()
    const el = toValue(containerRef)
    const content = toValue(contentRef)
    if (!el || !term.value.trim() || !content) return
    highlightTerm(el, term.value)
    timer = setTimeout(() => {
      unwrapMarks(el)
    }, duration)
  }

  watch([term, () => toValue(contentRef)], apply, { immediate: true })

  onUnmounted(() => {
    clear()
  })
}
