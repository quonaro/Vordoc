import { marked, type Token, type Tokens } from 'marked'
import { markedHighlight } from 'marked-highlight'
import hljs from 'highlight.js'
import { sanitize } from 'isomorphic-dompurify'

interface HrefParts {
  path: string
  query: string
  fragment: string
}

function splitHref(href: string): HrefParts {
  let path = href
  let query = ''
  let fragment = ''

  const hashIndex = path.indexOf('#')
  if (hashIndex !== -1) {
    fragment = path.slice(hashIndex)
    path = path.slice(0, hashIndex)
  }

  const queryIndex = path.indexOf('?')
  if (queryIndex !== -1) {
    query = path.slice(queryIndex)
    path = path.slice(0, queryIndex)
  }

  return { path, query, fragment }
}

function shouldTransform(href: string): boolean {
  if (!href || href.startsWith('#')) return false
  if (/^[a-z][a-z0-9+.-]*:/i.test(href)) return false

  const { path } = splitHref(href)
  return path.toLowerCase().endsWith('.md')
}

function baseDirFromFilePath(docName: string, filePath: string): string {
  const slashIndex = filePath.lastIndexOf('/')
  const dir = slashIndex === -1 ? '' : filePath.slice(0, slashIndex)
  return dir ? `${docName}/${dir}` : docName
}

function resolveRelativePath(baseDir: string, relativePath: string): string {
  const base = `http://localhost/${baseDir}/`
  const url = new URL(relativePath, base)
  return url.pathname.replace(/^\//, '')
}

function normalizeIndexPath(path: string): string {
  let normalized = path.replace(/\/+$/, '')
  if (normalized.endsWith('/index')) {
    normalized = normalized.slice(0, -6)
  } else if (normalized === 'index') {
    normalized = ''
  }
  return normalized
}

export function resolveMarkdownLink(
  href: string,
  docName: string,
  filePath: string,
): string | null {
  if (!shouldTransform(href)) return null

  const { path, query, fragment } = splitHref(href)
  const pathWithoutMd = path.slice(0, -3)

  let resolvedPath: string
  if (href.startsWith('/')) {
    resolvedPath = normalizeIndexPath(pathWithoutMd)
  } else {
    const baseDir = baseDirFromFilePath(docName, filePath)
    resolvedPath = resolveRelativePath(baseDir, pathWithoutMd)
    resolvedPath = normalizeIndexPath(resolvedPath)
  }

  if (!resolvedPath.startsWith('/')) {
    resolvedPath = `/${resolvedPath}`
  }

  return `${resolvedPath}${query}${fragment}`
}

function shouldTransformImage(src: string): boolean {
  if (!src || src.startsWith('#')) return false
  if (/^[a-z][a-z0-9+.-]*:/i.test(src)) return false
  if (src.startsWith('/api/') || src.startsWith('/assets/')) return false
  return true
}

function normalizeAssetPath(path: string): string {
  return path.replace(/^\/+/, '').replace(/\/+$/, '')
}

function assetBaseDirFromFilePath(filePath: string): string {
  const slashIndex = filePath.lastIndexOf('/')
  return slashIndex === -1 ? '' : filePath.slice(0, slashIndex)
}

export function resolveMarkdownImage(
  src: string,
  docName: string,
  filePath: string,
): string | null {
  if (!shouldTransformImage(src)) return null

  let resolvedPath: string
  if (src.startsWith('/')) {
    resolvedPath = normalizeAssetPath(src)
  } else {
    const baseDir = assetBaseDirFromFilePath(filePath)
    const base = baseDir ? `http://localhost/${baseDir}/` : 'http://localhost/'
    const url = new URL(src, base)
    resolvedPath = normalizeAssetPath(url.pathname)
  }

  return `/api/v1/assets/${docName}/${resolvedPath}`
}

function slugify(text: string): string {
  return text
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^[^a-z]+/, '')
    .slice(0, 64)
}

function extractPlainText(tokens: Token[]): string {
  let text = ''
  for (const token of tokens) {
    switch (token.type) {
      case 'text':
      case 'codespan':
        text += (token as Tokens.Text | Tokens.Codespan).text
        break
      case 'strong':
      case 'em':
      case 'del':
        text += extractPlainText(
          (token as Tokens.Strong | Tokens.Em | Tokens.Del).tokens,
        )
        break
      case 'link':
        text += extractPlainText((token as Tokens.Link).tokens)
        break
      case 'html':
        text += (token as Tokens.HTML).text
        break
    }
  }
  return text
}

function resolveMarkdownToken(
  token: Token,
  docName: string,
  filePath: string,
): void {
  if (token.type === 'link') {
    const link = token as Tokens.Link
    const resolved = resolveMarkdownLink(link.href, docName, filePath)
    if (resolved) {
      link.href = resolved
    }
    return
  }

  if (token.type === 'image') {
    const image = token as Tokens.Image
    const resolved = resolveMarkdownImage(image.href, docName, filePath)
    if (resolved) {
      image.href = resolved
    }
  }
}

marked.use(
  markedHighlight({
    emptyLangClass: 'hljs',
    langPrefix: 'hljs language-',
    highlight(code, lang) {
      const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
      return hljs.highlight(code, { language }).value
    },
  }),
  {
    renderer: {
      heading(token: Tokens.Heading) {
        const html = marked.Parser.parseInline(token.tokens, {
          async: false,
        }) as string
        const id = slugify(extractPlainText(token.tokens))
        return `<h${token.depth} id="${id}">${html}</h${token.depth}>`
      },
    },
  },
)

export function renderMarkdown(
  content: string,
  docName: string,
  filePath: string,
): string {
  const raw = marked.parse(content, {
    async: false,
    walkTokens: (token: Token) => resolveMarkdownToken(token, docName, filePath),
  }) as string

  return sanitize(raw, {
    ALLOWED_TAGS: [
      'h1',
      'h2',
      'h3',
      'h4',
      'h5',
      'h6',
      'p',
      'br',
      'hr',
      'a',
      'strong',
      'em',
      'del',
      's',
      'code',
      'pre',
      'blockquote',
      'ul',
      'ol',
      'li',
      'table',
      'thead',
      'tbody',
      'tr',
      'th',
      'td',
      'img',
      'div',
      'span',
      'sup',
      'sub',
    ],
    ALLOWED_ATTR: [
      'href',
      'title',
      'target',
      'rel',
      'id',
      'class',
      'src',
      'alt',
      'width',
      'height',
    ],
  }) as string
}
