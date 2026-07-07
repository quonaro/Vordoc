export default defineNuxtConfig({
  ssr: false,
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss'],
  css: ['~/assets/css/main.css'],
  components: [{ path: '~/components/ui', prefix: 'Ui' }, '~/components'],
  runtimeConfig: {
    public: {
      apiBase: (import.meta as { dev?: boolean }).dev
        ? 'http://localhost:8080'
        : '/api',
    },
  },
})
