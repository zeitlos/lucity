export default defineNuxtConfig({
  site: {
    url: 'https://lucity.cloud',
    name: 'Lucity'
  },
  modules: ['nuxt-vitalizer'],
  hooks: {
    'pages:extend'(pages) {
      // Remove the [[lang]]/[...slug] route from the Docus layer.
      // i18n is not enabled, so the optional :lang? prefix just creates
      // duplicate URLs (e.g. /en/getting-started/concepts) that confuse
      // search engines ("Alternate page with proper canonical tag").
      const idx = pages.findIndex(p => p.path === '/:lang?/:slug(.*)*');
      if (idx !== -1) pages.splice(idx, 1);
    }
  },
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
      script: process.env.NODE_ENV === 'production'
        ? [{ src: 'https://p.lucity.cloud/api/script.js', defer: true, 'data-site-id': '42cfa77ed96d' }]
        : [],
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
