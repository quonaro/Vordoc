import type { Tokens } from 'marked'
import { escapeHtmlAttribute } from '~/utils/markdown/escape'

export interface ImageToken extends Tokens.Generic {
  type: 'customImage'
  raw: string
  src: string
  alt: string
  background: boolean
  resolvedSrc?: string
}

const backgroundFlagPattern = /^(background|bg)(?:=(true|1|on|yes))?$/i

function parseBackgroundFlag(arg: string): boolean {
  return backgroundFlagPattern.test(arg.trim())
}

export function createImageExtension() {
  return {
    name: 'customImage',
    level: 'block' as const,
    start(src: string) {
      return src.match(/Image\[/)?.index
    },
    tokenizer(source: string) {
      const match = /^Image\[([^\]]+)\](?:\{([^}]*)\})?(?:\n|$)/.exec(source)
      if (!match) return undefined

      const rawArgs = (match[1] ?? '')
        .split(';')
        .map((s) => s.trim())
        .filter(Boolean)

      const src = rawArgs[0] ?? ''
      if (!src) return undefined

      const background = rawArgs.slice(1).some(parseBackgroundFlag)
      const alt = (match[2] ?? '').trim()

      return {
        type: 'customImage',
        raw: match[0],
        src,
        alt,
        background,
      } as ImageToken
    },
    renderer(token: ImageToken) {
      const src = escapeHtmlAttribute(token.resolvedSrc ?? token.src)
      const alt = escapeHtmlAttribute(token.alt)
      const bgClass = token.background ? ' markdown-image--background' : ''
      return `<img class="markdown-image${bgClass}" src="${src}" alt="${alt}" />`
    },
  }
}
