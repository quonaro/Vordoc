export default defineNuxtPlugin(async () => {
  const { load } = useRuntimePublic()
  await load()
})
