#!/usr/bin/env node
/**
 * Renders a release's social image with the generator on the dashboard's
 * /brand page, so the release header and the generator never drift apart.
 *
 * Without --background it picks the wallpaper used least recently: each
 * release body carries a hidden `<!-- social-background: Name -->` marker,
 * wallpapers no release has used come first (in /brand order), then the one
 * whose last use is oldest. A release that already has a marker keeps it.
 *
 * Writes the PNG, copies it to the clipboard on macOS and prints the marker
 * line to put under the image in the release body.
 *
 * Usage:
 *   node scripts/social-image.mjs --tag v26.11.5 --headline 'Scale to Zero &\nfaster Builds' \
 *     --tagline 'Plus smarter build queues and more secure sessions.'
 *   node scripts/social-image.mjs --list
 *
 * Options:
 *   --tag         release tag, also the kicker
 *   --headline    headline, \n starts a new line
 *   --tagline     line under the headline (default: the generator's tagline)
 *   --background  wallpaper name as on /brand, instead of the automatic pick
 *   --out         output file (default ../tmp/social-<tag>.png)
 *   --list        print the wallpapers and the releases that used them
 *
 * Configuration (env, all optional):
 *   LUCITY_URL       platform URL          (default https://lucity.dev)
 *   LUCITY_APP_PATH  dashboard path on it  (default /app, empty for a local dashboard)
 *   SOCIAL_BROWSER   browser channel       (default chrome, empty for Playwright's Chromium)
 */

import { execFileSync } from 'node:child_process';
import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { parseArgs } from 'node:util';
import { chromium } from 'playwright';

const repoDir = dirname(dirname(dirname(fileURLToPath(import.meta.url))));
const marker = /<!--\s*social-background:\s*(.+?)\s*-->/;

const { values: options } = parseArgs({
  options: {
    tag: { type: 'string' },
    headline: { type: 'string' },
    tagline: { type: 'string' },
    background: { type: 'string' },
    out: { type: 'string' },
    list: { type: 'boolean', default: false },
  },
});

if (!options.list && (!options.tag || !options.headline)) {
  console.error('--tag and --headline are required (or --list)');
  process.exit(1);
}

const base = (process.env.LUCITY_URL || 'https://lucity.dev').replace(/\/$/, '');
const appPath = (process.env.LUCITY_APP_PATH ?? '/app').replace(/\/$/, '');
const channel = process.env.SOCIAL_BROWSER ?? 'chrome';

function releases() {
  const output = execFileSync('gh', ['api', 'repos/{owner}/{repo}/releases?per_page=50', '--jq', '[.[] | {tag: .tag_name, body: (.body // "")}]'], { cwd: repoDir, encoding: 'utf8' });
  return JSON.parse(output).map(release => ({ tag: release.tag, background: release.body.match(marker)?.[1] }));
}

function pick(wallpapers, history) {
  const lastUse = new Map();
  history.forEach((release, index) => {
    if (release.background && !lastUse.has(release.background)) lastUse.set(release.background, index);
  });
  const unused = wallpapers.find(name => !lastUse.has(name));
  if (unused) return unused;
  return wallpapers.reduce((oldest, name) => (lastUse.get(name) > lastUse.get(oldest) ? name : oldest));
}

const browser = await chromium.launch({ channel: channel || undefined });
try {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1200 } });
  await page.goto(`${base}${appPath}/brand`, { waitUntil: 'networkidle' });
  const canvas = page.locator('canvas').first();
  await canvas.waitFor();
  const wallpapers = await page.locator('button[aria-pressed]:has(img)').evaluateAll(buttons => buttons.map(button => button.getAttribute('aria-label')));
  const history = releases();

  if (options.list) {
    for (const name of wallpapers) {
      const used = history.filter(release => release.background === name).map(release => release.tag);
      console.log(`${name}: ${used.join(', ') || 'unused'}`);
    }
  } else {
    await render(page, canvas, wallpapers, history);
  }
} finally {
  await browser.close();
}

async function render(page, canvas, wallpapers, history) {
  const current = history.find(release => release.tag === options.tag)?.background;
  const background = options.background ?? current ?? pick(wallpapers, history.filter(release => release.tag !== options.tag));
  if (!wallpapers.includes(background)) throw new Error(`unknown background "${background}", try --list`);

  await page.getByRole('button', { name: background, exact: true }).click();
  await page.locator('#social-kicker').fill(options.tag);
  await page.locator('#social-headline').fill(options.headline.replaceAll('\\n', '\n'));
  if (options.tagline !== undefined) await page.locator('#social-tagline').fill(options.tagline);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);

  const dataUrl = await canvas.evaluate(element => element.toDataURL('image/png'));
  const out = resolve(options.out ?? join(repoDir, 'tmp', `social-${options.tag}.png`));
  mkdirSync(dirname(out), { recursive: true });
  writeFileSync(out, Buffer.from(dataUrl.split(',')[1], 'base64'));

  if (process.platform === 'darwin') {
    execFileSync('osascript', ['-e', `set the clipboard to (read (POSIX file "${out}") as «class PNGf»)`]);
  }

  console.log(`image: ${out}`);
  console.log(`background: ${background}`);
  console.log(`marker: <!-- social-background: ${background} -->`);
}
