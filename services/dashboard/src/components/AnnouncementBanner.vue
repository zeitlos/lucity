<script setup lang="ts">
import { onMounted, ref, watchEffect } from 'vue';
import { useElementSize } from '@vueuse/core';
import { Info, OctagonAlert, TriangleAlert, X } from '@lucide/vue';
import { Button } from '@/components/ui/button';

interface Banner {
  id: string;
  message: string;
  severity: 'info' | 'warn' | 'danger';
  dismissible: boolean;
}

const DISMISSED_KEY = 'lucity-banner-dismissed';

const tones = {
  info: { icon: Info, color: 'var(--status-progress)' },
  warn: { icon: TriangleAlert, color: 'var(--status-warn)' },
  danger: { icon: OctagonAlert, color: 'var(--status-danger)' },
};

const banner = ref<Banner | null>(null);
const root = ref<HTMLElement>();
const { height } = useElementSize(root, undefined, { box: 'border-box' });

watchEffect(() => {
  document.documentElement.style.setProperty('--banner-height', `${height.value}px`);
});

onMounted(async () => {
  const res = await fetch(`${import.meta.env.BASE_URL}banner.json`, { cache: 'no-cache' }).catch(() => null);
  const data: Banner | null = res?.ok ? await res.json().catch(() => null) : null;
  if (!data?.message) return;
  if (data.dismissible && localStorage.getItem(DISMISSED_KEY) === data.id) return;
  banner.value = data;
});

function dismiss(id: string) {
  localStorage.setItem(DISMISSED_KEY, id);
  banner.value = null;
}
</script>

<template>
  <div
    v-if="banner"
    ref="root"
    role="status"
    class="banner relative z-1 flex items-center justify-center gap-2 border-b px-11 py-2 text-center text-xs font-medium text-foreground"
    :style="{ '--tone': tones[banner.severity].color }"
  >
    <component :is="tones[banner.severity].icon" :size="14" class="banner-icon shrink-0" />
    <span class="text-pretty">{{ banner.message }}</span>
    <Button
      v-if="banner.dismissible"
      variant="ghost"
      size="icon"
      class="absolute top-1/2 right-2 h-6 w-6 -translate-y-1/2 text-muted-foreground"
      aria-label="Dismiss"
      @click="dismiss(banner.id)"
    >
      <X />
    </Button>
  </div>
</template>

<style scoped>
.banner {
  background-color: color-mix(in oklab, var(--tone) 14%, var(--background));
  border-color: color-mix(in oklab, var(--tone) 35%, var(--background));
}

.banner-icon {
  color: var(--tone);
}
</style>
