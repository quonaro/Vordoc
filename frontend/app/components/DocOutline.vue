<script setup lang="ts">
import type { TocItem } from '~/composables/useToc'

defineProps<{
  items: TocItem[]
  activeLink: string | null
}>()

const { t } = useText()
</script>

<template>
  <aside class="hidden w-56 shrink-0 xl:block">
    <nav v-if="items.length" class="sticky top-8">
      <div class="rounded-lg border bg-card p-3 shadow-sm md:p-4">
        <p class="mb-3 text-sm font-semibold">
          {{ t('doc.onThisPage') }}
        </p>
        <TableOfContents :items="items" :active-link="activeLink" root />
      </div>
    </nav>
  </aside>
</template>
