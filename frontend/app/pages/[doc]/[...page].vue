<script setup lang="ts">
import type { HeaderConfig } from '~/composables/useSiteConfig'
import { findTocTitleByLink } from '~/composables/useToc'
import { renderMarkdown } from '~/utils/markdown'

const { t } = useText()

interface PageNode {
  path: string
  title: string
  access?: string
  access_scope?: string
  children?: PageNode[]
}

interface DocMeta {
  name: string
  title: string
  description?: string
  access?: string
  access_scope?: string
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
const passwordScope = ref('')
const contentRef = shallowRef<HTMLElement | null>(null)

const sidebarNodes = useSidebarNodes(
  computed(() => docMeta.value?.pages ?? []),
  computed(() => pageData.value?.path ?? ''),
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
    docMeta.value = await $fetch<DocMeta>(
      `${config.public.apiBase}/v1/${docName}`,
      { credentials: 'include' },
    )
    return true
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      data?: { password_required?: boolean; scope?: string }
    }
    if (err?.statusCode === 403 && err?.data?.password_required) {
      passwordRequired.value = true
      passwordScope.value = err?.data?.scope ?? ''
      return false
    }
    console.error('failed to fetch doc', e)
    return false
  }
}

async function fetchPage() {
  try {
    pageData.value = await $fetch<PageData>(
      `${config.public.apiBase}/v1/${docName}/${pagePath}`,
      { credentials: 'include' },
    )
    passwordRequired.value = false
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      data?: { password_required?: boolean; scope?: string }
    }
    if (err?.statusCode === 403 && err?.data?.password_required) {
      passwordRequired.value = true
      passwordScope.value = err?.data?.scope ?? ''
      return
    }
    console.error('failed to fetch doc page', e)
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

function findPageNode(
  nodes: PageNode[] | undefined,
  path: string,
): PageNode | undefined {
  if (!nodes) return undefined
  for (const node of nodes) {
    if (node.path === path) {
      return node
    }
    if (node.children?.length) {
      const found = findPageNode(node.children, path)
      if (found) return found
    }
  }
  return undefined
}

function checkPasswordRequired(): boolean {
  if (!docMeta.value) return false
  if (!pagePath) {
    if (docMeta.value.access === 'password') {
      passwordScope.value = docMeta.value.access_scope ?? ''
      return true
    }
    return false
  }
  const node = findPageNode(docMeta.value.pages, pagePath)
  if (node?.access === 'password') {
    passwordScope.value = node.access_scope ?? ''
    return true
  }
  return false
}

pageData.value = null

const unlocked = await loadDocMeta()
if (unlocked) {
  if (checkPasswordRequired()) {
    passwordRequired.value = true
  } else {
    await fetchPage()
  }
}
loading.value = false
</script>

<template>
  <div class="min-h-screen bg-background">
    <SiteHeader :header="docMeta?.header" />
    <div class="mx-auto flex max-w-7xl gap-8 p-8">
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
            :current-path="pageData?.path ?? ''"
          />
        </nav>
      </aside>

      <!-- Content -->
      <main class="min-w-0 flex-1">
        <Breadcrumbs
          :doc-title="docMeta?.title ?? docName"
          :doc-name="docName"
          :pages="docMeta?.pages ?? []"
          :current-path="pageData?.path ?? ''"
        />
        <PasswordForm
          v-if="passwordRequired"
          :doc="docName"
          :page-path="pagePath"
          :scope="passwordScope"
          @success="onUnlock"
          @close="navigateTo(`/${docName}`, { replace: true })"
        />
        <div
          v-if="renderedContent"
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
          :current-path="pageData?.path ?? ''"
        />
      </main>

      <!-- Table of contents -->
      <DocOutline :items="tocItems" :active-link="activeLink" />
    </div>
  </div>
</template>
