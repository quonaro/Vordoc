const API_CACHE_NAME = 'vordoc-api-v1'

/**
 * Persist an API response in the browser Cache API so the service worker
 * can serve it when the user is offline.
 */
export async function cacheApiResponse(
  url: string,
  data: unknown,
): Promise<void> {
  if (!import.meta.client || !('caches' in window)) {
    return
  }

  try {
    const cache = await caches.open(API_CACHE_NAME)
    const response = new Response(JSON.stringify(data), {
      status: 200,
      statusText: 'OK',
      headers: { 'Content-Type': 'application/json' },
    })
    const request = new Request(url, { credentials: 'include' })
    await cache.put(request, response)
  } catch (err) {
    console.error('failed to cache API response', err)
  }
}
