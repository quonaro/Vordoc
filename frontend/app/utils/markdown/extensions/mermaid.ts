import type { Token, Tokens } from 'marked'

export interface MermaidToken extends Tokens.Generic {
  type: 'mermaid'
  raw: string
  text: string
}

export function createMermaidExtension() {
  return {
    name: 'mermaid',
    level: 'block' as const,
    start(src: string) {
      return src.match(/```\s*mermaid/)?.index
    },
    tokenizer(src: string) {
      const match = /^(```)\s*mermaid\s*\n([\s\S]*?)\n\1\s*(?:\n|$)/.exec(src)
      if (!match) return undefined
      return {
        type: 'mermaid',
        raw: match[0],
        text: (match[2] ?? '').trim(),
      } as MermaidToken
    },
    renderer(token: Token) {
      const text = (token as MermaidToken).text
      return `<pre class="mermaid">${text}</pre>\n`
    },
  }
}
