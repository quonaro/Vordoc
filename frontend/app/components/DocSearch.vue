<script setup lang="ts">
import { Search, X, FileText } from '@lucide/vue'

const { t } = useText()

interface SearchResult {
  title: string
  path: string
  snippet?: string
}

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()

const query = ref('')
const results = ref<SearchResult[]>([])
const loading = ref(false)
const open = ref(false)
const activeIndex = ref(-1)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const wrapperRef = ref<HTMLElement | null>(null)
const scrolled = ref(false)
const SCROLL_THRESHOLD = 80
const floatingFocused = ref(false)
const floatingHovered = ref(false)
const floatingActive = computed(
  () => floatingFocused.value || floatingHovered.value,
)

function onScroll() {
  scrolled.value = window.scrollY > SCROLL_THRESHOLD
}

function onFloatingFocus() {
  open.value = true
  floatingFocused.value = true
}

const docName = computed(() => {
  const parts = route.path.split('/').filter(Boolean)
  return parts[0] || ''
})

const hasDoc = computed(() => !!docName.value && docName.value !== '')

watch(
  () => query.value,
  (value) => {
    if (debounceTimer) clearTimeout(debounceTimer)
    activeIndex.value = -1

    const trimmed = value.trim()
    if (trimmed.length < 2) {
      results.value = []
      loading.value = false
      return
    }

    loading.value = true
    debounceTimer = setTimeout(async () => {
      try {
        const data = await $fetch<{ results: SearchResult[] }>(
          `${config.public.apiBase}/v1/${docName.value}/search`,
          {
            query: { q: trimmed },
            credentials: 'include',
          },
        )
        results.value = data.results ?? []
      } catch (e) {
        console.error('search failed', e)
        results.value = []
      } finally {
        loading.value = false
      }
    }, 200)
  },
)

function navigateToResult(result: SearchResult) {
  const name = docName.value
  if (!name) return
  const path = result.path ? `/${name}/${result.path}` : `/${name}`
  const q = query.value.trim()
  router.push({ path, query: q ? { q } : undefined })
  close()
}

function close() {
  open.value = false
  query.value = ''
  results.value = []
  activeIndex.value = -1
}

function onKeydown(e: KeyboardEvent) {
  if (!open.value) return

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIndex.value =
      activeIndex.value < results.value.length - 1 ? activeIndex.value + 1 : 0
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIndex.value =
      activeIndex.value > 0 ? activeIndex.value - 1 : results.value.length - 1
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const result = results.value[activeIndex.value]
    if (result) navigateToResult(result)
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
  const listener = (event: MouseEvent) => {
    const el = wrapperRef.value
    if (!el || el.contains(event.target as Node)) return
    open.value = false
  }
  document.addEventListener('click', listener, true)

  window.addEventListener('scroll', onScroll, { passive: true })
  onScroll()

  onUnmounted(() => {
    document.removeEventListener('click', listener, true)
    window.removeEventListener('scroll', onScroll)
  })
})
</script>

<template>
  <div ref="wrapperRef" class="relative z-[40] w-full max-w-xl">
    <template v-if="hasDoc">
      <div
        :class="[
          'flex items-center transition-all duration-300',
          scrolled
            ? 'relative z-[100] w-full opacity-0 scale-[0.98]'
            : 'relative z-[100] w-full opacity-100 scale-100',
        ]"
      >
        <template v-if="!scrolled">
          <Search class="absolute left-3 h-4 w-4 text-muted-foreground" />
          <input
            v-model="query"
            type="text"
            :placeholder="t('search.placeholder')"
            class="h-9 w-full rounded-md border border-input bg-background py-2 pl-9 pr-8 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            @focus="open = true"
            @keydown="onKeydown"
          />
          <button
            v-if="query"
            type="button"
            class="absolute right-2 rounded p-1 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
            @click="query = ''"
          >
            <X class="h-3.5 w-3.5" />
          </button>

          <div
            v-if="open && (query.trim().length >= 2 || results.length > 0)"
            class="absolute left-0 top-full z-50 mt-2 w-full overflow-hidden rounded-md border bg-card shadow-lg"
          >
            <div v-if="loading" class="px-4 py-3 text-sm text-muted-foreground">
              {{ t('search.searching') }}
            </div>
            <div
              v-else-if="results.length === 0 && query.trim().length >= 2"
              class="px-4 py-3 text-sm text-muted-foreground"
            >
              {{ t('search.noResults') }}
            </div>
            <ul
              v-else-if="results.length > 0"
              class="max-h-80 overflow-y-auto py-1"
            >
              <li
                v-for="(result, idx) in results"
                :key="result.path"
                :class="[
                  'cursor-pointer border-l-2 border-transparent px-4 py-2 transition-all duration-150',
                  activeIndex === idx
                    ? 'search-result-active'
                    : 'hover:border-accent/50 hover:bg-muted/30 hover:pl-3',
                ]"
                @click="navigateToResult(result)"
                @mouseenter="activeIndex = idx"
              >
                <div class="flex items-start gap-2">
                  <FileText
                    class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground"
                  />
                  <div class="min-w-0">
                    <div class="text-sm font-medium">
                      {{ result.title }}
                    </div>
                    <!-- eslint-disable vue/no-v-html -->
                    <div
                      v-if="result.snippet"
                      class="mt-0.5 line-clamp-2 text-xs text-muted-foreground"
                      v-html="highlight(result.snippet, query)"
                    />
                    <!-- eslint-enable vue/no-v-html -->
                  </div>
                </div>
              </li>
            </ul>
          </div>
        </template>
      </div>

      <div
        :class="[
          'flex items-center transition-all duration-300',
          scrolled
            ? floatingActive
              ? 'fixed left-1/2 top-3 z-[100] w-[calc(100%-2rem)] max-w-none -translate-x-1/2 scale-100 rounded-md border border-input bg-background/95 opacity-100 shadow-xl backdrop-blur md:w-[90%] md:max-w-2xl md:top-4'
              : 'fixed left-1/2 top-3 z-[100] w-[calc(100%-2rem)] max-w-none -translate-x-1/2 scale-100 rounded-md border border-input bg-background/60 opacity-70 shadow-xl backdrop-blur md:w-[90%] md:max-w-2xl md:top-4'
            : 'fixed left-1/2 top-3 z-[100] w-[calc(100%-2rem)] max-w-none -translate-x-1/2 scale-[0.98] rounded-md border border-input bg-background/60 opacity-0 shadow-xl backdrop-blur pointer-events-none md:w-[90%] md:max-w-2xl md:top-4',
        ]"
        @mouseenter="floatingHovered = true"
        @mouseleave="floatingHovered = false"
        @focusin="onFloatingFocus"
        @focusout="floatingFocused = false"
      >
        <template v-if="scrolled">
          <Search class="absolute left-3 h-4 w-4 text-muted-foreground" />
          <input
            v-model="query"
            type="text"
            :placeholder="t('search.placeholder')"
            class="h-9 w-full rounded-md border-0 bg-transparent py-2 pl-9 pr-8 text-sm outline-none transition-colors placeholder:text-muted-foreground focus:outline-none focus-visible:outline-none focus-visible:ring-0"
            @focus="onFloatingFocus"
            @blur="floatingFocused = false"
            @keydown="onKeydown"
          />
          <button
            v-if="query"
            type="button"
            class="absolute right-2 rounded p-1 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
            @click="query = ''"
          >
            <X class="h-3.5 w-3.5" />
          </button>

          <div
            v-if="
              (open || floatingActive) &&
              (query.trim().length >= 2 || results.length > 0)
            "
            class="absolute left-0 top-full z-50 mt-2 w-full overflow-hidden rounded-md border bg-card shadow-lg"
          >
            <div v-if="loading" class="px-4 py-3 text-sm text-muted-foreground">
              {{ t('search.searching') }}
            </div>
            <div
              v-else-if="results.length === 0 && query.trim().length >= 2"
              class="px-4 py-3 text-sm text-muted-foreground"
            >
              {{ t('search.noResults') }}
            </div>
            <ul
              v-else-if="results.length > 0"
              class="max-h-80 overflow-y-auto py-1"
            >
              <li
                v-for="(result, idx) in results"
                :key="result.path"
                :class="[
                  'cursor-pointer border-l-2 border-transparent px-4 py-2 transition-all duration-150',
                  activeIndex === idx
                    ? 'search-result-active'
                    : 'hover:border-accent/50 hover:bg-muted/30 hover:pl-3',
                ]"
                @click="navigateToResult(result)"
                @mouseenter="activeIndex = idx"
              >
                <div class="flex items-start gap-2">
                  <FileText
                    class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground"
                  />
                  <div class="min-w-0">
                    <div class="text-sm font-medium">
                      {{ result.title }}
                    </div>
                    <!-- eslint-disable vue/no-v-html -->
                    <div
                      v-if="result.snippet"
                      class="mt-0.5 line-clamp-2 text-xs text-muted-foreground"
                      v-html="highlight(result.snippet, query)"
                    />
                    <!-- eslint-enable vue/no-v-html -->
                  </div>
                </div>
              </li>
            </ul>
          </div>
        </template>
      </div>

      <Transition
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-200 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="open || (scrolled && floatingActive)"
          class="fixed inset-0 z-[90] bg-black/40"
          @click="close"
        />
      </Transition>
    </template>
  </div>
</template>
