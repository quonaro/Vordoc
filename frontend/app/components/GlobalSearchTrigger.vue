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

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault()
    open()
  }
}
</script>

<template>
  <div
    class="group relative flex h-10 w-full cursor-pointer items-center rounded-md border border-input bg-background px-3 text-sm text-muted-foreground shadow-sm transition-all hover:border-accent hover:bg-accent/5 hover:text-foreground md:max-w-md"
    @click="open"
  >
    <Search class="pointer-events-none absolute left-3 h-4 w-4 shrink-0" />
    <input
      readonly
      type="text"
      :placeholder="t('search.globalPlaceholder')"
      class="h-full w-full cursor-pointer bg-transparent py-0 pl-7 pr-16 text-base outline-none placeholder:text-muted-foreground md:text-sm"
      @focus="open"
      @click="open"
      @keydown="onKeydown"
    />
    <kbd
      class="pointer-events-none absolute right-2 hidden rounded border bg-muted px-1.5 py-0.5 text-xs font-medium transition-colors group-hover:border-accent/50 group-hover:text-foreground md:inline-block"
    >
      {{ shortcut }}
    </kbd>
  </div>
</template>
