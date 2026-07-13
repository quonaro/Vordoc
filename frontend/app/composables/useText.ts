export type UIText = Record<string, unknown>

const STORAGE_KEY = 'vordoc:ui-text:v1'

function readStoredText(): UIText | null {
  if (!import.meta.client) return null
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as UIText
  } catch {
    return null
  }
}

function getByPath(obj: UIText | undefined, path: string): string | undefined {
  if (!obj || typeof obj !== 'object') return undefined

  const parts = path.split('.')
  let current: unknown = obj
  for (const part of parts) {
    if (current == null || typeof current !== 'object') return undefined
    current = (current as Record<string, unknown>)[part]
  }

  if (typeof current === 'string') return current
  return undefined
}

export function useText() {
  const config = useRuntimeConfig()
  const state = useState<UIText | null>('ui-text', () => readStoredText())

  async function load(): Promise<UIText | null> {
    try {
      const data = await $fetch<UIText>(`${config.public.apiBase}/v1/text`)
      state.value = data
      if (import.meta.client) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
      }
      return data
    } catch (err) {
      console.error('failed to load UI text', err)
      return state.value
    }
  }

  function t(path: string): string {
    return getByPath(state.value ?? {}, path) ?? path
  }

  return {
    text: computed(() => state.value ?? {}),
    t,
    load,
  }
}
