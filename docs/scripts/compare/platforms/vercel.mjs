import { driver, waitForWorkingBoard } from '../drive.mjs';

const base = 'https://vercel.com';
const repo = process.env.COMPARE_REPO || 'rolandderroboter/vouch';
const name = repo.split('/')[1];

const project = (target, path = '') => `${base}/${target.team}/${target.project}${path}`;
const store = (target, path = '') => `${base}/${target.team}/~/integrations/neon/${target.store}${path}`;

async function importProject(page) {
  const act = driver(page);
  await act.click(page.getByRole('link', { name: `Import ${name}` }));
  const create = page.getByRole('button', { name: 'Create Project' });
  await create.waitFor({ timeout: 30_000 });
  await act.pause(1_500);
  await act.click(create);
  const deploy = page.getByRole('button', { name: 'Deploy', exact: true });
  await deploy.waitFor({ timeout: 30_000 });
  await act.pause(1_000);
  await act.click(deploy);
  await act.pause(5_000);
  const verify = page.getByText('Verify Your Identity');
  if (await verify.isVisible()) {
    console.log('      Vercel asks for the emailed code, enter it in the window');
    await verify.waitFor({ state: 'hidden', timeout: 10 * 60_000 });
  }
  const team = new URL(page.url()).searchParams.get('teamSlug');
  const done = page.getByRole('button', { name: 'Continue to Project' }).or(page.getByRole('link', { name: 'Continue to Project' })).first();
  await done.waitFor({ timeout: 15 * 60_000 });
  await act.pause(2_500);
  await act.click(done);
  await page.waitForURL((url) => url.pathname === `/${team}/${name}`, { timeout: 60_000 });
  await page.getByText('Ready', { exact: true }).first().waitFor({ timeout: 5 * 60_000 });
  await act.pause(2_000);
  return team;
}

async function addDatabase(page) {
  const act = driver(page);
  await act.click(page.getByRole('link', { name: 'Storage', exact: true }).first());
  await act.click(page.getByRole('button', { name: 'Create Database' }).first());
  const dialog = page.getByRole('dialog').last();
  await act.click(dialog.getByText('Neon', { exact: true }));
  await act.click(dialog.getByRole('button', { name: 'Continue' }));
  await act.click(dialog.getByRole('searchbox', { name: 'Choose your option' }));
  await act.click(page.getByRole('option', { name: /^Frankfurt/ }));
  if (await dialog.getByTestId('toggle/input').first().isChecked()) {
    await act.click(dialog.locator('label').filter({ has: page.getByTestId('toggle/input') }).first());
  }
  await act.click(dialog.getByRole('button', { name: 'Continue' }));
  const resource = dialog.getByRole('textbox', { name: 'Resource Name' });
  await act.click(resource);
  await page.keyboard.press('Meta+a');
  await act.type('vouch-db');
  await act.click(dialog.getByRole('button', { name: 'Create' }));
  await dialog.getByText('Your Neon database is ready to use').waitFor({ timeout: 3 * 60_000 });
  await act.pause(1_000);
  await act.click(dialog.getByRole('button', { name: 'Continue' }));
  const connect = dialog.getByRole('button', { name: 'Connect' });
  for (let waited = 0; !(await connect.isEnabled()) && waited < 60_000; waited += 500) await page.waitForTimeout(500);
  await act.pause(1_000);
  await act.click(connect);
  await page.waitForURL(/\/resources\/storage\/store_/, { timeout: 60_000 });
  await act.pause(2_000);
  return page.url().match(/neon\/(icfg_[^/]+\/resources\/storage\/store_[^/?#]+)/)[1];
}

async function redeploy(page, team) {
  const act = driver(page);
  const back = page.getByRole('link', { name, exact: true }).first();
  if (await back.isVisible().catch(() => false)) await act.click(back);
  else await page.goto(`${base}/${team}/${name}`, { waitUntil: 'load' });
  await act.pause(2_000);
  await act.click(page.getByRole('link', { name: 'Deployments', exact: true }).first());
  await act.pause(2_000);
  await act.click(page.getByRole('button', { name: 'Deployment Actions' }).first());
  await act.click(page.getByRole('menuitem', { name: /Redeploy/ }).first());
  const dialog = page.getByRole('dialog').last();
  await act.click(dialog.getByRole('button', { name: 'Redeploy' }));
  await page.waitForURL((url) => new RegExp(`^/${team}/${name}/[A-Za-z0-9]{20,}$`).test(url.pathname), { timeout: 60_000 });
  await act.pause(2_000);
  return page.url().split('/').pop();
}

async function deploy(page) {
  const act = driver(page);
  const team = await importProject(page);
  const storeID = await addDatabase(page);
  const deployment = await redeploy(page, team);
  await page.getByText('Ready', { exact: true }).first().waitFor({ timeout: 10 * 60_000 });
  const visit = page.getByRole('link', { name: /\.vercel\.app$/ }).last();
  const appURL = `https://${(await visit.innerText()).trim()}`;
  await waitForWorkingBoard(appURL);
  await act.pause(1_500);
  await act.click(visit);
  return { team, project: name, store: storeID, deployment, appURL };
}

async function open(page, url, text) {
  await page.goto(url, { waitUntil: 'load' });
  if (text) await page.getByText(text).first().waitFor({ timeout: 20_000 });
  await page.waitForTimeout(3_000);
}

export default {
  title: 'Vercel',
  home: `${base}/dashboard`,
  targets: ['team', 'project', 'store', 'deployment'],
  async signedIn(page) {
    const url = new URL(page.url());
    return url.hostname === 'vercel.com' && !/^\/(login|signup|new\/?$)/.test(url.pathname) && (await page.getByRole('button', { name: /^(Continue with|Log in)/ }).count()) === 0;
  },
  shots: [
    {
      name: 'deploy',
      title: 'From repository to running app',
      video: true,
      needs: [],
      hint: `Import ${repo}, deploy, add a Neon database from Storage, connect it, redeploy.`,
      open: async (page) => {
        await open(page, `${base}/new`);
        await page.getByRole('link', { name: `Import ${name}` }).waitFor({ timeout: 30_000 });
      },
      drive: deploy,
    },
    {
      name: 'overview',
      title: 'The project',
      needs: ['team', 'project'],
      open: (page, target) => open(page, project(target), 'Production Deployment'),
    },
    {
      name: 'build-logs',
      title: 'Build logs',
      needs: ['team', 'project', 'deployment'],
      open: async (page, target) => {
        await open(page, project(target, `/${target.deployment}`), 'Build Logs');
        const logs = page.getByText('Build Logs', { exact: true }).first();
        await logs.click().catch(() => {});
        await page.waitForTimeout(3_000);
      },
    },
    {
      name: 'runtime-logs',
      title: 'Runtime logs',
      needs: ['team', 'project'],
      open: async (page, target) => {
        await fetch(target.appURL).catch(() => {});
        await open(page, project(target, '/logs'));
        await page.waitForTimeout(6_000);
      },
    },
    {
      name: 'variables',
      title: 'Variables',
      needs: ['team', 'project'],
      open: (page, target) => open(page, project(target, '/settings/environment-variables'), 'DATABASE_URL'),
    },
    {
      name: 'database',
      title: 'Browsing the database',
      needs: ['team', 'store'],
      open: async (page, target) => {
        await open(page, store(target, '/data'));
        await page.getByText('posts', { exact: true }).first().click().catch(() => {});
        await page.waitForTimeout(4_000);
      },
    },
    {
      name: 'database-backups',
      title: 'Database backups',
      needs: ['team', 'store'],
      open: (page, target) => open(page, store(target, '/settings')),
    },
  ],
};
