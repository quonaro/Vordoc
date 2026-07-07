const TEN_MINUTES_MS = 10 * 60 * 1000

interface RuntimePublicState {
  data: Record<string, unknown> | null
  fetchedAt: number | null
}

export function useRuntimePublic() {
  const config = useRuntimeConfig()
  const state = useState<RuntimePublicState>('runtime-public', () => ({
    data: null,
    fetchedAt: null,
  }))

  async function load(): Promise<Record<string, unknown>> {
    const now = Date.now()
    const cached = state.value.data
    const fetchedAt = state.value.fetchedAt
    if (cached && fetchedAt && now - fetchedAt < TEN_MINUTES_MS) {
      return cached
    }

    const data = await $fetch<Record<string, unknown>>(
      `${config.public.apiBase}/config/public`,
    )
    state.value = { data, fetchedAt: now }
    return data
  }

  function get<T = unknown>(key: string): T | undefined
  function get<T = unknown>(key: string, defaultValue: T): T
  function get<T = unknown>(key: string, defaultValue?: T): T | undefined {
    const value = state.value.data?.[key]
    return value !== undefined ? (value as T) : defaultValue
  }

  const data = computed(() => state.value.data)

  return {
    data: readonly(data),
    load,
    get,
  }
}
