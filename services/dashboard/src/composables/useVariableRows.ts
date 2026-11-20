import { computed, ref, watch } from 'vue';
import { duplicateKeys, liveRows, variableChanges, type VariableRow } from '@/lib/variables';

/**
 * Keeps the editable rows of a variables form next to the variables they were
 * loaded from, so a save sends only what changed.
 */
export function useVariableRows(source: () => VariableRow[] | undefined) {
  const baseline = ref<VariableRow[]>([]);
  const rows = ref<VariableRow[]>([]);

  const changes = computed(() => variableChanges(baseline.value, rows.value));
  const dirty = computed(() => changes.value.set.length > 0 || changes.value.unset.length > 0);
  const duplicate = computed(() => duplicateKeys(rows.value)[0]);

  function load(variables: VariableRow[]) {
    baseline.value = variables;
    rows.value = variables.map(variable => ({ ...variable, persisted: true }));
  }

  function reset() {
    const variables = source();
    if (variables) load(variables);
  }

  function commit() {
    load(liveRows(rows.value).map(row => ({
      key: row.key.trim(),
      value: row.ref ? '' : row.value,
      ref: row.ref ?? null,
    })));
  }

  watch(source, () => {
    if (!dirty.value) reset();
  }, { immediate: true });

  return { rows, changes, dirty, duplicate, reset, commit };
}
