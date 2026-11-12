/**
 * Helpers for scripted takes: a visible cursor that glides to what it clicks,
 * and typing at a pace a viewer can follow. The cursor is drawn by the page
 * itself because Chrome's screencast leaves the real pointer out.
 */

export function cursorScript() {
  if (window.top !== window) return;
  const draw = () => {
    if (!document.body) return requestAnimationFrame(draw);
    const cursor = document.createElement('div');
    cursor.setAttribute('data-compare-cursor', '');
    cursor.style.cssText = [
      'position:fixed', 'left:0', 'top:0', 'width:22px', 'height:22px', 'margin:-11px 0 0 -11px',
      'border-radius:50%', 'background:rgba(124,58,237,.28)', 'border:2px solid rgba(124,58,237,.9)',
      'z-index:2147483647', 'pointer-events:none', 'transition:transform .12s ease',
      `transform:translate(${window.__compareCursor?.x ?? -40}px,${window.__compareCursor?.y ?? -40}px)`,
    ].join(';');
    document.body.appendChild(cursor);
    const place = (x, y, scale = 1) => {
      cursor.style.transform = `translate(${x}px,${y}px) scale(${scale})`;
    };
    addEventListener('mousemove', (event) => place(event.clientX, event.clientY), true);
    addEventListener('mousedown', (event) => place(event.clientX, event.clientY, 0.7), true);
    addEventListener('mouseup', (event) => place(event.clientX, event.clientY), true);
  };
  draw();
}

/** Makes links and window.open load in the same tab, so the recorded tab follows them. */
export function sameTabScript() {
  addEventListener('click', (event) => {
    const link = event.target.closest?.('a[target]');
    if (link && link.target !== '_self') link.target = '_self';
  }, true);
  window.open = (url) => {
    if (url) location.href = url;
    return window;
  };
}

export function driver(page) {
  let position = { x: 640, y: 400 };

  async function moveTo(locator) {
    await locator.waitFor({ state: 'visible', timeout: 60_000 });
    await locator.scrollIntoViewIfNeeded().catch(() => {});
    const box = await locator.boundingBox();
    if (!box) throw new Error('element has no box');
    const target = { x: box.x + box.width / 2, y: box.y + box.height / 2 };
    const distance = Math.hypot(target.x - position.x, target.y - position.y);
    await page.mouse.move(target.x, target.y, { steps: Math.max(8, Math.round(distance / 18)) });
    position = target;
    await page.waitForTimeout(250);
  }

  return {
    async click(locator) {
      await moveTo(locator);
      await page.mouse.down();
      await page.waitForTimeout(80);
      await page.mouse.up();
      await page.waitForTimeout(700);
    },
    async type(text) {
      const editable = await page.evaluate(() => {
        const element = document.activeElement;
        return Boolean(element && (element.matches('input, textarea') || element.isContentEditable));
      });
      if (!editable) throw new Error(`refusing to type "${text}": no input has focus, the keys would reach the page's shortcuts`);
      await page.keyboard.type(text, { delay: 55 });
      await page.waitForTimeout(400);
    },
    async press(key) {
      await page.keyboard.press(key);
      await page.waitForTimeout(500);
    },
    pause: (milliseconds) => page.waitForTimeout(milliseconds),
  };
}

/** Waits until a vouch board at `url` serves the working page, polled from here, not the browser. */
export async function waitForWorkingBoard(url, timeout = 15 * 60_000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    try {
      const response = await fetch(url);
      const html = await response.text();
      if (response.ok && html.includes('Submit Feedback') && !html.includes('Database not connected')) return;
    } catch {}
    await new Promise((resolve) => setTimeout(resolve, 5_000));
  }
  throw new Error(`${url} did not serve a working board in time`);
}
