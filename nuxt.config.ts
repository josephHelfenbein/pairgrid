// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  devtools: { enabled: true },

  routeRules: {
    // prerender index route by default
    '/': { prerender: true },
  },
  runtimeConfig: {
    public:{
      pusherAppKey: process.env.PUSHER_APP_KEY,
    }
  },

  modules:['@nuxtjs/tailwindcss', 'shadcn-nuxt', '@clerk/nuxt'],
  compatibilityDate: '2024-12-20',
  app: {
    head: {
      titleTemplate: 'PairGrid - %s',
      meta: [
        {
          name: 'description',
          content: 'PairGrid is a real-time matchmaking platform that connects developers with similar interests and tech stacks for collaborative coding.',
        },
        {
          name: 'keywords',
          content: 'coding, developers, pair programming, real-time chat, video calls, collaboration, matchmaking',
        },
        {
          property: 'og:title',
          content: 'PairGrid - Find Your Perfect Coding Partner',
        },
        {
          property: 'og:description',
          content: 'PairGrid is a real-time matchmaking platform for developers to collaborate and connect with like-minded coders.',
        },
        {
          property: 'og:url',
          content: 'https://www.pairgrid.com',
        },
        {
          name: 'twitter:title',
          content: 'PairGrid - Find Your Perfect Coding Partner',
        },
        {
          name: 'twitter:description',
          content: 'PairGrid is a real-time matchmaking platform for developers to collaborate and connect with like-minded coders.',
        },
      ],
      link: [
        {
          rel: 'icon',
          type: 'image/png',
          href: '/favicon-96x96.png',
          sizes: '96x96',
        },
        {
          rel: 'icon',
          type: 'image/svg+xml',
          href: '/favicon.svg',
        },
        {
          rel: 'shortcut icon',
          href: '/favicon.ico',
        },
        {
          rel: 'apple-touch-icon',
          sizes: '180x180',
          href: '/apple-touch-icon.png',
        },
        {
          rel: 'manifest',
          href: '/site.webmanifest',
        },
      ],
    },
  },
});