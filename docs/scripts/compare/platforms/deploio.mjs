import { driver, waitForWorkingBoard } from '../drive.mjs';

const cockpit = 'https://cockpit.nine.ch';
const organisation = process.env.COMPARE_DEPLOIO_ORG || 'n10799';
const repo = process.env.COMPARE_REPO || 'rolandderroboter/vouch';
const base = `${cockpit}/en/${organisation}`;

const projectPage = (target) => `${base}/deploio/management/projects/${target.project}?current_namespace=${target.project}`;
const appPage = (target, tab = 'overview') => `${base}/deploio/apps/applications/${target.app}?namespace=${target.project}&tab=${tab}`;
const databasePage = (target, tab = 'detail') => `${base}/on_demand_services/databases/${target.project}/postgresdatabases/${target.database}?tab=${tab}`;

function stamp() {
  const now = new Date();
  return [now.getMonth() + 1, now.getDate(), now.getHours(), now.getMinutes()].map((part) => String(part).padStart(2, '0')).join('');
}

async function until(page, check, { every = 5_000, timeout = 10 * 60_000 } = {}) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    if (await check()) return;
    await page.waitForTimeout(every);
    await page.reload({ waitUntil: 'load' });
    await page.waitForTimeout(1_000);
  }
  throw new Error('timed out waiting on the Cockpit');
}

async function createProject(page) {
  const act = driver(page);
  const name = `vouch-${stamp()}`;
  await act.click(page.getByRole('link', { name: /New Project/ }));
  await act.click(page.getByRole('textbox', { name: 'Name', exact: true }));
  await act.type(name);
  await act.click(page.getByRole('textbox', { name: 'Display Name', exact: true }));
  await page.keyboard.press('Meta+a');
  await act.type('vouch');
  await act.click(page.getByRole('button', { name: 'Create Project' }));
  const project = `${organisation}-${name}`;
  await page.waitForURL((url) => url.href.includes(project), { timeout: 60_000 }).catch(async () => {
    await page.goto(`${base}/deploio/management/projects/${project}?current_namespace=${project}`, { waitUntil: 'load' });
  });
  await page.getByRole('button', { name: /New Resource/ }).waitFor({ timeout: 60_000 });
  await act.pause(1_500);
  return project;
}

async function newService(page, link) {
  const act = driver(page);
  const service = page.getByRole('link', { name: 'Service', exact: true });
  for (let attempt = 0; attempt < 3 && !(await service.isVisible()); attempt++) {
    await act.click(page.getByRole('button', { name: /New Resource/ }));
    await service.waitFor({ timeout: 2_000 }).catch(() => {});
  }
  await act.click(service);
  await act.click(page.getByRole('main').getByRole('link', { name: link }));
  await page.waitForLoadState('load');
  await act.pause(1_500);
}

async function createDatabase(page) {
  const act = driver(page);
  await newService(page, /^PostgreSQL Database/);
  await act.click(page.getByText('Economy', { exact: true }));
  await act.click(page.getByRole('button', { name: 'Continue' }));
  await act.click(page.getByRole('textbox', { name: 'Name', exact: true }));
  await act.type('vouch-db');
  await act.click(page.getByRole('button', { name: 'Save' }));
  await page.waitForURL(/postgresdatabases\/vouch-db/, { timeout: 60_000 });
  await until(page, async () => (await page.getByRole('cell', { name: 'available', exact: true }).count()) > 0);
  await act.pause(1_500);
  await act.click(page.getByRole('link', { name: 'Access', exact: true }));
  await page.getByRole('rowheader', { name: 'FQDN' }).waitFor({ timeout: 30_000 });
  await act.pause(2_500);
}

async function connectionString(page, project) {
  const side = await page.context().newPage();
  try {
    await side.goto(databasePage({ project, database: 'vouch-db' }, 'access'), { waitUntil: 'load' });
    const cell = (name) => side.getByRole('row').filter({ has: side.getByRole('rowheader', { name }) }).getByRole('cell');
    const user = (await cell('Database Name / Username').innerText()).trim();
    const host = (await cell('FQDN').innerText()).trim();
    await cell('Credentials').getByRole('button', { name: /Show/ }).click();
    const modal = side.getByRole('dialog').last();
    await modal.waitFor({ timeout: 10_000 });
    await modal.getByText('Loading secret').waitFor({ state: 'hidden', timeout: 30_000 });
    const lines = (await modal.innerText()).split('\n').map((line) => line.trim()).filter(Boolean);
    const password = lines[lines.indexOf(user) + 1];
    if (!user || !host || !password || /\s/.test(password) || password === 'Close') throw new Error('could not read the database access details');
    return `postgresql://${user}:${encodeURIComponent(password)}@${host}:5432/${user}?sslmode=no-verify`;
  } finally {
    await side.close();
  }
}

async function createApplication(page, databaseURL) {
  const act = driver(page);
  await act.click(page.getByRole('navigation', { name: 'Breadcrumb' }).getByRole('link').nth(1));
  await page.getByRole('button', { name: /New Resource/ }).waitFor({ timeout: 30_000 });
  await newService(page, /^Application/);
  const name = page.getByRole('textbox', { name: 'Name', exact: true });
  await act.click(name);
  await page.keyboard.press('Meta+a');
  await act.type('vouch');
  await act.click(page.getByRole('textbox', { name: 'URL', exact: true }));
  await act.type(`https://github.com/${repo}`);
  await act.press('Tab');
  await page.getByText('The provided URL is accessible.').waitFor({ timeout: 30_000 });
  await act.pause(1_000);
  const replicas = page.getByRole('spinbutton', { name: 'Replicas' });
  await act.click(replicas);
  await page.keyboard.press('Meta+a');
  await act.type('1');
  await act.click(page.getByRole('button', { name: /Add Variable/ }).first());
  const row = page.getByRole('table').first().getByRole('row').filter({ has: page.getByRole('button', { name: /Remove/ }) }).last();
  await act.click(row.getByRole('textbox').first());
  await act.type('DATABASE_URL');
  await act.click(row.getByRole('textbox').nth(1));
  await row.getByRole('textbox').nth(1).fill(databaseURL);
  await act.click(row.getByRole('checkbox'));
  await act.click(page.getByRole('button', { name: 'Create Application' }));
  await page.waitForURL(/\/applications\/vouch\?/, { timeout: 60_000 });
  await act.pause(2_000);
}

async function deploy(page) {
  const act = driver(page);
  const project = await createProject(page);
  await createDatabase(page);
  const databaseURL = await connectionString(page, project);
  await createApplication(page, databaseURL);
  const host = page.getByRole('link', { name: /\.deploio\.app$/ }).first();
  await until(page, async () => (await host.count()) > 0 && (await page.getByRole('link', { name: /^vouch-build-/ }).count()) > 0);
  const appURL = (await host.getAttribute('href')).replace(/\/$/, '');
  const build = (await page.getByRole('link', { name: /^vouch-build-/ }).first().innerText()).trim();
  await act.click(page.getByRole('link', { name: /^vouch-build-/ }).first());
  await act.pause(2_000);
  await waitForWorkingBoard(appURL, 20 * 60_000);
  await page.goto(appPage({ project, app: 'vouch' }), { waitUntil: 'load' });
  await act.pause(2_000);
  await act.click(page.getByRole('link', { name: /\.deploio\.app$/ }).first());
  return { project, app: 'vouch', database: 'vouch-db', build, appURL };
}

async function open(page, url, text) {
  await page.goto(url, { waitUntil: 'load' });
  if (text) await page.getByText(text).first().waitFor({ timeout: 20_000 });
  await page.waitForTimeout(3_000);
}

export default {
  title: 'Deploio',
  home: `${base}/deploio`,
  targets: ['project', 'app', 'database', 'build'],
  async signedIn(page) {
    const url = new URL(page.url());
    return url.hostname === 'cockpit.nine.ch' && (await page.getByRole('navigation').count()) > 0;
  },
  shots: [
    {
      name: 'deploy',
      title: 'From repository to running app',
      video: true,
      needs: [],
      hint: `Create a project, an Economy Postgres, then an application from https://github.com/${repo} with DATABASE_URL.`,
      open: (page) => open(page, `${base}/deploio`),
      drive: deploy,
    },
    {
      name: 'overview',
      title: 'The project',
      needs: ['project'],
      open: async (page, target) => {
        await open(page, projectPage(target), 'PostgreSQL Databases');
        await page.getByText('Loading...').first().waitFor({ state: 'hidden', timeout: 30_000 }).catch(() => {});
        await page.getByRole('button', { name: /New Resource/ }).evaluate((button) => button.scrollIntoView({ block: 'start' }));
        await page.mouse.wheel(0, -60);
        await page.waitForTimeout(2_000);
      },
    },
    {
      name: 'build-logs',
      title: 'Build logs',
      needs: ['project', 'build'],
      open: async (page, target) => {
        await open(page, `${base}/deploio/apps/builds/${target.build}?namespace=${target.project}&tab=logs`);
        await page.waitForTimeout(5_000);
      },
    },
    {
      name: 'runtime-logs',
      title: 'Runtime logs',
      needs: ['project', 'app'],
      open: async (page, target) => {
        await open(page, appPage(target, 'logs'));
        await page.waitForTimeout(5_000);
      },
    },
    {
      name: 'variables',
      title: 'Variables',
      needs: ['project', 'app'],
      open: async (page, target) => {
        await open(page, appPage(target, 'variables'));
        await page.getByRole('cell', { name: /^DATABASE_URL/ }).first().waitFor({ timeout: 20_000 });
        await page.waitForTimeout(1_000);
      },
    },
    {
      name: 'database',
      title: 'Browsing the database',
      needs: ['project', 'database'],
      open: (page, target) => open(page, databasePage(target, 'access'), 'FQDN'),
    },
    {
      name: 'database-backups',
      title: 'Database backups',
      needs: ['project', 'database'],
      open: (page, target) => open(page, databasePage(target, 'backup'), 'Daily Backups'),
    },
  ],
};
