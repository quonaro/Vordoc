<script setup lang="ts">
import type { HeaderConfig } from '~/composables/useSiteConfig'
import { findTocTitleByLink } from '~/composables/useToc'
import { renderMarkdown } from '~/utils/markdown'
import { cacheApiResponse } from '~/utils/apiCache'
import { prefetchDocPages } from '~/composables/usePagePrefetch'
import { LockKeyhole } from '@lucide/vue'

const { t } = useText()

interface PageNode {
  path: string
  title: string
  access?: string
  access_scope?: string
  lock_color?: string
  children?: PageNode[]
}

interface DocMeta {
  name: string
  title: string
  description?: string
  access?: string
  access_scope?: string
  lock_color?: string
  pages?: PageNode[]
  header?: HeaderConfig
}

interface PageData {
  doc: string
  path: string
  filePath: string
  title: string
  order?: number
  content?: string
}

const route = useRoute()
const config = useRuntimeConfig()
const docName = route.params.doc as string
const pagePath = (route.params.page as string[])?.join('/') ?? ''

const docMeta = ref<DocMeta | null>(null)
const pageData = useState<PageData | null>(`doc-page-${docName}`, () => null)
const loading = ref(true)
const passwordRequired = ref(false)
const passwordMode = ref<'doc' | 'page'>('page')
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

const sidebarNodes = useSidebarNodes(
  computed(() => docMeta.value?.pages ?? []),
  computed(() => pagePath),
)

useSearchHighlight(
  contentRef,
  computed(() => pageData.value?.content ?? ''),
)

const tocItems = useToc(computed(() => pageData.value?.content ?? ''))
const activeLink = useActiveAnchor(contentRef, tocItems)

const activeSectionTitle = computed(() => {
  if (!activeLink.value) return undefined
  return findTocTitleByLink(tocItems.value, activeLink.value)
})

const pageTitle = usePageTitle()
pageTitle.set(() => [
  docMeta.value?.title,
  pageData.value?.title,
  activeSectionTitle.value,
])

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
      passwordMode.value = 'doc'
      passwordRequired.value = true
      return false
    }
    if (err?.statusCode === 404 || err?.data?.error === 'doc_not_found') {
      throw createError({
        statusCode: 404,
        statusMessage: t('errors.doc_not_found'),
        fatal: true,
      })
    }
    console.error('failed to fetch doc', e)
    throw createError({
      statusCode: err?.statusCode ?? 500,
      statusMessage: t('errors.failed_to_get_doc'),
      fatal: true,
    })
  }
}

async function fetchPage() {
  try {
    const url = `${config.public.apiBase}/v1/${docName}/${pagePath}`
    pageData.value = await $fetch<PageData>(url, { credentials: 'include' })
    await cacheApiResponse(url, pageData.value)
    passwordRequired.value = false
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      data?: { password_required?: boolean; error?: string }
    }
    if (err?.statusCode === 403 && err?.data?.password_required) {
      passwordMode.value = 'page'
      passwordRequired.value = true
      return
    }
    if (err?.statusCode === 404 || err?.data?.error === 'page_not_found') {
      throw createError({
        statusCode: 404,
        statusMessage: t('errors.page_not_found'),
        fatal: true,
      })
    }
    console.error('failed to fetch doc page', e)
    throw createError({
      statusCode: err?.statusCode ?? 500,
      statusMessage: t('errors.failed_to_get_page'),
      fatal: true,
    })
  }
}

async function onUnlock() {
  const unlocked = await loadDocMeta()
  if (!unlocked) return
  await fetchPage()
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
  await fetchPage()
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
          :current-path="pagePath"
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
        <nav v-if="docMeta" class="sticky top-8">
          <div class="rounded-lg border bg-card p-3 shadow-sm md:p-4">
            <NuxtLink
              :to="`/${docName}`"
              class="flex items-center gap-2 text-base font-semibold hover:text-primary"
            >
              <span class="flex-1 truncate">{{ docMeta.title }}</span>
              <LockKeyhole
                v-if="docMeta.access === 'password'"
                class="h-3.5 w-3.5"
                :style="{ color: docMeta.lock_color }"
              />
            </NuxtLink>
            <div class="my-3 border-t border-border" />
            <SidebarBack
              :doc-name="docName"
              :current-path="pagePath"
              :pages="docMeta?.pages ?? []"
              class="mb-2"
            />
            <SidebarTree
              v-if="docMeta?.pages?.length"
              :nodes="sidebarNodes"
              :doc-name="docName"
              :current-path="pagePath"
            />
            <p v-else class="text-sm text-muted-foreground">
              {{ t('sidebar.empty') }}
            </p>
          </div>
        </nav>
      </aside>

      <!-- Content -->
      <main class="min-w-0 flex-1">
        <Breadcrumbs
          :doc-title="docMeta?.title ?? docName"
          :doc-name="docName"
          :pages="docMeta?.pages ?? []"
          :current-path="pagePath"
        />
        <PasswordForm
          v-if="passwordRequired"
          :doc="docName"
          :page-path="pagePath"
          :mode="passwordMode"
          @success="onUnlock"
          @close="navigateTo(`/${docName}`, { replace: true })"
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
          :current-path="pagePath"
          :loading="loading"
        />
      </main>

      <!-- Table of contents -->
      <DocOutline :items="tocItems" :active-link="activeLink" />
    </div>
  </div>
</template>
