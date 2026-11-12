import { driver, waitForWorkingBoard } from '../drive.mjs';

const dashboard = 'https://dashboard.render.com';
const repo = process.env.COMPARE_REPO || 'rolandderroboter/vouch';
const region = 'Frankfurt (EU Central)';

const service = (target, path = '') => `${dashboard}/web/${target.service}${path}`;
const database = (target, path = '') => `${dashboard}/d/${target.database}${path}`;

async function newItem(page, name) {
  const act = driver(page);
  await act.click(page.getByRole('button', { name: 'New', exact: true }));
  await act.click(page.getByRole('listbox', { name: 'New' }).getByRole('link', { name, exact: true }));
  await page.waitForLoadState('load');
  await act.pause(1_500);
}

async function pickRegion(page) {
  const act = driver(page);
  const existing = page.getByRole('radio', { name: new RegExp(`^${region.replace(/[()]/g, '\\$&')}`) });
  if (await existing.isVisible().catch(() => false)) {
    if (!(await existing.isChecked())) await act.click(existing);
    return;
  }
  await act.click(page.getByRole('button', { name: /^Region Selector/ }));
  await act.click(page.getByRole('option', { name: region }));
}

async function createDatabase(page) {
  const act = driver(page);
  await newItem(page, 'Postgres');
  await act.click(page.getByRole('textbox', { name: 'Name' }));
  await act.type('vouch-db');
  await pickRegion(page);
  await act.click(page.getByRole('radio', { name: /Free$/ }));
  await act.click(page.getByRole('button', { name: 'Create database' }));
  await page.waitForURL(/\/d\/dpg-/, { timeout: 60_000 });
  const id = page.url().match(/\/d\/(dpg-[^/?#]+)/)[1];
  await page.getByText(/^available$/i).first().waitFor({ timeout: 10 * 60_000 });
  await act.pause(1_500);
  const field = page.getByRole('textbox', { name: 'Internal Database URL' });
  await field.scrollIntoViewIfNeeded();
  const row = page.locator('div').filter({ has: field }).filter({ has: page.getByRole('button', { name: 'Copy' }) }).last();
  await act.click(row.getByRole('button', { name: 'Copy' }));
  const url = await page.evaluate(() => navigator.clipboard.readText());
  if (!/^postgres(ql)?:\/\//.test(url)) throw new Error('the internal database URL did not reach the clipboard');
  return { id, url };
}

async function createService(page, databaseURL) {
  const act = driver(page);
  await newItem(page, 'Web Service');
  await act.click(page.getByRole('tab', { name: 'Public Git Repository' }));
  await act.click(page.getByRole('textbox', { name: 'Repository URL' }));
  await act.type(`https://github.com/${repo}`);
  await act.click(page.getByRole('button', { name: 'Connect' }));
  await page.getByRole('button', { name: 'Deploy web service' }).waitFor({ timeout: 30_000 });
  await act.pause(1_000);
  await pickRegion(page);
  await act.click(page.getByRole('radio', { name: /Free$/ }));
  await act.click(page.getByRole('button', { name: 'Add Environment Variable' }));
  await act.click(page.getByRole('textbox', { name: 'Key' }).last());
  await act.type('DATABASE_URL');
  const value = page.getByRole('textbox', { name: 'Value' }).last();
  await act.click(value);
  await page.keyboard.press('Meta+v');
  await act.pause(400);
  if ((await value.inputValue()) !== databaseURL) await value.fill(databaseURL);
  await act.press('Tab');
  await act.click(page.getByRole('button', { name: 'Deploy web service' }));
  await page.waitForURL(/\/web\/srv-/, { timeout: 60_000 });
  await act.pause(2_000);
  return page.url().match(/\/web\/(srv-[^/?#]+)/)[1];
}

async function appAddress(page, serviceID) {
  const side = await page.context().newPage();
  try {
    await side.goto(`${dashboard}/web/${serviceID}`, { waitUntil: 'load' });
    const link = side.getByRole('link', { name: /\.onrender\.com$/ }).first();
    await link.waitFor({ timeout: 60_000 });
    return (await link.getAttribute('href')).replace(/\/$/, '');
  } finally {
    await side.close();
  }
}

async function deploy(page) {
  const act = driver(page);
  const postgres = await createDatabase(page);
  const serviceID = await createService(page, postgres.url);
  const deployID = page.url().match(/\/deploys\/(dep-[^/?#]+)/)?.[1];
  const appURL = await appAddress(page, serviceID);
  await waitForWorkingBoard(appURL);
  await act.pause(2_000);
  await act.click(page.getByRole('link', { name: 'vouch', exact: true }).first());
  await act.pause(1_500);
  await act.click(page.getByRole('link', { name: /\.onrender\.com$/ }).first());
  return { service: serviceID, database: postgres.id, deploy: deployID, appURL };
}

async function open(page, url, text) {
  await page.goto(url, { waitUntil: 'load' });
  if (text) await page.getByText(text).first().waitFor({ timeout: 20_000 });
  await page.waitForTimeout(3_000);
}

export default {
  title: 'Render',
  home: dashboard,
  targets: ['service', 'deploy', 'database'],
  async tidy(page) {
    const close = page.getByRole('button', { name: 'Close notification' });
    if (await close.isVisible().catch(() => false)) await close.click().catch(() => {});
  },
  async signedIn(page) {
    const url = new URL(page.url());
    return url.hostname === 'dashboard.render.com' && !/^\/(login|register|sign-?up|verify)/.test(url.pathname);
  },
  shots: [
    {
      name: 'deploy',
      title: 'From repository to running app',
      video: true,
      needs: [],
      hint: `Create a free Postgres, then a web service from https://github.com/${repo} with DATABASE_URL set to its internal URL.`,
      prepare: (page) => page.context().grantPermissions(['clipboard-read', 'clipboard-write'], { origin: dashboard }),
      open: (page) => open(page, dashboard),
      drive: deploy,
    },
    {
      name: 'overview',
      title: 'The project',
      needs: [],
      open: (page) => open(page, dashboard, 'vouch-db'),
    },
    {
      name: 'build-logs',
      title: 'Build logs',
      needs: ['service', 'deploy'],
      open: (page, target) => open(page, service(target, `/deploys/${target.deploy}`), /build successful|Your service is live/i),
    },
    {
      name: 'runtime-logs',
      title: 'Runtime logs',
      needs: ['service'],
      open: (page, target) => open(page, service(target, '/logs'), /Ready in|next start/i),
    },
    {
      name: 'variables',
      title: 'Variables',
      needs: ['service'],
      open: (page, target) => open(page, service(target, '/env'), 'Secret Files'),
    },
    {
      name: 'database',
      title: 'Browsing the database',
      needs: ['database'],
      open: (page, target) => open(page, database(target), 'Internal Database URL'),
    },
    {
      name: 'database-backups',
      title: 'Database backups',
      needs: ['database'],
      open: (page, target) => open(page, database(target, '/recovery'), /recovery|backup|export/i),
    },
  ],
};
