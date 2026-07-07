<script setup lang="ts">
import { LockKeyhole } from '@lucide/vue'

interface DocMeta {
  name: string
  title: string
  description?: string
  access?: string
}

const runtimePublic = useRuntimePublic()
const docs = ref<DocMeta[]>([])
const loading = ref(true)
const enableRootPage = ref(true)

onMounted(async () => {
  try {
    await runtimePublic.load()
    enableRootPage.value = runtimePublic.get<boolean>('enable_root_page', true)
    if (!enableRootPage.value) return

    const list = await $fetch<{ docs: { name: string; access?: string }[] }>(
      'http://localhost:8080/api/v1',
    )
    const items = list.docs ?? []
    if (items.length === 0) return

    const results = await Promise.all(
      items.map(async (item) => {
        try {
          const meta = await $fetch<DocMeta>(
            `http://localhost:8080/api/v1/${item.name}`,
          )
          return { ...meta, access: meta.access || item.access }
        } catch {
          return { name: item.name, title: item.name, access: item.access }
        }
      }),
    )
    docs.value = results
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
  <div class="min-h-screen bg-background p-8">
    <div class="mx-auto max-w-4xl">
      <h1 class="mb-8 text-4xl font-bold">Vordoc</h1>
      <p class="mb-8 text-muted-foreground">Available documentation</p>

      <div v-if="enableRootPage" class="grid gap-4 sm:grid-cols-2">
        <NuxtLink
          v-for="doc in docs"
          :key="doc.name"
          :to="`/${doc.name}`"
          class="rounded-lg border bg-card p-6 transition-colors hover:bg-accent"
        >
          <div class="flex items-center gap-2">
            <LockKeyhole
              v-if="isProtected(doc)"
              class="h-4 w-4 text-muted-foreground"
            />
            <h2 class="text-xl font-semibold">{{ doc.title }}</h2>
          </div>
          <p v-if="doc.description" class="mt-2 text-sm text-muted-foreground">
            {{ doc.description }}
          </p>
        </NuxtLink>
      </div>

      <p v-if="!enableRootPage" class="text-muted-foreground">
        Documentation selection is disabled.
      </p>
      <p v-else-if="loading" class="text-muted-foreground">Loading...</p>
      <p v-else-if="docs.length === 0" class="text-muted-foreground">
        No documentation found.
      </p>
    </div>
  </div>
</template>
