export default defineNuxtPlugin(async () => {
  const { load } = useSiteConfig()
  const { load: loadText } = useText()

  await load()
  await loadText()
})
