<script setup lang="ts">
import { Anchor, X } from '@lucide/vue'
import type { TocItem } from '~/composables/useToc'

const props = defineProps<{
  items: TocItem[]
  activeLink: string | null
}>()

const { t } = useText()
const open = ref(false)

function close() {
  open.value = false
}

watch(() => props.activeLink, close)
</script>

<template>
  <div class="xl:hidden">
    <button
      type="button"
      class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
      :title="t('doc.onThisPage')"
      :disabled="!items.length"
      @click="open = true"
    >
      <Anchor class="h-4 w-4" />
      <span class="sr-only">{{ t('doc.onThisPage') }}</span>
    </button>

    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-200 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="open"
          class="fixed inset-0 z-[60] bg-black/50 backdrop-blur-sm"
          @click="close"
        />
      </Transition>

      <aside
        :class="[
          'fixed right-0 top-0 z-[70] h-full w-[min(80vw,20rem)] overflow-y-auto border-l bg-background p-4 shadow-xl transition-transform duration-300 ease-out',
          open ? 'translate-x-0' : 'translate-x-full',
        ]"
      >
        <div class="mb-4 flex items-center justify-between gap-2">
          <span class="font-semibold">{{ t('doc.onThisPage') }}</span>
          <button
            type="button"
            class="rounded p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            @click="close"
          >
            <X class="h-4 w-4" />
            <span class="sr-only">{{ t('menu.close') }}</span>
          </button>
        </div>
        <TableOfContents :items="items" :active-link="activeLink" root />
      </aside>
    </Teleport>
  </div>
</template>
