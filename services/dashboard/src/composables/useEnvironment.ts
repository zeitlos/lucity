import { ref, computed } from 'vue';

export interface DeploymentInfo {
  id: string;
  imageTag: string;
  active: boolean;
  timestamp?: string;
  sourceCommitMessage?: string;
  sourceUrl?: string;
}

export interface DomainInfo {
  hostname: string;
  type: 'PLATFORM' | 'CUSTOM';
  dnsStatus: 'VALID' | 'PENDING' | 'MISCONFIGURED' | 'ERROR';
}

export interface AutoscalingInfo {
  enabled: boolean;
  minReplicas: number;
  maxReplicas: number;
  targetCPU: number;
}

export interface ScalingInfo {
  replicas: number;
  autoscaling?: AutoscalingInfo;
}

export interface ResourcesInfo {
  cpuMillicores: number;
  memoryMB: number;
  cpuLimitMillicores: number;
  memoryLimitMB: number;
}

export interface ServiceInstance {
  name: string;
  environment: string;
  image: string;
  port?: number;
  framework?: string;
  sourceUrl?: string;
  contextPath?: string;
  startCommand?: string;
  customStartCommand?: string;
  imageTag: string;
  ready: boolean;
  replicas: number;
  scaling?: ScalingInfo;
  resources?: ResourcesInfo;
  domains: DomainInfo[];
  deployments: DeploymentInfo[];
}

export interface VolumeInfo {
  name: string;
  size: string;
  requestedSize: string;
  usedBytes: number;
  capacityBytes: number;
}

export interface DatabaseInstance {
  name: string;
  environment: string;
  ready: boolean;
  instances: number;
  version: string;
  size: string;
  volume?: VolumeInfo;
}

export interface Environment {
  id: string;
  name: string;
  namespace: string;
  ephemeral: boolean;
  syncStatus: string;
  resourceTier?: string;
  services: ServiceInstance[];
  databases: DatabaseInstance[];
}

const activeEnvironment = ref<Environment | null>(null);
const environments = ref<Environment[]>([]);

export function useEnvironment() {
  function setEnvironments(envs: Environment[], preferredEnvName?: string) {
    environments.value = envs;

    if (preferredEnvName) {
      const preferred = envs.find(e => e.name === preferredEnvName);
      if (preferred) {
        activeEnvironment.value = preferred;
        return;
      }
    }

    if (!activeEnvironment.value || !envs.find(e => e.id === activeEnvironment.value!.id)) {
      const nonEphemeral = envs.find(e => !e.ephemeral);
      activeEnvironment.value = nonEphemeral ?? envs[0] ?? null;
    }
  }

  function setEnvironment(env: Environment) {
    activeEnvironment.value = env;
  }

  function setEnvironmentByName(name: string) {
    const env = environments.value.find(e => e.name === name);
    if (env) {
      activeEnvironment.value = env;
    }
  }

  const activeEnvServices = computed(() => activeEnvironment.value?.services ?? []);
  const activeEnvDatabases = computed(() => activeEnvironment.value?.databases ?? []);

  function refreshActiveEnvironment(envs: Environment[]) {
    if (activeEnvironment.value) {
      const updated = envs.find(e => e.id === activeEnvironment.value!.id);
      if (updated) {
        activeEnvironment.value = updated;
      }
    }
  }

  function updateServiceDomains(serviceName: string, domains: DomainInfo[]) {
    if (!activeEnvironment.value) return;
    activeEnvironment.value = {
      ...activeEnvironment.value,
      services: activeEnvironment.value.services.map(s =>
        s.name === serviceName ? { ...s, domains } : s,
      ),
    };
  }

  return {
    activeEnvironment,
    environments,
    activeEnvServices,
    activeEnvDatabases,
    setEnvironments,
    setEnvironment,
    setEnvironmentByName,
    refreshActiveEnvironment,
    updateServiceDomains,
  };
}
