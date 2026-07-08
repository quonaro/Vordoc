export function useGlobalSearchShortcut(open: () => void) {
  if (!import.meta.client) return

  const onKeydown = (e: KeyboardEvent) => {
    const target = e.target as HTMLElement | null
    const isTyping =
      target?.tagName === 'INPUT' ||
      target?.tagName === 'TEXTAREA' ||
      target?.isContentEditable
    const isModifier = e.metaKey || e.ctrlKey
    const isKKey = e.code === 'KeyK'

    if (isModifier && isKKey && !isTyping) {
      e.preventDefault()
      e.stopPropagation()
      open()
    }
  }

  document.addEventListener('keydown', onKeydown, true)
  onUnmounted(() => {
    document.removeEventListener('keydown', onKeydown, true)
  })
}
