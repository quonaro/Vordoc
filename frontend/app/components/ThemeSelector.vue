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
  <div class="flex items-center gap-0.5 rounded-md bg-background p-1">
    <UiButton
      v-for="option in options"
      :key="option"
      variant="ghost"
      :class="
        cn([
          'h-8 w-8 flex items-center justify-center transition-colors',
          theme === option
            ? 'bg-accent text-accent-foreground'
            : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
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
