import { nextTick, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'
import type { TocItem } from '~/composables/useToc'

const TOP_OFFSET = 80

export function useActiveAnchor(
  contentRef: Ref<HTMLElement | null>,
  tocItems: Ref<TocItem[]>,
) {
  const activeLink = ref<string | null>(null)

  const links = computed(() => collectLinks(tocItems.value))

  watch(
    () => tocItems.value,
    () => {
      nextTick(updateActiveAnchor)
    },
  )

  onMounted(() => {
    window.addEventListener('scroll', updateActiveAnchor, { passive: true })
    updateActiveAnchor()
  })

  onUnmounted(() => {
    window.removeEventListener('scroll', updateActiveAnchor)
  })

  function updateActiveAnchor() {
    const container = contentRef.value
    if (!container) {
      activeLink.value = null
      return
    }

    const headings = links.value
      .map((link) => container.querySelector(`[id="${link.slice(1)}"]`))
      .filter((el): el is HTMLElement => el !== null)

    if (!headings.length) {
      activeLink.value = null
      return
    }

    const scrollY = window.scrollY
    let active: string | null = null

    for (const heading of headings) {
      const top = heading.getBoundingClientRect().top + scrollY - TOP_OFFSET
      if (top <= scrollY + 1) {
        active = `#${heading.id}`
      } else {
        break
      }
    }

    activeLink.value = active
  }

  return activeLink
}

function collectLinks(items: TocItem[]): string[] {
  const result: string[] = []
  for (const item of items) {
    result.push(item.link)
    result.push(...collectLinks(item.children))
  }
  return result
}
