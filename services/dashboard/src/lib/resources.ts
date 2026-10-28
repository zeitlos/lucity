export type ResourceKind = 'repositoryService' | 'imageService' | 'database' | 'keyValueStore' | 'bucket' | 'volume';

export function serviceKind(sourceUrl: string): ResourceKind {
  return sourceUrl ? 'repositoryService' : 'imageService';
}

export function repositoryName(sourceUrl: string): string {
  return sourceUrl.replace('https://github.com/', '').replace(/\.git$/, '');
}
