<script setup lang="ts">
import type { HeaderConfig } from '~/composables/useSiteConfig'
import { renderMarkdown } from '~/utils/markdown'

const { t } = useText()

interface PageNode {
  path: string
  title: string
  access?: string
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
const contentRef = shallowRef<HTMLElement | null>(null)

const sidebarNodes = useSidebarNodes(
  computed(() => docMeta.value?.pages ?? []),
  computed(() => pageData.value?.path ?? ''),
)

const tocItems = useToc(computed(() => pageData.value?.content ?? ''))
const activeLink = useActiveAnchor(contentRef, tocItems)

useSearchHighlight(
  contentRef,
  computed(() => pageData.value?.content ?? ''),
)

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
  return renderMarkdown(
    pageData.value.content,
    docName,
    pageData.value.filePath,
  )
})

pageData.value = null

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

      <!-- Table of contents -->
      <DocOutline :items="tocItems" :active-link="activeLink" />
    </div>
  </div>
</template>
