<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const replicas = ref(2);
const auto = ref(false);

const pods = computed(() => Array.from({ length: replicas.value }, (_, i) => i));

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { replicas.value = 3; }, 500);
  setTimeout(() => { replicas.value = 5; }, 1100);
  setTimeout(() => { auto.value = true; }, 1550);
  setTimeout(() => { replicas.value = 8; }, 2000);
});
</script>

<template>
  <div
    ref="root"
    class="bento-scale"
  >
    <TransitionGroup
      name="bento-pod"
      tag="div"
      class="bento-pods"
    >
      <div
        v-for="i in pods"
        :key="i"
        class="bento-pod"
      />
    </TransitionGroup>

    <div class="bento-scale-meta">
      <span class="bento-scale-count">{{ replicas }} replicas</span>
      <span
        class="bento-scale-badge"
        :class="{ 'bento-scale-badge-on': auto }"
      >
        <UIcon
          name="i-lucide-zap"
          class="size-2.5"
        />
        auto-scaling
      </span>
    </div>
  </div>
</template>

<style scoped>
.bento-scale {
  height: 230px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18px;
  padding: 20px 24px;
}

.bento-pods {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
  align-content: center;
  max-width: 180px;
}

.bento-pod {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--bento-accent-subtle) 0%, var(--ui-bg-elevated) 100%);
  border: 1.5px solid var(--bento-accent);
}

.bento-scale-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
}

.bento-scale-count {
  font-weight: 600;
  color: var(--ui-text);
  font-variant-numeric: tabular-nums;
}

.bento-scale-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 3px 9px;
  border-radius: 10px;
  border: 1px solid var(--ui-border);
  color: var(--ui-text-muted);
  font-weight: 500;
  transition: all 0.4s ease;
}

.bento-scale-badge-on {
  border-color: transparent;
  background: var(--bento-accent-subtle);
  color: var(--bento-accent);
}

/* Pod enter animation */
.bento-pod-enter-active {
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.bento-pod-enter-from {
  opacity: 0;
  transform: scale(0.4);
}

.bento-pod-move {
  transition: transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}
</style>
