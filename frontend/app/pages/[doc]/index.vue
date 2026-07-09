<script setup lang="ts">
import type { HeaderConfig } from '~/composables/useSiteConfig'
import { findTocTitleByLink } from '~/composables/useToc'
import { renderMarkdown } from '~/utils/markdown'
import { cacheApiResponse } from '~/utils/apiCache'
import { prefetchDocPages } from '~/composables/usePagePrefetch'

const { t } = useText()

interface PageNode {
  path: string
  title: string
  access?: string
  access_scope?: string
  lock_color?: string
  has_index?: boolean
  show?: boolean
  children?: PageNode[]
}

interface PageData {
  doc: string
  path: string
  filePath: string
  title: string
  order?: number
  content?: string
}

interface DocMeta {
  name: string
  title: string
  description?: string
  access?: string
  access_scope?: string
  lock_color?: string
  pages?: PageNode[]
  index_page?: PageData
  header?: HeaderConfig
}

const route = useRoute()
const config = useRuntimeConfig()
const docName = route.params.doc as string

const docMeta = ref<DocMeta | null>(null)
const pageData = useState<PageData | null>(`doc-page-${docName}`, () => null)
const loading = ref(true)
const passwordRequired = ref(false)
const contentRef = shallowRef<HTMLElement | null>(null)

const theme = useTheme()
const effectiveTheme = computed(() => {
  if (theme.theme.value === 'dark') return 'dark'
  if (theme.theme.value === 'light') return 'light'
  if (!import.meta.client) return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light'
})

useCopyCode(contentRef)
useMermaid(contentRef, effectiveTheme)

const currentPath = ''

const sidebarNodes = useSidebarNodes(
  computed(() => docMeta.value?.pages ?? []),
  computed(() => currentPath),
)

const tocItems = useToc(computed(() => pageData.value?.content ?? ''))
const activeLink = useActiveAnchor(contentRef, tocItems)

useSearchHighlight(
  contentRef,
  computed(() => pageData.value?.content ?? ''),
)

const activeSectionTitle = computed(() => {
  if (!activeLink.value) return undefined
  return findTocTitleByLink(tocItems.value, activeLink.value)
})

const pageTitle = usePageTitle()
pageTitle.set(() => [docMeta.value?.title, activeSectionTitle.value])

async function loadDocMeta(): Promise<boolean> {
  try {
    const url = `${config.public.apiBase}/v1/${docName}`
    docMeta.value = await $fetch<DocMeta>(url, { credentials: 'include' })
    await cacheApiResponse(url, docMeta.value)
    return true
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      data?: { password_required?: boolean; error?: string }
    }
    if (err?.statusCode === 403 && err?.data?.password_required) {
      passwordRequired.value = true
      return false
    }
    if (err?.statusCode === 404 || err?.data?.error === 'doc_not_found') {
      throw createError({
        statusCode: 404,
        statusMessage: t('errors.doc_not_found'),
      })
    }
    console.error('failed to fetch doc', e)
    throw createError({
      statusCode: err?.statusCode ?? 500,
      statusMessage: t('errors.failed_to_get_doc'),
    })
  }
}

async function fetchPage() {
  try {
    const url = `${config.public.apiBase}/v1/${docName}/`
    pageData.value = await $fetch<PageData>(url, { credentials: 'include' })
    await cacheApiResponse(url, pageData.value)
    passwordRequired.value = false
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      data?: { password_required?: boolean }
    }
    if (err?.statusCode === 403 && err?.data?.password_required) {
      passwordRequired.value = true
      return
    }
    console.error('failed to fetch doc page', e)
  }
}

async function onUnlock() {
  const unlocked = await loadDocMeta()
  if (!unlocked) return
  if (docMeta.value?.index_page) {
    pageData.value = docMeta.value.index_page
    passwordRequired.value = false
  } else {
    await fetchPage()
  }
}

const renderedContent = computed(() => {
  if (!pageData.value?.content) return ''
  return renderMarkdown(
    pageData.value.content,
    docName,
    pageData.value.filePath,
  )
})

const unlocked = await loadDocMeta()
if (unlocked) {
  void prefetchDocPages(docName, docMeta.value?.pages ?? [])
  if (docMeta.value?.index_page) {
    pageData.value = docMeta.value.index_page
    passwordRequired.value = false
  } else {
    await fetchPage()
  }
}
loading.value = false
</script>

<template>
  <div class="min-h-screen bg-background">
    <SiteHeader :header="docMeta?.header">
      <template #mobile-leading>
        <MobileDocNav
          :doc-name="docName"
          :doc-title="docMeta?.title ?? docName"
          :pages="docMeta?.pages ?? []"
          :current-path="currentPath"
          :access="docMeta?.access"
          :lock-color="docMeta?.lock_color"
        />
      </template>
      <template #mobile-trailing>
        <MobileTocNav :items="tocItems" :active-link="activeLink" />
      </template>
    </SiteHeader>
    <div class="mx-auto flex max-w-7xl gap-4 p-4 md:gap-8 md:p-8">
      <!-- Sidebar -->
      <aside class="hidden w-64 shrink-0 lg:block">
        <nav v-if="docMeta?.pages?.length" class="sticky top-8 space-y-1">
          <NuxtLink
            :to="`/${docName}`"
            class="mb-4 block font-semibold hover:text-primary"
          >
            {{ docMeta.title }}
          </NuxtLink>
          <SidebarTree
            :nodes="sidebarNodes"
            :doc-name="docName"
            :current-path="currentPath"
          />
        </nav>
      </aside>

      <!-- Content -->
      <main class="min-w-0 flex-1">
        <Breadcrumbs
          :doc-title="docMeta?.title ?? docName"
          :doc-name="docName"
          :pages="docMeta?.pages ?? []"
          :current-path="currentPath"
        />
        <PasswordForm
          v-if="passwordRequired"
          :doc="docName"
          :page-path="''"
          @success="onUnlock"
          @close="navigateTo('/', { replace: true })"
        />
        <div
          v-if="loading"
          class="space-y-4"
          aria-busy="true"
          aria-label="Loading page"
        >
          <div class="h-8 w-2/3 animate-pulse rounded bg-muted" />
          <div class="h-4 w-full animate-pulse rounded bg-muted" />
          <div class="h-4 w-5/6 animate-pulse rounded bg-muted" />
          <div class="h-32 w-full animate-pulse rounded bg-muted" />
          <div class="h-4 w-4/5 animate-pulse rounded bg-muted" />
        </div>
        <div
          v-else-if="renderedContent"
          ref="contentRef"
          class="prose prose-slate max-w-none dark:prose-invert"
        >
          <div v-html="renderedContent" />
        </div>
        <p v-else-if="!passwordRequired" class="text-muted-foreground">
          {{ t('doc.noContent') }}
        </p>
        <PageNavigation
          v-if="!passwordRequired"
          :doc-name="docName"
          :doc-title="docMeta?.title ?? docName"
          :pages="docMeta?.pages ?? []"
          :current-path="currentPath"
          :loading="loading"
        />
      </main>

      <!-- Table of contents -->
      <DocOutline :items="tocItems" :active-link="activeLink" />
    </div>
  </div>
</template>
