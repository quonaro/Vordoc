<script setup lang="ts">
import { LockKeyhole } from '@lucide/vue'

const { t } = useText()

interface DocMeta {
  name: string
  title: string
  description?: string
  access?: string
}

const siteConfig = useSiteConfig()
const config = useRuntimeConfig()
const docs = ref<DocMeta[]>([])
const loading = ref(true)

const rootPage = computed(() => siteConfig.data.value?.root)
const enableRootPage = computed(() => rootPage.value?.enable ?? true)
const rootTitle = computed(() => rootPage.value?.title || t('app.title'))
const header = computed(() => siteConfig.data.value?.header)

onMounted(async () => {
  try {
    await siteConfig.load()
    if (!enableRootPage.value) return

    const list = await $fetch<{ docs: DocMeta[] }>(
      `${config.public.apiBase}/v1/docs`,
    )
    docs.value = list.docs ?? []
  } catch (e) {
    console.error('failed to fetch docs', e)
  } finally {
    loading.value = false
  }
})

function isProtected(doc: DocMeta): boolean {
  return doc.access === 'password'
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <SiteHeader :header="header" />
    <div class="p-8">
      <div class="mx-auto max-w-4xl">
        <div v-if="enableRootPage" class="grid gap-4 sm:grid-cols-2">
          <NuxtLink
            v-for="doc in docs"
            :key="doc.name"
            :to="`/${doc.name}`"
            class="group rounded-lg border bg-card p-6 transition-colors hover:bg-accent hover:text-accent-foreground"
          >
            <div class="flex items-center gap-2">
              <LockKeyhole
                v-if="isProtected(doc)"
                class="h-4 w-4 text-muted-foreground group-hover:text-accent-foreground"
              />
              <h2
                class="text-xl font-semibold group-hover:text-accent-foreground"
              >
                {{ doc.title }}
              </h2>
            </div>
            <p
              v-if="doc.description"
              class="mt-2 text-sm text-muted-foreground group-hover:text-accent-foreground"
            >
              {{ doc.description }}
            </p>
          </NuxtLink>
        </div>

        <p v-if="!enableRootPage" class="text-muted-foreground">
          {{ t('root.disabled') }}
        </p>
        <p v-else-if="loading" class="text-muted-foreground">
          {{ t('root.loading') }}
        </p>
        <p v-else-if="docs.length === 0" class="text-muted-foreground">
          {{ t('root.noDocumentation') }}
        </p>
      </div>
    </div>
  </div>
</template>
