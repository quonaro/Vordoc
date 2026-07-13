<script setup lang="ts">
import { AlertCircle } from '@lucide/vue'

defineOptions({ inheritAttrs: false })

const { t } = useText()
const siteConfig = useSiteConfig()
const error = useError()
const router = useRouter()
const globalSearchOpen = useGlobalSearchState()

useGlobalSearchShortcut(() => {
  globalSearchOpen.value = true
})

onMounted(() => {
  siteConfig.load().catch(() => {})
})

const statusCode = computed(() => {
  const err = error.value as { statusCode?: number } | null
  return err?.statusCode ?? 500
})

const isNotFound = computed(() => statusCode.value === 404)

const title = computed(() =>
  isNotFound.value ? t('errors.notFoundTitle') : t('errors.genericTitle'),
)

const description = computed(() => {
  if (isNotFound.value) return t('errors.notFoundDescription')
  const err = error.value as { statusMessage?: string; message?: string } | null
  return err?.statusMessage || err?.message || t('errors.genericDescription')
})

const homePath = computed(() => {
  const root = siteConfig.data.value?.root
  if (root?.enable === false) return undefined
  return '/'
})

function clearErrorAndNavigate() {
  clearError()
  if (homePath.value) {
    navigateTo(homePath.value)
  }
}

function clearErrorAndGoBack() {
  clearError()
  router.back()
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <SiteHeader />
    <main
      class="mx-auto flex max-w-7xl flex-col items-center justify-center p-4 md:p-8"
    >
      <div class="flex max-w-md flex-col items-center text-center">
        <AlertCircle
          class="mb-4 h-12 w-12 text-muted-foreground md:h-16 md:w-16"
        />
        <h1
          class="text-3xl font-bold tracking-tight text-foreground md:text-4xl"
        >
          {{ statusCode }}
        </h1>
        <p class="mt-2 text-lg font-semibold text-foreground md:text-xl">
          {{ title }}
        </p>
        <p class="mt-2 text-muted-foreground">
          {{ description }}
        </p>
        <div class="mt-6 flex flex-wrap items-center justify-center gap-3">
          <UiButton variant="outline" @click="clearErrorAndGoBack">
            {{ t('errors.goBack') }}
          </UiButton>
          <UiButton v-if="homePath" @click="clearErrorAndNavigate">
            {{ t('errors.goHome') }}
          </UiButton>
        </div>
      </div>
    </main>
  </div>
  <GlobalSearch v-if="globalSearchOpen" @close="globalSearchOpen = false" />
</template>
