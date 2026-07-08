<script setup lang="ts">
import { AlertCircle } from '@lucide/vue'

const { t } = useText()
const siteConfig = useSiteConfig()
const error = useError()

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
</script>

<template>
  <div class="min-h-screen bg-background">
    <SiteHeader />
    <main
      class="mx-auto flex max-w-7xl flex-col items-center justify-center p-8"
    >
      <div class="flex max-w-md flex-col items-center text-center">
        <AlertCircle class="mb-4 h-16 w-16 text-muted-foreground" />
        <h1 class="text-4xl font-bold tracking-tight text-foreground">
          {{ statusCode }}
        </h1>
        <p class="mt-2 text-xl font-semibold text-foreground">
          {{ title }}
        </p>
        <p class="mt-2 text-muted-foreground">
          {{ description }}
        </p>
        <UiButton v-if="homePath" class="mt-6" @click="clearErrorAndNavigate">
          {{ t('errors.goHome') }}
        </UiButton>
      </div>
    </main>
  </div>
</template>
