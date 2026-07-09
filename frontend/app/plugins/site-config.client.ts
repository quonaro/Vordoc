export default defineNuxtPlugin(async () => {
  const { load } = useSiteConfig()
  const { load: loadText } = useText()

  try {
    await load()
  } catch (err) {
    console.error('failed to load site config', err)
  }

  try {
    await loadText()
  } catch (err) {
    console.error('failed to load UI text', err)
  }
})
