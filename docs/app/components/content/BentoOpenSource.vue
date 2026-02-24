<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const shown = ref(0);

const colorMode = useColorMode();
const isDark = computed(() => colorMode.value === 'dark');

const tools = [
  { name: 'ArgoCD', icon: 'i-simple-icons-argo', color: 'oklch(0.52 0.18 30)', darkColor: 'oklch(0.70 0.16 30)' },
  { name: 'Helm', icon: 'i-simple-icons-helm', color: 'oklch(0.50 0.16 250)', darkColor: 'oklch(0.70 0.14 250)' },
  { name: 'CloudNativePG', icon: 'i-simple-icons-postgresql', color: 'oklch(0.50 0.16 250)', darkColor: 'oklch(0.70 0.14 250)' },
  { name: 'Zot', icon: 'i-lucide-package', color: 'oklch(0.55 0.16 30)', darkColor: 'oklch(0.72 0.14 30)' },
  { name: 'Soft-serve', icon: 'i-lucide-git-branch', color: 'oklch(0.50 0.16 140)', darkColor: 'oklch(0.70 0.14 140)' },
  { name: 'Gateway API', icon: 'i-simple-icons-kubernetes', color: 'oklch(0.50 0.16 250)', darkColor: 'oklch(0.70 0.14 250)' },
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
    <div class="bento-tools-grid">
      <div
        v-for="(tool, i) in tools"
        :key="tool.name"
        class="bento-tool"
        :class="{ 'bento-tool-visible': shown > i }"
        :style="{ animationDelay: `${i * 80}ms` }"
      >
        <UIcon
          :name="tool.icon"
          class="size-5 sm:size-7"
          :style="{ color: isDark ? tool.darkColor : tool.color }"
        />
        <span class="text-xs font-medium sm:text-sm">{{ tool.name }}</span>
      </div>
    </div>
    <div
      v-if="shown >= tools.length"
      class="bento-license"
    >
      AGPL-3.0 &middot; Forever open
    </div>

    <!-- Gopher mascot overflowing bottom-right -->
    <img
      src="/img/gopher_ship.webp"
      alt=""
      class="bento-gopher"
      :style="{ opacity: isDark ? 0.10 : 0.18 }"
    >
  </div>
</template>

<style scoped>
/* Fixed height — prevents the grid from shifting when the
   tool pills and license badge animate in. */
.bento-opensource {
  height: 180px;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  gap: 14px;
  overflow: visible;
}

.bento-tools-grid {
  display: grid;
  grid-template-columns: repeat(3, auto);
  gap: 10px;
  justify-content: center;
  position: relative;
  z-index: 1;
}

@media (min-width: 640px) {
  .bento-tools-grid {
    gap: 12px;
  }
}

.bento-tool {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  color: var(--ui-text);
  opacity: 0;
}

@media (min-width: 640px) {
  .bento-tool {
    padding: 8px 18px;
    gap: 10px;
  }
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
  position: relative;
  z-index: 1;
}

/* Gopher image — overflows bottom-right corner */
.bento-gopher {
  position: absolute;
  bottom: -16px;
  right: -12px;
  width: 90px;
  opacity: 0.18;
  pointer-events: none;
  z-index: 0;
  /* Dark mode opacity handled via inline :style binding (isDark ternary).
     DO NOT use :global(.dark) in scoped CSS — it breaks scoping and
     applies styles to <html class="dark"> itself, lowering the opacity
     of the entire page. */
}

@media (min-width: 640px) {
  .bento-gopher {
    width: 130px;
    bottom: -20px;
    right: -16px;
  }
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
