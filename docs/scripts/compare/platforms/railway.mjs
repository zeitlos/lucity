import { driver, waitForWorkingBoard } from '../drive.mjs';

const base = 'https://railway.com';
const repo = process.env.COMPARE_REPO || 'rolandderroboter/vouch';

const project = (target) => `${base}/project/${target.project}?environmentId=${target.environment}`;
const service = (target, id, tab = '') => `${base}/project/${target.project}/service/${id}${tab}?environmentId=${target.environment}`;
const deployment = (target, logs) => `${service(target, target.service)}&id=${target.deployment}#${logs}`;

async function lightTheme(page) {
  const dark = await page.evaluate(() => {
    const [r, g, b] = getComputedStyle(document.body).backgroundColor.match(/\d+/g).map(Number);
    return r + g + b < 384;
  });
  if (!dark) return;
  await page.keyboard.press('Meta+k');
  await page.keyboard.type('theme');
  await page.waitForTimeout(600);
  await page.keyboard.press('Enter');
  await page.getByRole('option', { name: 'Light', exact: true }).click();
  await page.waitForTimeout(1_000);
}

function ids(page) {
  const url = new URL(page.url());
  return {
    project: url.pathname.match(/\/project\/([^/]+)/)?.[1],
    service: url.pathname.match(/\/service\/([^/]+)/)?.[1],
    environment: url.searchParams.get('environmentId'),
  };
}

async function deploy(page) {
  await createProject(page);
  const postgres = await addDatabase(page);
  await linkDatabase(page);
  const appURL = await publicDomain(page);
  const found = { ...ids(page), database: (await postgres.getAttribute('href'))?.match(/\/service\/([^/?]+)/)?.[1], appURL };
  await openApp(page, appURL);
  return found;
}

async function createProject(page) {
  const act = driver(page);
  await act.click(page.getByRole('option', { name: 'GitHub Repository', exact: true }));
  await act.type('vouch');
  await act.click(page.getByRole('option', { name: repo, exact: true }));
  await page.waitForURL(/\/project\//, { timeout: 60_000 });
  await act.pause(3_000);
}

async function addDatabase(page) {
  const act = driver(page);
  await act.press('Escape');
  await act.click(page.getByRole('button', { name: 'Add', exact: true }));
  await act.click(page.getByRole('option', { name: 'Database', exact: true }));
  await act.click(page.getByRole('option', { name: 'PostgreSQL', exact: true }));
  const postgres = page.locator("a[href*='/service/']").filter({ hasText: 'Postgres' }).first();
  await postgres.waitFor({ timeout: 60_000 });
  await act.pause(2_000);
  await act.press('Escape');
  return postgres;
}

async function linkDatabase(page) {
  const act = driver(page);
  await act.click(page.locator("a[href*='/service/']").filter({ hasText: 'vouch' }).first());
  await act.click(page.getByRole('link', { name: 'Variables', exact: true }));
  await act.click(page.getByText('New Variable', { exact: true }));
  await act.click(page.getByPlaceholder('VARIABLE_NAME').first());
  await act.type('DATABASE_URL');
  await act.press('Tab');
  const value = page.getByPlaceholder('VALUE or ${{REF}}').first();
  await act.type('${{Postgres.DATABASE_URL}}');
  if ((await value.inputValue()) !== '${{Postgres.DATABASE_URL}}') throw new Error('the DATABASE_URL reference did not land in the value field');
  await act.press('Escape');
  await act.click(value.locator('xpath=ancestor::form[1]').getByRole('button', { name: 'Add', exact: true }));
  await applyChanges(page);
}

async function applyChanges(page) {
  const act = driver(page);
  const staged = page.getByText(/^Apply \d+ changes?$/).first();
  await staged.waitFor({ timeout: 30_000 });
  await act.pause(800);
  const bar = staged.locator('xpath=ancestor::*[.//button[starts-with(normalize-space(), "Deploy")]][1]');
  await act.click(bar.getByRole('button', { name: /^Deploy/ }).first());
  await staged.waitFor({ state: 'hidden', timeout: 30_000 }).catch(() => {
    throw new Error('the staged DATABASE_URL change was not deployed');
  });
  await act.pause(1_500);
}

async function publicDomain(page) {
  const act = driver(page);
  await act.click(page.locator("a[href*='/service/'][href*='/settings']").first());
  await act.click(page.getByRole('button', { name: 'Generate Domain', exact: true }).first());
  const port = page.getByText('Enter the port your app is listening on');
  const domain = page.getByRole('link', { name: /\.up\.railway\.app$/ }).first();
  await Promise.race([port.waitFor({ timeout: 30_000 }).catch(() => {}), domain.waitFor({ timeout: 30_000 }).catch(() => {})]);
  if (await port.isVisible()) {
    const form = port.locator('xpath=ancestor::*[.//button[normalize-space()="Generate Domain"]][1]');
    await act.click(form.locator('input').first());
    await act.type('8080');
    await act.click(form.getByRole('button', { name: 'Generate Domain', exact: true }));
  }
  await domain.waitFor({ timeout: 60_000 });
  return `https://${(await domain.innerText()).trim()}`;
}

async function openApp(page, appURL) {
  const act = driver(page);
  if (await page.getByText(/^Apply \d+ changes?$/).first().isVisible()) throw new Error('changes are still staged, nothing would redeploy');
  await act.click(page.locator("a[href*='/service/']").filter({ hasText: /^Deployments$/ }).first());
  await waitForWorkingBoard(appURL);
  await act.pause(1_500);
  await act.click(page.getByRole('link', { name: /\.up\.railway\.app$/ }).first());
}

export const stages = { createProject, addDatabase, linkDatabase, applyChanges, publicDomain, openApp };

async function open(page, url, text) {
  await page.goto(url, { waitUntil: 'load' });
  if (text) await page.getByText(text).first().waitFor({ timeout: 20_000 });
  await page.waitForTimeout(3_000);
}

export default {
  title: 'Railway',
  home: `${base}/dashboard`,
  targets: ['project', 'environment', 'service', 'database', 'deployment'],
  async signedIn(page) {
    const url = new URL(page.url());
    if (url.hostname !== 'railway.com' || /^\/(login|signup)/.test(url.pathname)) return false;
    return (await page.getByText(/^(Continue with GitHub|Log in using email)$/).count()) === 0;
  },
  shots: [
    {
      name: 'deploy',
      title: 'From repository to running app',
      video: true,
      needs: [],
      hint: `Create a project from GitHub (${repo}), add Postgres, link DATABASE_URL, generate a domain.`,
      prepare: lightTheme,
      open: (page) => open(page, `${base}/new`),
      drive: deploy,
    },
    {
      name: 'overview',
      title: 'The project',
      needs: ['project', 'environment'],
      open: (page, target) => open(page, project(target)),
    },
    {
      name: 'build-logs',
      title: 'Build logs',
      needs: ['project', 'environment', 'service', 'deployment'],
      open: (page, target) => open(page, deployment(target, 'build'), /build logs/i),
    },
    {
      name: 'runtime-logs',
      title: 'Runtime logs',
      needs: ['project', 'environment', 'service', 'deployment'],
      open: (page, target) => open(page, deployment(target, 'deploy'), /deploy logs/i),
    },
    {
      name: 'variables',
      title: 'Variables',
      needs: ['project', 'environment', 'service'],
      open: (page, target) => open(page, service(target, target.service, '/variables'), 'DATABASE_URL'),
    },
    {
      name: 'database',
      title: 'Browsing the database',
      needs: ['project', 'environment', 'database'],
      open: (page, target) => open(page, service(target, target.database, '/data'), 'posts'),
    },
    {
      name: 'database-backups',
      title: 'Database backups',
      needs: ['project', 'environment', 'database'],
      open: (page, target) => open(page, service(target, target.database, '/backups'), /backup/i),
    },
  ],
};
