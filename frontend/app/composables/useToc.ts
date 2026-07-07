import { computed, unref, type Ref } from 'vue'
import { marked, type Tokens } from 'marked'
import { extractPlainText, slugify } from '~/utils/markdown'

export interface TocItem {
  title: string
  link: string
  level: number
  children: TocItem[]
}

export function useToc(content: Ref<string> | string) {
  return computed(() => buildTocTree(unref(content)))
}

export function buildTocTree(content: string): TocItem[] {
  const tokens = marked.lexer(content)
  const headers = tokens
    .filter((token): token is Tokens.Heading => token.type === 'heading')
    .map((token) => {
      const title = extractPlainText(token.tokens)
      return {
        title,
        link: `#${slugify(title) || `heading-${token.depth}`}`,
        level: token.depth,
      }
    })

  return buildTree(headers)
}

export function findTocTitleByLink(
  items: TocItem[],
  link: string,
): string | undefined {
  for (const item of items) {
    if (item.link === link) return item.title
    const child = findTocTitleByLink(item.children, link)
    if (child) return child
  }
  return undefined
}

function buildTree(headers: Omit<TocItem, 'children'>[]): TocItem[] {
  const result: TocItem[] = []
  const stack: TocItem[] = []

  for (const header of headers) {
    if (header.level < 2 || header.level > 3) continue

    const node: TocItem = { ...header, children: [] }

    let parent = stack[stack.length - 1]
    while (parent && parent.level >= node.level) {
      stack.pop()
      parent = stack[stack.length - 1]
    }
    if (parent) {
      parent.children.push(node)
    } else {
      result.push(node)
    }

    stack.push(node)
  }

  return result
}
