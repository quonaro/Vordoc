<script setup lang="ts">
import { Menu, X, LockKeyhole } from '@lucide/vue'
import type { PageNode } from '~/composables/useSidebarNodes'

const props = defineProps<{
  docName: string
  docTitle: string
  pages: PageNode[]
  currentPath: string
  access?: string
  lockColor?: string
}>()

const { t } = useText()
const open = ref(false)

const sidebarNodes = useSidebarNodes(
  computed(() => props.pages),
  computed(() => props.currentPath),
)

function close() {
  open.value = false
}

watch(() => props.currentPath, close)
</script>

<template>
  <div class="lg:hidden">
    <button
      type="button"
      class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
      :title="t('menu.open')"
      @click="open = true"
    >
      <Menu class="h-4 w-4" />
      <span class="sr-only">{{ t('menu.open') }}</span>
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
          'fixed left-0 top-0 z-[70] h-full w-[min(80vw,20rem)] overflow-y-auto border-r bg-background p-4 shadow-xl transition-transform duration-300 ease-out',
          open ? 'translate-x-0' : '-translate-x-full',
        ]"
      >
        <div class="mb-4 flex items-center justify-between gap-2">
          <NuxtLink
            :to="`/${docName}`"
            class="flex min-w-0 items-center gap-2 font-semibold hover:text-primary"
            @click="close"
          >
            <span class="truncate">{{ docTitle }}</span>
            <LockKeyhole
              v-if="access === 'password'"
              class="h-3.5 w-3.5 shrink-0"
              :style="{ color: lockColor }"
            />
          </NuxtLink>
          <button
            type="button"
            class="rounded p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            @click="close"
          >
            <X class="h-4 w-4" />
            <span class="sr-only">{{ t('menu.close') }}</span>
          </button>
        </div>
        <SidebarTree
          :nodes="sidebarNodes"
          :doc-name="docName"
          :current-path="currentPath"
        />
      </aside>
    </Teleport>
  </div>
</template>
