import { computed, unref } from 'vue'

export interface PageNode {
  path: string
  title: string
  access?: string
  access_scope?: string
  has_index?: boolean
  show?: boolean
  children?: PageNode[]
}

export function useSidebarNodes(
  pages: Ref<PageNode[] | undefined> | PageNode[],
  currentPath: Ref<string> | string,
) {
  return computed(() => {
    const nodes = unref(pages) ?? []
    const path = unref(currentPath)
    return buildSidebarNodes(nodes, path)
  })
}

export function buildSidebarNodes(
  pages: PageNode[],
  currentPath: string,
): PageNode[] {
  const directoryNodes = getDirectoryNodes(pages, currentPath)
  return trimDepth(directoryNodes, 2)
}

function trimDepth(nodes: PageNode[], depth: number): PageNode[] {
  if (depth <= 0) return []

  return nodes.map((node) => {
    const copy: PageNode = { ...node }
    if (copy.children) {
      copy.children = trimDepth(copy.children, depth - 1)
    }
    return copy
  })
}

function getDirectoryNodes(pages: PageNode[], currentPath: string): PageNode[] {
  if (!currentPath) {
    return pages
  }

  const current = findNode(pages, currentPath)
  if (!current) {
    return pages
  }

  if (isDirectory(current)) {
    return current.children ?? []
  }

  const parentPath = currentPath.split('/').slice(0, -1).join('/')
  if (!parentPath) {
    return pages
  }

  const parent = findNode(pages, parentPath)
  return parent?.children ?? []
}

function isDirectory(node: PageNode): boolean {
  return node.has_index === true || (node.children?.length ?? 0) > 0
}

function findNode(
  nodes: PageNode[] | undefined,
  path: string,
): PageNode | undefined {
  if (!nodes) return undefined
  for (const node of nodes) {
    if (node.path === path) return node
    if (node.children) {
      const found = findNode(node.children, path)
      if (found) return found
    }
  }
  return undefined
}
