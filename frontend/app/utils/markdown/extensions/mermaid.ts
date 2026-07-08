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
      const match = /^(```)\s*mermaid\s*\n(?:(?!\1\s*$)[\s\S])*\1\s*$/m.exec(
        src,
      )
      if (!match) return undefined
      const code = match[0]
        .replace(/^```\s*mermaid\s*\n/, '')
        .replace(/```\s*$/, '')
      return {
        type: 'mermaid',
        raw: match[0],
        text: code.trim(),
      } as MermaidToken
    },
    renderer(token: Token) {
      const text = (token as MermaidToken).text
      return `<pre class="mermaid">${text}</pre>\n`
    },
  }
}
