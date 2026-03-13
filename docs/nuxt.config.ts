export default defineNuxtConfig({
  site: {
    url: 'https://lucity.cloud',
    name: 'Lucity'
  },
  nitro: {
    prerender: {
      routes: ['/llms.txt', '/llms-full.txt']
    }
  },
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Caveat:wght@400;700&family=Fira+Code:wght@400;500&family=Instrument+Serif&family=Sora:wght@300..700&family=VT323&display=swap' }
      ],
      meta: [
        { property: 'og:image', content: 'https://lucity.cloud/img/og.jpg' },
        { property: 'og:image:width', content: '1280' },
        { property: 'og:image:height', content: '640' },
        { property: 'og:image:type', content: 'image/jpeg' },
        { property: 'og:type', content: 'website' },
        { property: 'og:site_name', content: 'Lucity' },
        { name: 'twitter:card', content: 'summary_large_image' },
        { name: 'twitter:image', content: 'https://lucity.cloud/img/og.jpg' }
      ]
    }
  }
});
