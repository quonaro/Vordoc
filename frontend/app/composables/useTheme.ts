const STORAGE_KEY = 'vordoc:theme'

export type ThemeMode = 'system' | 'light' | 'dark'

export function useTheme() {
  const siteConfig = useSiteConfig()
  const defaultTheme = computed<ThemeMode>(() => {
    const raw = siteConfig.data.value?.theme?.default
    if (raw === 'light' || raw === 'dark') return raw
    return 'system'
  })

  const systemDark = useState<boolean>('theme-system-dark', () => false)
  const theme = useState<ThemeMode>('theme-mode', () => 'system')
  const isInitialized = useState<boolean>('theme-initialized', () => false)

  function readStoredTheme(): ThemeMode {
    if (!import.meta.client) return defaultTheme.value
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw === 'light' || raw === 'dark' || raw === 'system') return raw
    return defaultTheme.value
  }

  function applyTheme() {
    if (!import.meta.client) return
    const isDark =
      theme.value === 'dark' || (theme.value === 'system' && systemDark.value)
    const html = document.documentElement
    if (isDark) {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }
    localStorage.setItem(STORAGE_KEY, theme.value)
  }

  function setTheme(value: ThemeMode) {
    theme.value = value
    applyTheme()
  }

  function updateSystemDark() {
    if (!import.meta.client) return
    systemDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }

  onMounted(() => {
    if (isInitialized.value) return
    isInitialized.value = true

    updateSystemDark()
    theme.value = readStoredTheme()
    applyTheme()

    if (import.meta.client) {
      const media = window.matchMedia('(prefers-color-scheme: dark)')
      media.addEventListener('change', () => {
        updateSystemDark()
        applyTheme()
      })
    }
  })

  return {
    theme: readonly(theme),
    setTheme,
    options: ['system', 'light', 'dark'] as ThemeMode[],
  }
}
