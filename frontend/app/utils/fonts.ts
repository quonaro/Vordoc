export interface ResolvedFont {
  family: string
  url?: string
  isCustom: boolean
  googleFontsUrl?: string
}

const systemFonts = new Set([
  'arial',
  'helvetica',
  'times new roman',
  'times',
  'courier new',
  'courier',
  'verdana',
  'georgia',
  'palatino',
  'garamond',
  'bookman',
  'comic sans ms',
  'trebuchet ms',
  'impact',
  'system-ui',
  'ui-serif',
  'ui-sans-serif',
  'ui-monospace',
  'ui-rounded',
  'sans-serif',
  'serif',
  'monospace',
  'cursive',
  'fantasy',
  'inherit',
  'initial',
  'revert',
  'unset',
])

export function resolveFont(name: string): ResolvedFont {
  const trimmed = name.trim()
  const lower = trimmed.toLowerCase()
  if (lower.endsWith('.ttf') || lower.endsWith('.otf')) {
    const url = trimmed.startsWith('/') ? trimmed : `/fonts/${trimmed}`
    const base = trimmed.split('/').pop() || trimmed
    const family = base.replace(/\.(ttf|otf)$/i, '')
    return { family, url, isCustom: true }
  }
  if (systemFonts.has(lower)) {
    return { family: trimmed, isCustom: false }
  }
  const googleFamily = trimmed.replace(/\s+/g, '+')
  const googleFontsUrl = `https://fonts.googleapis.com/css2?family=${googleFamily}:wght@400;700&display=swap`
  return { family: trimmed, isCustom: false, googleFontsUrl }
}
