const base = 'https://fly.io';

export default {
  title: 'Fly.io',
  home: `${base}/dashboard`,
  targets: ['app'],
  async signedIn(page) {
    const url = new URL(page.url());
    return url.hostname === 'fly.io' && url.pathname.startsWith('/dashboard');
  },
  shots: [],
};
