export default defineNuxtConfig({
  site: {
    url: 'https://lucity.cloud',
    name: 'Lucity'
  },
  modules: ['nuxt-vitalizer'],
  vitalizer: {
    disableStylesheets: 'entry'
  },
  llms: {
    domain: 'https://lucity.cloud',
    title: 'Lucity',
    description: 'Open-source PaaS on Kubernetes with full ejectability. Git push to deploy, environments out of the box, and a real exit door.'
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
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }
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
