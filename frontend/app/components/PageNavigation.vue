<script setup lang="ts">
import { ChevronLeft, ChevronRight } from '@lucide/vue'

interface PageNode {
  path: string
  title: string
  access?: string
  has_index?: boolean
  children?: PageNode[]
}

const props = defineProps<{
  docName: string
  docTitle: string
  pages: PageNode[]
  currentPath: string
  loading?: boolean
}>()

const { t } = useText()

const linkClass = (align: 'start' | 'end'): string => {
  const base =
    'group flex max-w-[calc(50%-0.5rem)] flex-col gap-1 rounded-lg border border-border bg-background p-4 transition-colors hover:border-primary'
  const alignClass = align === 'start' ? 'items-start' : 'items-end text-right'
  const loadingClass = props.loading ? 'cursor-wait opacity-60' : ''
  return `${base} ${alignClass} ${loadingClass}`
}

const neighbors = usePageNeighbors(
  computed(() => props.pages),
  computed(() => props.docTitle),
  computed(() => props.currentPath),
)
</script>

<template>
  <nav
    v-if="neighbors.prev || neighbors.next"
    class="mt-12 flex justify-between gap-4 border-t border-border pt-8"
    aria-label="page"
    :aria-busy="loading"
  >
    <NuxtLink
      v-if="neighbors.prev"
      :to="`/${docName}/${neighbors.prev.path}`"
      prefetch
      :class="linkClass('start')"
      :aria-disabled="loading"
    >
      <span class="flex items-center gap-1 text-xs text-muted-foreground">
        <ChevronLeft class="h-3.5 w-3.5" />
        {{ t('doc.previousPage') }}
      </span>
      <span class="font-medium text-primary">
        {{ neighbors.prev.title }}
      </span>
    </NuxtLink>
    <div v-else class="flex-1" />

    <NuxtLink
      v-if="neighbors.next"
      :to="`/${docName}/${neighbors.next.path}`"
      prefetch
      :class="linkClass('end')"
      :aria-disabled="loading"
    >
      <span class="flex items-center gap-1 text-xs text-muted-foreground">
        {{ t('doc.nextPage') }}
        <ChevronRight class="h-3.5 w-3.5" />
      </span>
      <span class="font-medium text-primary">
        {{ neighbors.next.title }}
      </span>
    </NuxtLink>
  </nav>
</template>
