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
  dnsStatus: 'VALID' | 'PENDING' | 'ERROR';
}

export interface ServiceInstance {
  name: string;
  environment: string;
  imageTag: string;
  ready: boolean;
  replicas: number;
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
  services: ServiceInstance[];
  databases: DatabaseInstance[];
}

const activeEnvironment = ref<Environment | null>(null);
const environments = ref<Environment[]>([]);

export function useEnvironment() {
  function setEnvironments(envs: Environment[]) {
    environments.value = envs;

    if (!activeEnvironment.value || !envs.find(e => e.id === activeEnvironment.value!.id)) {
      const nonEphemeral = envs.find(e => !e.ephemeral);
      activeEnvironment.value = nonEphemeral ?? envs[0] ?? null;
    }
  }

  function setEnvironment(env: Environment) {
    activeEnvironment.value = env;
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
    const svc = activeEnvironment.value.services.find(s => s.name === serviceName);
    if (svc) {
      svc.domains = domains;
    }
  }

  return {
    activeEnvironment,
    environments,
    activeEnvServices,
    activeEnvDatabases,
    setEnvironments,
    setEnvironment,
    refreshActiveEnvironment,
    updateServiceDomains,
  };
}
