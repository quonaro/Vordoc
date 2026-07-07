<script setup lang="ts">
import type { TocItem } from '~/composables/useToc'

const props = defineProps<{
  items: TocItem[]
  activeLink: string | null
  root?: boolean
}>()

function isActive(link: string): boolean {
  return props.activeLink === link
}

function indentClass(level: number): string {
  const sizes: Record<number, string> = {
    2: 'pl-0',
    3: 'pl-3',
    4: 'pl-6',
    5: 'pl-9',
    6: 'pl-12',
  }
  return sizes[level] ?? 'pl-0'
}
</script>

<template>
  <ul class="space-y-1" :class="root ? '' : 'border-l border-border'">
    <li v-for="item in items" :key="item.link">
      <a
        :href="item.link"
        :title="item.title"
        class="block rounded-md py-1 text-sm transition-colors"
        :class="[
          indentClass(item.level),
          isActive(item.link)
            ? 'font-medium text-primary'
            : 'text-muted-foreground hover:text-foreground',
        ]"
      >
        {{ item.title }}
      </a>
      <TableOfContents
        v-if="item.children.length"
        :items="item.children"
        :active-link="activeLink"
      />
    </li>
  </ul>
</template>
