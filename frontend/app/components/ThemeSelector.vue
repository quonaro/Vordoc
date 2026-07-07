<script setup lang="ts">
import { Moon, Sun, Monitor } from '@lucide/vue'
import type { ThemeMode } from '~/composables/useTheme'

const { t } = useText()
const { theme, setTheme, options } = useTheme()

const labels = computed<Record<ThemeMode, string>>(() => ({
  system: t('theme.system'),
  light: t('theme.light'),
  dark: t('theme.dark'),
}))

const icons: Record<ThemeMode, typeof Monitor> = {
  system: Monitor,
  light: Sun,
  dark: Moon,
}
</script>

<template>
  <div
    class="group flex items-center rounded-md bg-background p-1 transition-all duration-200 hover:gap-1"
  >
    <UiButton
      v-for="option in options"
      :key="option"
      variant="ghost"
      :class="
        cn([
          'h-8 px-0 flex items-center justify-center transition-all duration-200 overflow-hidden',
          theme === option
            ? 'w-8 bg-accent text-accent-foreground'
            : 'w-0 opacity-0 pointer-events-none group-hover:w-8 group-hover:opacity-100 group-hover:pointer-events-auto text-muted-foreground hover:text-foreground',
        ])
      "
      :title="labels[option]"
      @click="setTheme(option)"
    >
      <component :is="icons[option]" class="h-4 w-4 shrink-0" />
      <span class="sr-only">{{ labels[option] }}</span>
    </UiButton>
  </div>
</template>
