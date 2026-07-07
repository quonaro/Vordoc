<script setup lang="ts">
import { Moon, Sun, Monitor } from '@lucide/vue'
import type { ThemeMode } from '~/composables/useTheme'

const { theme, setTheme, options } = useTheme()

const labels: Record<ThemeMode, string> = {
  system: 'System',
  light: 'Light',
  dark: 'Dark',
}

const icons: Record<ThemeMode, typeof Monitor> = {
  system: Monitor,
  light: Sun,
  dark: Moon,
}
</script>

<template>
  <div class="flex items-center gap-1 rounded-md border bg-background p-1">
    <UiButton
      v-for="option in options"
      :key="option"
      variant="ghost"
      size="icon"
      :class="
        cn([
          'h-8 w-8',
          theme === option
            ? 'bg-accent text-accent-foreground'
            : 'text-muted-foreground hover:text-foreground',
        ])
      "
      :title="labels[option]"
      @click="setTheme(option)"
    >
      <component :is="icons[option]" class="h-4 w-4" />
      <span class="sr-only">{{ labels[option] }}</span>
    </UiButton>
  </div>
</template>
