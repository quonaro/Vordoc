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
  const starts: Record<HeaderElement, string> = {
    logo: 'md:col-start-1',
    search: 'md:col-start-2',
    'theme-switch': 'md:col-start-3',
  }
  const aligns: Record<HeaderElement, string> = {
    logo: 'md:justify-self-start',
    search: 'md:justify-self-center',
    'theme-switch': 'md:justify-self-end',
  }
  return `${starts[id]} ${aligns[id]}`
}
</script>

<template>
  <header v-if="headerEnabled" class="border-b px-4 py-3">
    <!-- Mobile layout -->
    <div class="flex flex-col gap-2 md:hidden">
      <div class="flex items-center justify-between gap-2">
        <div class="flex min-w-0 items-center gap-3">
          <img
            :src="logo"
            :alt="t('app.logoAlt')"
            :style="{ height: `${logoSize}px`, width: 'auto' }"
          />
          <span
            class="site-header-title min-w-0 truncate leading-none"
            :style="{
              fontFamily: `${font.family}, ui-serif, Georgia, serif`,
              '--header-font-size': `${fontSize}px`,
            }"
            >{{ title }}</span
          >
        </div>
        <ThemeSelector />
      </div>
      <div class="flex items-center gap-2">
        <div v-if="$slots['mobile-leading']">
          <slot name="mobile-leading" />
        </div>
        <GlobalSearchTrigger class="min-w-0 flex-1" />
        <div v-if="$slots['mobile-trailing']">
          <slot name="mobile-trailing" />
        </div>
      </div>
    </div>

    <!-- Desktop layout -->
    <div
      class="hidden md:grid md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] md:items-center md:gap-4"
    >
      <div
        :class="cn('flex min-w-0 items-center gap-3', positionClass('logo'))"
      >
        <img
          :src="logo"
          :alt="t('app.logoAlt')"
          :style="{ height: `${logoSize}px`, width: 'auto' }"
        />
        <span
          class="site-header-title min-w-0 truncate leading-none"
          :style="{
            fontFamily: `${font.family}, ui-serif, Georgia, serif`,
            '--header-font-size': `${fontSize}px`,
          }"
          >{{ title }}</span
        >
      </div>
      <GlobalSearchTrigger
        :class="cn('w-full max-w-xl', positionClass('search'))"
      />
      <ThemeSelector :class="positionClass('theme-switch')" />
    </div>
  </header>
</template>

<style scoped>
.site-header-title {
  font-weight: 700;
  font-size: calc(var(--header-font-size) * 0.75);
}

@media (min-width: 768px) {
  .site-header-title {
    font-size: var(--header-font-size);
  }
}
</style>
