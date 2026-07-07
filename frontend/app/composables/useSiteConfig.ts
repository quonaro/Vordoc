const STORAGE_KEY = 'vordoc:site-config:v2'

export interface LogoConfig {
  path?: string
  size?: number
}

export interface FontConfig {
  name?: string
  size?: number
}

export interface HeaderConfig {
  enable: boolean
  selector?: boolean
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
  }))

  async function load(): Promise<SiteConfig> {
    const data = await $fetch<SiteConfig>(`${config.public.apiBase}/v1/config`)
    state.value = { data }
    if (import.meta.client) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
    }
    return data
  }

  async function refresh(): Promise<SiteConfig> {
    return load()
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
