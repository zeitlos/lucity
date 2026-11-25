<script setup lang="ts">
import { computed } from 'vue';
import { Server } from '@lucide/vue';
import { cn } from '@/lib/utils';
import { modelProviderBrand, type ModelProviderBrand, type ModelProviderIdentity } from '@/lib/agents';

const props = withDefaults(defineProps<{
  brand?: ModelProviderBrand;
  provider?: ModelProviderIdentity;
  size?: 'sm' | 'md';
}>(), {
  brand: undefined,
  provider: undefined,
  size: 'md',
});

interface Mark {
  path: string;
  viewBox?: string;
  background: string;
  color: string;
}

const marks: Record<Exclude<ModelProviderBrand, 'generic'>, Mark> = {
  anthropic: {
    path: 'M17.3041 3.541h-3.6718l6.696 16.918H24Zm-10.6082 0L0 20.459h3.7442l1.3693-3.5527h7.0052l1.3693 3.5528h3.7442L10.5363 3.5409Zm-.3712 10.2232 2.2914-5.9456 2.2914 5.9456Z',
    background: '#D97757',
    color: '#141413',
  },
  openai: {
    path: 'M22.2819 9.8211a5.9847 5.9847 0 0 0-.5157-4.9108 6.0462 6.0462 0 0 0-6.5098-2.9A6.0651 6.0651 0 0 0 4.9807 4.1818a5.9847 5.9847 0 0 0-3.9977 2.9 6.0462 6.0462 0 0 0 .7427 7.0966 5.98 5.98 0 0 0 .511 4.9107 6.051 6.051 0 0 0 6.5146 2.9001A5.9847 5.9847 0 0 0 13.2599 24a6.0557 6.0557 0 0 0 5.7718-4.2058 5.9894 5.9894 0 0 0 3.9977-2.9001 6.0557 6.0557 0 0 0-.7475-7.0729zm-9.022 12.6081a4.4755 4.4755 0 0 1-2.8764-1.0408l.1419-.0804 4.7783-2.7582a.7948.7948 0 0 0 .3927-.6813v-6.7369l2.02 1.1686a.071.071 0 0 1 .038.052v5.5826a4.504 4.504 0 0 1-4.4945 4.4944zm-9.6607-4.1254a4.4708 4.4708 0 0 1-.5346-3.0137l.142.0852 4.783 2.7582a.7712.7712 0 0 0 .7806 0l5.8428-3.3685v2.3324a.0804.0804 0 0 1-.0332.0615L9.74 19.9502a4.4992 4.4992 0 0 1-6.1408-1.6464zM2.3408 7.8956a4.485 4.485 0 0 1 2.3655-1.9728V11.6a.7664.7664 0 0 0 .3879.6765l5.8144 3.3543-2.0201 1.1685a.0757.0757 0 0 1-.071 0l-4.8303-2.7865A4.504 4.504 0 0 1 2.3408 7.872zm16.5963 3.8558L13.1038 8.364 15.1192 7.2a.0757.0757 0 0 1 .071 0l4.8303 2.7913a4.4944 4.4944 0 0 1-.6765 8.1042v-5.6772a.79.79 0 0 0-.407-.667zm2.0107-3.0231l-.142-.0852-4.7735-2.7818a.7759.7759 0 0 0-.7854 0L9.409 9.2297V6.8974a.0662.0662 0 0 1 .0284-.0615l4.8303-2.7866a4.4992 4.4992 0 0 1 6.6802 4.66zM8.3065 12.863l-2.02-1.1638a.0804.0804 0 0 1-.038-.0567V6.0742a4.4992 4.4992 0 0 1 7.3757-3.4537l-.142.0805L8.704 5.459a.7948.7948 0 0 0-.3927.6813zm1.0976-2.3654l2.602-1.4998 2.6069 1.4998v2.9994l-2.5974 1.4997-2.6067-1.4997Z',
    background: 'var(--foreground)',
    color: 'var(--background)',
  },
  mistral: {
    path: 'M17.143 3.429v3.428h-3.429v3.429h-3.428V6.857H6.857V3.43H3.43v13.714H0v3.428h10.286v-3.428H6.857v-3.429h3.429v3.429h3.429v-3.429h3.428v3.429h-3.428v3.428H24v-3.428h-3.43V3.429z',
    background: '#FA520F',
    color: '#FFFFFF',
  },
  infomaniak: {
    path: 'M5.12 0h4.992v11.7119l3.648-4.128h6l-4.58 4.3506 4.868 8.1296h-5.52l-2.5938-5.0211L10.112 16.8v3.264H5.12Z',
    viewBox: '0.58 -1.97 24 24',
    background: '#0098FF',
    color: '#FFFFFF',
  },
  scaleway: {
    path: 'M16.605 11.11v5.72a1.77 1.77 0 01-1.54 1.69h-4a1.43 1.43 0 01-1.31-1.22 1.09 1.09 0 010-.18 1.37 1.37 0 011.37-1.36h1.74a1 1 0 001-1v-3.62a1.4 1.4 0 011.18-1.39h.17a1.37 1.37 0 011.39 1.36zm-6.46 1.74V9.26a1 1 0 011-1h1.85a1.37 1.37 0 001.37-1.37 1 1 0 000-.17 1.45 1.45 0 00-1.41-1.2h-3.96a1.81 1.81 0 00-1.58 1.66v5.7a1.37 1.37 0 001.37 1.37h.21a1.4 1.4 0 001.15-1.4zm12-4.29V20a4.53 4.53 0 01-4.15 4h-7.58a8.57 8.57 0 01-8.56-8.57V4.54A4.54 4.54 0 016.395 0h7.18a8.56 8.56 0 018.56 8.56zm-2.74 0a5.83 5.83 0 00-5.82-5.82h-7.19a1.79 1.79 0 00-1.8 1.8v10.89a5.83 5.83 0 005.82 5.8h7.44a1.79 1.79 0 001.54-1.48z',
    background: '#4F0599',
    color: '#FFFFFF',
  },
  hetzner: {
    path: 'M4.602 4.025h2.244c.509 0 .716.215.716.717v5.64h8.883v-5.64c0-.509.215-.717.717-.717h2.229c.5 0 .71.23.724.717v14.516c0 .509-.215.717-.717.717h-2.23c-.51 0-.717-.215-.717-.717v-5.735H7.562v5.735c0 .516-.215.717-.716.717H4.602c-.51 0-.717-.208-.717-.717V4.742c0-.509.207-.717.717-.717z',
    background: '#D50C2D',
    color: '#FFFFFF',
  },
  opencode: {
    path: 'M22 24H2V0h20zM17 4.8H7v14.4h10z',
    background: 'var(--foreground)',
    color: 'var(--background)',
  },
};

const resolved = computed<ModelProviderBrand>(() => props.brand ?? (props.provider ? modelProviderBrand(props.provider) : 'generic'));
const mark = computed(() => (resolved.value === 'generic' ? null : marks[resolved.value]));
</script>

<template>
  <span
    :class="cn(
      'flex shrink-0 items-center justify-center rounded-sm',
      size === 'sm' ? 'size-5' : 'size-8',
      !mark && 'bg-muted text-muted-foreground',
    )"
    :style="mark ? { background: mark.background, color: mark.color } : undefined"
    aria-hidden="true"
  >
    <svg
      v-if="mark"
      :viewBox="mark.viewBox ?? '0 0 24 24'"
      :class="size === 'sm' ? 'size-3' : 'size-4.5'"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path :d="mark.path" />
    </svg>
    <Server v-else :class="size === 'sm' ? 'size-3' : 'size-4.5'" />
  </span>
</template>
