#!/usr/bin/env node
/**
 * Seeds a deployed vouch app with the same demo posts and votes on every
 * platform, through the app's own API. Safe to run again: existing titles are
 * kept and only the missing votes are added.
 *
 * Usage: node scripts/compare/seed.mjs https://vouch-compare.example.app
 */

const posts = [
  ['Dark mode for the board', 'Half the team reads this at night and the white background is rough. A dark theme would help.', 48],
  ['Keyboard shortcuts for voting', 'Let me arrow through the list and hit V to vote instead of reaching for the mouse.', 31],
  ['Filter requests by status', 'Past twenty requests it gets hard to see what is actually planned versus still open.', 24],
  ['Attach images to a request', 'A screenshot explains a layout bug faster than three paragraphs of description.', 21],
  ['Weekly digest email', 'A Monday summary of new requests and status changes, so nobody has to check the board.', 17],
  ['Chart tooltips overlap on narrow screens', 'On a 380px viewport the tooltip covers the bar it describes.', 9],
];

const base = (process.argv[2] || '').replace(/\/$/, '');
if (!/^https?:\/\//.test(base)) {
  console.error('usage: seed.mjs <app url>');
  process.exit(2);
}

async function call(path, init) {
  const response = await fetch(`${base}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json' },
  });
  if (!response.ok) throw new Error(`${init?.method ?? 'GET'} ${path}: ${response.status} ${await response.text()}`);
  return response.json();
}

const existing = await call('/api/posts');

for (const [title, description, votes] of [...posts].reverse()) {
  const post = existing.find((candidate) => candidate.title === title) ?? (await call('/api/posts', { method: 'POST', body: JSON.stringify({ title, description }) }));
  for (let count = Number(post.votes ?? 0); count < votes; count++) {
    await call('/api/votes', { method: 'POST', body: JSON.stringify({ postId: post.id }) });
  }
  console.log(`${String(votes).padStart(3)}  ${title}`);
}
