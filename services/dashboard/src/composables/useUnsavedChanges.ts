import { computed, onBeforeUnmount, onMounted, ref, toValue, type MaybeRefOrGetter } from 'vue';

const sources = new Set<() => boolean>();
const pending = ref<((discard: boolean) => void) | null>(null);
let listening = false;

function hasUnsavedChanges() {
  return [...sources].some(isDirty => isDirty());
}

function onBeforeUnload(event: BeforeUnloadEvent) {
  if (hasUnsavedChanges()) event.preventDefault();
}

/**
 * Registers a form's unsaved state while it is mounted, so leaving it asks first.
 */
export function useUnsavedChanges(dirty: MaybeRefOrGetter<boolean>) {
  const isDirty = () => toValue(dirty);

  onMounted(() => {
    sources.add(isDirty);
    if (!listening) {
      window.addEventListener('beforeunload', onBeforeUnload);
      listening = true;
    }
  });

  onBeforeUnmount(() => {
    sources.delete(isDirty);
  });
}

/**
 * Resolves true when nothing is unsaved, or once the user chooses to discard.
 */
export function confirmDiscard(): Promise<boolean> {
  if (!hasUnsavedChanges()) return Promise.resolve(true);
  pending.value?.(false);
  return new Promise(resolve => {
    pending.value = resolve;
  });
}

export function useDiscardPrompt() {
  return {
    open: computed(() => pending.value !== null),
    answer(discard: boolean) {
      const resolve = pending.value;
      pending.value = null;
      resolve?.(discard);
    },
  };
}

/**
 * A tab value that only changes once unsaved changes on the current tab are kept or discarded.
 */
export function useGuardedTab(initial: string) {
  const tab = ref(initial);

  async function selectTab(value: string | number) {
    const next = String(value);
    if (next === tab.value) return;
    if (await confirmDiscard()) tab.value = next;
  }

  return { tab, selectTab };
}
