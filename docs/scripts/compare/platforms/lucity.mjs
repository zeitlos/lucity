const base = process.env.LUCITY_URL || 'https://lucity.dev';
const dashboard = `${base}/app`;

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
      hint: 'Create a project from GitHub (rolandderroboter/vouch), add a database, link DATABASE_URL.',
      open: async (page) => {
        await page.goto(dashboard, { waitUntil: 'load' });
        await page.waitForTimeout(2_000);
      },
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
