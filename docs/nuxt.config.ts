export default defineNuxtConfig({
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      script: [
        {
          innerHTML: ";(function(){var sites={'lucity.cloud':'42cfa77ed96d'};var id=sites[location.hostname];if(!id)return;var el=document.createElement('script');el.defer=true;el.src='https://p.lucity.cloud/api/script.js';el.dataset.siteId=id;document.head.appendChild(el);})()",
          type: 'text/javascript'
        }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Caveat:wght@400;700&family=Fira+Code:wght@400;500&family=Instrument+Serif&family=Sora:wght@300..700&family=VT323&display=swap' }
      ]
    }
  }
});
