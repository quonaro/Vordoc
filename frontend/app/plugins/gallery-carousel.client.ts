export default defineNuxtPlugin(async () => {
  const { t, load } = useText()
  await load()

  function scrollGallery(gallery: HTMLElement, direction: number) {
    const item = gallery.querySelector('.gallery__item') as HTMLElement | null
    const gap = parseInt(getComputedStyle(gallery).gap, 10) || 16
    const amount = item ? item.offsetWidth + gap : gallery.clientWidth
    gallery.scrollBy({ left: direction * amount, behavior: 'smooth' })
  }

  function createLightbox() {
    const overlay = document.createElement('div')
    overlay.className = 'gallery-lightbox hidden'
    overlay.setAttribute('role', 'dialog')
    overlay.setAttribute('aria-modal', 'true')

    const img = document.createElement('img')
    img.alt = ''

    const closeButton = document.createElement('button')
    closeButton.type = 'button'
    closeButton.className = 'gallery-lightbox__close'
    closeButton.setAttribute('aria-label', t('gallery.close'))

    overlay.appendChild(img)
    overlay.appendChild(closeButton)
    document.body.appendChild(overlay)

    function close() {
      overlay.classList.add('hidden')
      img.src = ''
      document.removeEventListener('keydown', onKeyDown)
    }

    function open(src: string) {
      img.src = src
      overlay.classList.remove('hidden')
      document.addEventListener('keydown', onKeyDown)
    }

    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') close()
    }

    overlay.addEventListener('click', (e) => {
      if (e.target === overlay) close()
    })
    closeButton.addEventListener('click', (e) => {
      e.stopPropagation()
      close()
    })

    return { open, close }
  }

  const lightbox = createLightbox()

  function decorateGallery(gallery: HTMLElement) {
    if (gallery.dataset.carousel) return
    gallery.dataset.carousel = 'true'

    const wrapper = document.createElement('div')
    wrapper.className = 'gallery-wrapper'
    gallery.parentNode!.insertBefore(wrapper, gallery)
    wrapper.appendChild(gallery)

    gallery.querySelectorAll('img').forEach((img) => {
      const item = document.createElement('div')
      item.className = 'gallery__item'
      img.parentNode!.insertBefore(item, img)
      item.appendChild(img)

      img.setAttribute('role', 'button')
      img.setAttribute('aria-label', t('gallery.openPreview'))
      img.addEventListener('click', () => lightbox.open(img.src))

      const downloadLink = document.createElement('a')
      downloadLink.href = img.src
      downloadLink.download = ''
      downloadLink.className = 'gallery__download'
      downloadLink.setAttribute('aria-label', t('gallery.download'))
      downloadLink.setAttribute('title', t('gallery.download'))
      item.appendChild(downloadLink)
    })

    const prevButton = document.createElement('button')
    prevButton.type = 'button'
    prevButton.className = 'gallery__prev'
    prevButton.setAttribute('aria-label', t('gallery.previous'))

    const nextButton = document.createElement('button')
    nextButton.type = 'button'
    nextButton.className = 'gallery__next'
    nextButton.setAttribute('aria-label', t('gallery.next'))

    wrapper.appendChild(prevButton)
    wrapper.appendChild(nextButton)

    prevButton.addEventListener('click', () => scrollGallery(gallery, -1))
    nextButton.addEventListener('click', () => scrollGallery(gallery, 1))
  }

  const observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      for (const node of mutation.addedNodes) {
        if (node.nodeType !== Node.ELEMENT_NODE) continue
        const el = node as HTMLElement
        if (el.classList.contains('gallery')) {
          decorateGallery(el)
        }
        el.querySelectorAll('.gallery').forEach((gallery) => {
          decorateGallery(gallery as HTMLElement)
        })
      }
    }
  })

  observer.observe(document.body, { childList: true, subtree: true })

  document.querySelectorAll('.gallery').forEach((gallery) => {
    decorateGallery(gallery as HTMLElement)
  })
})
