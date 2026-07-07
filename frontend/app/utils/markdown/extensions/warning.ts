import { marked, type Token, type Tokens } from 'marked'
import { escapeHtmlAttribute } from '~/utils/markdown/escape'

export interface WarningToken extends Tokens.Generic {
  type: 'warning'
  raw: string
  title: string
  bodyTokens: Token[]
}

export function createWarningExtension() {
  return {
    name: 'warning',
    level: 'block' as const,
    start(src: string) {
      return src.match(/Warning\[/)?.index
    },
    tokenizer(src: string) {
      const match = /^Warning\[([^\]]+)\]{([^}]*)}(?:\n|$)/.exec(src)
      if (!match) return undefined
      const title = (match[1] ?? '').trim()
      const bodyText = (match[2] ?? '').trim()
      return {
        type: 'warning',
        raw: match[0],
        title,
        bodyTokens: marked.Lexer.lexInline(bodyText),
      } as WarningToken
    },
    renderer(token: WarningToken) {
      const body = marked.Parser.parseInline(token.bodyTokens, {
        async: false,
      }) as string
      return `<div class="callout warning"><div class="callout-title">${escapeHtmlAttribute(token.title)}</div><div class="callout-body">${body}</div></div>`
    },
    childTokens: ['bodyTokens'],
  }
}
