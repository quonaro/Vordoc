<script setup lang="ts">
import { marked } from 'marked'
import type { HeaderConfig } from '~/composables/useSiteConfig'

interface PageNode {
  path: string
  title: string
  access?: string
  children?: PageNode[]
}

interface PageData {
  doc: string
  path: string
  title: string
  order?: number
  content?: string
}

interface DocMeta {
  name: string
  title: string
  description?: string
  sidebar?: string[]
  access?: string
  pages?: PageNode[]
  index_page?: PageData
  header?: HeaderConfig
}

const route = useRoute()
const config = useRuntimeConfig()
const docName = route.params.doc as string

const { data: docMeta } = await useFetch<DocMeta>(
  `${config.public.apiBase}/v1/${docName}`,
  { key: `doc-meta-${docName}` },
)
const pageData = useState<PageData | null>(`doc-page-${docName}`, () => null)
const loading = ref(true)
const passwordRequired = ref(false)

async function fetchPage() {
  try {
    pageData.value = await $fetch<PageData>(
      `${config.public.apiBase}/v1/${docName}/`,
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

onMounted(async () => {
  try {
    if (docMeta.value?.access === 'password') {
      passwordRequired.value = true
    } else if (docMeta.value?.index_page) {
      pageData.value = docMeta.value.index_page
      passwordRequired.value = false
    } else {
      await fetchPage()
    }
  } catch (e) {
    console.error('failed to fetch doc', e)
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
      <main class="flex-1 min-w-0">
        <Breadcrumbs
          :doc-title="docMeta?.title ?? docName"
          :doc-name="docName"
          :pages="docMeta?.pages ?? []"
          :current-path="pageData?.path ?? ''"
        />
        <PasswordForm
          v-if="passwordRequired"
          :doc="docName"
          :page-path="''"
          @success="fetchPage"
          @close="navigateTo('/', { replace: true })"
        />
        <h1 class="mb-4 text-3xl font-bold">
          {{ pageData?.title ?? docMeta?.title ?? docName }}
        </h1>
        <div
          v-if="renderedContent"
          class="prose prose-slate max-w-none dark:prose-invert"
        >
          <div v-html="renderedContent" />
        </div>
        <p v-else-if="!passwordRequired" class="text-muted-foreground">
          No content available.
        </p>
      </main>
    </div>
  </div>
</template>
