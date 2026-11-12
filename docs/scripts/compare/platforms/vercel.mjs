const base = 'https://vercel.com';

export default {
  title: 'Vercel',
  home: `${base}/dashboard`,
  targets: ['team', 'project'],
  async signedIn(page) {
    const url = new URL(page.url());
    return url.hostname === 'vercel.com' && !/^\/(login|signup|new\/?$)/.test(url.pathname) && (await page.getByRole('button', { name: /^(Continue with|Log in)/ }).count()) === 0;
  },
  shots: [],
};
