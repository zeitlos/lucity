import { ref, computed } from 'vue';
import type {
  BuildStatus,
  BucketStatus,
  DatabaseStatus,
  DeployStatus,
  ScanStatus,
  ResourceTier,
  ServiceStatus,
  Protocol,
  DeploymentStatus,
  DnsStatus,
  DnsRecordType,
  TlsStatus,
  EndpointType,
  ReleaseStatus,
  ReleaseTriggerKind,
  SourceProvider,
} from '@/gql/graphql';

export interface DnsRecord {
  type: DnsRecordType;
  host: string;
  value: string;
}

export interface DnsState {
  status: DnsStatus;
  requiredRecords: DnsRecord[];
}

export interface Endpoint {
  host: string;
  port: number;
  protocol: Protocol;
  dns: DnsState;
  tls: TlsStatus;
  type: EndpointType;
}

export interface ReplicaCount {
  desired: number;
  ready: number;
}

export interface AutoscalingSettings {
  minReplicas: number;
  maxReplicas: number;
  targetCpu: number;
}

export interface Resources {
  cpu: string;
  memory: string;
}

export interface Deployment {
  id: string;
  image: string;
  imageDigest?: string | null;
  commit: string;
  commitMessage: string;
  ref: string;
  status: DeploymentStatus;
  createdAt: string;
}

export interface Build {
  id: string;
  status: BuildStatus;
  startedAt?: string | null;
  finishedAt?: string | null;
}

export interface Deploy {
  id: string;
  status: DeployStatus;
  startedAt?: string | null;
  finishedAt?: string | null;
}

export interface Scan {
  id: string;
  scanner: string;
  status: ScanStatus;
  findingsCount?: number | null;
  startedAt?: string | null;
  finishedAt?: string | null;
}

export interface ReleaseCommit {
  sha: string;
  message: string;
  url?: string | null;
}

export interface GitSource {
  provider: SourceProvider;
  repository: string;
  url: string;
  ref: string;
  contextPath: string;
  commit: ReleaseCommit;
}

export interface ReleaseTrigger {
  kind: ReleaseTriggerKind;
  actor?: string | null;
}

export interface Release {
  id: string;
  status: ReleaseStatus;
  source?: GitSource | null;
  trigger: ReleaseTrigger;
  build?: Build | null;
  deploy?: Deploy | null;
  scans: Scan[];
  deployment?: Deployment | null;
  createdAt: string;
}

export interface Service {
  id: string;
  name: string;
  status: ServiceStatus;
  replicas: ReplicaCount;
  autoscaling?: AutoscalingSettings | null;
  port: number;
  endpoints: Endpoint[];
  sourceUrl: string;
  contextPath: string;
  resources: Resources;
  command: string;
  defaultCommand: string;
  activeDeployment?: Deployment | null;
  deployments: Deployment[];
  builds: Build[];
  releases: Release[];
  lastDeployedAt?: string | null;
  createdAt: string;
}

export interface Database {
  id: string;
  name: string;
  version: string;
  instances: number;
  status: DatabaseStatus;
  size: string;
  createdAt: string;
}

export interface KeyValueStore {
  id: string;
  name: string;
  version: string;
  status: DatabaseStatus;
  size: string;
  createdAt: string;
}

export interface Bucket {
  id: string;
  name: string;
  region: string;
  endpoint: string;
  status: BucketStatus;
  sizeBytes: number;
  objectCount: number;
  createdAt: string;
}

export interface Mount {
  service: string;
  path: string;
}

export interface Volume {
  id: string;
  name: string;
  size: string;
  mount?: Mount | null;
  usagePercent?: number | null;
}

export interface Environment {
  id: string;
  name: string;
  resourceTier: ResourceTier;
  services: Service[];
  databases: Database[];
  keyValueStores: KeyValueStore[];
  buckets: Bucket[];
  volumes: Volume[];
}

const activeEnvironment = ref<Environment | null>(null);
const environments = ref<Environment[]>([]);

const TERMINAL_BUILD_STATUSES = new Set<string>(['SUCCEEDED', 'FAILED', 'CANCELLED']);

export function activeBuild(service: Pick<Service, 'builds'>): Build | null {
  const list = service.builds ?? [];
  return list.find(b => !TERMINAL_BUILD_STATUSES.has(b.status)) ?? null;
}

export function useEnvironment() {
  function setEnvironments(envs: Environment[], preferredEnvId?: string) {
    environments.value = envs;

    if (preferredEnvId) {
      const preferred = envs.find(e => e.id === preferredEnvId);
      if (preferred) {
        activeEnvironment.value = preferred;
        return;
      }
    }

    if (!activeEnvironment.value || !envs.find(e => e.id === activeEnvironment.value!.id)) {
      activeEnvironment.value = envs[0] ?? null;
    }
  }

  function setEnvironment(env: Environment) {
    activeEnvironment.value = env;
  }

  function setEnvironmentById(id: string) {
    const env = environments.value.find(e => e.id === id);
    if (env) {
      activeEnvironment.value = env;
    }
  }

  const activeEnvServices = computed(() => activeEnvironment.value?.services ?? []);
  const activeEnvDatabases = computed(() => activeEnvironment.value?.databases ?? []);
  const activeEnvKeyValueStores = computed(() => activeEnvironment.value?.keyValueStores ?? []);
  const activeEnvBuckets = computed(() => activeEnvironment.value?.buckets ?? []);
  const activeEnvVolumes = computed(() => activeEnvironment.value?.volumes ?? []);

  return {
    activeEnvironment,
    environments,
    activeEnvServices,
    activeEnvDatabases,
    activeEnvKeyValueStores,
    activeEnvBuckets,
    activeEnvVolumes,
    setEnvironments,
    setEnvironment,
    setEnvironmentById,
  };
}
