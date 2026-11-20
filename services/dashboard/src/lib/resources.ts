export type ResourceKind = 'repositoryService' | 'imageService' | 'database' | 'keyValueStore' | 'bucket' | 'volume';

export function serviceKind(sourceUrl: string): ResourceKind {
  return sourceUrl ? 'repositoryService' : 'imageService';
}

export function repositoryName(sourceUrl: string): string {
  return sourceUrl.replace('https://github.com/', '').replace(/\.git$/, '');
}

export interface EnvironmentContents {
  services: { id: string; name: string; sourceUrl: string }[];
  databases: { id: string; name: string }[];
  keyValueStores: { id: string; name: string }[];
  buckets: { id: string; name: string }[];
  volumes: { id: string; name: string }[];
}

export interface ResourceRow {
  id: string;
  kind: ResourceKind;
  name: string;
}

export function environmentResources(environment: EnvironmentContents): ResourceRow[] {
  return [
    ...environment.services.map(s => ({ id: s.id, kind: serviceKind(s.sourceUrl), name: s.name })),
    ...environment.databases.map(d => ({ id: d.id, kind: 'database' as const, name: d.name })),
    ...environment.keyValueStores.map(k => ({ id: k.id, kind: 'keyValueStore' as const, name: k.name })),
    ...environment.buckets.map(b => ({ id: b.id, kind: 'bucket' as const, name: b.name })),
    ...environment.volumes.map(v => ({ id: v.id, kind: 'volume' as const, name: v.name })),
  ];
}

const nouns: [keyof EnvironmentContents, string, string][] = [
  ['services', 'service', 'services'],
  ['databases', 'database', 'databases'],
  ['keyValueStores', 'key-value store', 'key-value stores'],
  ['buckets', 'bucket', 'buckets'],
  ['volumes', 'volume', 'volumes'],
];

export function describeResources(environment: EnvironmentContents): string {
  const parts = nouns
    .map(([key, one, many]) => [environment[key].length, one, many] as const)
    .filter(([count]) => count > 0)
    .map(([count, one, many]) => `${count} ${count === 1 ? one : many}`);

  return new Intl.ListFormat('en-GB', { type: 'conjunction' }).format(parts);
}
