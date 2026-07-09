import { onMounted, onUnmounted, ref } from 'vue'

export function useOnline() {
  const isOnline = ref(true)

  if (import.meta.client) {
    isOnline.value = navigator.onLine
  }

  const update = () => {
    isOnline.value = navigator.onLine
  }

  onMounted(() => {
    window.addEventListener('online', update)
    window.addEventListener('offline', update)
  })

  onUnmounted(() => {
    window.removeEventListener('online', update)
    window.removeEventListener('offline', update)
  })

  return isOnline
}
