import { computed, unref } from 'vue'

interface PageNode {
  path: string
  title: string
  access?: string
  has_index?: boolean
  children?: PageNode[]
}

export interface NavPage {
  path: string
  title: string
}

export interface PageNeighbors {
  prev?: NavPage
  next?: NavPage
}

export function usePageNeighbors(
  pages: Ref<PageNode[] | undefined> | PageNode[],
  docTitle: Ref<string | undefined> | string | undefined,
  currentPath: Ref<string> | string,
) {
  return computed<PageNeighbors>(() => {
    const all = flattenPages(unref(pages) ?? [], unref(docTitle) ?? '')
    const path = unref(currentPath)
    const idx = all.findIndex((page) => page.path === path)
    return {
      prev: idx > 0 ? all[idx - 1] : undefined,
      next:
        idx >= 0 && idx < all.length - 1
          ? all[idx + 1]
          : undefined,
    }
  })
}

function flattenPages(
  nodes: PageNode[],
  rootTitle: string,
  parentPath: string = '',
): NavPage[] {
  const result: NavPage[] = []

  if (parentPath === '') {
    result.push({ path: '', title: rootTitle })
  }

  for (const node of nodes) {
    const hasChildren = (node.children?.length ?? 0) > 0
    const hasIndex = node.has_index === true

    if (hasIndex || !hasChildren) {
      result.push({ path: node.path, title: node.title })
    }

    if (hasChildren) {
      result.push(...flattenPages(node.children ?? [], node.title, node.path))
    }
  }

  return result
}
