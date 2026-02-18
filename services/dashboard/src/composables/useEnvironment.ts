import { ref, computed, watch } from 'vue';

export interface Environment {
  id: string;
  name: string;
  namespace: string;
  ephemeral: boolean;
  syncStatus: string;
  services: {
    name: string;
    imageTag: string;
    ready: boolean;
    replicas: number;
  }[];
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

  function refreshActiveEnvironment(envs: Environment[]) {
    if (activeEnvironment.value) {
      const updated = envs.find(e => e.id === activeEnvironment.value!.id);
      if (updated) {
        activeEnvironment.value = updated;
      }
    }
  }

  return {
    activeEnvironment,
    environments,
    activeEnvServices,
    setEnvironments,
    setEnvironment,
    refreshActiveEnvironment,
  };
}
