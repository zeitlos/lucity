export type VariableSourceTypename = 'DatabaseSource' | 'KeyValueStoreSource' | 'BucketSource' | 'SharedSource';

export interface VariableRow {
  key: string;
  value: string;
  ref?: string | null;
  revealed?: boolean;
}

export interface VariableReference {
  id: string;
  key: string;
  typename: VariableSourceTypename;
  sourceName: string;
}
