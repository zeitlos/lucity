import { computed, ref, watch } from 'vue';
import { ID_MAX_LENGTH, deriveId, isValidId } from '@/lib/id';

export type IdFieldSource = {
  name: string;
  id: string;
  fallback: string;
  taken: ReadonlySet<string>;
  scope?: string;
  reserved?: boolean;
};

export function useIdField(source: IdFieldSource, setId: (id: string) => void) {
  const edited = ref(false);

  const base = computed(() => deriveId(source.name, source.fallback));
  const derived = computed(() => deriveId(source.name, source.fallback, source.taken));
  const reason = computed(() => (source.reserved ? 'reserved' : 'taken'));
  const numbered = computed(() => !edited.value && derived.value !== base.value);
  const isValid = computed(() => isValidId(source.id) && !source.taken.has(source.id));

  const error = computed(() => {
    if (!edited.value) return null;
    if (!isValidId(source.id)) {
      return `Use 2 to ${ID_MAX_LENGTH} characters, starting and ending with a letter or digit.`;
    }
    if (!source.taken.has(source.id)) return null;
    if (source.reserved) return `"${source.id}" is reserved.`;
    return source.scope ? `"${source.id}" is taken in this ${source.scope}.` : `"${source.id}" is taken.`;
  });

  watch(derived, (id) => {
    if (!edited.value) setId(id);
  }, { immediate: true });

  function input(id: string) {
    edited.value = true;
    setId(id);
  }

  function reset() {
    edited.value = false;
    setId(derived.value);
  }

  return { edited, base, derived, reason, numbered, isValid, error, input, reset };
}
