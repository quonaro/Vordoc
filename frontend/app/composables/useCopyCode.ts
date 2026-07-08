import type { Ref } from 'vue'

export function useCopyCode(containerRef: Ref<HTMLElement | null>) {
  const { t } = useText()

  async function copy(button: HTMLButtonElement) {
    const block = button.closest('.vordoc-code-block')
    const codeEl = block?.querySelector('code')
    if (!codeEl) return

    try {
      await navigator.clipboard.writeText(codeEl.textContent ?? '')
      const original = button.textContent
      button.textContent = t('copyCode.copied')
      setTimeout(() => {
        button.textContent = original
      }, 2000)
    } catch {
      button.textContent = t('copyCode.failed')
    }
  }

  function attach() {
    const container = containerRef.value
    if (!container) return
    const buttons =
      container.querySelectorAll<HTMLButtonElement>('.vordoc-copy-btn')
    for (const btn of buttons) {
      if ((btn as unknown as { __copyAttached?: boolean }).__copyAttached)
        continue
      ;(btn as unknown as { __copyAttached?: boolean }).__copyAttached = true
      btn.textContent = t('copyCode.copy')
      btn.setAttribute('aria-label', t('copyCode.copy'))
      btn.addEventListener('click', () => copy(btn))
    }
  }

  onMounted(() => {
    attach()
  })

  watch(
    () => containerRef.value?.innerHTML,
    () => {
      nextTick(() => attach())
    },
  )
}
