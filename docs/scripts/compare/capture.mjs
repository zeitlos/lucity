#!/usr/bin/env node
/**
 * Captures the comparison page screenshots and videos, the same tasks on every
 * platform, from a real browser you signed in to yourself.
 *
 * --login opens the platform in a visible Chrome window and waits until you
 * have signed in (OAuth, two-factor, whatever it asks). The browser state,
 * session cookies included, is then saved to
 * ~/Library/Application Support/lucity-compare/<platform>.json (mode 600,
 * outside the repo). Stills run headless on that session; --review and
 * --record use a visible window. No password ever passes through this script.
 *
 * Usage:
 *   node scripts/compare/capture.mjs railway --login
 *   node scripts/compare/capture.mjs railway --target project=<id> --target service=<id> ...
 *   node scripts/compare/capture.mjs railway                  # every still, headless
 *   node scripts/compare/capture.mjs railway --headed         # the same in a visible window
 *   node scripts/compare/capture.mjs railway --only variables,build-logs
 *   node scripts/compare/capture.mjs railway --review         # pause before each capture to fix the view by hand
 *   node scripts/compare/capture.mjs railway --record deploy  # real-time video, clock added afterwards
 *   node scripts/compare/capture.mjs railway --record deploy --auto  # scripted take, headless (add --headed to watch)
 *   node scripts/compare/capture.mjs railway --finish deploy --trim 8  # redo the clock on the raw take, from 8s in
 *   node scripts/compare/capture.mjs railway --list
 *
 * Targets (project ids, app names) are saved to docs/.shots/compare/<platform>.json,
 * which the setup scripts write. --target key=value overrides and saves one.
 *
 * Output: stills in docs/public/img/compare/<platform>/<shot>.webp, and
 * docs/public/img/compare/manifest.json with the capture dates. Videos and their
 * posters go to the public media bucket on Lucity (COMPARE_BUCKET, uploaded with
 * the CLI's session) under content-hashed names, and the manifest holds their URLs.
 * Raw takes without the clock stay in docs/.shots/compare/raw/, so the start
 * can be trimmed later. Only the start: a take is never cut in the middle.
 *
 * Env: COMPARE_BROWSER (default chrome, an installed browser channel),
 *      COMPARE_THEME (light | dark, default light).
 */

import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, renameSync, rmSync, writeFileSync } from 'node:fs';
import { homedir, tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { createInterface } from 'node:readline/promises';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';

import { cursorScript, sameTabScript } from './drive.mjs';
import { fingerprint, publish } from './publish.mjs';
import { encodeVideo, finishVideo, startRecording } from './recorder.mjs';

import deploio from './platforms/deploio.mjs';
import fly from './platforms/fly.mjs';
import heroku from './platforms/heroku.mjs';
import lucity from './platforms/lucity.mjs';
import railway from './platforms/railway.mjs';
import render from './platforms/render.mjs';
import vercel from './platforms/vercel.mjs';

const platforms = { railway, heroku, render, deploio, fly, vercel, lucity };

const docs = dirname(dirname(dirname(fileURLToPath(import.meta.url))));
const outputRoot = join(docs, 'public', 'img', 'compare');
const manifestPath = join(outputRoot, 'manifest.json');
const stateRoot = join(docs, '.shots', 'compare');
const profileRoot = join(homedir(), 'Library', 'Application Support', 'lucity-compare');

const viewport = { width: 1280, height: 800 };
const deviceScaleFactor = 2;

const args = process.argv.slice(2);
const flag = (name) => args.includes(`--${name}`);
const option = (name) => {
  const index = args.indexOf(`--${name}`);
  return index !== -1 ? args[index + 1] : undefined;
};
const options = (name) => args.flatMap((value, index) => (args[index - 1] === `--${name}` ? [value] : []));

const platformName = args[0];
const platform = platforms[platformName];
if (!platform) {
  console.error(`usage: capture.mjs <${Object.keys(platforms).join('|')}> [--login] [--list] [--only a,b] [--review] [--record <shot>] [--finish <shot> --trim <s>] [--target key=value]`);
  process.exit(2);
}

const statePath = join(stateRoot, `${platformName}.json`);
const target = existsSync(statePath) ? JSON.parse(readFileSync(statePath, 'utf8')) : {};
for (const pair of options('target')) {
  const index = pair.indexOf('=');
  if (index < 1) throw new Error(`--target expects key=value, got ${pair}`);
  target[pair.slice(0, index)] = pair.slice(index + 1);
}
mkdirSync(stateRoot, { recursive: true });
writeFileSync(statePath, `${JSON.stringify(target, null, 2)}\n`);

if (flag('list')) {
  for (const shot of platform.shots) {
    console.log(`${shot.video ? 'video' : 'still'}  ${shot.name.padEnd(18)} ${shot.title}`);
  }
  console.log(`\ntargets needed: ${platform.targets.join(', ')}`);
  console.log(`targets saved:  ${JSON.stringify(target)}`);
  process.exit(0);
}

const interactive = process.stdin.isTTY && process.stdout.isTTY;

async function prompt(question, signal) {
  if (!interactive) throw new Error('needs a terminal to pause, run it in your own terminal');
  const readline = createInterface({ input: process.stdin, output: process.stdout });
  try {
    await readline.question(question, { signal });
  } catch (error) {
    if (error.name !== 'AbortError') throw error;
    process.stdout.write('\n');
  } finally {
    readline.close();
  }
}

async function appWorking(page) {
  const board = await page.getByRole('button', { name: 'Submit Feedback' }).count().catch(() => 0);
  const broken = await page.getByText(/Database not connected|Application error|Application failed to respond/i).count().catch(() => 1);
  return board > 0 && broken === 0;
}

async function waitForWorkingApp(page, signal) {
  while (!signal.aborted) {
    if (await appWorking(page)) {
      await page.waitForTimeout(3_000);
      if (await appWorking(page)) return;
    }
    await new Promise((resolve) => setTimeout(resolve, 1_000));
  }
}

async function declineCookies(page) {
  const button = page.getByRole('button', { name: /^(reject non-essential|reject all|decline all|necessary only|only necessary)/i }).first();
  if (await button.isVisible().catch(() => false)) {
    await button.click().catch(() => {});
    await page.waitForTimeout(500);
  }
  await platform.tidy?.(page);
}

async function waitForSignIn(page) {
  await page.goto(platform.home, { waitUntil: 'load' }).catch(() => {});
  await page.waitForTimeout(3_000);
  await declineCookies(page);
  if (await platform.signedIn(page)) return saveSession(page.context());
  console.log(`Sign in to ${platform.title} in the browser window. Waiting up to 30 minutes...`);
  const deadline = Date.now() + 30 * 60_000;
  while (Date.now() < deadline) {
    await page.waitForTimeout(2_000);
    if (page.isClosed()) throw new Error('the browser window was closed before signing in');
    await declineCookies(page);
    if (await platform.signedIn(page)) {
      console.log('Signed in.');
      await page.waitForTimeout(3_000);
      return saveSession(page.context());
    }
  }
  throw new Error(`still not signed in to ${platform.title} after 30 minutes`);
}

function missingTargets(shot) {
  return (shot.needs ?? platform.targets).filter((key) => !target[key]);
}

function readManifest() {
  return existsSync(manifestPath) ? JSON.parse(readFileSync(manifestPath, 'utf8')) : {};
}

function record(name, entry) {
  const manifest = readManifest();
  manifest[platformName] = { ...(manifest[platformName] ?? {}), [name]: entry };
  mkdirSync(outputRoot, { recursive: true });
  writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
}

const sessionPath = join(profileRoot, `${platformName}.json`);

async function launch({ headless = false, ...extra } = {}) {
  const browser = await chromium.launch({
    channel: process.env.COMPARE_BROWSER || 'chrome',
    headless,
    args: ['--lang=en-US', `--force-device-scale-factor=${deviceScaleFactor}`],
  });
  const context = await browser.newContext({
    viewport,
    deviceScaleFactor,
    locale: 'en-US',
    timezoneId: 'Europe/Zurich',
    colorScheme: process.env.COMPARE_THEME === 'dark' ? 'dark' : 'light',
    storageState: existsSync(sessionPath) ? sessionPath : undefined,
    ...extra,
  });
  context.on('close', () => browser.close().catch(() => {}));
  return context;
}

async function saveSession(context) {
  mkdirSync(profileRoot, { recursive: true, mode: 0o700 });
  writeFileSync(sessionPath, JSON.stringify(await context.storageState()), { mode: 0o600 });
}

const rawPath = (shot) => join(stateRoot, 'raw', `${platformName}-${shot.name}.mp4`);

async function finish(shot, trim) {
  const raw = rawPath(shot);
  if (!existsSync(raw)) throw new Error(`no raw take at ${raw}, record it first`);
  const out = join(stateRoot, 'out');
  mkdirSync(out, { recursive: true });
  const mp4 = join(out, `${platformName}-${shot.name}.mp4`);
  const poster = join(out, `${platformName}-${shot.name}.webp`);
  const seconds = await finishVideo({ raw, mp4, poster, trim, scale: deviceScaleFactor });

  const previous = readManifest()[platformName]?.[shot.name];
  const capturedAt = previous?.capturedAt ?? new Date().toISOString();
  const stem = `compare/${platformName}/${shot.name}-${capturedAt.slice(0, 10)}`;
  const file = await publish(mp4, `${stem}-${fingerprint(mp4)}.mp4`, 'video/mp4');
  const posterURL = await publish(poster, `${stem}-${fingerprint(poster)}.webp`, 'image/webp');

  for (const extension of ['mp4', 'webp']) rmSync(join(outputRoot, platformName, `${shot.name}.${extension}`), { force: true });
  record(shot.name, {
    kind: 'video',
    title: shot.title,
    file,
    poster: posterURL,
    seconds,
    trimmed: trim,
    capturedAt,
    viewport,
    deviceScaleFactor,
  });
  console.log(`ok    ${shot.name} -> ${file} (${seconds}s${trim ? `, first ${trim}s trimmed` : ''})`);
}

async function captureVideo(shot) {
  const scratch = mkdtempSync(join(tmpdir(), 'lucity-compare-'));
  const size = { width: viewport.width * deviceScaleFactor, height: viewport.height * deviceScaleFactor };
  const auto = flag('auto');
  if (auto && !shot.drive) throw new Error(`${platformName} has no scripted ${shot.name} yet, record it by hand`);
  const context = await launch({ headless: auto && !flag('headed') });
  const page = context.pages()[0] ?? (await context.newPage());
  const began = Date.now();
  const since = () => `${Math.round((Date.now() - began) / 1000)}s in`;
  let finished = false;
  page.on('crash', () => console.log(`      the page crashed, ${since()}`));
  page.on('close', () => finished || console.log(`      the page closed, ${since()}`));
  context.browser()?.on('disconnected', () => finished || console.log(`      the browser disconnected, ${since()}`));
  try {
    await waitForSignIn(page);
    if (shot.prepare) await shot.prepare(page);
    await context.addInitScript(cursorScript);
    await context.addInitScript(sameTabScript);
    await shot.open(page, target);
    await page.waitForLoadState('networkidle', { timeout: 15_000 }).catch(() => {});
    await page.waitForTimeout(2_000);
    await declineCookies(page);
    if (!auto && interactive) await prompt('Set up the window (theme, banners), then press Enter to start recording. ');
    const recording = await startRecording(page, join(scratch, 'frames'), size);

    console.log(`Recording "${shot.title}"${auto ? ', scripted' : `. ${shot.hint ?? ''}`}`);
    const stop = new AbortController();
    if (auto) {
      try {
        const found = await shot.drive(page, target);
        Object.assign(target, Object.fromEntries(Object.entries(found ?? {}).filter(([, value]) => value)));
        writeFileSync(statePath, `${JSON.stringify(target, null, 2)}\n`);
      } catch (error) {
        await page.screenshot({ path: join(stateRoot, `${platformName}-${shot.name}-failed.png`) }).catch(() => {});
        await recording.stop();
        throw new Error(`scripted take failed at ${page.url()}: ${error.message.split('\n')[0]} (screenshot in docs/.shots/compare/)`);
      }
      const timeout = new Promise((_, reject) => setTimeout(() => reject(new Error('the app did not show the working board within 5 minutes')), 5 * 60_000));
      await Promise.race([waitForWorkingApp(page, stop.signal), timeout]);
      console.log('The app works, stopping.');
    } else {
      console.log('Stops by itself once this tab shows the working vouch board.');
      await Promise.race([
        waitForWorkingApp(page, stop.signal).then(() => console.log('The app works, stopping.')),
        ...(interactive ? [prompt('Press Enter to stop early. ', stop.signal)] : []),
      ]);
    }
    stop.abort();
    const recorded = await recording.stop();
    await saveSession(context);
    finished = true;
    await context.close();

    mkdirSync(dirname(rawPath(shot)), { recursive: true });
    encodeVideo(recorded, scratch, size, rawPath(shot));
    record(shot.name, { ...(readManifest()[platformName]?.[shot.name] ?? {}), capturedAt: new Date().toISOString() });
    await finish(shot, Number(option('trim') ?? 0));
  } finally {
    await context.close().catch(() => {});
    rmSync(scratch, { recursive: true, force: true });
  }
}

async function captureStills(shots) {
  const scratch = mkdtempSync(join(tmpdir(), 'lucity-compare-'));
  const headless = !flag('headed') && !flag('review');
  const context = await launch({ headless });
  const page = context.pages()[0] ?? (await context.newPage());
  const failed = [];
  try {
    if (headless) {
      await page.goto(platform.home, { waitUntil: 'load' }).catch(() => {});
      await page.waitForTimeout(3_000);
      await declineCookies(page);
      if (!(await platform.signedIn(page))) throw new Error(`not signed in to ${platform.title}, run --login first`);
    } else {
      await waitForSignIn(page);
    }
    for (const shot of shots) {
      const missing = missingTargets(shot);
      if (missing.length > 0) {
        failed.push([shot.name, `missing target ${missing.join(', ')}`]);
        console.log(`skip  ${shot.name} (missing target ${missing.join(', ')})`);
        continue;
      }
      try {
        let problem;
        try {
          await shot.open(page, target);
        } catch (error) {
          problem = error.message.split('\n')[0];
        }
        if (problem && (headless || !interactive)) throw new Error(`${problem} (retry with --only ${shot.name} --review)`);
        if (problem || flag('review')) {
          if (problem) console.log(`      ${shot.name}: ${problem}`);
          await prompt(`      Set up "${shot.title}" in the browser, then press Enter to capture. `);
        }
        await declineCookies(page);
        await page.waitForTimeout(500);
        const png = join(scratch, `${shot.name}.png`);
        await page.screenshot({ path: png, animations: 'disabled' });
        const directory = join(outputRoot, platformName);
        mkdirSync(directory, { recursive: true });
        const webp = join(directory, `${shot.name}.webp`);
        execFileSync('cwebp', ['-quiet', '-q', '88', png, '-o', `${webp}.tmp`]);
        renameSync(`${webp}.tmp`, webp);
        record(shot.name, {
          kind: 'still',
          title: shot.title,
          file: `/img/compare/${platformName}/${shot.name}.webp`,
          capturedAt: new Date().toISOString(),
          viewport,
          deviceScaleFactor,
        });
        console.log(`ok    ${shot.name} -> public/img/compare/${platformName}/${shot.name}.webp`);
      } catch (error) {
        failed.push([shot.name, error.message.split('\n')[0]]);
        console.log(`fail  ${shot.name}: ${error.message.split('\n')[0]}`);
      }
    }
    await saveSession(context);
  } finally {
    await context.close().catch(() => {});
    rmSync(scratch, { recursive: true, force: true });
  }
  return failed;
}

if (flag('login')) {
  const context = await launch();
  const page = context.pages()[0] ?? (await context.newPage());
  await waitForSignIn(page);
  await context.close();
  console.log(`Session saved for ${platform.title}.`);
  process.exit(0);
}

const finishName = option('finish');
if (finishName) {
  const shot = platform.shots.find((candidate) => candidate.name === finishName && candidate.video);
  if (!shot) throw new Error(`no video shot named ${finishName} for ${platformName}, see --list`);
  await finish(shot, Number(option('trim') ?? 0));
  process.exit(0);
}

const recordName = option('record');
if (recordName) {
  const shot = platform.shots.find((candidate) => candidate.name === recordName && candidate.video);
  if (!shot) throw new Error(`no video shot named ${recordName} for ${platformName}, see --list`);
  await captureVideo(shot);
  process.exit(0);
}

const only = option('only')?.split(',').map((name) => name.trim()).filter(Boolean);
const selected = platform.shots.filter((shot) => !shot.video && (!only || only.includes(shot.name)));
if (only) {
  const unknown = only.filter((name) => !selected.some((shot) => shot.name === name));
  if (unknown.length > 0) throw new Error(`unknown still ${unknown.join(', ')}, see --list`);
}

const failed = await captureStills(selected);
console.log(`\ncaptured ${selected.length - failed.length}/${selected.length}`);
if (failed.length > 0) {
  for (const [name, reason] of failed) console.log(`  ${name}: ${reason}`);
  process.exit(1);
}
