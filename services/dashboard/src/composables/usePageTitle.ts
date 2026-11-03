import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { useTitle } from '@vueuse/core';

/**
 * Sets the browser tab title from the most specific part to the least, like
 * "api · production · shop | Lucity". Missing parts are skipped.
 */
export function usePageTitle(parts: MaybeRefOrGetter<(string | null | undefined)[]>) {
  const title = computed(() => {
    const page = toValue(parts).filter(Boolean).join(' · ');
    return page ? `${page} | Lucity` : 'Lucity';
  });
  useTitle(title, { restoreOnUnmount: false });
}
