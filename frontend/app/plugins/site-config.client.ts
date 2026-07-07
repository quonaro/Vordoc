export default defineNuxtPlugin(async () => {
  const { load } = useSiteConfig()
  await load()
})
