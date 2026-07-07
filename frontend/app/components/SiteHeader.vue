<script setup lang="ts">
import type { HeaderConfig } from '~/composables/useSiteConfig'

const props = defineProps<{
  header?: HeaderConfig
}>()

const title = computed(() => props.header?.title || 'Vordoc')
const logo = computed(() => props.header?.logo || '/api/v1/logo')
</script>

<template>
  <header v-if="header?.enable" class="border-b bg-background">
    <div class="mx-auto flex max-w-6xl items-center gap-4 px-4 py-3">
      <img :src="logo" alt="logo" class="h-8 w-auto" />
      <span class="text-lg font-semibold">
        {{ title }}
      </span>
    </div>
  </header>
</template>
