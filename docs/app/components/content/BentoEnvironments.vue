<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const step = ref(0);

const envs = [
  { name: 'production', color: 'var(--ui-primary)' },
  { name: 'staging', color: 'oklch(0.75 0.15 85)' },
  { name: 'development', color: 'oklch(0.65 0.15 250)' },
];

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { step.value = 1; }, 300);
  setTimeout(() => { step.value = 2; }, 900);
  setTimeout(() => { step.value = 3; }, 1500);
});
</script>

<template>
  <div
    ref="root"
    class="bento-envs"
  >
    <div class="flex flex-col gap-2.5">
      <!-- Permanent envs -->
      <div
        v-for="(env, i) in envs"
        :key="env.name"
        class="bento-env"
        :class="{ 'bento-env-active': step > 0 }"
        :style="{ animationDelay: `${i * 120}ms` }"
      >
        <span
          class="bento-env-dot"
          :style="{ background: env.color }"
        />
        <span class="text-xs font-medium">{{ env.name }}</span>
      </div>

      <!-- Clone animation -->
      <div
        v-if="step >= 2"
        class="bento-env bento-env-clone"
      >
        <span
          class="bento-env-dot"
          :style="{ background: 'oklch(0.70 0.12 310)' }"
        />
        <span class="text-xs font-medium">pr-142</span>
        <span
          v-if="step >= 3"
          class="bento-badge"
        >
          <UIcon
            name="i-lucide-check"
            class="size-2.5"
          />
          Ready in 8s
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bento-envs {
  min-height: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 24px;
}

.bento-env {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  border-radius: 20px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  color: var(--ui-text);
  animation: bento-slide-in 0.4s cubic-bezier(0.16, 1, 0.3, 1) both;
}

.bento-env-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.bento-env-clone {
  animation: bento-clone-in 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
  border-style: dashed;
}

.bento-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  margin-left: auto;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--bento-accent-subtle);
  color: var(--bento-accent);
  font-size: 10px;
  font-weight: 500;
  animation: bento-fade-in 0.3s ease both;
}

@keyframes bento-slide-in {
  from { opacity: 0; transform: translateX(-8px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes bento-clone-in {
  from { opacity: 0; transform: translateY(-8px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

@keyframes bento-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
