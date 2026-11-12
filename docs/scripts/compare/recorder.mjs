import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { chromium } from 'playwright';

/**
 * Records a page through Chrome's screencast: a frame whenever the page
 * repaints, so motion stays smooth and idle stretches cost nothing. Frames
 * come at the browser's real scale, which --force-device-scale-factor sets;
 * Playwright's deviceScaleFactor alone leaves them at 1×.
 */
export async function startRecording(page, directory, size) {
  mkdirSync(directory, { recursive: true });
  const session = await page.context().newCDPSession(page);
  const frames = [];

  session.on('Page.screencastFrame', ({ data, metadata, sessionId }) => {
    const file = join(directory, `${String(frames.length).padStart(6, '0')}.jpg`);
    writeFileSync(file, Buffer.from(data, 'base64'));
    frames.push({ file, timestamp: metadata.timestamp });
    session.send('Page.screencastFrameAck', { sessionId }).catch(() => {});
  });

  await session.send('Page.startScreencast', {
    format: 'jpeg',
    quality: 92,
    maxWidth: size.width,
    maxHeight: size.height,
    everyNthFrame: 2,
  });

  return {
    async stop() {
      await session.send('Page.stopScreencast').catch(() => {});
      await session.detach().catch(() => {});
      return { frames, end: Date.now() / 1000 };
    },
  };
}

/** Turns recorded frames into a raw H.264 mp4 at 60 fps, keeping their real timing. */
export function encodeVideo({ frames, end }, directory, size, mp4) {
  if (frames.length === 0) throw new Error('no frames were recorded');
  const lines = [];
  frames.forEach((frame, index) => {
    const next = frames[index + 1]?.timestamp ?? end;
    lines.push(`file '${frame.file}'`, `duration ${Math.max(next - frame.timestamp, 0.001).toFixed(3)}`);
  });
  lines.push(`file '${frames.at(-1).file}'`);
  const list = join(directory, 'frames.txt');
  writeFileSync(list, `${lines.join('\n')}\n`);

  execFileSync('ffmpeg', [
    '-y', '-loglevel', 'error',
    '-f', 'concat', '-safe', '0', '-i', list,
    '-vf', `scale=${size.width}:${size.height}:flags=lanczos,format=yuv420p`,
    '-fps_mode', 'cfr', '-r', '60',
    '-c:v', 'libx264', '-crf', '16', '-preset', 'slow',
    '-movflags', '+faststart', '-an', mp4,
  ]);
}

function duration(file) {
  return Number(execFileSync('ffprobe', ['-v', 'error', '-show_entries', 'format=duration', '-of', 'csv=p=0', file], { encoding: 'utf8' }).trim());
}

function clockLabel(seconds) {
  const minutes = String(Math.floor(seconds / 60)).padStart(2, '0');
  return `${minutes}:${String(seconds % 60).padStart(2, '0')} real time`;
}

async function renderClock(directory, seconds, scale) {
  mkdirSync(directory, { recursive: true });
  const browser = await chromium.launch({ channel: process.env.COMPARE_BROWSER || 'chrome', headless: true });
  const page = await browser.newPage({ viewport: { width: 400, height: 100 }, deviceScaleFactor: scale });
  await page.setContent(`<div id="clock" style="display:inline-block;font:600 14px/1 ui-monospace,SFMono-Regular,Menlo,monospace;color:#fff;background:rgba(17,17,17,.82);padding:8px 11px;border-radius:7px"></div>`);
  const clock = page.locator('#clock');
  const files = [];
  for (let second = 0; second <= seconds; second++) {
    await clock.evaluate((element, text) => { element.textContent = text; }, clockLabel(second));
    const file = join(directory, `${String(second).padStart(5, '0')}.png`);
    await clock.screenshot({ path: file, omitBackground: true });
    files.push(file);
  }
  await browser.close();
  return files;
}

/**
 * Burns a running real-time clock into a raw take, starting at 0:00 after
 * `trim` seconds, and writes the mp4 and a webp poster of its last frame.
 * Returns the length of the finished video in seconds.
 */
export async function finishVideo({ raw, mp4, poster, trim = 0, scale, margin = 24 }) {
  const length = duration(raw) - trim;
  if (length <= 0) throw new Error(`--trim ${trim} is longer than the take`);
  const work = mkdtempSync(join(tmpdir(), 'lucity-clock-'));
  try {
    const files = await renderClock(join(work, 'clock'), Math.ceil(length), scale);
    const list = join(work, 'clock.txt');
    writeFileSync(list, `${files.flatMap((file) => [`file '${file}'`, 'duration 1']).join('\n')}\nfile '${files.at(-1)}'\n`);
    execFileSync('ffmpeg', [
      '-y', '-loglevel', 'error',
      '-ss', String(trim), '-i', raw,
      '-f', 'concat', '-safe', '0', '-i', list,
      '-filter_complex', `[0:v][1:v]overlay=W-w-${margin}:H-h-${margin}:shortest=1,format=yuv420p`,
      '-fps_mode', 'cfr', '-r', '60',
      '-c:v', 'libx264', '-crf', '20', '-preset', 'slow',
      '-movflags', '+faststart', '-an', mp4,
    ]);
    const frame = join(work, 'poster.png');
    execFileSync('ffmpeg', ['-y', '-loglevel', 'error', '-sseof', '-0.5', '-i', mp4, '-frames:v', '1', frame]);
    execFileSync('cwebp', ['-quiet', '-q', '88', frame, '-o', poster]);
    return Math.round(duration(mp4));
  } finally {
    rmSync(work, { recursive: true, force: true });
  }
}
