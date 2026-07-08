<script setup lang="ts">
import type { HeaderConfig, HeaderElement } from '~/composables/useSiteConfig'
import { resolveFont } from '~/utils/fonts'

const { t } = useText()

const props = defineProps<{
  header?: HeaderConfig
}>()

const siteConfig = useSiteConfig()

const resolvedHeader = computed(
  () => props.header ?? siteConfig.data.value?.header,
)
const headerEnabled = computed(() => resolvedHeader.value?.enable ?? false)

const title = computed(() => resolvedHeader.value?.title || t('app.title'))
const logo = computed(() => resolvedHeader.value?.logo?.path || '/api/v1/logo')
const logoSize = computed(() => resolvedHeader.value?.logo?.size ?? 40)

const defaultElements: HeaderElement[] = ['logo', 'search', 'theme-switch']
const elements = computed(() => {
  const raw = resolvedHeader.value?.elements
  if (raw === undefined || raw === null) return defaultElements
  return raw
})

const font = computed(() => {
  const raw = resolvedHeader.value?.font?.name
  if (!raw) return resolveFont('FabergeDigital.otf')
  return resolveFont(raw)
})

const fontSize = computed(() => resolvedHeader.value?.font?.size ?? 24)

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

const positionClass = (id: HeaderElement): string => {
  const idx = elements.value.indexOf(id)
  if (idx === -1) return 'hidden'
  const starts = ['col-start-1', 'col-start-2', 'col-start-3']
  const aligns = [
    'justify-self-start',
    'justify-self-center',
    'justify-self-end',
  ]
  return `${starts[idx]} ${aligns[idx]}`
}
</script>

<template>
  <header
    v-if="headerEnabled"
    class="grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center border-b px-4 py-3"
  >
    <div :class="cn('flex items-center gap-3', positionClass('logo'))">
      <img
        :src="logo"
        :alt="t('app.logoAlt')"
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
    <GlobalSearchTrigger
      :class="cn('w-full max-w-xl', positionClass('search'))"
    />
    <ThemeSelector :class="positionClass('theme-switch')" />
  </header>
</template>

<style scoped>
.site-header-title {
  font-weight: 700;
}
</style>
