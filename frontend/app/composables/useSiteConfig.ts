const STORAGE_KEY = 'vordoc:site-config:v2'

export interface LogoConfig {
  path?: string
  size?: number
  link?: string
}

export interface FontConfig {
  name?: string
  size?: number
}

export type HeaderElement = 'logo' | 'search' | 'theme-switch'

export interface HeaderConfig {
  enable: boolean
  elements?: readonly HeaderElement[]
  title?: string
  logo?: LogoConfig
  font?: FontConfig
}

export interface ThemeConfig {
  default?: string
  accent_color?: string
}

export interface RootPageConfig {
  enable: boolean
  title?: string
}

export interface SiteConfig {
  root: RootPageConfig
  favicon?: string
  header?: HeaderConfig
  theme?: ThemeConfig
}

interface SiteConfigState {
  data: SiteConfig | null
  loading: boolean
  error: Error | null
  loadPromise: Promise<SiteConfig> | null
}

function readStoredData(): SiteConfig | null {
  if (!import.meta.client) return null
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    const data = JSON.parse(raw) as SiteConfig
    if (!data || typeof data !== 'object' || !data.root) return null
    return data
  } catch {
    return null
  }
}

export function useSiteConfig() {
  const config = useRuntimeConfig()
  const state = useState<SiteConfigState>('site-config', () => ({
    data: readStoredData(),
    loading: false,
    error: null,
    loadPromise: null,
  }))

  async function load(force = false): Promise<SiteConfig> {
    if (!force && state.value.data) return state.value.data
    if (state.value.loadPromise) return state.value.loadPromise

    state.value.loading = true
    state.value.error = null

    const promise = $fetch<SiteConfig>(`${config.public.apiBase}/v1/config`)
      .then((data) => {
        state.value.data = data
        state.value.loading = false
        state.value.loadPromise = null
        if (import.meta.client) {
          localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
        }
        return data
      })
      .catch((err) => {
        state.value.error = err
        state.value.loading = false
        state.value.loadPromise = null
        throw err
      })

    state.value.loadPromise = promise
    return promise
  }

  async function refresh(): Promise<SiteConfig> {
    return load(true)
  }

  function get<T = unknown>(key: keyof SiteConfig): T | undefined
  function get<T = unknown>(key: keyof SiteConfig, defaultValue: T): T
  function get<T = unknown>(
    key: keyof SiteConfig,
    defaultValue?: T,
  ): T | undefined {
    const value = state.value.data?.[key]
    return value !== undefined ? (value as T) : defaultValue
  }

  const data = computed(() => state.value.data)

  return {
    data: readonly(data),
    load,
    refresh,
    get,
  }
}
