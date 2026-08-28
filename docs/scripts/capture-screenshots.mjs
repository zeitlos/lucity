#!/usr/bin/env node
/**
 * Captures the documentation screenshots from the live platform.
 *
 * Auth uses the CLI session: `lucity token` provides the workspace bearer that
 * the browser sends on every request, and `lucity token --account` provides the
 * account token the conductor needs for GitHub-backed calls (the repo picker).
 *
 * Usage:
 *   node scripts/capture-screenshots.mjs                 # capture everything
 *   node scripts/capture-screenshots.mjs --list          # show the shot list
 *   node scripts/capture-screenshots.mjs --only canvas-complete,bucket-files
 *   node scripts/capture-screenshots.mjs --headed --slow 250
 *
 * Configuration (env, all optional):
 *   LUCITY_CLI        path to the lucity binary (default: lucity on PATH)
 *   LUCITY_URL        platform URL              (default https://lucity.cloud)
 *   LUCITY_WORKSPACE  workspace slug            (default: `lucity workspace`)
 *   SHOTS_PROJECT     demo project              (default vouch)
 *   SHOTS_ENV         demo environment          (default development)
 *   SHOTS_SERVICE     demo service              (default vouch)
 *   SHOTS_DATABASE    demo database             (default feedback)
 *   SHOTS_BUCKET      demo bucket               (default attachments)
 *   SHOTS_TABLE       table to open in the database (default posts)
 *   SHOTS_BUCKET_PREFIX folder to open in the bucket (default uploads)
 *   SHOTS_VARIABLE_KEY key typed into the demo variable row (default DATABASE_URL)
 *   SHOTS_APP_URL     public URL of the demo app (default: discovered)
 *   SHOTS_REPO_URL    repo used for the fork shot (default zeitlos/vouch)
 *   SHOTS_REPO_OWNER  GitHub account to pick in the palette (default zeitlos)
 *   SHOTS_REPO        repository to filter for (default vouch)
 *   SHOTS_USER_NAME   display name to show in the dashboard header
 *   SHOTS_USER_EMAIL  email to show in the dashboard header
 *   SHOTS_THEME       light | dark              (default light)
 */

import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync, writeFileSync, mkdirSync, unlinkSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const imagesDir = join(root, 'public/img/docs');
const outputDir = join(imagesDir, 'quickstart');

/** Shots default to the quickstart folder; `dir` puts them somewhere else. */
function shotDir(shot) {
  return join(imagesDir, shot.dir ?? 'quickstart');
}

function shotPath(shot) {
  return `img/docs/${shot.dir ?? 'quickstart'}/${shot.name}.png`;
}
const quickstartPage = join(root, 'content/01.quickstart.md');

const args = process.argv.slice(2);
const flag = (name) => args.includes(`--${name}`);
const option = (name, fallback) => {
  const index = args.indexOf(`--${name}`);
  return index !== -1 && args[index + 1] ? args[index + 1] : fallback;
};

const config = {
  base: (process.env.LUCITY_URL || 'https://lucity.cloud').replace(/\/$/, ''),
  workspace: process.env.LUCITY_WORKSPACE || '',
  project: process.env.SHOTS_PROJECT || 'vouch',
  environment: process.env.SHOTS_ENV || 'development',
  service: process.env.SHOTS_SERVICE || 'vouch',
  database: process.env.SHOTS_DATABASE || 'feedback',
  bucket: process.env.SHOTS_BUCKET || 'attachments',
  table: process.env.SHOTS_TABLE || 'posts',
  query: process.env.SHOTS_QUERY || 'SELECT title, votes, status FROM posts ORDER BY votes DESC;',
  bucketPrefix: process.env.SHOTS_BUCKET_PREFIX || 'uploads',
  variableKey: process.env.SHOTS_VARIABLE_KEY || 'DATABASE_URL',
  appURL: process.env.SHOTS_APP_URL || '',
  repoURL: process.env.SHOTS_REPO_URL || 'https://github.com/zeitlos/vouch',
  repoOwner: process.env.SHOTS_REPO_OWNER || 'zeitlos',
  repo: process.env.SHOTS_REPO || 'vouch',
  theme: process.env.SHOTS_THEME || option('theme', 'light'),
  headed: flag('headed'),
  slowMo: Number(option('slow', '0')),
  updateDocs: !flag('no-update-docs'),
  prunePlaceholders: flag('prune-placeholders'),
};

// Deliberately small: the dashboard uses fixed pixel type, so a narrow
// viewport makes text read larger in the finished screenshot. The device
// scale factor keeps it sharp.
const viewport = { width: 1120, height: 760 };
const deviceScaleFactor = 2;

// ── Helpers ───────────────────────────────────────────────────────────────

function cli(cliArgs) {
  return execFileSync(process.env.LUCITY_CLI || 'lucity', cliArgs, {
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  }).trim();
}

function environmentID() {
  return `${config.workspace}/${config.project}/${config.environment}`;
}

function environmentURL() {
  const project = encodeURIComponent(`${config.workspace}/${config.project}`);
  return `${config.base}/app/projects/${project}/environments/${encodeURIComponent(environmentID())}`;
}

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

/** Opens the environment canvas and waits for its nodes to settle. */
async function openCanvas(page) {
  if (!page.url().startsWith(environmentURL())) {
    await page.goto(environmentURL(), { waitUntil: 'domcontentloaded' });
  }
  await page.locator('.vue-flow__node').first().waitFor({ timeout: 30_000 });
  await page.waitForTimeout(1_500);

}

/** Drags a canvas node until its top-left sits on the given viewport point. */
async function placeNode(page, name, to) {
  const node = page.locator(`.vue-flow__node[data-id$="/${name}"]`).first();

  for (let attempt = 0; attempt < 3; attempt++) {
    const box = await node.boundingBox();
    if (!box) return;

    const dx = to.x - box.x;
    const dy = to.y - box.y;
    if (Math.abs(dx) < 6 && Math.abs(dy) < 6) return;

    // Grab low on the card: the title row selects instead of dragging.
    const grabX = box.x + 40;
    const grabY = box.y + box.height - 20;
    await page.mouse.move(grabX, grabY);
    await page.mouse.down();
    await page.mouse.move(grabX + dx, grabY + dy, { steps: 12 });
    await page.mouse.up();
    await page.waitForTimeout(400);
  }
}

/** Trims the outer edges of the viewport, keeping the frame's shape. */
function viewportRect(page, { left = 0, top = 0, right = 0, bottom = 0 } = {}) {
  const { width, height } = page.viewportSize();
  return { x: left, y: top, width: width - left - right, height: height - top - bottom };
}

/** Union of every canvas node, padded, for a tight canvas crop. */
async function nodesRect(page, padding = 56) {
  const boxes = await page.locator('.vue-flow__node').evaluateAll((nodes) =>
    nodes.map((node) => {
      const rect = node.getBoundingClientRect();
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
  const node = page.locator(`.vue-flow__node[data-id$="/${name}"]`).first();
  await node.waitFor({ timeout: 30_000 });
  await node.click();
  await page.locator('[role="tablist"]').first().waitFor({ timeout: 15_000 });
  await page.waitForTimeout(1_000);
}

async function openTab(page, name) {
  await page.getByRole('tab', { name, exact: true }).click();
  await page.waitForTimeout(1_500);
}

/** The detail panel, used as the clip region for panel shots. */
function panel(page) {
  return page.locator('div.rounded-lg.border.bg-card:has(h2):has([role="tablist"])').last();
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

/**
 * A clip rect covering an element down to where its content actually ends.
 * Panels are full height, so a plain element screenshot leaves dead space
 * under short tabs. Leaf elements approximate the real content bottom.
 */
async function contentRect(page, locator, padding = 24) {
  const box = await locator.boundingBox();
  if (!box) throw new Error('element has no box to clip');

  const contentBottom = await locator.evaluate((element) => {
    let bottom = 0;
    for (const node of element.querySelectorAll('*')) {
      if (node.children.length > 0) continue;

      // Only things that actually draw: text, icons, images, form controls.
      // Empty layout fillers stretch to the panel bottom and would defeat this.
      const paints =
        node.textContent.trim().length > 0 ||
        ['IMG', 'SVG', 'INPUT', 'TEXTAREA'].includes(node.tagName.toUpperCase());
      if (!paints) continue;

      const rect = node.getBoundingClientRect();
      if (rect.width === 0 || rect.height === 0) continue;
      bottom = Math.max(bottom, rect.bottom);
    }
    return bottom;
  });

  const viewportHeight = page.viewportSize().height;
  const height = Math.min(box.height, Math.max(contentBottom + padding - box.y, 240));

  return {
    x: Math.max(0, box.x),
    y: Math.max(0, box.y),
    width: box.width,
    height: Math.min(height, viewportHeight - box.y),
  };
}

// ── Shots ─────────────────────────────────────────────────────────────────

const explorerCrop = { left: 110, top: 8, right: 10, bottom: 70 };
const explorerRect = (page) => viewportRect(page, explorerCrop);

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
      await page.goto(`${config.base}/app/`, { waitUntil: 'domcontentloaded' });
      await page.getByRole('button', { name: 'New', exact: true }).click();
      await page.getByText('GitHub Repository', { exact: true }).click();
      await page.waitForTimeout(2_500);
      for (const [text, reason] of [
        ['Connect your GitHub account', 'the account token was rejected — is the CLI session current?'],
        ['No GitHub App installations found', 'the Lucity GitHub App is not installed on any account'],
      ]) {
        if (await page.getByText(text).isVisible().catch(() => false)) throw new Error(reason);
      }
      // Switch to the account that owns the demo repo, then filter for it.
      // The account row is the avatar button inside the palette; the only other
      // avatar button on the page is the header's single-letter user menu.
      const accountPicker = page
        .locator('button:visible:has(img)')
        .filter({ hasNotText: /^\s*\S\s*$/ })
        .first();

      if (await accountPicker.isVisible().catch(() => false)) {
        const current = (await accountPicker.innerText()).trim();
        if (!current.startsWith(config.repoOwner)) {
          await accountPicker.click();
          await page.waitForTimeout(750);
          await page.getByRole('button', { name: new RegExp(`^${config.repoOwner}\\b`) }).last().click();
          await page.waitForTimeout(2_000);
        }
      }

      const search = page.getByPlaceholder(/search/i).first();
      if (await search.isVisible().catch(() => false)) {
        await search.fill(config.repo);
        await page.waitForTimeout(1_500);
      }

      if (await page.getByText('No repositories found.').isVisible().catch(() => false)) {
        throw new Error(`no repository matching "${config.repo}" under ${config.repoOwner}`);
      }

      const palette = page.locator('div.rounded-xl.border.bg-popover').first();
      return { rect: await paddedRect(page, palette, { x: 40, y: 40 }) };
    },
  },
  {
    name: 'first-build',
    description: 'Build logs streaming from the active deployment',
    async capture(page) {
      await openPanel(page, config.service);
      await openTab(page, 'Deployments');

      // Expand the active deployment, then open the logs of its build step.
      await page.getByText('ACTIVE').first().click();
      await page.waitForTimeout(1_000);

      // Steps are Build, Secrets, Deploy, Rollout; the guide is about the build.
      const buildStep = page
        .locator('div')
        .filter({ hasText: /^Build/ })
        .filter({ has: page.getByRole('button', { name: 'Logs', exact: true }) })
        .last();
      const logs = buildStep.getByRole('button', { name: 'Logs', exact: true }).first();
      await logs.waitFor({ timeout: 15_000 });
      await logs.click();
      await page.waitForTimeout(5_000);

      // The tail of a finished build is layer digests. Scroll back to the part
      // that reads like a build: dependency install and framework output.
      const interesting = page
        .getByText(/Creating an optimized production build|Compiled successfully|Route \(app\)|added \d+ packages/)
        .first();
      if (await interesting.isVisible().catch(() => false)) {
        await interesting.scrollIntoViewIfNeeded();
        await page.waitForTimeout(1_500);
      }
      return {};
    },
  },
  {
    name: 'generate-domain',
    description: 'Platform Domain section of the service settings',
    async capture(page) {
      await openPanel(page, config.service);
      await openTab(page, 'Settings');
      const heading = page.getByText('Platform Domain', { exact: true }).first();
      await heading.waitFor({ timeout: 15_000 });
      await heading.scrollIntoViewIfNeeded();
      await page.waitForTimeout(800);
      // Frame whole cards around the domain section so no row is sliced.
      const cardFor = (label) =>
        page.getByText(label, { exact: true }).first().locator('xpath=ancestor::div[contains(@class,"rounded-lg")][1]');

      const top = await cardFor('Health Check').boundingBox();
      const bottom = await cardFor('Custom Domains').boundingBox();
      const panelBox = await panel(page).boundingBox();
      const viewportHeight = page.viewportSize().height;

      const y = Math.max(0, top.y - 16);
      return {
        rect: {
          x: panelBox.x,
          y,
          width: panelBox.width,
          height: Math.min(bottom.y + bottom.height + 16 - y, viewportHeight - y),
        },
      };
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
      const rows = page.locator('div.flex.items-center.gap-2:has(input[placeholder="KEY"])');
      let row = null;
      for (let index = 0; index < (await rows.count()); index++) {
        const candidate = rows.nth(index);
        const key = await candidate.locator('input[placeholder="KEY"]').inputValue();
        if (key === config.variableKey) {
          row = candidate;
          break;
        }
      }
      if (!row) throw new Error(`no ${config.variableKey} variable on ${config.service}`);

      const detach = row.locator('button[title="Detach from dynamic value"]');
      if (await detach.isVisible().catch(() => false)) {
        await detach.click();
        await page.waitForTimeout(500);
      }

      const link = row.locator('button[title="Link to a dynamic value"]');
      await link.waitFor({ timeout: 15_000 });
      await link.click();
      await page.getByPlaceholder('Search references...').waitFor({ timeout: 10_000 });
      await page.waitForTimeout(800);
      return {};
    },
  },
  {
    name: 'database-tables',
    description: 'Database panel on the Tables tab',
    async capture(page) {
      await openPanel(page, config.database);
      await openTab(page, 'Tables');
      const table = page.getByText(config.table, { exact: true }).last();
      if (await table.isVisible().catch(() => false)) {
        await table.click();
        await page.waitForTimeout(2_000);
      }
      return { rect: await contentRect(page, panel(page)) };
    },
  },
  // Both explorer shots crop identically, trimming the empty canvas on the left
  // and the strip below the panel while keeping the viewport's proportions.
  // The explorer shots keep the whole dashboard in frame: the panel on its own
  // reads as a floating card, and a tall crop eats vertical space in the docs.
  {
    name: 'explorer-tables',
    dir: 'postgres',
    description: 'Database explorer on the Tables tab',
    async capture(page) {
      await openPanel(page, config.database);
      await openTab(page, 'Tables');
      const table = page.getByText(config.table, { exact: true }).last();
      if (await table.isVisible().catch(() => false)) {
        await table.click();
        await page.waitForTimeout(2_000);
      }
      return { rect: explorerRect(page) };
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
      await panel(page).getByRole('button', { name: 'Run Query' }).click();
      await page.waitForTimeout(3_000);
      return { rect: explorerRect(page) };
    },
  },
  {
    name: 'files-explorer',
    dir: 'object-storage',
    description: 'Bucket file explorer on the Files tab',
    async capture(page) {
      await openPanel(page, config.bucket);
      await openTab(page, 'Files');
      const prefix = panel(page)
        .locator('button')
        .filter({ hasText: new RegExp(`^${config.bucketPrefix}$`) })
        .first();
      if (await prefix.isVisible().catch(() => false)) {
        await prefix.click();
        await page.waitForTimeout(2_000);
      }
      return { rect: explorerRect(page) };
    },
  },
  {
    name: 'bucket-files',
    description: 'Bucket panel on the Files tab',
    async capture(page) {
      await openPanel(page, config.bucket);
      await openTab(page, 'Files');
      const prefix = panel(page)
        .locator('button')
        .filter({ hasText: new RegExp(`^${config.bucketPrefix}$`) })
        .first();
      if (await prefix.isVisible().catch(() => false)) {
        await prefix.click();
        await page.waitForTimeout(2_000);
      }
      return { rect: await contentRect(page, panel(page)) };
    },
  },
  {
    name: 'canvas-complete',
    description: 'Project canvas with service, database and bucket',
    viewport: { width: 1280, height: 900 },
    async capture(page) {
      await openCanvas(page);

      // The default layout spreads the resources wider than a documentation
      // image wants, and the dashboard's own fit-view control does not move
      // the pane, so lay them out in two columns by hand.
      await placeNode(page, config.service, { x: 120, y: 250 });
      await placeNode(page, config.database, { x: 560, y: 200 });
      await placeNode(page, config.bucket, { x: 560, y: 440 });

      // Dragging leaves the last node selected; click empty canvas to clear it.
      await page.mouse.click(1_180, 820);
      await page.waitForTimeout(800);

      return { rect: await nodesRect(page) };
    },
  },
];

async function discoverAppURL() {
  const data = await graphql(
    'query($e: EnvironmentID!) { environment(environment: $e) { services { name endpoints { host } } } }',
    { e: environmentID() },
  );
  const service = data.environment.services.find((s) => s.name === config.service);
  const host = service?.endpoints.map((e) => e.host).find((h) => !h.endsWith('.svc.cluster.local'));
  if (!host) throw new Error(`service ${config.service} has no public domain — set SHOTS_APP_URL`);
  return `https://${host}`;
}

// ── Doc references ────────────────────────────────────────────────────────

function pointDocsAt(names) {
  if (!existsSync(quickstartPage)) return [];
  let markdown = readFileSync(quickstartPage, 'utf8');
  const switched = [];

  for (const name of names) {
    const placeholder = `/img/docs/quickstart/${name}.svg`;
    if (!markdown.includes(placeholder)) continue;
    markdown = markdown.split(placeholder).join(`/img/docs/quickstart/${name}.png`);
    switched.push(name);
  }

  if (switched.length > 0) writeFileSync(quickstartPage, markdown);
  return switched;
}

function prunePlaceholder(shot) {
  const placeholder = join(shotDir(shot), `${shot.name}.svg`);
  if (existsSync(placeholder)) unlinkSync(placeholder);
}

// ── Run ───────────────────────────────────────────────────────────────────

if (flag('list')) {
  for (const shot of shots) console.log(`${shot.name.padEnd(20)} ${shot.description}`);
  process.exit(0);
}

const only = option('only', '')
  .split(',')
  .map((s) => s.trim())
  .filter(Boolean);

const selected = only.length > 0 ? shots.filter((s) => only.includes(s.name)) : shots;
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

if (tokens.account === tokens.bearer) {
  // A CLI without `token --account` ignores the flag and prints the workspace
  // bearer instead, which the Account API rejects.
  tokens.account = '';
  console.warn('note: this lucity build has no `token --account`, GitHub-backed shots will be skipped');
  console.warn('      build a current one and point LUCITY_CLI at it');
}

if (!config.workspace) {
  config.workspace = cli(['workspace']);
}

mkdirSync(outputDir, { recursive: true });

console.log(`platform:  ${config.base}`);
console.log(`workspace: ${config.workspace}`);
console.log(`demo:      ${config.project}/${config.environment} (${config.service}, ${config.database}, ${config.bucket})`);
console.log('');

const browser = await chromium.launch({ headless: !config.headed, slowMo: config.slowMo });

const authedContext = await browser.newContext({
  viewport,
  deviceScaleFactor,
  colorScheme: config.theme === 'dark' ? 'dark' : 'light',
  extraHTTPHeaders: {
    Authorization: `Bearer ${tokens.bearer}`,
    'X-Lucity-Workspace': config.workspace,
    ...(tokens.account ? { 'X-Lucity-Account-Token': tokens.account } : {}),
  },
});

await authedContext.addInitScript(
  ({ workspace, theme }) => {
    localStorage.setItem('lucity_workspace', workspace);
    localStorage.setItem('lucity-theme', theme);
    localStorage.setItem(`lucity_onboarding_${workspace}_dismissed`, 'true');
  },
  { workspace: config.workspace, theme: config.theme },
);

if (process.env.SHOTS_USER_NAME || process.env.SHOTS_USER_EMAIL) {
  await authedContext.route('**/auth/me', async (route) => {
    const response = await route.fetch();
    const me = await response.json();
    route.fulfill({
      response,
      json: {
        ...me,
        name: process.env.SHOTS_USER_NAME || me.name,
        email: process.env.SHOTS_USER_EMAIL || me.email,
      },
    });
  });
}

const plainContext = await browser.newContext({
  viewport,
  deviceScaleFactor,
  colorScheme: config.theme === 'dark' ? 'dark' : 'light',
});

const captured = [];
const failed = [];

for (const shot of selected) {
  if (shot.needsAccountToken && !tokens.account) {
    failed.push([shot.name, 'needs an account token']);
    console.log(`skip  ${shot.name} (needs an account token)`);
    continue;
  }

  const context = shot.auth === false ? plainContext : authedContext;
  const page = await context.newPage();
  const dir = shotDir(shot);
  mkdirSync(dir, { recursive: true });
  const target = join(dir, `${shot.name}.png`);

  if (shot.viewport) await page.setViewportSize(shot.viewport);

  try {
    const { clip, rect } = (await shot.capture(page)) ?? {};
    if (clip) {
      await clip.screenshot({ path: target, animations: 'disabled' });
    } else {
      await page.screenshot({ path: target, animations: 'disabled', ...(rect ? { clip: rect } : {}) });
    }
    captured.push(shot.name);
    console.log(`ok    ${shot.name} -> public/${shotPath(shot)}`);
  } catch (error) {
    failed.push([shot.name, error.message.split('\n')[0]]);
    console.log(`fail  ${shot.name}: ${error.message.split('\n')[0]}`);
  } finally {
    await page.close();
  }
}

await browser.close();

if (config.updateDocs && captured.length > 0) {
  const switched = pointDocsAt(captured);
  if (switched.length > 0) {
    console.log(`\nquickstart now points at: ${switched.join(', ')}`);
  }
  if (config.prunePlaceholders) {
    for (const shot of selected) {
      if (captured.includes(shot.name)) prunePlaceholder(shot);
    }
  }
}

console.log(`\ncaptured ${captured.length}/${selected.length}`);
if (failed.length > 0) {
  console.log('still placeholders:');
  for (const [name, reason] of failed) console.log(`  ${name}: ${reason}`);
  process.exit(1);
}
