import { ref } from 'vue';

export const environments = ['dev', 'prod'];

export const environment = ref<string | null>(null);

export const environmentNames: Record<string, string> = {
  dev: 'Development',
  prod: 'Production',
};
