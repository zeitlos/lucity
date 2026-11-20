import { useEventListener, useIntervalFn } from '@vueuse/core';
import { infoToast } from '@/components/ui/sonner';

const checkEvery = 5 * 60 * 1000;
const checkAtMostEvery = 60 * 1000;

function entryScript(doc: Document): string | null {
  return doc.querySelector('script[type="module"][src]')?.getAttribute('src') ?? null;
}

export function useNewVersion() {
  if (!import.meta.env.PROD) return;

  const loaded = entryScript(document);
  if (!loaded) return;

  let announced: string | null = null;
  let checkedAt = 0;

  async function check() {
    if (document.visibilityState !== 'visible') return;
    if (Date.now() - checkedAt < checkAtMostEvery) return;
    checkedAt = Date.now();

    try {
      const res = await fetch(import.meta.env.BASE_URL, { cache: 'no-store' });
      if (!res.ok) return;

      const latest = entryScript(new DOMParser().parseFromString(await res.text(), 'text/html'));
      if (!latest || latest === loaded || latest === announced) return;

      announced = latest;
      infoToast('A new version of Lucity is available', {
        description: 'Reload to get the latest changes.',
        duration: Infinity,
        action: { label: 'Reload', onClick: () => window.location.reload() },
      });
    } catch {
      return;
    }
  }

  useIntervalFn(check, checkEvery);
  useEventListener(document, 'visibilitychange', check);
  useEventListener(window, 'focus', check);
}
