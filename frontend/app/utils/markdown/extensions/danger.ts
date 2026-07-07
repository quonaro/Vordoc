import { marked, type Token, type Tokens } from 'marked'
import { escapeHtmlAttribute } from '~/utils/markdown/escape'

export interface DangerToken extends Tokens.Generic {
  type: 'danger'
  raw: string
  title: string
  bodyTokens: Token[]
}

export function createDangerExtension() {
  return {
    name: 'danger',
    level: 'block' as const,
    start(src: string) {
      return src.match(/Danger\[/)?.index
    },
    tokenizer(src: string) {
      const match = /^Danger\[([^\]]+)\]{([^}]*)}(?:\n|$)/.exec(src)
      if (!match) return undefined
      const title = (match[1] ?? '').trim()
      const bodyText = (match[2] ?? '').trim()
      return {
        type: 'danger',
        raw: match[0],
        title,
        bodyTokens: marked.Lexer.lexInline(bodyText),
      } as DangerToken
    },
    renderer(token: DangerToken) {
      const body = marked.Parser.parseInline(token.bodyTokens, {
        async: false,
      }) as string
      return `<div class="callout danger"><div class="callout-title">${escapeHtmlAttribute(token.title)}</div><div class="callout-body">${body}</div></div>`
    },
    childTokens: ['bodyTokens'],
  }
}
