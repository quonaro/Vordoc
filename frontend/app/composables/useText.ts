export type UIText = Record<string, unknown>

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
  const state = useState<UIText | null>('ui-text', () => null)

  async function load(): Promise<UIText> {
    if (state.value) return state.value
    const data = await $fetch<UIText>(`${config.public.apiBase}/v1/text`)
    state.value = data
    return data
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
