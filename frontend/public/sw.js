const CACHE_VERSION = '1'

const SHELL_CACHE = `vordoc-shell-${CACHE_VERSION}`
const API_CACHE = `vordoc-api-${CACHE_VERSION}`
const ASSETS_CACHE = `vordoc-assets-${CACHE_VERSION}`

const SHELL_URLS = ['/']

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches
      .open(SHELL_CACHE)
      .then((cache) => cache.addAll(SHELL_URLS))
      .then(() => self.skipWaiting()),
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(
          keys
            .filter(
              (key) =>
                key !== SHELL_CACHE &&
                key !== API_CACHE &&
                key !== ASSETS_CACHE,
            )
            .map((key) => caches.delete(key)),
        ),
      )
      .then(() => self.clients.claim()),
  )
})

function isApiRequest(url) {
  return url.pathname.startsWith('/api/v1/')
}

function isStaticAsset(url) {
  if (
    url.pathname.startsWith('/_nuxt/') ||
    url.pathname.startsWith('/fonts/') ||
    url.pathname.startsWith('/assets/')
  ) {
    return true
  }

  return /\.(js|css|svg|png|jpg|jpeg|gif|webp|ico|woff|woff2|ttf|otf|eot)$/i.test(
    url.pathname,
  )
}

self.addEventListener('fetch', (event) => {
  const { request } = event

  if (request.method !== 'GET') {
    return
  }

  const url = new URL(request.url)

  if (isApiRequest(url)) {
    event.respondWith(
      fetch(request)
        .then(async (response) => {
          if (response.ok) {
            const cache = await caches.open(API_CACHE)
            await cache.put(request, response.clone())
          }
          return response
        })
        .catch(async () => {
          const cached = await caches.match(request)
          if (cached) {
            return cached
          }

          return new Response(
            JSON.stringify({ error: 'offline', cached: false }),
            {
              status: 503,
              headers: { 'Content-Type': 'application/json' },
            },
          )
        }),
    )
    return
  }

  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request)
        .then(async (response) => {
          if (response.ok) {
            const cache = await caches.open(SHELL_CACHE)
            await cache.put(request, response.clone())
          }
          return response
        })
        .catch(async () => {
          const cached =
            (await caches.match(request)) ||
            (await caches.match('/')) ||
            (await caches.match('/index.html'))

          if (cached) {
            return cached
          }

          return new Response('Offline', {
            status: 503,
            headers: { 'Content-Type': 'text/plain' },
          })
        }),
    )
    return
  }

  if (isStaticAsset(url)) {
    event.respondWith(
      fetch(request)
        .then(async (response) => {
          if (response.ok) {
            const cache = await caches.open(ASSETS_CACHE)
            await cache.put(request, response.clone())
          }
          return response
        })
        .catch(async () => {
          const cached = await caches.match(request)
          if (cached) {
            return cached
          }

          return new Response('Offline', {
            status: 503,
            headers: { 'Content-Type': 'text/plain' },
          })
        }),
    )
    return
  }

  event.respondWith(
    fetch(request)
      .then(async (response) => {
        if (response.ok) {
          const cache = await caches.open(ASSETS_CACHE)
          await cache.put(request, response.clone())
        }
        return response
      })
      .catch(() => caches.match(request)),
  )
})
