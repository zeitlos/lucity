<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const step = ref(0);

const nodes = [
  { icon: 'i-lucide-code', label: 'Code' },
  { icon: 'i-simple-icons-github', label: 'GitHub' },
  { icon: 'i-lucide-hammer', label: 'Build' },
  { icon: 'i-lucide-container', label: 'Registry' },
  { icon: 'i-simple-icons-kubernetes', label: 'K8s' },
];

watch(visible, (v) => {
  if (!v) return;
  const delays = [0, 600, 1200, 1800, 2400];
  delays.forEach((delay, i) => {
    setTimeout(() => { step.value = i + 1; }, delay);
  });
});
</script>

<template>
  <div
    ref="root"
    class="bento-deploy"
  >
    <div class="flex items-center justify-between gap-1 px-2 py-6 sm:gap-2 sm:px-4">
      <template
        v-for="(node, i) in nodes"
        :key="node.label"
      >
        <!-- Node -->
        <div
          class="bento-node"
          :class="{ 'bento-node-active': step > i, 'bento-node-current': step === i + 1 }"
        >
          <UIcon
            :name="node.icon"
            class="size-4 sm:size-5"
          />
          <span class="text-[10px] font-medium sm:text-xs">{{ node.label }}</span>
          <!-- Checkmark on final node -->
          <span
            v-if="i === nodes.length - 1 && step >= nodes.length"
            class="bento-check"
          >
            <UIcon
              name="i-lucide-check"
              class="size-2.5"
            />
          </span>
        </div>

        <!-- Connector line -->
        <div
          v-if="i < nodes.length - 1"
          class="bento-line"
          :class="{ 'bento-line-active': step > i + 1 }"
        >
          <div
            v-if="step === i + 2"
            class="bento-dot"
          />
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.bento-deploy {
  min-height: 80px;
}

.bento-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 10px;
  border-radius: 10px;
  border: 1.5px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  color: var(--ui-text-muted);
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
  position: relative;
  min-width: 56px;
}

.bento-node-active {
  border-color: var(--ui-primary);
  color: var(--ui-text);
}

.bento-node-current {
  box-shadow: 0 0 12px oklch(0.75 0.18 160 / 0.25);
}

.bento-check {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--ui-primary);
  color: white;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: bento-scale-in 0.3s cubic-bezier(0.16, 1, 0.3, 1) both;
}

.bento-line {
  flex: 1;
  height: 2px;
  background: var(--ui-border);
  border-radius: 1px;
  position: relative;
  transition: background 0.3s ease;
  min-width: 12px;
}

.bento-line-active {
  background: var(--ui-primary);
}

.bento-dot {
  position: absolute;
  top: 50%;
  left: 0;
  width: 8px;
  height: 8px;
  margin-top: -4px;
  border-radius: 50%;
  background: var(--ui-primary);
  animation: bento-travel 0.5s ease-in-out forwards;
  box-shadow: 0 0 8px oklch(0.75 0.18 160 / 0.5);
}

@keyframes bento-travel {
  from { left: 0; }
  to { left: calc(100% - 8px); }
}

@keyframes bento-scale-in {
  from { opacity: 0; transform: scale(0); }
  to { opacity: 1; transform: scale(1); }
}
</style>
