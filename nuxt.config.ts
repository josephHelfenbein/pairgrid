// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  devtools: { enabled: true },

  routeRules: {
    // prerender index route by default
    '/': { prerender: true },
  },

  modules:['@nuxtjs/tailwindcss', 'shadcn-nuxt', '@clerk/nuxt'],
  compatibilityDate: '2024-12-20',
});