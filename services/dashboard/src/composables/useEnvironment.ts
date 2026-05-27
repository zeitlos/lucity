import { ref, computed } from 'vue';
import type {
  BuildStatus,
  DatabaseStatus,
  ResourceTier,
  ServiceStatus,
  Protocol,
  DeploymentStatus,
} from '@/gql/graphql';

export interface Endpoint {
  host: string;
  port: number;
  protocol: Protocol;
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
  deployedBy: string;
}

export interface Build {
  id: string;
  status: BuildStatus;
  startedAt: string;
  finishedAt?: string | null;
}

export interface Service {
  id: string;
  name: string;
  status: ServiceStatus;
  replicas: ReplicaCount;
  autoscaling?: AutoscalingSettings | null;
  endpoints: Endpoint[];
  sourceUrl: string;
  contextPath: string;
  resources: Resources;
  command: string;
  defaultCommand: string;
  activeDeployment?: Deployment | null;
  deployments: Deployment[];
  builds: Build[];
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

export interface Environment {
  id: string;
  name: string;
  resourceTier: ResourceTier;
  services: Service[];
  databases: Database[];
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

  return {
    activeEnvironment,
    environments,
    activeEnvServices,
    activeEnvDatabases,
    setEnvironments,
    setEnvironment,
    setEnvironmentById,
  };
}
