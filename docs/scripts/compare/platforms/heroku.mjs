import { driver, waitForWorkingBoard } from '../drive.mjs';

const dashboard = 'https://dashboard.heroku.com';
const repo = process.env.COMPARE_REPO || 'rolandderroboter/vouch';

const inputHolds = (page, value, timeout) =>
  page.waitForFunction((wanted) => [...document.querySelectorAll('input')].some((input) => input.value === wanted), value, { timeout });

const app = (target, path = '') => `${dashboard}/apps/${target.app}${path}`;

function appName() {
  const now = new Date();
  const stamp = `${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}${String(now.getHours()).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}`;
  return `vouch-compare-${stamp}`;
}

async function tab(page, name) {
  const act = driver(page);
  await act.click(page.getByRole('navigation', { name: 'App Navigation Menu' }).getByRole('link', { name, exact: true }));
  await page.waitForLoadState('load');
  await act.pause(2_000);
}

async function createApp(page) {
  const act = driver(page);
  const name = appName();
  await act.click(page.getByRole('button', { name: 'New', exact: true }));
  await act.click(page.getByRole('link', { name: 'Create new app' }));
  await act.click(page.getByRole('textbox', { name: 'App name' }));
  await act.type(name);
  await act.click(page.locator('label.runtime-card').filter({ hasText: 'Europe' }));
  const create = page.getByRole('button', { name: 'Create app' });
  await page.waitForFunction((button) => !button.disabled, await create.elementHandle(), { timeout: 30_000 });
  await act.click(create);
  await page.waitForURL(/\/apps\/[^/]+\/deploy/, { timeout: 60_000 });
  await act.pause(2_000);
  return name;
}

async function connectRepository(page) {
  const act = driver(page);
  await act.click(page.getByRole('link', { name: /^GitHub/ }));
  const search = page.getByRole('searchbox', { name: 'Repository' });
  await act.click(search);
  await act.type(repo.split('/')[1]);
  await act.press('Enter');
  await page.getByText(repo, { exact: true }).waitFor({ timeout: 30_000 });
  await act.click(page.getByRole('button', { name: 'Connect', exact: true }).first());
  await page.getByRole('button', { name: 'Deploy Branch' }).waitFor({ timeout: 60_000 });
  await act.pause(1_500);
}

async function addPostgres(page) {
  const act = driver(page);
  await tab(page, 'Resources');
  const addons = page.getByRole('combobox', { name: 'Quickly add add-ons from Elements' });
  await act.click(addons);
  await act.type('postgres');
  await act.click(page.getByRole('option', { name: 'Heroku Postgres', exact: true }));
  const order = page.getByRole('dialog', { name: 'Online Order Form' });
  await order.getByRole('button', { name: 'Submit Order Form' }).waitFor({ timeout: 30_000 });
  await act.pause(1_500);
  await act.click(order.getByRole('button', { name: 'Submit Order Form' }));
  await order.waitFor({ state: 'hidden', timeout: 60_000 });
  await page.getByRole('link', { name: /Heroku Postgres/ }).first().waitFor({ timeout: 60_000 });
  await act.pause(2_000);
}

async function addConfigVar(page, key, value) {
  const act = driver(page);
  await act.click(page.getByRole('textbox', { name: 'KEY' }));
  await act.type(key);
  await act.click(page.getByRole('textbox', { name: 'VALUE' }));
  await act.type(value);
  await act.click(page.getByRole('table', { name: 'Config Vars' }).getByRole('button', { name: 'Add', exact: true }));
  await inputHolds(page, key, 30_000);
  await act.pause(1_000);
}

async function configVars(page) {
  const act = driver(page);
  await tab(page, 'Settings');
  await act.click(page.getByRole('button', { name: 'Reveal Config Vars' }));
  await inputHolds(page, 'DATABASE_URL', 30_000);
  await addConfigVar(page, 'PGSSLMODE', 'no-verify');
  await addConfigVar(page, 'NPM_CONFIG_PRODUCTION', 'false');
  await act.pause(500);
}

async function deployBranch(page) {
  const act = driver(page);
  await tab(page, 'Deploy');
  await act.click(page.getByRole('button', { name: 'Deploy Branch' }));
  await page.getByText(/Your app was successfully deployed/).waitFor({ timeout: 15 * 60_000 });
  await act.pause(1_500);
}

async function deploy(page) {
  const act = driver(page);
  const name = await createApp(page);
  await connectRepository(page);
  await addPostgres(page);
  await configVars(page);
  await deployBranch(page);
  const view = page.getByRole('link', { name: 'View', exact: true }).last();
  const appURL = (await view.getAttribute('href')).replace(/\/$/, '');
  await waitForWorkingBoard(appURL);
  await act.click(view);
  return { app: name, appURL };
}

async function open(page, url, text) {
  await page.goto(url, { waitUntil: 'load' });
  if (text) await page.getByText(text).first().waitFor({ timeout: 20_000 });
  await page.waitForTimeout(3_000);
}

async function datastore(page, target) {
  await open(page, app(target, '/resources'), 'Heroku Postgres');
  const href = await page.getByRole('link', { name: /Heroku Postgres/ }).first().getAttribute('href');
  const id = href.match(/datastores\/([0-9a-f-]{36})/)?.[1];
  if (!id) throw new Error(`no datastore id in ${href}`);
  return id;
}

export default {
  title: 'Heroku',
  home: `${dashboard}/apps`,
  targets: ['app'],
  async tidy(page) {
    const dismiss = page.getByRole('button', { name: 'Dismiss', exact: true });
    if (await dismiss.isVisible().catch(() => false)) await dismiss.click().catch(() => {});
  },
  async signedIn(page) {
    return new URL(page.url()).hostname === 'dashboard.heroku.com';
  },
  shots: [
    {
      name: 'deploy',
      title: 'From repository to running app',
      video: true,
      needs: [],
      hint: `Create an app, connect GitHub (${repo}) on the Deploy tab, add Heroku Postgres, set PGSSLMODE and NPM_CONFIG_PRODUCTION=false, deploy the branch.`,
      open: (page) => open(page, `${dashboard}/apps`),
      drive: deploy,
    },
    {
      name: 'overview',
      title: 'The app',
      open: (page, target) => open(page, app(target, '/resources'), 'Heroku Postgres'),
    },
    {
      name: 'build-logs',
      title: 'Build logs',
      open: async (page, target) => {
        await open(page, app(target, '/activity'));
        const log = page.getByRole('link', { name: 'View build log' }).first();
        await log.waitFor({ timeout: 20_000 });
        await open(page, `${dashboard}${await log.getAttribute('href')}`, 'Building on the');
      },
    },
    {
      name: 'runtime-logs',
      title: 'Runtime logs',
      open: async (page, target) => {
        await fetch(target.appURL).catch(() => {});
        await open(page, app(target, '/logs'));
        await page.waitForTimeout(8_000);
      },
    },
    {
      name: 'variables',
      title: 'Variables',
      open: async (page, target) => {
        await open(page, app(target, '/settings'), /config vars/i);
        await page.getByRole('button', { name: 'Reveal Config Vars' }).click();
        await inputHolds(page, 'PGSSLMODE', 20_000);
        await page.getByRole('heading', { name: 'Config Vars' }).scrollIntoViewIfNeeded();
        await page.waitForTimeout(1_500);
      },
    },
    {
      name: 'database',
      title: 'Browsing the database',
      open: async (page, target) => {
        const addon = await datastore(page, target);
        await open(page, app(target, `/datastores/${addon}/dataclips`), /dataclip/i);
      },
    },
    {
      name: 'database-backups',
      title: 'Database backups',
      open: async (page, target) => {
        const addon = await datastore(page, target);
        await open(page, app(target, `/datastores/${addon}/durability`), /rollback|backup/i);
      },
    },
  ],
};
