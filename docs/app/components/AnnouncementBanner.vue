<script setup lang="ts">
interface Banner {
  id: string;
  message: string;
  severity: 'info' | 'warn' | 'danger';
  dismissible: boolean;
}

const DISMISSED_KEY = 'lucity-banner-dismissed';

const tones = {
  info: { icon: 'i-lucide-info', color: 'var(--ui-info)' },
  warn: { icon: 'i-lucide-triangle-alert', color: 'var(--ui-warning)' },
  danger: { icon: 'i-lucide-octagon-alert', color: 'var(--ui-error)' },
};

const banner = ref<Banner | null>(null);

onMounted(async () => {
  const res = await fetch('/banner.json', { cache: 'no-cache' }).catch(() => null);
  const data: Banner | null = res?.ok ? await res.json().catch(() => null) : null;
  if (!data?.message) return;
  try {
    if (data.dismissible && localStorage.getItem(DISMISSED_KEY) === data.id) return;
  } catch {}
  banner.value = data;
});

function dismiss(id: string) {
  banner.value = null;
  try {
    localStorage.setItem(DISMISSED_KEY, id);
  } catch {}
}
</script>

<template>
  <div
    v-if="banner"
    role="status"
    class="banner relative flex items-center justify-center gap-2 border-b px-12 py-2.5 text-center text-sm font-medium text-highlighted"
    :style="{ '--tone': tones[banner.severity].color }"
  >
    <UIcon :name="tones[banner.severity].icon" class="banner-icon size-4 shrink-0" />
    <span class="text-pretty">{{ banner.message }}</span>
    <UButton
      v-if="banner.dismissible"
      icon="i-lucide-x"
      color="neutral"
      variant="ghost"
      size="sm"
      aria-label="Dismiss"
      class="absolute top-1/2 right-3 -translate-y-1/2"
      @click="dismiss(banner.id)"
    />
  </div>
</template>

<style scoped>
.banner {
  background-color: color-mix(in oklab, var(--tone) 14%, var(--ui-bg));
  border-color: color-mix(in oklab, var(--tone) 35%, var(--ui-bg));
}

.banner-icon {
  color: var(--tone);
}
</style>
