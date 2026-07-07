import type { RouterScrollBehavior } from 'vue-router'

export default {
  scrollBehavior: ((to, from, savedPosition) => {
    if (to.hash) {
      const el = document.querySelector(to.hash)
      if (el) {
        return { el, top: 0, behavior: 'smooth' }
      }
      return false
    }

    if (savedPosition) {
      return savedPosition
    }

    return { top: 0, behavior: 'smooth' }
  }) satisfies RouterScrollBehavior,
}
