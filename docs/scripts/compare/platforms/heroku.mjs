const dashboard = 'https://dashboard.heroku.com';

const app = (target, path = '') => `${dashboard}/apps/${target.app}${path}`;

async function open(page, url, text) {
  await page.goto(url, { waitUntil: 'load' });
  if (text) await page.getByText(text).first().waitFor({ timeout: 20_000 });
  await page.waitForTimeout(3_000);
}

export default {
  title: 'Heroku',
  home: `${dashboard}/apps`,
  targets: ['app', 'build', 'addon'],
  async signedIn(page) {
    return new URL(page.url()).hostname === 'dashboard.heroku.com';
  },
  shots: [
    {
      name: 'deploy',
      title: 'From repository to running app',
      video: true,
      needs: [],
      hint: 'Create an app, connect GitHub (rolandderroboter/vouch) on the Deploy tab, add Heroku Postgres, deploy the branch.',
      open: (page) => open(page, `${dashboard}/new-app`),
    },
    {
      name: 'overview',
      title: 'The app',
      needs: ['app'],
      open: (page, target) => open(page, app(target, '/resources'), /dynos|eco|basic/i),
    },
    {
      name: 'build-logs',
      title: 'Build logs',
      needs: ['app', 'build'],
      open: (page, target) => open(page, app(target, `/activity/builds/${target.build}`), /build/i),
    },
    {
      name: 'runtime-logs',
      title: 'Runtime logs',
      needs: ['app'],
      open: async (page, target) => {
        await open(page, app(target, '/logs'));
        await page.waitForTimeout(8_000);
      },
    },
    {
      name: 'variables',
      title: 'Variables',
      needs: ['app'],
      open: (page, target) => open(page, app(target, '/settings'), /config vars/i),
    },
    {
      name: 'database',
      title: 'Browsing the database',
      needs: ['addon'],
      open: (page, target) => open(page, `https://data.heroku.com/datastores/${target.addon}#dataclips`),
    },
    {
      name: 'database-backups',
      title: 'Database backups',
      needs: ['addon'],
      open: (page, target) => open(page, `https://data.heroku.com/datastores/${target.addon}#durability`, /rollback|backup/i),
    },
  ],
};
