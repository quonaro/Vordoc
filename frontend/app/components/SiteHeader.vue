<script setup lang="ts">
import type { HeaderConfig } from '~/composables/useSiteConfig'
import { resolveFont } from '~/utils/fonts'

const props = defineProps<{
  header?: HeaderConfig
}>()

const title = computed(() => props.header?.title || 'Vordoc')
const logo = computed(() => props.header?.logo?.path || '/api/v1/logo')
const logoSize = computed(() => props.header?.logo?.size ?? 40)
const showThemeSelector = computed(() => props.header?.selector ?? true)

const font = computed(() => {
  const raw = props.header?.font?.name
  if (!raw) return resolveFont('FabergeDigital.otf')
  return resolveFont(raw)
})

const fontSize = computed(() => props.header?.font?.size ?? 24)

const fontFace = computed(() => {
  if (!font.value.isCustom || !font.value.url) return ''
  const format = font.value.url.toLowerCase().endsWith('.otf')
    ? 'opentype'
    : 'truetype'
  return `@font-face {
  font-family: "${font.value.family}";
  src: url("${font.value.url}") format("${format}");
  font-weight: normal;
  font-style: normal;
  font-display: swap;
}`
})

useHead(() => ({
  style: fontFace.value
    ? [{ innerHTML: fontFace.value, type: 'text/css' }]
    : [],
}))
</script>

<template>
  <header
    v-if="header?.enable"
    class="grid grid-cols-[1fr_auto_1fr] items-center border-b px-4 py-3"
  >
    <div />
    <div class="flex items-center gap-3">
      <img
        :src="logo"
        alt="logo"
        :style="{ height: `${logoSize}px`, width: 'auto' }"
      />
      <span
        class="site-header-title leading-none"
        :style="{
          fontFamily: `${font.family}, ui-serif, Georgia, serif`,
          fontSize: `${fontSize}px`,
        }"
        >{{ title }}</span
      >
    </div>
    <ThemeSelector v-if="showThemeSelector" class="justify-self-end" />
  </header>
</template>

<style scoped>
.site-header-title {
  font-weight: 700;
}
</style>
