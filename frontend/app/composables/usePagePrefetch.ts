import { cacheApiResponse } from '~/utils/apiCache'

interface PageNode {
  path: string
  has_index?: boolean
  children?: PageNode[]
}

function isPageNode(node: PageNode): boolean {
  return node.has_index === true || (node.children?.length ?? 0) === 0
}

function collectPageUrls(nodes: PageNode[], baseUrl: string): string[] {
  const urls: string[] = []

  for (const node of nodes) {
    if (isPageNode(node)) {
      urls.push(`${baseUrl}${node.path}`)
    }

    if (node.children?.length) {
      urls.push(...collectPageUrls(node.children, baseUrl))
    }
  }

  return urls
}

async function prefetchPageUrl(url: string): Promise<void> {
  try {
    const data = await $fetch(url, { credentials: 'include' })
    await cacheApiResponse(url, data)
  } catch {
    // Password-protected pages or missing pages are skipped silently.
  }
}

/**
 * Warm the API cache for every page of the current documentation.
 * This makes the whole doc available offline after it has been opened once
 * while the user is online.
 */
export async function prefetchDocPages(
  docName: string,
  pages: PageNode[],
): Promise<void> {
  if (!import.meta.client || !navigator.onLine) {
    return
  }

  const config = useRuntimeConfig()
  const baseUrl = `${config.public.apiBase}/v1/${docName}/`

  const pageUrls = collectPageUrls(pages, baseUrl)
  pageUrls.push(baseUrl)

  const uniqueUrls = [...new Set(pageUrls)]
  let index = 0
  const concurrency = 4

  async function worker(): Promise<void> {
    while (index < uniqueUrls.length) {
      const url = uniqueUrls[index++]
      if (!url) continue
      await prefetchPageUrl(url)
    }
  }

  await Promise.all(Array.from({ length: concurrency }, worker))
}
