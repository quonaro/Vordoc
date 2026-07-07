<script setup lang="ts">
import { hexToHsl, hslCssValue, foregroundForHsl } from '~/utils/colors'

const siteConfig = useSiteConfig()
useTheme()

onMounted(async () => {
  try {
    await siteConfig.load()
  } catch (e) {
    console.error('failed to load site config', e)
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

useHead(() => ({
  title: pageTitle.fullTitle,
  link: [
    ...(favicon.value
      ? [{ rel: 'icon', type: 'image/x-icon', href: favicon.value }]
      : []),
  ],
  style: accentStyle.value
    ? [{ innerHTML: accentStyle.value, type: 'text/css' }]
    : [],
}))
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
