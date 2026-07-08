<script setup lang="ts">
import { Search } from '@lucide/vue'

const { t } = useText()
const globalSearchOpen = useGlobalSearchState()

const shortcut = computed(() => {
  if (!import.meta.client) return 'Ctrl K'
  return navigator.platform.toLowerCase().includes('mac') ? '⌘ K' : 'Ctrl K'
})

function open() {
  globalSearchOpen.value = true
}
</script>

<template>
  <button
    type="button"
    class="group flex w-full max-w-md items-center justify-between gap-2 rounded-md border border-input bg-background px-3 py-2 text-sm text-muted-foreground shadow-sm transition-all hover:border-accent hover:bg-accent/5 hover:text-foreground"
    @click="open"
  >
    <span class="flex items-center gap-2">
      <Search class="h-4 w-4 shrink-0" />
      <span class="truncate">{{ t('search.globalPlaceholder') }}</span>
    </span>
    <kbd
      class="hidden rounded border bg-muted px-1.5 py-0.5 text-xs font-medium transition-colors group-hover:border-accent/50 group-hover:text-foreground sm:inline-block"
    >
      {{ shortcut }}
    </kbd>
  </button>
</template>
