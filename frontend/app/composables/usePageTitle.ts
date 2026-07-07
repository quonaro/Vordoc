import {
  computed,
  onScopeDispose,
  readonly,
  toValue,
  watch,
  type MaybeRefOrGetter,
} from 'vue'

export function usePageTitle() {
  const siteConfig = useSiteConfig()
  const { t } = useText()

  const parts = useState<string[]>('page-title-parts', () => [])

  let stopWatcher: (() => void) | null = null

  function set(value: MaybeRefOrGetter<(string | null | undefined)[]>) {
    stopWatcher?.()

    stopWatcher = watch(
      () => toValue(value),
      (resolved) => {
        parts.value = resolved.filter((part): part is string => !!part)
      },
      { immediate: true },
    )
  }

  onScopeDispose(() => {
    stopWatcher?.()
  })

  const fullTitle = computed(() => {
    const siteTitle = siteConfig.data.value?.root?.title ?? t('app.title')
    const separator = t('app.titleSeparator') ?? '·'
    if (!parts.value.length) return siteTitle
    return [...parts.value, siteTitle].join(` ${separator} `)
  })

  return {
    set,
    fullTitle,
    parts: readonly(parts),
  }
}
