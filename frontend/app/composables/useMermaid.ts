import type { Ref } from 'vue'
import mermaid from 'mermaid'

export function useMermaid(
  containerRef: Ref<HTMLElement | null>,
  theme: Ref<string>,
) {
  let initialized = false

  function init() {
    if (!import.meta.client) return
    mermaid.initialize({
      startOnLoad: false,
      theme: theme.value === 'dark' ? 'dark' : 'default',
      securityLevel: 'strict',
    })
    initialized = true
  }

  async function render() {
    if (!import.meta.client) return
    const container = containerRef.value
    if (!container) return

    const nodes = container.querySelectorAll<HTMLElement>(
      '.mermaid:not([data-processed="true"])',
    )
    if (!nodes.length) return

    if (!initialized) {
      init()
    }

    await mermaid.run({ nodes: Array.from(nodes) })
  }

  onMounted(() => {
    nextTick(() => render())
  })

  watch(
    () => containerRef.value?.innerHTML,
    () => {
      nextTick(() => render())
    },
  )

  watch(theme, () => {
    init()
    const container = containerRef.value
    if (!container) return
    const nodes = container.querySelectorAll<HTMLElement>(
      '.mermaid[data-processed="true"]',
    )
    for (const node of nodes) {
      node.removeAttribute('data-processed')
    }
    render()
  })
}
