<script setup lang="ts">
import { Home, ChevronRight } from '@lucide/vue'

interface PageNode {
  path: string
  title: string
  has_index?: boolean
  children?: PageNode[]
}

const props = defineProps<{
  docTitle: string
  docName: string
  pages: PageNode[]
  currentPath: string
}>()

function buildCrumbs(
  nodes: PageNode[],
  target: string,
  acc: { path: string; title: string; has_index?: boolean }[] = [],
): { path: string; title: string; has_index?: boolean }[] | null {
  for (const node of nodes) {
    if (node.path === target) {
      return [
        ...acc,
        { path: node.path, title: node.title, has_index: node.has_index },
      ]
    }
    if (node.children) {
      const result = buildCrumbs(node.children, target, [
        ...acc,
        { path: node.path, title: node.title, has_index: node.has_index },
      ])
      if (result) return result
    }
  }
  return null
}

const crumbs = computed(() => {
  const out: { to: string; title: string; has_index?: boolean }[] = []
  if (!props.currentPath) return out
  const found = buildCrumbs(props.pages, props.currentPath)
  if (found) {
    for (const item of found) {
      out.push({
        to: `/${props.docName}/${item.path}`,
        title: item.title,
        has_index: item.has_index,
      })
    }
  }
  return out
})
</script>

<template>
  <nav
    class="mb-6 inline-flex items-center gap-1 rounded-lg border bg-card px-4 py-2 text-sm shadow-sm"
  >
    <NuxtLink
      :to="`/${docName}`"
      class="flex items-center gap-1 text-muted-foreground transition-colors hover:text-foreground"
      :class="{ 'font-medium text-primary': !crumbs.length }"
    >
      <Home :size="16" />
      <span>{{ docTitle }}</span>
    </NuxtLink>

    <template v-for="(crumb, i) in crumbs" :key="crumb.to">
      <ChevronRight :size="14" class="text-muted-foreground/50" />
      <NuxtLink
        v-if="i < crumbs.length - 1 && crumb.has_index"
        :to="crumb.to"
        class="text-muted-foreground transition-colors hover:text-foreground"
      >
        {{ crumb.title }}
      </NuxtLink>
      <span v-else-if="i < crumbs.length - 1" class="text-muted-foreground">
        {{ crumb.title }}
      </span>
      <span v-else class="font-medium text-primary">
        {{ crumb.title }}
      </span>
    </template>
  </nav>
</template>
