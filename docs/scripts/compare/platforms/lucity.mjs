import { driver, waitForWorkingBoard } from '../drive.mjs';

const base = process.env.LUCITY_URL || 'https://lucity.cloud';
const dashboard = `${base}/app`;
const repo = process.env.COMPARE_REPO || 'rolandderroboter/vouch';

const visiblePanel = (page) => page.locator('[role="tabpanel"]:visible').last();

async function createProject(page) {
  const act = driver(page);
  await act.click(page.getByRole('button', { name: 'New project' }).first());
  const dialog = page.getByRole('dialog').last();
  await act.click(dialog.getByRole('button', { name: 'GitHub repository' }));
  await act.click(dialog.getByRole('button', { name: new RegExp(`^${repo}\\b`) }));
  await act.click(dialog.getByRole('button', { name: 'Continue' }));
  await act.click(dialog.getByRole('button', { name: /Create service/ }));
  await page.waitForURL(/\/app\/[^/]+\/[^/]+\/[^/~]+$/, { timeout: 60_000 });
  await page.locator('.vue-flow__node[data-id^="service:"]').first().waitFor({ timeout: 60_000 });
  await act.pause(2_000);
  const [workspace, project, environment] = new URL(page.url()).pathname.split('/').slice(2, 5);
  return { workspace, project, environment };
}

async function addDatabase(page) {
  const act = driver(page);
  await act.click(page.getByRole('button', { name: 'New', exact: true }));
  const dialog = page.getByRole('dialog').last();
  await act.click(dialog.getByRole('button', { name: 'PostgreSQL' }));
  await act.click(dialog.getByRole('textbox', { name: 'Database name' }));
  await act.type('db');
  await act.click(dialog.getByRole('button', { name: 'Create database' }));
  await page.locator('.vue-flow__node[data-id$="/db"]').first().waitFor({ timeout: 60_000 });
  await act.pause(2_000);
}

async function linkDatabase(page, service) {
  const act = driver(page);
  await act.click(page.locator(`.vue-flow__node[data-id$="/${service}"]`).first());
  await act.click(page.getByRole('tab', { name: 'Variables', exact: true }));
  await act.click(visiblePanel(page).getByRole('button', { name: 'Add variable' }).first());
  await act.click(visiblePanel(page).getByRole('textbox', { name: 'Key' }).last());
  await act.type('DATABASE_URL');
  await act.click(visiblePanel(page).getByRole('button', { name: 'Link to a dynamic value' }).last());
  await act.click(page.getByRole('option', { name: 'fqdn-uri' }));
  await act.click(visiblePanel(page).getByRole('button', { name: 'Save' }));
  await act.pause(2_500);
}

async function publicDomain(page) {
  const act = driver(page);
  await act.click(page.getByRole('tab', { name: 'Settings', exact: true }));
  await act.click(visiblePanel(page).getByRole('button', { name: 'Generate domain' }));
  const domain = visiblePanel(page).getByText(/\.lucity\.app$/).first();
  await domain.waitFor({ timeout: 60_000 });
  await act.pause(1_000);
  return `https://${(await domain.innerText()).trim()}`;
}

async function deploy(page) {
  const act = driver(page);
  const found = await createProject(page);
  await addDatabase(page);
  await linkDatabase(page, 'vouch');
  const appURL = await publicDomain(page);
  await waitForWorkingBoard(appURL);
  await act.pause(1_500);
  const link = page.getByRole('link', { name: new URL(appURL).host }).first();
  if (await link.isVisible().catch(() => false)) await act.click(link);
  else await page.goto(appURL, { waitUntil: 'load' });
  return { ...found, service: 'vouch', database: 'db', appURL };
}

const environment = (target) => `${dashboard}/${target.workspace}/${target.project}/${target.environment}`;
const panel = (page) => page.locator('.rounded-lg.border.bg-card:has(> header h2)').last();

async function openCanvas(page, target) {
  await page.goto(environment(target), { waitUntil: 'load' });
  await page.locator('.vue-flow__node').first().waitFor({ timeout: 30_000 });
  await page.waitForTimeout(1_500);
}

async function openPanel(page, target, name) {
  await openCanvas(page, target);
  await page.locator(`.vue-flow__node[data-id$="/${name}"]`).first().click();
  await page.locator('[role="tablist"]').first().waitFor({ timeout: 15_000 });
  await page.waitForTimeout(1_000);
}

async function openTab(page, name) {
  await page.getByRole('tab', { name, exact: true }).click();
  await page.waitForTimeout(2_000);
}

export default {
  title: 'Lucity',
  home: dashboard,
  targets: ['workspace', 'project', 'environment', 'service', 'database'],
  async signedIn(page) {
    return page.url().startsWith(dashboard) && (await page.locator('nav, .vue-flow').count()) > 0;
  },
  shots: [
    {
      name: 'deploy',
      title: 'From repository to running app',
      video: true,
      needs: [],
      hint: `Create a project from GitHub (${repo}), add a database, link DATABASE_URL, generate a domain.`,
      open: async (page) => {
        await page.goto(dashboard, { waitUntil: 'load' });
        await page.getByRole('button', { name: 'New project' }).first().waitFor({ timeout: 30_000 });
        await page.waitForTimeout(2_000);
      },
      drive: deploy,
    },
    {
      name: 'overview',
      title: 'The project',
      needs: ['workspace', 'project', 'environment'],
      open: openCanvas,
    },
    {
      name: 'build-logs',
      title: 'Build logs',
      open: async (page, target) => {
        await openPanel(page, target, target.service);
        await openTab(page, 'Deployments');
        await panel(page).getByText(/^active$/i).first().click();
        await page.waitForTimeout(1_500);
        await panel(page).locator('[role="tabpanel"]').getByRole('button', { name: 'Logs', exact: true }).first().click();
        await page.waitForTimeout(3_000);
      },
    },
    {
      name: 'runtime-logs',
      title: 'Runtime logs',
      open: async (page, target) => {
        await openPanel(page, target, target.service);
        await page.locator('[role="tablist"]').getByRole('button', { name: /Logs/ }).click();
        await page.getByText('Runtime logs').first().waitFor({ timeout: 15_000 });
        await page.waitForTimeout(5_000);
      },
    },
    {
      name: 'variables',
      title: 'Variables',
      open: async (page, target) => {
        await openPanel(page, target, target.service);
        await openTab(page, 'Variables');
      },
    },
    {
      name: 'database',
      title: 'Browsing the database',
      open: async (page, target) => {
        await openPanel(page, target, target.database);
        await openTab(page, 'Tables');
        await page.locator('[role="tabpanel"]:visible').last().getByText('posts', { exact: true }).first().click();
        await page.getByText('Dark mode for the board').first().waitFor({ timeout: 20_000 });
        await page.waitForTimeout(1_500);
      },
    },
    {
      name: 'database-backups',
      title: 'Database backups',
      open: async (page, target) => {
        await openPanel(page, target, target.database);
        await openTab(page, 'Backups');
      },
    },
  ],
};
