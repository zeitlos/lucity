import { BucketStatus, DatabaseStatus, ServiceStatus } from '@/gql/graphql';

export type ResourceKind = 'repositoryService' | 'imageService' | 'database' | 'keyValueStore' | 'bucket' | 'volume';

export function serviceKind(sourceUrl: string): ResourceKind {
  return sourceUrl ? 'repositoryService' : 'imageService';
}

export type ResourceTone = 'ok' | 'warn' | 'danger' | 'progress' | 'neutral';

export function serviceTone(status: ServiceStatus): ResourceTone {
  switch (status) {
    case ServiceStatus.Healthy:
      return 'ok';
    case ServiceStatus.Failed:
      return 'danger';
    case ServiceStatus.Degraded:
      return 'warn';
    case ServiceStatus.Deploying:
      return 'progress';
    default:
      return 'neutral';
  }
}

export function databaseTone(status: DatabaseStatus): ResourceTone {
  switch (status) {
    case DatabaseStatus.Healthy:
      return 'ok';
    case DatabaseStatus.Failed:
      return 'danger';
    case DatabaseStatus.Degraded:
      return 'warn';
    case DatabaseStatus.Pending:
    case DatabaseStatus.Updating:
      return 'progress';
    default:
      return 'neutral';
  }
}

export function bucketTone(status: BucketStatus): ResourceTone {
  switch (status) {
    case BucketStatus.Ready:
      return 'ok';
    case BucketStatus.Failed:
      return 'danger';
    case BucketStatus.Pending:
      return 'progress';
    default:
      return 'neutral';
  }
}

export function repositoryName(sourceUrl: string): string {
  return sourceUrl.replace('https://github.com/', '').replace(/\.git$/, '');
}
