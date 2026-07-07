export interface ResolvedFont {
  family: string
  url?: string
  isCustom: boolean
}

export function resolveFont(name: string): ResolvedFont {
  const trimmed = name.trim()
  const lower = trimmed.toLowerCase()
  if (lower.endsWith('.ttf') || lower.endsWith('.otf')) {
    const url = trimmed.startsWith('/') ? trimmed : `/fonts/${trimmed}`
    const base = trimmed.split('/').pop() || trimmed
    const family = base.replace(/\.(ttf|otf)$/i, '')
    return { family, url, isCustom: true }
  }
  return { family: trimmed, isCustom: false }
}
