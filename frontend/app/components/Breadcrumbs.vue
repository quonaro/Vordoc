<script setup lang="ts">
interface PageNode {
  path: string
  title: string
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
  acc: { path: string; title: string }[] = [],
): { path: string; title: string }[] | null {
  for (const node of nodes) {
    if (node.path === target) {
      return [...acc, { path: node.path, title: node.title }]
    }
    if (node.children) {
      const result = buildCrumbs(node.children, target, [
        ...acc,
        { path: node.path, title: node.title },
      ])
      if (result) return result
    }
  }
  return null
}

const crumbs = computed(() => {
  const out: { to: string; title: string }[] = [
    { to: `/${props.docName}`, title: props.docTitle || props.docName },
  ]
  if (!props.currentPath) return out
  const found = buildCrumbs(props.pages, props.currentPath)
  if (found) {
    for (const item of found) {
      out.push({ to: `/${props.docName}/${item.path}`, title: item.title })
    }
  }
  return out
})
</script>

<template>
  <nav class="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
    <template v-for="(crumb, i) in crumbs" :key="crumb.to">
      <span v-if="i > 0">/</span>
      <NuxtLink
        :to="crumb.to"
        class="transition-colors hover:text-foreground"
        :class="{ 'text-foreground': i === crumbs.length - 1 }"
      >
        {{ crumb.title }}
      </NuxtLink>
    </template>
  </nav>
</template>
