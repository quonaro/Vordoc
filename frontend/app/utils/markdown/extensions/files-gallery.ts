import type { Tokens } from 'marked'
import { escapeHtmlAttribute } from '~/utils/markdown/escape'
import { getFileCategory, getFileIconSvg } from './file-icons'

export interface FilesGalleryItem {
  src: string
  title: string
  format: string
  resolvedSrc?: string
  category: string
}

export interface FilesGalleryToken extends Tokens.Generic {
  type: 'filesGallery'
  raw: string
  files: FilesGalleryItem[]
}

const externalHrefPattern = /^[a-z][a-z0-9+.-]*:/i

function isExternalHref(href: string): boolean {
  return externalHrefPattern.test(href)
}

function extractExtension(src: string): string {
  try {
    const url = new URL(src, 'http://localhost')
    const path = url.pathname
    const dot = path.lastIndexOf('.')
    return dot === -1 ? '' : path.slice(dot + 1).toLowerCase()
  } catch {
    return ''
  }
}

function basenameFromSrc(src: string): string {
  try {
    const url = new URL(src, 'http://localhost')
    return url.pathname.split('/').pop() ?? src
  } catch {
    return src
  }
}

function parseFileArg(arg: string): FilesGalleryItem {
  const separatorIndex = arg.indexOf('|')
  const src = (separatorIndex === -1 ? arg : arg.slice(0, separatorIndex)).trim()
  const title = (separatorIndex === -1 ? '' : arg.slice(separatorIndex + 1)).trim()
  const extension = extractExtension(src)
  return {
    src,
    title: title || basenameFromSrc(src),
    format: extension.toUpperCase() || (isExternalHref(src) ? 'URL' : 'FILE'),
    category: getFileCategory(extension),
  }
}

export function createFilesGalleryExtension() {
  return {
    name: 'filesGallery',
    level: 'block' as const,
    start(src: string) {
      return src.match(/FilesGallery\[/)?.index
    },
    tokenizer(src: string) {
      const match = /^FilesGallery\[([^\]]+)\](?:\n|$)/.exec(src)
      if (!match) return undefined
      const rawArgs = (match[1] ?? '')
        .split(';')
        .map((s) => s.trim())
        .filter(Boolean)
      return {
        type: 'filesGallery',
        raw: match[0],
        files: rawArgs.map(parseFileArg),
      } as FilesGalleryToken
    },
    renderer(token: FilesGalleryToken) {
      const items = token.files
        .map((file) => {
          const href = escapeHtmlAttribute(file.resolvedSrc ?? file.src)
          const isExternal = isExternalHref(file.src)
          const externalAttrs = isExternal
            ? ' target="_blank" rel="noopener noreferrer"'
            : ''
          const downloadAttr = isExternal ? '' : ' download'
          const icon = getFileIconSvg(file.category)
          const formatLabel = file.format
            ? `<span class="files-gallery__format">${escapeHtmlAttribute(file.format)}</span>`
            : ''
          return `<a class="files-gallery__item group files-gallery__item--${escapeHtmlAttribute(file.category)}" href="${href}" title="${escapeHtmlAttribute(file.title)}"${downloadAttr}${externalAttrs}>${formatLabel}<span class="files-gallery__icon">${icon}</span><span class="files-gallery__title">${escapeHtmlAttribute(file.title)}</span></a>`
        })
        .join('')
      return `<div class="files-gallery">${items}</div>`
    },
  }
}
