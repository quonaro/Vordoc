export default defineNuxtConfig({
  ssr: false,
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss'],
  css: ['~/assets/css/main.css'],
  components: [{ path: '~/components/ui', prefix: 'Ui' }, '~/components'],
  app: {
    head: {
      title: 'Vordoc',
    },
  },
  runtimeConfig: {
    public: {
      apiBase: '/api',
    },
  },
  devServer: {
    port: 12301,
  },
  vite: {
    server: {
      allowedHosts: true,
      proxy: {
        '/api': {
          target: 'http://localhost:12300',
          changeOrigin: true,
        },
      },
    },
  },
})
