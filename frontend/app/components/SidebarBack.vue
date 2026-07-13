<script setup lang="ts">
import { ArrowLeft } from '@lucide/vue'

interface PageNode {
  path: string
  title: string
  has_index?: boolean
  children?: PageNode[]
}

const props = defineProps<{
  docName: string
  currentPath: string
  pages: PageNode[]
}>()

const { t } = useText()

function findNode(nodes: PageNode[], target: string): PageNode | null {
  for (const node of nodes) {
    if (node.path === target) return node
    if (node.children) {
      const found = findNode(node.children, target)
      if (found) return found
    }
  }
  return null
}

const parentLink = computed(() => {
  if (!props.currentPath) return null
  const segments = props.currentPath.split('/')
  while (segments.length > 0) {
    segments.pop()
    const parentPath = segments.join('/')
    if (parentPath === '') return `/${props.docName}`
    const parentNode = findNode(props.pages, parentPath)
    if (parentNode?.has_index) return `/${props.docName}/${parentPath}`
  }
  return `/${props.docName}`
})
</script>

<template>
  <NuxtLink
    v-if="parentLink"
    :to="parentLink"
    class="flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
  >
    <ArrowLeft :size="16" />
    <span>{{ t('sidebar.back') }}</span>
  </NuxtLink>
</template>
