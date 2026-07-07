<script setup lang="ts">
import { marked } from 'marked'
import type { HeaderConfig } from '~/composables/useSiteConfig'

const { t } = useText()

interface PageNode {
  path: string
  title: string
  access?: string
  children?: PageNode[]
}

interface DocMeta {
  name: string
  title: string
  description?: string
  access?: string
  pages?: PageNode[]
  header?: HeaderConfig
}

interface PageData {
  doc: string
  path: string
  title: string
  order?: number
  content?: string
}

const route = useRoute()
const config = useRuntimeConfig()
const docName = route.params.doc as string
const pagePath = (route.params.page as string[])?.join('/') ?? ''

const { data: docMeta } = await useFetch<DocMeta>(
  `${config.public.apiBase}/v1/${docName}`,
  { key: `doc-meta-${docName}` },
)
const pageData = useState<PageData | null>(`doc-page-${docName}`, () => null)
const loading = ref(true)
const passwordRequired = ref(false)
const contentRef = shallowRef<HTMLElement | null>(null)

useSearchHighlight(
  contentRef,
  computed(() => pageData.value?.content ?? ''),
)

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
      data?: { password_required?: boolean }
    }
    if (err?.statusCode === 403 && err?.data?.password_required) {
      passwordRequired.value = true
      return
    }
    console.error('failed to fetch doc page', e)
  }
}

const renderedContent = computed(() => {
  if (!pageData.value?.content) return ''
  return marked.parse(pageData.value.content)
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

function requiresPassword(): boolean {
  if (!docMeta.value) return false
  if (!pagePath) {
    return docMeta.value.access === 'password'
  }
  const node = findPageNode(docMeta.value.pages, pagePath)
  return node?.access === 'password'
}

onMounted(async () => {
  try {
    if (requiresPassword()) {
      passwordRequired.value = true
    } else {
      await fetchPage()
    }
  } catch (e) {
    console.error('failed to fetch doc page', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="min-h-screen bg-background">
    <SiteHeader :header="docMeta?.header" />
    <div class="mx-auto flex max-w-6xl gap-8 p-8">
      <!-- Sidebar -->
      <aside
        v-if="docMeta?.pages?.length"
        class="hidden w-64 shrink-0 lg:block"
      >
        <nav class="sticky top-8 space-y-1">
          <h3 class="mb-4 font-semibold">{{ docMeta.title }}</h3>
          <SidebarTree
            :nodes="docMeta.pages"
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
          @success="fetchPage"
          @close="navigateTo(`/${docName}`, { replace: true })"
        />
        <h1 class="mb-4 text-3xl font-bold">
          {{ pageData?.title ?? docMeta?.title ?? docName }}
        </h1>
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
      </main>
    </div>
  </div>
</template>
