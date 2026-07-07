const STORAGE_KEY = 'vordoc:site-config'

export interface HeaderConfig {
  enable: boolean
  title?: string
  logo?: string
}

export interface SiteConfig {
  enable_root_page: boolean
  header?: HeaderConfig
}

interface SiteConfigState {
  data: SiteConfig | null
}

function readStoredData(): SiteConfig | null {
  if (!import.meta.client) return null
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as SiteConfig
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
