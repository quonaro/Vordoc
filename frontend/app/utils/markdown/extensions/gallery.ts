import type { Tokens } from 'marked'
import { escapeHtmlAttribute } from '~/utils/markdown/escape'

export interface GalleryToken extends Tokens.Generic {
  type: 'gallery'
  raw: string
  items: string[]
  resolvedItems?: string[]
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
      const items = (match[1] ?? '')
        .split(';')
        .map((s) => s.trim())
        .filter(Boolean)
      return {
        type: 'gallery',
        raw: match[0],
        items,
      } as GalleryToken
    },
    renderer(token: GalleryToken) {
      const images = (token.resolvedItems ?? token.items)
        .map((src) => `<img src="${escapeHtmlAttribute(src)}" alt="" />`)
        .join('')
      return `<div class="gallery">${images}</div>`
    },
  }
}
