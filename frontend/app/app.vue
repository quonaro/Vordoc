<script setup lang="ts">
import { hexToHsl, hslCssValue, foregroundForHsl } from '~/utils/colors'
import { useText } from '~/composables/useText'
import { useOnline } from '~/composables/useOnline'
import { WifiOff } from '@lucide/vue'

const { t } = useText()
const siteConfig = useSiteConfig()
const isOnline = useOnline()
useTheme()

onMounted(async () => {
  try {
    await siteConfig.load()
  } catch (e) {
    console.error('failed to load site config', e)
  }

  if (import.meta.env.PROD && 'serviceWorker' in navigator) {
    try {
      const registration = await navigator.serviceWorker.register('/sw.js')
      console.log('service worker registered:', registration.scope)
    } catch (e) {
      console.error('service worker registration failed:', e)
    }
  }
})

const favicon = computed(() => {
  const raw = siteConfig.data.value?.favicon
  if (!raw) return '/favicon.ico'
  if (raw.startsWith('http') || raw.startsWith('/')) return raw
  return `/${raw}`
})

const accentStyle = computed(() => {
  const raw = siteConfig.data.value?.theme?.accent_color
  if (!raw) return ''
  try {
    const hsl = hexToHsl(raw)
    const value = hslCssValue(hsl)
    const foreground = foregroundForHsl(hsl)
    return `:root, .dark {
  --accent: ${value};
  --accent-foreground: ${foreground};
  --primary: ${value};
  --primary-foreground: ${foreground};
  --ring: ${value};
}`
  } catch {
    return ''
  }
})

const pageTitle = usePageTitle()
const globalSearchOpen = useGlobalSearchState()

useGlobalSearchShortcut(() => {
  globalSearchOpen.value = true
})

useHead(() => ({
  title: pageTitle.fullTitle,
  link: favicon.value
    ? [{ rel: 'icon', type: 'image/x-icon', href: favicon.value }]
    : [],
  style: accentStyle.value
    ? [{ innerHTML: accentStyle.value, type: 'text/css' }]
    : [],
}))
</script>

<template>
  <div
    v-if="!isOnline"
    class="sticky top-0 z-50 flex items-center justify-center gap-2 bg-yellow-500 px-4 py-2 text-sm font-medium text-black"
    role="status"
    aria-live="polite"
  >
    <WifiOff class="h-4 w-4" aria-hidden="true" />
    <span>{{ t('offline.warning') }}</span>
  </div>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <GlobalSearch v-if="globalSearchOpen" @close="globalSearchOpen = false" />
</template>
