<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const step = ref(0);

const topRow = [
  { icon: 'i-lucide-code', label: 'Code' },
  { icon: 'i-simple-icons-github', label: 'GitHub' },
  { icon: 'i-lucide-hammer', label: 'Build' },
];

const bottomRow = [
  { icon: 'i-lucide-container', label: 'Registry' },
  { icon: 'i-simple-icons-argo', label: 'ArgoCD' },
  { icon: 'i-simple-icons-kubernetes', label: 'K8s' },
];

watch(visible, (v) => {
  if (!v) return;
  /* Steps 1-3: top row, step 4: snake connector, steps 5-7: bottom row */
  const delays = [0, 600, 1200, 1900, 2500, 3100, 3700];
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
    <div class="bento-pipeline">
      <!-- Top row: Code → GitHub → Build -->
      <div class="bento-row">
        <template
          v-for="(node, i) in topRow"
          :key="node.label"
        >
          <div
            class="bento-node"
            :class="{ 'bento-node-active': step > i, 'bento-node-current': step === i + 1 }"
          >
            <UIcon
              :name="node.icon"
              class="size-4 sm:size-5"
            />
            <span class="text-[10px] font-medium sm:text-xs">{{ node.label }}</span>
          </div>
          <div
            v-if="i < topRow.length - 1"
            class="bento-hline"
            :class="{ 'bento-line-active': step > i + 1 }"
          >
            <div
              v-if="step === i + 2"
              class="bento-hdot"
            />
          </div>
        </template>
      </div>

      <!-- Snake connector: Build → down → left → up → Registry.
           Wrapper has same padding as rows so horizontal positions align. -->
      <div class="bento-snake-wrap">
        <svg
          class="bento-snake"
          viewBox="0 0 200 28"
          preserveAspectRatio="none"
        >
          <!-- Background path — x=186 aligns with Build center, x=14 with Registry -->
          <path
            d="M 186 0 V 14 H 14 V 28"
            fill="none"
            :stroke="step > 3 ? 'var(--bento-accent)' : 'var(--ui-border)'"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="bento-snake-path"
          />
          <!-- Traveling dot -->
          <circle
            v-if="step === 4"
            r="4"
            :fill="'var(--bento-accent)'"
            class="bento-snake-dot"
          >
            <animateMotion
              dur="0.6s"
              fill="freeze"
              path="M 186 0 V 14 H 14 V 28"
            />
          </circle>
        </svg>
      </div>

      <!-- Bottom row: Registry → ArgoCD → K8s -->
      <div class="bento-row">
        <template
          v-for="(node, i) in bottomRow"
          :key="node.label"
        >
          <div
            class="bento-node"
            :class="{
              'bento-node-active': step > i + 4,
              'bento-node-current': step === i + 5,
            }"
          >
            <UIcon
              :name="node.icon"
              class="size-4 sm:size-5"
            />
            <span class="text-[10px] font-medium sm:text-xs">{{ node.label }}</span>
            <!-- Checkmark on final node -->
            <span
              v-if="i === bottomRow.length - 1 && step >= 7"
              class="bento-check"
            >
              <UIcon
                name="i-lucide-check"
                class="size-2.5"
              />
            </span>
          </div>
          <div
            v-if="i < bottomRow.length - 1"
            class="bento-hline"
            :class="{ 'bento-line-active': step > i + 5 }"
          >
            <div
              v-if="step === i + 6"
              class="bento-hdot"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- Tagline -->
    <div
      v-if="step >= 7"
      class="bento-tagline"
    >
      Zero YAML. No Dockerfile required.
    </div>
  </div>
</template>

<style scoped>
/* Fixed height prevents the grid from bumping down when the
   tagline animates in after the pipeline completes. */
.bento-deploy {
  height: 235px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 44px 12px 16px;
}

.bento-pipeline {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  width: 100%;
  max-width: 420px;
  margin: 0 auto;
}

.bento-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1px;
  padding: 0 4px;
}

@media (min-width: 640px) {
  .bento-row {
    gap: 2px;
    padding: 0 8px;
  }
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

@media (min-width: 640px) {
  .bento-node {
    padding: 10px 14px;
    min-width: 64px;
  }
}

.bento-node-active {
  border-color: var(--bento-accent);
  color: var(--ui-text);
  background: linear-gradient(135deg, var(--bento-accent-subtle) 0%, var(--ui-bg-elevated) 100%);
}

.bento-node-current {
  box-shadow: 0 0 14px var(--bento-accent-glow);
}

.bento-check {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--bento-accent);
  color: white;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: bento-scale-in 0.3s cubic-bezier(0.16, 1, 0.3, 1) both;
}

/* Horizontal connector lines */
.bento-hline {
  flex: 1;
  height: 2px;
  background: var(--ui-border);
  border-radius: 1px;
  position: relative;
  transition: background 0.3s ease;
  min-width: 12px;
}

.bento-line-active {
  background: var(--bento-accent);
}

.bento-hdot {
  position: absolute;
  top: 50%;
  left: 0;
  width: 8px;
  height: 8px;
  margin-top: -4px;
  border-radius: 50%;
  background: var(--bento-accent);
  animation: bento-travel-h 0.5s ease-in-out forwards;
  box-shadow: 0 0 8px var(--bento-accent-glow);
}

/* Snake SVG connector — wrapper matches row padding for alignment */
.bento-snake-wrap {
  padding: 0 4px;
}

@media (min-width: 640px) {
  .bento-snake-wrap {
    padding: 0 8px;
  }
}

.bento-snake {
  width: 100%;
  height: 28px;
  display: block;
}

.bento-snake-path {
  transition: stroke 0.3s ease;
  vector-effect: non-scaling-stroke;
}

.bento-snake-dot {
  filter: drop-shadow(0 0 6px var(--bento-accent-glow));
}

.bento-tagline {
  text-align: center;
  font-size: 11px;
  font-weight: 500;
  color: var(--bento-accent);
  animation: bento-fade-in 0.6s ease both;
}

@keyframes bento-travel-h {
  from { left: 0; }
  to { left: calc(100% - 8px); }
}

@keyframes bento-scale-in {
  from { opacity: 0; transform: scale(0); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes bento-fade-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
