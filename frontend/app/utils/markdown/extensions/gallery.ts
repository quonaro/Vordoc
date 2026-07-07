import type { Tokens } from 'marked'
import { escapeHtmlAttribute } from '~/utils/markdown/escape'

export interface GalleryToken extends Tokens.Generic {
  type: 'gallery'
  raw: string
  items: string[]
  resolvedItems?: string[]
  height?: string
}

const lengthPattern = /^\d+(?:\.\d+)?(?:px|rem|em|%|vh|vw|cm|mm|in|pt|pc|ex|ch)$/i
const plainNumberPattern = /^\d+$/
const heightArgPattern = /^height[:=](.+)$/i

function parseHeightArg(arg: string): string | null {
  const value = arg.trim()
  const match = heightArgPattern.exec(value)
  if (match) {
    const height = (match[1] ?? '').trim()
    if (lengthPattern.test(height) || plainNumberPattern.test(height)) {
      return plainNumberPattern.test(height) ? `${height}px` : height
    }
    return null
  }

  if (lengthPattern.test(value) || plainNumberPattern.test(value)) {
    return plainNumberPattern.test(value) ? `${value}px` : value
  }

  return null
}

export function createGalleryExtension() {
  return {
    name: 'gallery',
    level: 'block' as const,
    start(src: string) {
      return src.match(/Gallery\[/)?.index
    },
    tokenizer(src: string) {
      const match = /^Gallery\[([^\]]+)\](?:\n|$)/.exec(src)
      if (!match) return undefined
      const rawArgs = (match[1] ?? '')
        .split(';')
        .map((s) => s.trim())
        .filter(Boolean)

      let height: string | undefined
      const firstHeight = parseHeightArg(rawArgs[0] ?? '')
      if (firstHeight) {
        height = firstHeight
        rawArgs.shift()
      }

      const items = rawArgs
      return {
        type: 'gallery',
        raw: match[0],
        items,
        height,
      } as GalleryToken
    },
    renderer(token: GalleryToken) {
      const images = (token.resolvedItems ?? token.items)
        .map((src) => `<img src="${escapeHtmlAttribute(src)}" alt="" />`)
        .join('')
      const style = token.height
        ? ` style="height: ${escapeHtmlAttribute(token.height)}"`
        : ''
      return `<div class="gallery"${style}>${images}</div>`
    },
  }
}
