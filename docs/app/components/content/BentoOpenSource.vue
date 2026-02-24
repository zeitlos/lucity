<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const shown = ref(0);

const tools = [
  { name: 'ArgoCD', icon: 'i-simple-icons-argo' },
  { name: 'Helm', icon: 'i-simple-icons-helm' },
  { name: 'CloudNativePG', icon: 'i-simple-icons-postgresql' },
  { name: 'Zot', icon: 'i-lucide-package' },
  { name: 'Soft-serve', icon: 'i-lucide-git-branch' },
  { name: 'Gateway API', icon: 'i-simple-icons-kubernetes' },
];

watch(visible, (v) => {
  if (!v) return;
  tools.forEach((_, i) => {
    setTimeout(() => { shown.value = i + 1; }, 200 + i * 150);
  });
});
</script>

<template>
  <div
    ref="root"
    class="bento-opensource"
  >
    <div class="bento-grid">
      <div
        v-for="(tool, i) in tools"
        :key="tool.name"
        class="bento-tool"
        :class="{ 'bento-tool-visible': shown > i }"
        :style="{ animationDelay: `${i * 80}ms` }"
      >
        <UIcon
          :name="tool.icon"
          class="size-3.5 text-(--ui-text-muted)"
        />
        <span class="text-[11px] font-medium">{{ tool.name }}</span>
      </div>
    </div>
    <div
      v-if="shown >= tools.length"
      class="bento-license"
    >
      AGPL-3.0 &middot; Forever open
    </div>
  </div>
</template>

<style scoped>
.bento-opensource {
  min-height: 140px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  gap: 12px;
}

.bento-grid {
  display: grid;
  grid-template-columns: repeat(3, auto);
  gap: 8px;
  justify-content: center;
}

.bento-tool {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 16px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  color: var(--ui-text);
  opacity: 0;
}

.bento-tool-visible {
  animation: bento-tool-in 0.35s cubic-bezier(0.16, 1, 0.3, 1) both;
}

.bento-license {
  font-size: 11px;
  color: var(--bento-accent);
  font-weight: 500;
  letter-spacing: 0.02em;
  animation: bento-fade-in 0.5s ease both;
  animation-delay: 0.3s;
}

@keyframes bento-tool-in {
  from { opacity: 0; transform: scale(0.8); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes bento-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
