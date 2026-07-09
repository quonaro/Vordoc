<script setup lang="ts">
import { Search, X, FileText, LockKeyhole, BookOpen } from '@lucide/vue'

const { t } = useText()

interface PageResult {
  title: string
  path: string
  snippet?: string
}

interface DocResult {
  name: string
  title: string
  description?: string
  access?: string
  pages?: PageResult[]
}

const props = withDefaults(
  defineProps<{
    inline?: boolean
  }>(),
  {
    inline: false,
  },
)

const emit = defineEmits<{
  close: []
}>()

const config = useRuntimeConfig()
const router = useRouter()

const query = ref('')
const results = ref<DocResult[]>([])
const loading = ref(false)
const activeIndex = ref(0)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const inputRef = ref<HTMLInputElement | null>(null)
const resultsRef = ref<HTMLElement | null>(null)

const flatItems = computed(() => {
  const items: { doc: DocResult; page?: PageResult; index: number }[] = []
  let index = 0
  for (const doc of results.value) {
    items.push({ doc, index: index++ })
    if (doc.pages) {
      for (const page of doc.pages) {
        items.push({ doc, page, index: index++ })
      }
    }
  }
  return items
})

watch(
  () => query.value,
  (value) => {
    if (debounceTimer) clearTimeout(debounceTimer)
    activeIndex.value = 0

    const trimmed = value.trim()
    if (trimmed.length < 2) {
      results.value = []
      loading.value = false
      return
    }

    loading.value = true
    debounceTimer = setTimeout(async () => {
      try {
        const data = await $fetch<{ results: DocResult[] }>(
          `${config.public.apiBase}/v1/search`,
          {
            query: { q: trimmed },
            credentials: 'include',
          },
        )
        results.value = data.results ?? []
      } catch (e) {
        console.error('global search failed', e)
        results.value = []
      } finally {
        loading.value = false
      }
    }, 200)
  },
)

watch(activeIndex, () => {
  nextTick(() => {
    const active = resultsRef.value?.querySelector<HTMLElement>(
      '[data-active="true"]',
    )
    active?.scrollIntoView({ block: 'nearest' })
  })
})

function navigateTo(item: { doc: DocResult; page?: PageResult }) {
  const docName = item.doc.name
  const pagePath = item.page?.path ?? ''
  const target = pagePath ? `/${docName}/${pagePath}` : `/${docName}`
  const q = query.value.trim()
  router.push({ path: target, query: q ? { q } : undefined })
  close()
}

function close() {
  query.value = ''
  results.value = []
  activeIndex.value = 0
  emit('close')
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIndex.value =
      activeIndex.value < flatItems.value.length - 1 ? activeIndex.value + 1 : 0
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIndex.value =
      activeIndex.value > 0 ? activeIndex.value - 1 : flatItems.value.length - 1
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const item = flatItems.value[activeIndex.value]
    if (item) navigateTo(item)
  } else if (e.key === 'Escape') {
    close()
  }
}

function highlight(text: string | undefined, q: string): string {
  if (!text) return ''
  const terms = q
    .toLowerCase()
    .split(/\s+/)
    .filter((t) => t.length >= 2)
  if (!terms.length) return text
  const pattern = new RegExp(
    `(${terms.map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`,
    'gi',
  )
  return text.replace(
    pattern,
    (match) => `<span class="search-mark">${match}</span>`,
  )
}

onMounted(() => {
  if (!props.inline) {
    nextTick(() => inputRef.value?.focus())
  }
})
</script>

<template>
  <ClientOnly :fallback-tag="'span'">
    <Teleport to="body" :disabled="inline">
      <div
        :class="[
          inline
            ? 'relative w-full'
            : 'fixed inset-0 z-[100] w-full bg-black/50 p-0 md:p-4 backdrop-blur-sm',
        ]"
        @click.self="!inline && close()"
      >
        <div
          :class="[
            'search-dialog flex w-full flex-col overflow-hidden border bg-card shadow-2xl',
            inline
              ? 'max-w-none rounded-xl'
              : 'fixed bottom-0 left-0 right-0 h-[85vh] rounded-t-xl md:absolute md:bottom-auto md:left-1/2 md:right-auto md:top-1/2 md:h-auto md:max-h-[80vh] md:w-full md:max-w-2xl md:-translate-x-1/2 md:-translate-y-1/2 md:rounded-xl',
          ]"
          @keydown="onKeydown"
        >
          <div
            v-if="!inline"
            class="mx-auto mt-2 h-1.5 w-12 shrink-0 rounded-full bg-muted md:hidden"
            aria-hidden="true"
          />

          <div class="relative flex items-center gap-3 border-b px-4 py-3">
            <Search class="h-5 w-5 shrink-0 text-muted-foreground" />
            <input
              ref="inputRef"
              v-model="query"
              type="text"
              :placeholder="t('search.globalPlaceholder')"
              class="min-w-0 flex-1 bg-transparent text-base outline-none placeholder:text-muted-foreground"
            />
            <button
              v-if="query"
              type="button"
              class="rounded p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
              @click="query = ''"
            >
              <X class="h-4 w-4" />
            </button>
            <span
              v-if="!inline"
              class="hidden shrink-0 rounded border bg-muted px-1.5 py-0.5 text-xs text-muted-foreground sm:inline-block"
            >
              ESC
            </span>
          </div>

          <div
            ref="resultsRef"
            :class="[
              'overflow-y-auto',
              inline ? 'max-h-[50vh]' : 'flex-1 md:max-h-[60vh]',
            ]"
          >
            <div
              v-if="loading"
              class="px-4 py-8 text-center text-sm text-muted-foreground"
            >
              {{ t('search.searching') }}
            </div>

            <div
              v-else-if="query.trim().length >= 2 && results.length === 0"
              class="px-4 py-8 text-center text-sm text-muted-foreground"
            >
              {{ t('search.noResults') }}
            </div>

            <ul v-else-if="results.length > 0" class="py-2">
              <template v-for="doc in results" :key="doc.name">
                <li
                  :data-active="
                    activeIndex ===
                    flatItems.find((i) => i.doc === doc && !i.page)?.index
                  "
                  :class="[
                    'flex cursor-pointer items-center gap-3 border-b px-4 py-2.5 text-sm font-medium transition-colors',
                    activeIndex ===
                    flatItems.find((i) => i.doc === doc && !i.page)?.index
                      ? 'bg-accent text-accent-foreground'
                      : 'bg-muted/40 text-foreground hover:bg-muted',
                  ]"
                  @mouseenter="
                    activeIndex =
                      flatItems.find((i) => i.doc === doc && !i.page)?.index ??
                      0
                  "
                  @click="navigateTo({ doc })"
                >
                  <BookOpen class="h-4 w-4 shrink-0 opacity-70" />
                  <span class="flex-1 truncate">{{ doc.title }}</span>
                  <span
                    v-if="doc.access === 'password'"
                    class="flex items-center gap-1 rounded bg-background px-1.5 py-0.5 text-xs text-muted-foreground"
                  >
                    <LockKeyhole class="h-3 w-3" />
                    {{ t('password.title') }}
                  </span>
                </li>

                <li
                  v-for="page in doc.pages"
                  :key="page.path"
                  :data-active="
                    activeIndex ===
                    flatItems.find((i) => i.doc === doc && i.page === page)
                      ?.index
                  "
                  :class="[
                    'flex cursor-pointer items-start gap-3 border-l-2 border-transparent px-4 py-2 transition-all duration-150',
                    activeIndex ===
                    flatItems.find((i) => i.doc === doc && i.page === page)
                      ?.index
                      ? 'search-result-active border-l-accent'
                      : 'hover:border-l-accent/50 hover:bg-muted/30',
                  ]"
                  @mouseenter="
                    activeIndex =
                      flatItems.find((i) => i.doc === doc && i.page === page)
                        ?.index ?? 0
                  "
                  @click="navigateTo({ doc, page })"
                >
                  <FileText
                    class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground"
                  />
                  <div class="min-w-0 flex-1">
                    <div class="text-sm font-medium">{{ page.title }}</div>
                    <!-- eslint-disable vue/no-v-html -->
                    <div
                      v-if="page.snippet"
                      class="mt-0.5 line-clamp-2 text-xs text-muted-foreground"
                      v-html="highlight(page.snippet, query)"
                    />
                    <!-- eslint-enable vue/no-v-html -->
                  </div>
                </li>
              </template>
            </ul>
          </div>

          <div
            v-if="!inline"
            class="hidden items-center justify-between border-t px-4 py-2 text-xs text-muted-foreground md:flex"
          >
            <span>{{ t('search.globalHint') }}</span>
            <span class="flex items-center gap-1">
              <kbd class="rounded border bg-muted px-1.5 py-0.5">↑</kbd>
              <kbd class="rounded border bg-muted px-1.5 py-0.5">↓</kbd>
              <kbd class="rounded border bg-muted px-1.5 py-0.5">↵</kbd>
              <span class="mx-1">{{ t('search.selectHint') }}</span>
              <kbd class="rounded border bg-muted px-1.5 py-0.5">esc</kbd>
              <span>{{ t('search.closeHint') }}</span>
            </span>
          </div>
        </div>
      </div>
    </Teleport>
  </ClientOnly>
</template>

<style scoped>
.search-dialog {
  padding-bottom: env(safe-area-inset-bottom);
  animation: sheet-in 0.25s ease-out;
}

@media (min-width: 768px) {
  .search-dialog {
    animation: dialog-in 0.2s ease-out;
  }
}

@keyframes sheet-in {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}

@keyframes dialog-in {
  from {
    opacity: 0;
    transform: translate(-50%, -48%) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
}
</style>
