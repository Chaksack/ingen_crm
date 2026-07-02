import tailwindcss from '@tailwindcss/vite'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  // App shell runs as an SPA behind auth; SSR is reserved for future public/portal pages.
  ssr: false,

  modules: ['@pinia/nuxt', '@vite-pwa/nuxt', '@nuxt/icon', '@nuxtjs/color-mode', '@vueuse/nuxt'],

  colorMode: {
    classSuffix: '',
  },

  components: [
    { path: '~/components/ui', pathPrefix: false },
    '~/components',
  ],

  css: ['~/assets/css/tailwind.css'],
  vite: {
    plugins: [tailwindcss()],
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
      wsBase: process.env.NUXT_PUBLIC_WS_BASE || 'ws://localhost:8080',
    },
  },

  pwa: {
    registerType: 'autoUpdate',
    manifest: {
      name: 'Ingen One',
      short_name: 'IngenOne',
      description: 'Unified Business Applications Platform',
      theme_color: '#0f172a',
      background_color: '#0f172a',
      display: 'standalone',
      icons: [
        { src: 'pwa-192x192.png', sizes: '192x192', type: 'image/png' },
        { src: 'pwa-512x512.png', sizes: '512x512', type: 'image/png' },
      ],
    },
    workbox: {
      navigateFallback: '/',
      globPatterns: ['**/*.{js,css,html,png,svg,ico}'],
    },
    devOptions: {
      enabled: true,
    },
  },
})
