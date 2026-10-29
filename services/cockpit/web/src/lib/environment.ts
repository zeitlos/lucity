import { ref } from 'vue';

export const environments = ['dev', 'prod'];

export const environment = ref<string | null>(null);
