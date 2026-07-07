export interface HSL {
  h: number
  s: number
  l: number
}

export function hexToHsl(hex: string): HSL {
  let sanitized = hex.trim().toLowerCase()
  if (sanitized.startsWith('#')) {
    sanitized = sanitized.slice(1)
  }
  if (!/^[0-9a-f]{3}$|^[0-9a-f]{6}$/.test(sanitized)) {
    throw new Error(`invalid hex color: ${hex}`)
  }
  if (sanitized.length === 3) {
    sanitized = sanitized
      .split('')
      .map((c) => c + c)
      .join('')
  }

  const r = Number.parseInt(sanitized.slice(0, 2), 16) / 255
  const g = Number.parseInt(sanitized.slice(2, 4), 16) / 255
  const b = Number.parseInt(sanitized.slice(4, 6), 16) / 255

  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  const l = (max + min) / 2

  if (max === min) {
    return { h: 0, s: 0, l: Math.round(l * 100) }
  }

  const d = max - min
  const s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
  let h = 0
  switch (max) {
    case r:
      h = (g - b) / d + (g < b ? 6 : 0)
      break
    case g:
      h = (b - r) / d + 2
      break
    case b:
      h = (r - g) / d + 4
      break
  }
  h /= 6

  return {
    h: Math.round(h * 360),
    s: Math.round(s * 100),
    l: Math.round(l * 100),
  }
}

export function hslCssValue({ h, s, l }: HSL): string {
  return `${h} ${s}% ${l}%`
}

export function foregroundForHsl({ l }: HSL): string {
  return l > 60 ? '0 0% 0%' : '0 0% 100%'
}
