#!/usr/bin/env node
/**
 * Captures the documentation and landing page screenshots from a live dashboard.
 *
 * Auth uses the CLI session: `lucity token` provides the workspace bearer and
 * `lucity token --account` the account token the conductor needs for
 * GitHub-backed calls (the repo picker). Both are only sent to the platform.
 *
 * Usage:
 *   node scripts/capture-screenshots.mjs                 # capture everything
 *   node scripts/capture-screenshots.mjs --list          # show the shot list
 *   node scripts/capture-screenshots.mjs --only canvas-complete,bucket-files
 *   node scripts/capture-screenshots.mjs --headed --slow 250
 *
 * Configuration (env, all optional):
 *   LUCITY_CLI        path to the lucity binary (default: lucity on PATH)
 *   LUCITY_URL        platform URL              (default https://lucity.dev)
 *   LUCITY_APP_PATH   dashboard path on it      (default /app, empty for a local dashboard)
 *   LUCITY_WORKSPACE  workspace id              (default: `lucity workspace`)
 *   SHOTS_PROJECT     demo project id           (default vouch)
 *   SHOTS_ENV         demo environment id       (default development)
 *   SHOTS_SERVICE     demo service              (default vouch)
 *   SHOTS_DATABASE    demo database             (default feedbacks)
 *   SHOTS_BUCKET      demo bucket               (default vouch-uploads)
 *   SHOTS_TABLE       table to open in the database (default posts)
 *   SHOTS_BUCKET_PREFIX folder to open in the bucket (default uploads)
 *   SHOTS_VARIABLE_KEY key whose reference picker is shown (default DATABASE_URL)
 *   SHOTS_HIDE_HOSTS  custom domains left out of the shots (default cf-strict.lucity.ch,lucity.ch)
 *   SHOTS_HIDE_VARIABLES variables left out of every shot (default sali,SPASS)
 *   SHOTS_APP_URL     public URL of the demo app (default: discovered)
 *   SHOTS_REPO_URL    repo used for the fork shot (default zeitlos/vouch)
 *   SHOTS_REPO_OWNER  GitHub account to pick in the palette (default zeitlos)
 *   SHOTS_REPO        repository to filter for (default vouch)
 *   SHOTS_THEME       light | dark              (default light)
 *   SHOTS_BROWSER     installed browser channel, such as chrome (default: Playwright's Chromium)
 *   SHOTS_CHROMIUM    path to a Chromium executable, in place of Playwright's own
 *   SHOTS_METRICS_WINDOW range shown on the database metrics shot (default 24h)
 */

import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const publicDir = join(root, 'public');

const args = process.argv.slice(2);
const flag = (name) => args.includes(`--${name}`);
const option = (name, fallback) => {
  const index = args.indexOf(`--${name}`);
  return index !== -1 && args[index + 1] ? args[index + 1] : fallback;
};

const list = (value) => value.split(',').map((s) => s.trim()).filter(Boolean);

const config = {
  base: (process.env.LUCITY_URL || 'https://lucity.dev').replace(/\/$/, ''),
  appPath: (process.env.LUCITY_APP_PATH ?? '/app').replace(/\/$/, ''),
  workspace: process.env.LUCITY_WORKSPACE || '',
  project: process.env.SHOTS_PROJECT || 'vouch',
  environment: process.env.SHOTS_ENV || 'development',
  service: process.env.SHOTS_SERVICE || 'vouch',
  database: process.env.SHOTS_DATABASE || 'feedbacks',
  bucket: process.env.SHOTS_BUCKET || 'vouch-uploads',
  table: process.env.SHOTS_TABLE || 'posts',
  workspaceNames: list(process.env.SHOTS_WORKSPACE_NAMES || 'Acme,Acme Labs,Side projects'),
  query: process.env.SHOTS_QUERY || 'SELECT title, votes, status FROM posts ORDER BY votes DESC;',
  bucketPrefix: process.env.SHOTS_BUCKET_PREFIX || 'uploads',
  variableKey: process.env.SHOTS_VARIABLE_KEY || 'DATABASE_URL',
  hiddenHosts: list(process.env.SHOTS_HIDE_HOSTS ?? 'cf-strict.lucity.ch,lucity.ch'),
  hiddenVariables: list(process.env.SHOTS_HIDE_VARIABLES ?? 'sali,SPASS'),
  appURL: process.env.SHOTS_APP_URL || '',
  repoURL: process.env.SHOTS_REPO_URL || 'https://github.com/zeitlos/vouch',
  repoOwner: process.env.SHOTS_REPO_OWNER || 'zeitlos',
  repo: process.env.SHOTS_REPO || 'vouch',
  theme: process.env.SHOTS_THEME || option('theme', 'light'),
  browserChannel: process.env.SHOTS_BROWSER || undefined,
  executablePath: process.env.SHOTS_CHROMIUM || undefined,
  metricsWindow: process.env.SHOTS_METRICS_WINDOW || '24h',
  headed: flag('headed'),
  slowMo: Number(option('slow', '0')),
};

// Deliberately small: the dashboard uses fixed pixel type, so a narrow
// viewport makes text read larger in the finished screenshot. The device
// scale factor keeps it sharp.
const viewport = { width: 1120, height: 760 };
const deviceScaleFactor = 2;

// Landing shots are crops of a panel, shown at most 776 css px wide.
const landingViewport = { width: 1600, height: 1000 };
const landingWidth = 776;

// ── Helpers ───────────────────────────────────────────────────────────────

function cli(cliArgs) {
  return execFileSync(process.env.LUCITY_CLI || 'lucity', cliArgs, {
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  }).trim();
}

const app = (path = '') => `${config.base}${config.appPath}${path}`;
const environmentID = () => `${config.workspace}/${config.project}/${config.environment}`;
const environmentURL = () => app(`/${environmentID()}`);

async function graphql(query, variables = {}) {
  const response = await fetch(`${config.base}/graphql`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${tokens.bearer}`,
      'X-Lucity-Workspace': config.workspace,
    },
    body: JSON.stringify({ query, variables }),
  });
  const payload = await response.json();
  if (payload.errors) throw new Error(payload.errors.map((e) => e.message).join(', '));
  return payload.data;
}

const node = (page, name) => page.locator(`.vue-flow__node[data-id$="/${name}"]`).first();

/** Opens the environment canvas and waits for its nodes to settle. */
async function openCanvas(page) {
  await page.goto(environmentURL(), { waitUntil: 'load' });
  await page.locator('.vue-flow__node').first().waitFor({ timeout: 30_000 });
  await page.waitForTimeout(1_500);
}

/** Union of every canvas node, padded, for a tight canvas crop. */
async function nodesRect(page, padding = 48) {
  const boxes = await page.locator('.vue-flow__node').evaluateAll((nodes) =>
    nodes.map((element) => {
      const rect = element.getBoundingClientRect();
      return { left: rect.left, top: rect.top, right: rect.right, bottom: rect.bottom };
    }),
  );
  if (boxes.length === 0) throw new Error('no canvas nodes to frame');

  const { width, height } = page.viewportSize();
  const left = Math.max(0, Math.min(...boxes.map((b) => b.left)) - padding);
  const top = Math.max(0, Math.min(...boxes.map((b) => b.top)) - padding);

  return {
    x: left,
    y: top,
    width: Math.min(Math.max(...boxes.map((b) => b.right)) + padding, width) - left,
    height: Math.min(Math.max(...boxes.map((b) => b.bottom)) + padding, height) - top,
  };
}

/** Opens the detail panel for a canvas node by resource name. */
async function openPanel(page, name) {
  await openCanvas(page);
  const target = node(page, name);
  await target.waitFor({ timeout: 30_000 });
  await target.click();
  await page.locator('[role="tablist"]').first().waitFor({ timeout: 15_000 });
  await page.waitForTimeout(1_000);
  if (page.landing) await fitPanel(page);
}

/** Resizes the viewport until the panel is as wide as the landing page shows it. */
async function fitPanel(page) {
  for (let attempt = 0; attempt < 3; attempt++) {
    const box = await panel(page).boundingBox();
    if (Math.abs(box.width - landingWidth) < 2) return;
    const size = page.viewportSize();
    await page.setViewportSize({ width: Math.round(size.width + landingWidth - box.width), height: size.height });
    await page.waitForTimeout(600);
  }
}

/** Scrolls an element to the top of its scroll container, leaving some room above it. */
async function scrollToTop(locator, room = 24) {
  await locator.evaluate((element, room) => {
    element.scrollIntoView({ block: 'start' });
    let parent = element.parentElement;
    while (parent && parent.scrollHeight <= parent.clientHeight) parent = parent.parentElement;
    if (parent) parent.scrollTop -= room;
  }, room);
  await locator.page().waitForTimeout(600);
}

async function openTab(page, name) {
  await page.getByRole('tab', { name, exact: true }).click();
  await page.waitForTimeout(2_000);
}

/** The detail panel or drawer: the card holding the open title. */
function panel(page) {
  return page.locator('.rounded-lg.border.bg-card:has(> header h2)').last();
}

/** A clip rect around an element, padded and clamped to the viewport. */
async function paddedRect(page, locator, padding = { x: 48, y: 48 }) {
  const box = await locator.boundingBox();
  if (!box) throw new Error('element has no box to clip');

  const { width, height } = page.viewportSize();
  const x = Math.max(0, box.x - padding.x);
  const y = Math.max(0, box.y - padding.y);

  return {
    x,
    y,
    width: Math.min(box.width + padding.x * 2, width - x),
    height: Math.min(box.height + padding.y * 2, height - y),
  };
}

/** The union of two boxes, padded. */
function unionRect(a, b, padding) {
  const left = Math.max(0, Math.min(a.x, b.x) - padding.left);
  const top = Math.max(0, Math.min(a.y, b.y) - padding.top);
  return {
    x: left,
    y: top,
    width: Math.max(a.x + a.width, b.x + b.width) + padding.right - left,
    height: Math.max(a.y + a.height, b.y + b.height) + padding.bottom - top,
  };
}

/**
 * A landing crop: the panel from its tabs down, or from `from` down, at most
 * `height` css px tall.
 */
async function landingRect(page, { from, height }) {
  const box = await panel(page).boundingBox();
  const start = from ? await from.boundingBox() : await panel(page).locator('[role="tablist"]').first().boundingBox();
  const top = from ? start.y - 6 : start.y;
  return {
    x: box.x + 1,
    y: top,
    width: box.width - 2,
    height: Math.min(height, box.y + box.height - 1 - top),
  };
}

/** Expands the active deployment on the Deployments tab. */
async function expandActiveDeployment(page) {
  await openTab(page, 'Deployments');
  await panel(page).getByText(/^active$/i).first().click();
  await page.waitForTimeout(1_500);
}

/** The Logs buttons of the deployment steps, not the panel's Logs tab. */
function stepLogs(page) {
  return panel(page).locator('[role="tabpanel"]').getByRole('button', { name: 'Logs', exact: true });
}

/** Opens a service Settings section by its heading and scrolls it to the top. */
async function settingsSection(page, title, open = true) {
  await openTab(page, 'Settings');
  const heading = panel(page).getByText(title, { exact: true }).first();
  await heading.waitFor({ timeout: 15_000 });
  if (open) {
    const row = panel(page).locator('button[aria-expanded]').filter({ hasText: title }).first();
    if ((await row.count()) > 0 && (await row.getAttribute('aria-expanded')) === 'false') {
      await row.click();
      await page.waitForTimeout(600);
    }
  }
  await heading.evaluate((element) => element.scrollIntoView({ block: 'start' }));
  await page.waitForTimeout(800);
  return heading;
}

// ── Shots ─────────────────────────────────────────────────────────────────

const shots = [
  {
    name: 'fork-button',
    description: 'Fork button on the Vouch repository',
    auth: false,
    viewport: { width: 1000, height: 820 },
    async capture(page) {
      const response = await page.goto(config.repoURL, { waitUntil: 'domcontentloaded' });
      if (!response?.ok() || page.url().includes('/login')) {
        throw new Error(`${config.repoURL} is not public — readers cannot fork it either`);
      }
      await page.locator('#repository-container-header').first().waitFor({ timeout: 20_000 });
      await page.waitForTimeout(1_500);
      return { rect: { x: 0, y: 0, width: 1000, height: 520 } };
    },
  },
  {
    name: 'create-project',
    description: 'Create palette with the GitHub repository picker',
    needsAccountToken: true,
    async capture(page) {
      await page.goto(app(`/${config.workspace}`), { waitUntil: 'load' });
      await page.getByRole('button', { name: 'New project' }).click();
      await page.getByText('GitHub repository', { exact: true }).click();
      await page.waitForTimeout(2_500);
      for (const [text, reason] of [
        ['Connect your GitHub account', 'the account token was rejected — is the CLI session current?'],
        ['No GitHub App installations found', 'the Lucity GitHub App is not installed on any account'],
      ]) {
        if (await page.getByText(text).isVisible().catch(() => false)) throw new Error(reason);
      }

      const palette = page.locator('[role="dialog"]').last();
      const accountPicker = palette.locator('button:visible:has(img)').first();
      if (await accountPicker.isVisible().catch(() => false)) {
        const current = (await accountPicker.innerText()).trim();
        if (!current.startsWith(config.repoOwner)) {
          await accountPicker.click();
          await page.waitForTimeout(750);
          await page.getByRole('menuitem', { name: new RegExp(`^${config.repoOwner}\\b`) }).first().click();
          await page.waitForTimeout(2_000);
        }
      }

      const search = palette.locator('input').first();
      await search.fill(config.repo);
      await page.waitForTimeout(1_500);

      if (await page.getByText('No repositories found.').isVisible().catch(() => false)) {
        throw new Error(`no repository matching "${config.repo}" under ${config.repoOwner}`);
      }
      return { rect: await paddedRect(page, palette.locator('.rounded-xl.border').first(), { x: 40, y: 40 }) };
    },
  },
  {
    name: 'first-build',
    description: 'Build logs of the active deployment',
    async capture(page) {
      await openPanel(page, config.service);
      await expandActiveDeployment(page);
      const logs = stepLogs(page).first();
      await logs.waitFor({ timeout: 15_000 }).catch(() => {
        throw new Error('the active deployment is over a week old and its build logs are gone, deploy it again');
      });
      await logs.click();
      await page.waitForTimeout(5_000);

      // The tail of a finished build is layer digests. Scroll back to the part
      // that reads like a build: dependency install and framework output.
      for (const pattern of [/Creating an optimized production build/, /Compiled successfully/, /Route \(app\)/]) {
        const line = page.getByText(pattern).first();
        if ((await line.count()) > 0) {
          await line.evaluate((element) => element.scrollIntoView({ block: 'center' }));
          await page.waitForTimeout(1_500);
          break;
        }
      }
      return {};
    },
  },
  {
    name: 'generate-domain',
    description: 'Platform domain in the service settings',
    async capture(page) {
      await openPanel(page, config.service);
      await settingsSection(page, 'Networking');
      const domain = panel(page).locator('button[aria-expanded]').filter({ hasText: 'Platform domain' }).first();
      if ((await domain.getAttribute('aria-expanded')) === 'false') await domain.click();
      await domain.evaluate((element) => element.scrollIntoView({ block: 'center' }));
      await page.waitForTimeout(800);
      return {};
    },
  },
  {
    name: 'app-live',
    description: 'The deployed demo app on its platform domain',
    auth: false,
    async capture(page) {
      const url = config.appURL || (await discoverAppURL());
      await page.goto(url, { waitUntil: 'domcontentloaded' });
      await page.waitForTimeout(2_500);
      return {};
    },
  },
  {
    name: 'variable-reference',
    description: 'Variables tab with the reference picker open',
    async capture(page) {
      await openPanel(page, config.service);
      await openTab(page, 'Variables');

      // Detach the row the guide walks through so its picker can be shown,
      // rather than adding a second row with the same key. Nothing is saved:
      // the change lives in the browser until Save, which is never clicked.
      const rows = panel(page).locator('fieldset > div').filter({ has: page.locator('input[placeholder="KEY"]') });
      let target = null;
      for (let index = 0; index < (await rows.count()); index++) {
        const candidate = rows.nth(index);
        if ((await candidate.locator('input[placeholder="KEY"]').inputValue()) === config.variableKey) target = candidate;
      }
      if (!target) throw new Error(`no ${config.variableKey} variable on ${config.service}`);

      const detach = target.locator('[aria-label="Detach from dynamic value"]');
      if (await detach.isVisible().catch(() => false)) {
        await detach.click();
        await page.waitForTimeout(500);
      }
      await target.locator('[aria-label="Link to a dynamic value"]').click();
      await page.getByPlaceholder(/search/i).last().waitFor({ timeout: 10_000 });
      await page.waitForTimeout(800);
      return {};
    },
  },
  {
    name: 'database-tables',
    description: 'Database panel on the Tables tab',
    async capture(page) {
      await openTable(page);
      return {};
    },
  },
  {
    name: 'explorer-tables',
    dir: 'postgres',
    description: 'Database explorer on the Tables tab',
    async capture(page) {
      await openTable(page);
      return {};
    },
  },
  {
    name: 'explorer-query',
    dir: 'postgres',
    description: 'Database explorer running a query',
    async capture(page) {
      await openPanel(page, config.database);
      await openTab(page, 'Query');
      const editor = panel(page).locator('textarea').first();
      await editor.waitFor({ timeout: 15_000 });
      await editor.fill(config.query);
      await panel(page).getByRole('button', { name: /run query/i }).click();
      await page.waitForTimeout(3_000);
      return {};
    },
  },
  {
    name: 'top-queries',
    dir: 'postgres',
    description: 'Top queries on the database Metrics tab, first row opened',
    async capture(page) {
      await openPanel(page, config.database);
      await openTab(page, 'Metrics');
      const table = panel(page).locator('table').filter({ hasText: 'Query' }).last();
      await table.waitFor({ timeout: 20_000 });
      await table.getByRole('button', { name: 'Show details' }).first().click();
      await page.waitForTimeout(800);
      await table.evaluate((element, keep) => {
        let statements = 0;
        for (const row of element.querySelectorAll('tbody tr')) {
          if (row.querySelector('button[aria-expanded]')) statements++;
          if (statements > keep) row.remove();
        }
      }, 6);
      await panel(page).getByText('Top queries').first().evaluate((element) => element.scrollIntoView({ block: 'start' }));
      await page.waitForTimeout(1_000);
      return {};
    },
  },
  {
    name: 'public-access',
    dir: 'postgres',
    description: 'Internet access card with allowed address ranges',
    async capture(page) {
      await answerGraphQL(page, {
        DatabasePublic: {
          database: {
            __typename: 'Database',
            id: `${environmentID()}/${config.database}`,
            public: true,
            allowRules: [
              { __typename: 'DatabaseAllowRule', range: '203.0.113.0/24', description: 'Office' },
              { __typename: 'DatabaseAllowRule', range: '198.51.100.7/32', description: 'CI runner' },
            ],
          },
          clientAddress: '192.0.2.44',
        },
      });
      await openPanel(page, config.database);
      await openTab(page, 'Connect');
      await panel(page).getByText('Office').waitFor({ timeout: 15_000 });
      await page.waitForTimeout(800);
      return {};
    },
  },
  {
    name: 'backups',
    dir: 'postgres',
    description: 'Backups tab with the restore window and recent base backups',
    async capture(page) {
      await answerBackups(page);
      await openPanel(page, config.database);
      await openTab(page, 'Backups');
      await panel(page).getByText(/base backups/i).first().waitFor({ timeout: 15_000 });
      await page.waitForTimeout(1_000);
      return {};
    },
  },
  {
    name: 'files-explorer',
    dir: 'object-storage',
    description: 'Bucket file explorer on the Files tab',
    async capture(page) {
      await openBucketFolder(page);
      return {};
    },
  },
  {
    name: 'bucket-files',
    description: 'Bucket panel on the Files tab',
    async capture(page) {
      await openBucketFolder(page);
      return {};
    },
  },
  {
    name: 'workspace-switcher',
    dir: 'workspaces',
    description: 'Workspace switcher open in the header',
    async capture(page) {
      await openCanvas(page);
      const trigger = page.locator('header [aria-label="Switch workspace"]:visible').first();
      await trigger.click();
      const menu = page.locator('[role="menu"]').first();
      await menu.waitFor({ timeout: 10_000 });
      await page.waitForTimeout(600);

      // The dropdown lists every workspace the account belongs to, which on a
      // real account means other people's names. Swap them for placeholders.
      await menu.evaluate((element, placeholders) => {
        let next = 0;
        for (const item of element.querySelectorAll('[role="menuitem"], [role="menuitemradio"]')) {
          if (item.getAttribute('aria-checked') === 'true' || item.querySelector('svg.lucide-check')) continue;
          const spans = [...item.querySelectorAll('span')].filter((span) => span.children.length === 0);
          const label = spans.find((span) => span.textContent.trim().length > 1);
          if (!label || /new workspace/i.test(item.textContent)) continue;
          if (next >= placeholders.length) { item.remove(); continue; }
          label.textContent = placeholders[next];
          const letter = spans.find((span) => span.textContent.trim().length === 1);
          if (letter) letter.textContent = placeholders[next][0];
          next++;
        }
      }, config.workspaceNames);
      await page.waitForTimeout(200);

      const brand = await page.locator('header').first().boundingBox();
      return { rect: unionRect(brand, await menu.boundingBox(), { left: 0, top: 0, right: 200, bottom: 32 }) };
    },
  },
  {
    name: 'canvas-complete',
    description: 'Project canvas with service, database and bucket',
    async capture(page) {
      await openCanvas(page);
      return { rect: await nodesRect(page) };
    },
  },
  {
    name: 'create-palette',
    dir: 'projects',
    description: 'Create palette with the three ways to start a project',
    async capture(page) {
      await page.goto(app(`/${config.workspace}`), { waitUntil: 'load' });
      await page.getByRole('button', { name: 'New project' }).click();
      await page.getByText('Empty project', { exact: true }).waitFor({ timeout: 15_000 });
      await page.waitForTimeout(800);
      const palette = page.locator('[role="dialog"]').last().locator('.rounded-xl.border').first();
      return { rect: await paddedRect(page, palette, { x: 40, y: 40 }) };
    },
  },
  {
    name: 'canvas',
    dir: 'projects',
    description: 'Environment canvas in the full dashboard',
    async capture(page) {
      await openCanvas(page);
      return {};
    },
  },
  {
    name: 'project-settings',
    dir: 'projects',
    description: 'Project settings drawer',
    async capture(page) {
      await page.goto(app(`/${config.workspace}/${config.project}/~/settings`), { waitUntil: 'load' });
      await page.getByText('Delete project', { exact: true }).waitFor({ timeout: 30_000 });
      await page.waitForTimeout(800);
      return {};
    },
  },
  {
    name: 'eject-dialog',
    dir: 'eject',
    description: 'Eject dialog opened from the project rail',
    async capture(page) {
      await openCanvas(page);
      await page.locator('nav svg.lucide-download, aside svg.lucide-download').first().click();
      const dialog = page.getByRole('dialog').filter({ hasText: 'Eject project' }).first();
      await dialog.waitFor({ timeout: 10_000 });
      await page.waitForTimeout(800);
      return {};
    },
  },
  {
    name: 'environment-switcher',
    dir: 'environments',
    description: 'Environment switcher open in the header, with example environments',
    rewrite(operation, data) {
      const projects = operation === 'ProjectEnvironments' ? [data?.project] : operation === 'ProjectsForNav' ? data?.projects : [];
      const projectID = `${config.workspace}/${config.project}`;
      for (const project of projects ?? []) {
        if (project?.id !== projectID) continue;
        for (const name of ['staging', 'production']) {
          if (project.environments.some((e) => e.name === name)) continue;
          project.environments.push({ __typename: 'Environment', id: `${projectID}/${name}`, name });
        }
      }
    },
    async capture(page) {
      await openCanvas(page);
      await page.locator('header [aria-label="Switch environment"]:visible').first().click();
      const menu = page.locator('[role="menu"]').first();
      await menu.getByText(/new environment/i).waitFor({ timeout: 10_000 });
      await page.waitForTimeout(600);
      const brand = await page.locator('header').first().boundingBox();
      return { rect: unionRect(brand, await menu.boundingBox(), { left: 0, top: 0, right: 0, bottom: 32 }) };
    },
  },
  {
    name: 'service-metrics',
    dir: 'metrics',
    description: 'Service panel on the Metrics tab',
    async capture(page) {
      await openPanel(page, config.service);
      await openTab(page, 'Metrics');
      await panel(page).getByRole('button', { name: '7d', exact: true }).click();
      await panel(page).getByText('Memory').first().waitFor({ timeout: 20_000 });
      await page.waitForTimeout(2_500);
      return {};
    },
  },
  {
    name: 'database-metrics',
    dir: 'metrics',
    description: 'Database panel on the Metrics tab',
    async capture(page) {
      await openDatabaseMetrics(page);
      return {};
    },
  },
  {
    name: 'runtime-logs',
    dir: 'logs',
    description: 'Runtime logs of one running instance next to the service panel',
    async capture(page) {
      await openRuntimeLogs(page);
      const lanes = page.locator('[aria-label^="Show only "]');
      const running = await lanes.evaluateAll((buttons) =>
        buttons.findIndex((button) => {
          let row = button;
          while (row.parentElement && !/Running|Stopped/.test(row.textContent)) row = row.parentElement;
          return /Running/.test(row.textContent) && !/Stopped/.test(row.textContent);
        }),
      );
      await lanes.nth(Math.max(running, 0)).click();
      await page.waitForTimeout(3_000);
      return {};
    },
  },
  {
    name: 'deployment-logs',
    dir: 'logs',
    description: 'Active deployment expanded, with a Logs button per step',
    async capture(page) {
      await openPanel(page, config.service);
      await expandActiveDeployment(page);
      await stepLogs(page)
        .first()
        .waitFor({ timeout: 15_000 })
        .catch(() => {
          throw new Error('the active deployment is over a week old and its step logs are gone, deploy it again');
        });
      await page.waitForTimeout(800);
      return {};
    },
  },
  {
    name: 'billing-settings',
    dir: 'billing',
    description: 'Billing drawer of the workspace, with example figures',
    viewport: { width: 1120, height: 1000 },
    async capture(page) {
      await answerBilling(page, {
        subscription: {
          plan: 'HOBBY',
          status: 'ACTIVE',
          currentPeriodEnd: daysFromNow(17),
          creditAmountCents: 500,
          creditExpiry: null,
          hasPaymentMethod: true,
        },
        usageSummary: { resourceCostCents: 312, creditsCents: 312, estimatedTotalCents: 500 },
      });
      await page.goto(app(`/${config.workspace}/~/billing`), { waitUntil: 'load' });
      await page.getByText('Estimated total', { exact: true }).waitFor({ timeout: 30_000 });
      await page.waitForTimeout(1_000);
      return {};
    },
  },
  {
    name: 'trial-badge',
    dir: 'billing',
    description: 'Trial badge in the header with its popover open, with example figures',
    async capture(page) {
      await answerBilling(page, {
        subscription: {
          plan: null,
          status: 'TRIALING',
          currentPeriodEnd: daysFromNow(9),
          creditAmountCents: 500,
          creditExpiry: daysFromNow(9),
          hasPaymentMethod: false,
        },
        usageSummary: { resourceCostCents: 184, creditsCents: 184, estimatedTotalCents: 0 },
      });
      await openCanvas(page);
      const badge = page.locator('header button:visible').filter({ hasText: /\d+d\b/ }).first();
      await badge.click();
      const popover = page.getByRole('dialog').last();
      await popover.waitFor({ timeout: 10_000 });
      await page.waitForTimeout(600);

      const [a, b] = [await badge.boundingBox(), await popover.boundingBox()];
      const left = Math.max(0, Math.min(a.x, b.x) - 120);
      return {
        rect: {
          x: left,
          y: 0,
          width: page.viewportSize().width - left,
          height: Math.max(a.y + a.height, b.y + b.height) + 28,
        },
      };
    },
  },

  // ── Landing page (webp, panel crops) ─────────────────────────────────────

  {
    name: 'database-metrics',
    landing: true,
    description: 'Landing: database metrics',
    async capture(page) {
      await openDatabaseMetrics(page);
      return { rect: await landingRect(page, { height: 690 }) };
    },
  },
  {
    name: 'variables',
    landing: true,
    description: 'Landing: variables linked to the database and bucket',
    async capture(page) {
      await openPanel(page, config.service);
      await openTab(page, 'Variables');
      const footer = panel(page).getByText(/Available at build time and runtime/).first();
      await footer.waitFor({ timeout: 15_000 });
      const box = await footer.boundingBox();
      const tabs = await panel(page).locator('[role="tablist"]').first().boundingBox();
      return { rect: await landingRect(page, { height: box.y + box.height + 18 - tabs.y }) };
    },
  },
  {
    name: 'scaling',
    landing: true,
    description: 'Landing: scaling and compute settings',
    async capture(page) {
      await openPanel(page, config.service);
      const heading = await settingsSection(page, 'Scaling');
      const replicas = panel(page).locator('button[aria-expanded]').filter({ hasText: 'Replicas' }).first();
      if ((await replicas.getAttribute('aria-expanded')) === 'false') await replicas.click();
      await page.waitForTimeout(600);
      await scrollToTop(heading, 16);
      return { rect: await landingRect(page, { from: heading, height: 545 }) };
    },
  },
  {
    name: 'deployments',
    landing: true,
    description: 'Landing: the Deployments tab',
    async capture(page) {
      await openPanel(page, config.service);
      await openTab(page, 'Deployments');
      return { rect: await landingRect(page, { height: 380 }) };
    },
  },
  {
    name: 'domains',
    landing: true,
    keepHosts: ['lucity.ch'],
    description: 'Landing: platform and custom domains',
    async capture(page) {
      await openPanel(page, config.service);
      await settingsSection(page, 'Networking');
      const platform = panel(page).locator('button[aria-expanded]').filter({ hasText: 'Platform domain' }).first();
      const custom = panel(page).locator('button[aria-expanded]').filter({ hasText: 'Custom domains' }).first();
      for (const row of [platform, custom]) {
        if ((await row.getAttribute('aria-expanded')) === 'false') await row.click();
      }
      await page.waitForTimeout(600);
      const card = platform.locator('xpath=..');
      await scrollToTop(card, 24);
      return { rect: await landingRect(page, { from: card, height: 325 }) };
    },
  },
  {
    name: 'logs',
    landing: true,
    description: 'Landing: the log explorer (dark)',
    async capture(page) {
      await openRuntimeLogs(page);
      await page.getByRole('button', { name: '24h', exact: true }).click();
      await page.waitForTimeout(4_000);
      const box = await page.getByText('Runtime logs', { exact: true }).first().evaluate((element) => {
        let node = element;
        while (node.parentElement && node.getBoundingClientRect().height < 400) node = node.parentElement;
        const rect = node.getBoundingClientRect();
        return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
      });
      return { rect: { x: box.x, y: box.y, width: Math.min(box.width, 734), height: Math.min(box.height, 660) } };
    },
  },
  {
    name: 'security',
    landing: true,
    theme: 'dark',
    description: 'Landing: the Security tab (dark)',
    async capture(page) {
      await openPanel(page, config.service);
      await openTab(page, 'Security');
      await panel(page).getByText(/leaked credentials|leaked secrets/i).first().waitFor({ timeout: 20_000 });
      await page.waitForTimeout(800);
      return { rect: await landingRect(page, { height: 575 }) };
    },
  },
];

async function openTable(page) {
  await openPanel(page, config.database);
  await openTab(page, 'Tables');
  const table = panel(page).getByText(config.table, { exact: true }).first();
  await table.waitFor({ timeout: 20_000 });
  await table.click();
  await page.waitForTimeout(2_500);
}

async function openBucketFolder(page) {
  await openPanel(page, config.bucket);
  await openTab(page, 'Files');
  const prefix = panel(page).getByText(config.bucketPrefix, { exact: true }).first();
  if (await prefix.isVisible().catch(() => false)) {
    await prefix.click();
    await page.waitForTimeout(2_000);
  }
}

async function openDatabaseMetrics(page) {
  await openPanel(page, config.database);
  await openTab(page, 'Metrics');
  await panel(page).getByRole('button', { name: config.metricsWindow, exact: true }).click();
  await panel(page).getByText('Connections').first().waitFor({ timeout: 20_000 });
  await page.waitForTimeout(2_500);
}

async function openRuntimeLogs(page) {
  await openPanel(page, config.service);
  await page.locator('[role="tablist"]').getByRole('button', { name: /Logs/ }).click();
  await page.getByText('Runtime logs').first().waitFor({ timeout: 15_000 });
  await page.waitForTimeout(5_000);
}

async function answerBackups(page) {
  const backup = (started, trigger) => ({
    __typename: 'DatabaseBackup',
    id: `${config.database}-${started.toISOString().slice(0, 16)}`,
    status: 'COMPLETED',
    trigger,
    createdAt: started.toISOString(),
    startedAt: started.toISOString(),
    finishedAt: new Date(started.getTime() + 47_000).toISOString(),
    error: null,
  });
  const sunday = (weeksAgo) => {
    const day = new Date();
    day.setUTCDate(day.getUTCDate() - day.getUTCDay() - 7 * weeksAgo);
    day.setUTCHours(2, 0, 0, 0);
    return day;
  };
  const backups = [
    backup(new Date(Date.now() - 3 * 3_600_000), 'MANUAL'),
    backup(sunday(0), 'SCHEDULED'),
    backup(sunday(1), 'SCHEDULED'),
    backup(sunday(2), 'SCHEDULED'),
  ];
  await answerGraphQL(page, {
    DatabaseBackups: {
      database: {
        __typename: 'Database',
        id: `${environmentID()}/${config.database}`,
        name: config.database,
        backups: {
          __typename: 'DatabaseBackups',
          enabled: true,
          retentionDays: 30,
          earliestRestorePoint: daysFromNow(-18),
          latestRestorePoint: new Date(Date.now() - 120_000).toISOString(),
          lastBackupAt: backups[0].finishedAt,
          backups,
        },
      },
    },
  });
}

async function answerBilling(page, { subscription, usageSummary }) {
  await answerGraphQL(page, {
    Subscription: { subscription: { __typename: 'BillingSubscription', ...subscription } },
    UsageSummary: { usageSummary: { __typename: 'UsageSummary', ...usageSummary } },
  });
}

/** Per-page answers for named GraphQL operations, checked before the context's rewrites. */
async function answerGraphQL(page, answers) {
  page.answers = { ...(page.answers ?? {}), ...answers };
}

function daysFromNow(days) {
  return new Date(Date.now() + days * 86_400_000).toISOString();
}

async function discoverAppURL() {
  const data = await graphql(
    'query($e: EnvironmentID!) { environment(environment: $e) { services { name endpoints { host type } } } }',
    { e: environmentID() },
  );
  const service = data.environment.services.find((s) => s.name === config.service);
  const host = service?.endpoints.find((e) => e.type === 'PLATFORM')?.host;
  if (!host) throw new Error(`service ${config.service} has no platform domain — set SHOTS_APP_URL`);
  return `https://${host}`;
}

// ── Response rewrites ─────────────────────────────────────────────────────

const demoPosts = [
  ['Dark mode for the board', 'Half the team reads this at night and the white background is rough. A dark theme would help.', 48, 'planned', 6],
  ['Keyboard shortcuts for voting', 'Let me arrow through the list and hit V to vote instead of reaching for the mouse.', 31, 'open', 4],
  ['Filter requests by status', 'Past twenty requests it gets hard to see what is actually planned versus still open.', 24, 'in-progress', 3],
  ['Attach images to a request', 'A screenshot explains a layout bug faster than three paragraphs of description.', 21, 'done', 8],
  ['Weekly digest email', 'A Monday summary of new requests and status changes, so nobody has to check the board.', 17, 'open', 2],
  ['Chart tooltips overlap on narrow screens', 'On a 380px viewport the tooltip covers the bar it describes.', 9, 'open', 1],
].map(([title, description, votes, status, daysAgo], index) => ({
  id: String(index + 1),
  title,
  description,
  votes: String(votes),
  status,
  created_at: new Date(Date.now() - daysAgo * 86_400_000).toISOString().replace('T', ' ').replace(/\.\d+Z$/, ' +0000 UTC'),
  image_key: null,
}));

/** Replaces rows of the demo table with the demo posts, keeping the columns the platform returned. */
function demoTable(result) {
  if (!result?.columns?.every((column) => column in demoPosts[0])) return;
  result.rows = demoPosts.map((post) => result.columns.map((column) => post[column]));
  if ('totalEstimatedRows' in result) result.totalEstimatedRows = demoPosts.length;
  if ('affectedRows' in result) result.affectedRows = demoPosts.length;
}

/**
 * Walks a response and tidies what should not be in a public screenshot:
 * test domains, scratch variables, agents on the canvas, and release actors
 * that only resolved to an account id.
 */
function tidy(operation, data, shot) {
  const hiddenHosts = config.hiddenHosts.filter((host) => !(shot.keepHosts ?? []).includes(host));
  if (operation === 'EnvironmentAgents' && data?.environment) data.environment.agents = [];
  if (operation === 'DatabaseTableData') demoTable(data?.databaseTableData);
  if (operation === 'ExecuteQuery') demoTable(data?.executeQuery);
  if (operation === 'DatabaseTables') {
    for (const table of data?.databaseTables ?? []) {
      if (table.name === config.table) table.estimatedRows = demoPosts.length;
    }
  }

  const actors = [];
  const visit = (value) => {
    if (Array.isArray(value)) {
      for (let index = value.length - 1; index >= 0; index--) {
        const item = value[index];
        if (item && typeof item === 'object') {
          if (typeof item.host === 'string' && hiddenHosts.includes(item.host)) { value.splice(index, 1); continue; }
          if (typeof item.key === 'string' && config.hiddenVariables.includes(item.key)) { value.splice(index, 1); continue; }
        }
        visit(item);
      }
      return;
    }
    if (!value || typeof value !== 'object') return;
    if (typeof value.actor === 'string') actors.push(value);
    for (const key of Object.keys(value)) visit(value[key]);
  };
  visit(data);

  const named = actors.map((a) => a.actor).find((actor) => /\s/.test(actor));
  if (named) {
    for (const trigger of actors) {
      if (/^[a-z0-9]{10,}$/.test(trigger.actor)) trigger.actor = named;
    }
  }
  return shot.rewrite ? shot.rewrite(operation, data) ?? data : data;
}

async function handleGraphQL(route, page, shot) {
  const request = route.request();
  const headers = {
    ...request.headers(),
    authorization: `Bearer ${tokens.bearer}`,
    'x-lucity-workspace': config.workspace,
    ...(tokens.account ? { 'x-lucity-account-token': tokens.account } : {}),
  };
  if (request.method() !== 'POST') return route.continue({ headers });

  let body;
  try {
    body = request.postDataJSON();
  } catch {
    return route.continue({ headers });
  }
  const operation = body?.operationName;
  const answer = page?.answers?.[operation];
  if (answer) return route.fulfill({ json: { data: answer } });

  // Only the read-only query runner may write; nothing else reaches the platform.
  if (/^\s*mutation\b/.test(body?.query ?? '') && operation !== 'ExecuteQuery') {
    return route.fulfill({ json: { data: null, errors: [{ message: 'blocked while taking screenshots' }] } });
  }

  const response = await route.fetch({ headers });
  let json;
  try {
    json = await response.json();
  } catch {
    return route.fulfill({ response });
  }
  if (json?.data) json.data = tidy(operation, json.data, shot);
  return route.fulfill({ response, json });
}

// ── Run ───────────────────────────────────────────────────────────────────

const shotKey = (shot) => (shot.landing ? `landing/${shot.name}` : shot.name);

if (flag('list')) {
  for (const shot of shots) console.log(`${shotKey(shot).padEnd(28)} ${shot.description}`);
  process.exit(0);
}

const only = list(option('only', ''));
const selected = only.length > 0 ? shots.filter((s) => only.includes(shotKey(s))) : shots;
if (selected.length === 0) {
  console.error(`no shots match --only ${only.join(',')} (try --list)`);
  process.exit(2);
}

const tokens = { bearer: '', account: '' };

try {
  tokens.bearer = cli(['token']);
} catch {
  console.error('could not get a token from the CLI — run `lucity login` first');
  process.exit(1);
}

try {
  tokens.account = cli(['token', '--account']);
} catch {
  console.warn('note: `lucity token --account` failed, GitHub-backed shots will be skipped');
}
if (tokens.account === tokens.bearer) tokens.account = '';

if (!config.workspace) config.workspace = cli(['workspace']);

console.log(`platform:  ${app()}`);
console.log(`workspace: ${config.workspace}`);
console.log(`demo:      ${config.project}/${config.environment} (${config.service}, ${config.database}, ${config.bucket})`);
console.log('');

const browser = await chromium.launch({
  headless: !config.headed,
  slowMo: config.slowMo,
  channel: config.browserChannel,
  executablePath: config.executablePath,
  args: ['--lang=en-US'],
});

const platformPattern = new RegExp(`^${config.base.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}/(graphql|auth|api|version)`);
const contexts = {};

async function authedContext(theme) {
  if (contexts[theme]) return contexts[theme];
  const context = await browser.newContext({
    viewport,
    deviceScaleFactor,
    locale: 'en-US',
    colorScheme: theme === 'dark' ? 'dark' : 'light',
  });
  await context.addInitScript(
    ({ workspace, theme }) => {
      localStorage.setItem('lucity_workspace', workspace);
      localStorage.setItem('lucity-theme', theme);
      localStorage.setItem(`lucity_onboarding_${workspace}_dismissed`, 'true');
      document.addEventListener('DOMContentLoaded', () => {
        const style = document.createElement('style');
        style.textContent = 'div.banner[role="status"] { display: none !important; }';
        document.head.appendChild(style);
      });
    },
    { workspace: config.workspace, theme },
  );
  await context.routeWebSocket(new RegExp(`${config.base.replace(/^http/, 'ws').replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}/graphql`), (socket) => {
    const server = socket.connectToServer();
    socket.onMessage((message) => {
      try {
        const frame = JSON.parse(message);
        if (frame.type === 'connection_init') {
          frame.payload = { ...(frame.payload ?? {}), Authorization: `Bearer ${tokens.bearer}`, 'X-Lucity-Workspace': config.workspace };
          return server.send(JSON.stringify(frame));
        }
      } catch {}
      server.send(message);
    });
    server.onMessage((message) => socket.send(message));
  });
  contexts[theme] = context;
  return context;
}

const plainContext = await browser.newContext({
  viewport,
  deviceScaleFactor,
  locale: 'en-US',
  colorScheme: config.theme === 'dark' ? 'dark' : 'light',
});

const scratch = mkdtempSync(join(tmpdir(), 'lucity-shots-'));
const captured = [];
const failed = [];

for (const shot of selected) {
  if (shot.needsAccountToken && !tokens.account) {
    failed.push([shotKey(shot), 'needs an account token']);
    console.log(`skip  ${shotKey(shot)} (needs an account token)`);
    continue;
  }

  const context = shot.auth === false ? plainContext : await authedContext(shot.theme ?? config.theme);
  const page = await context.newPage();
  page.landing = Boolean(shot.landing);
  if (shot.auth !== false) await page.route(platformPattern, (route) => handleGraphQL(route, page, shot));

  const relative = shot.landing ? `img/landing/${shot.name}.webp` : `img/docs/${shot.dir ?? 'quickstart'}/${shot.name}.png`;
  const target = join(publicDir, relative);
  mkdirSync(dirname(target), { recursive: true });
  const png = shot.landing ? join(scratch, `${shot.name}.png`) : target;

  await page.setViewportSize(shot.viewport ?? (shot.landing ? landingViewport : viewport));

  try {
    const { rect } = (await shot.capture(page)) ?? {};
    await page.screenshot({ path: png, animations: 'disabled', ...(rect ? { clip: rect } : {}) });
    if (shot.landing) execFileSync('cwebp', ['-quiet', '-q', '88', png, '-o', target]);
    captured.push(shotKey(shot));
    console.log(`ok    ${shotKey(shot)} -> public/${relative}`);
  } catch (error) {
    failed.push([shotKey(shot), error.message.split('\n')[0]]);
    console.log(`fail  ${shotKey(shot)}: ${error.message.split('\n')[0]}`);
  } finally {
    await page.close();
  }
}

await browser.close();
rmSync(scratch, { recursive: true, force: true });

console.log(`\ncaptured ${captured.length}/${selected.length}`);
if (failed.length > 0) {
  console.log('failed:');
  for (const [name, reason] of failed) console.log(`  ${name}: ${reason}`);
  process.exit(1);
}
