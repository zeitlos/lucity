<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const phase = ref<'spin' | 'explode' | 'done'>('spin');

const outputs = [
  { icon: 'i-simple-icons-helm', label: 'Helm chart', path: 'ejected/chart/' },
  { icon: 'i-simple-icons-argo', label: 'ArgoCD apps', path: 'ejected/argocd/' },
  { icon: 'i-lucide-settings', label: 'Env values', path: 'ejected/environments/' },
  { icon: 'i-lucide-file-text', label: 'README', path: 'ejected/README.md' },
];

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { phase.value = 'explode'; }, 1200);
  setTimeout(() => { phase.value = 'done'; }, 1800);
});
</script>

<template>
  <div
    ref="root"
    class="bento-eject"
  >
    <!-- Spinner phase -->
    <div
      v-if="phase === 'spin'"
      class="bento-spinner-wrap"
    >
      <svg
        class="bento-spinner"
        viewBox="0 0 40 40"
        width="40"
        height="40"
      >
        <circle
          cx="20"
          cy="20"
          r="16"
          fill="none"
          stroke="var(--ui-border)"
          stroke-width="3"
        />
        <circle
          cx="20"
          cy="20"
          r="16"
          fill="none"
          stroke="var(--ui-primary)"
          stroke-width="3"
          stroke-linecap="round"
          stroke-dasharray="80"
          stroke-dashoffset="60"
          class="bento-spinner-arc"
        />
      </svg>
    </div>

    <!-- Explode phase -->
    <div
      v-else
      class="bento-outputs"
    >
      <div
        v-for="(item, i) in outputs"
        :key="item.label"
        class="bento-output"
        :style="{ animationDelay: `${i * 100}ms` }"
      >
        <UIcon
          :name="item.icon"
          class="size-4 shrink-0 text-(--ui-primary)"
        />
        <div class="min-w-0">
          <div class="text-xs font-medium text-(--ui-text)">
            {{ item.label }}
          </div>
          <div class="truncate font-mono text-[10px] text-(--ui-text-muted)">
            {{ item.path }}
          </div>
        </div>
      </div>
      <div
        v-if="phase === 'done'"
        class="bento-tagline"
      >
        Your infrastructure is yours.
      </div>
    </div>
  </div>
</template>

<style scoped>
.bento-eject {
  min-height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.bento-spinner-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
}

.bento-spinner-arc {
  animation: bento-spin 1s linear infinite;
  transform-origin: center;
}

.bento-outputs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  width: 100%;
  max-width: 360px;
}

.bento-output {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  animation: bento-scale-in 0.4s cubic-bezier(0.16, 1, 0.3, 1) both;
}

.bento-tagline {
  grid-column: 1 / -1;
  text-align: center;
  font-family: var(--font-serif);
  font-size: 14px;
  color: var(--ui-text-muted);
  animation: bento-fade-in 0.6s ease both;
  animation-delay: 0.5s;
}

@keyframes bento-spin {
  to { transform: rotate(360deg); }
}

@keyframes bento-scale-in {
  from { opacity: 0; transform: scale(0.5); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes bento-fade-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
