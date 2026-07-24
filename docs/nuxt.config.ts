export default defineNuxtConfig({
  site: {
    url: 'https://lucity.cloud',
    name: 'Lucity'
  },
  fonts: {
    families: [
      { name: 'Mona Sans', provider: 'none' },
      { name: 'Mona Sans Condensed', provider: 'none' },
      { name: 'Redaction', provider: 'none' },
      { name: 'Redaction 10', provider: 'none' },
      { name: 'Redaction 20', provider: 'none' },
      { name: 'Redaction 35', provider: 'none' },
      { name: 'Redaction 50', provider: 'none' },
      { name: 'Redaction 70', provider: 'none' },
      { name: 'Redaction 100', provider: 'none' }
    ]
  },
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
      script: [
        {
          type: 'application/ld+json',
          innerHTML: JSON.stringify({
            '@context': 'https://schema.org',
            '@type': 'Organization',
            '@id': 'https://lucity.cloud/#organization',
            name: 'Lucity',
            url: 'https://lucity.cloud',
            logo: 'https://lucity.cloud/logo-light.png',
            description: 'The European open source alternative to Vercel, Heroku, and Railway. Deploy anything, own your stack.',
            sameAs: [
              'https://github.com/zeitlos/lucity',
              'https://www.linkedin.com/company/lucity/'
            ],
            parentOrganization: {
              '@type': 'Organization',
              name: 'zeitlos',
              sameAs: [
                'https://www.linkedin.com/company/zeitlossoftware',
                'https://github.com/zeitlos'
              ]
            }
          })
        },
        ...(process.env.NODE_ENV === 'production'
          ? [{ src: 'https://p.lucity.cloud/api/script.js', defer: true, 'data-site-id': '42cfa77ed96d' }]
          : [])
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' },
        { rel: 'manifest', href: '/site.webmanifest' }
      ],
      meta: [
        { property: 'og:image', content: 'https://lucity.cloud/img/og.jpg' },
        { property: 'og:image:width', content: '1280' },
        { property: 'og:image:height', content: '640' },
        { property: 'og:image:type', content: 'image/jpeg' },
        { property: 'og:image:alt', content: 'Lucity: Deploy anything. Own your stack.' },
        { property: 'og:type', content: 'website' },
        { property: 'og:site_name', content: 'Lucity' },
        { property: 'og:locale', content: 'en' },
        { name: 'twitter:card', content: 'summary_large_image' },
        { name: 'twitter:image', content: 'https://lucity.cloud/img/og.jpg' },
        { name: 'twitter:image:alt', content: 'Lucity: Deploy anything. Own your stack.' },
        { name: 'theme-color', content: '#301c0e' }
      ]
    }
  }
});
