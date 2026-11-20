export type VariableSourceTypename = 'DatabaseSource' | 'KeyValueStoreSource' | 'BucketSource' | 'SharedSource';

export interface VariableRow {
  key: string;
  value: string;
  ref?: string | null;
  revealed?: boolean;
  persisted?: boolean;
  removed?: boolean;
}

export interface VariableReference {
  id: string;
  key: string;
  typename: VariableSourceTypename;
  sourceName: string;
  value?: string;
}

export interface VariableChange {
  key: string;
  value: string;
  ref: string | null;
}

export interface VariableChanges {
  set: VariableChange[];
  unset: string[];
}

export function liveRows(rows: VariableRow[]): VariableRow[] {
  return rows.filter(row => !row.removed && row.key.trim());
}

export function duplicateKeys(rows: VariableRow[]): string[] {
  const seen = new Set<string>();
  const duplicates = new Set<string>();
  for (const row of liveRows(rows)) {
    const key = row.key.trim();
    if (seen.has(key)) duplicates.add(key);
    seen.add(key);
  }
  return [...duplicates];
}

export function variableChanges(baseline: VariableRow[], rows: VariableRow[]): VariableChanges {
  const before = new Map(baseline.map(row => [row.key, row]));
  const after = new Map(liveRows(rows).map(row => [row.key.trim(), row]));

  const set: VariableChange[] = [];
  for (const [key, row] of after) {
    const ref = row.ref ?? null;
    const previous = before.get(key);
    if (!previous || (previous.ref ?? null) !== ref || (!ref && previous.value !== row.value)) {
      set.push({ key, value: row.value, ref });
    }
  }

  const unset = [...before.keys()].filter(key => !after.has(key));

  return { set, unset };
}
