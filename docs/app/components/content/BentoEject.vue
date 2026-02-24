<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const phase = ref<'idle' | 'door' | 'lines' | 'done'>('idle');
const artifactCount = ref(0);

const outputs = [
  { icon: 'i-simple-icons-helm', label: 'Helm chart', path: 'ejected/chart/' },
  { icon: 'i-simple-icons-argo', label: 'ArgoCD apps', path: 'ejected/argocd/' },
  { icon: 'i-lucide-settings', label: 'Env values', path: 'ejected/environments/' },
  { icon: 'i-lucide-file-text', label: 'README', path: 'ejected/README.md' },
];

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { phase.value = 'door'; }, 300);
  setTimeout(() => { phase.value = 'lines'; }, 900);
  setTimeout(() => { artifactCount.value = 1; }, 1100);
  setTimeout(() => { artifactCount.value = 2; }, 1350);
  setTimeout(() => { artifactCount.value = 3; }, 1600);
  setTimeout(() => { artifactCount.value = 4; }, 1850);
  setTimeout(() => { phase.value = 'done'; }, 2200);
});
</script>

<template>
  <div
    ref="root"
    class="bento-eject"
  >
    <div class="bento-eject-layout">
      <!-- Left: door + command -->
      <div class="bento-door-section">
        <div
          class="bento-door"
          :class="{ 'bento-door-visible': phase !== 'idle' }"
        >
          <UIcon
            name="i-lucide-door-open"
            class="size-8 sm:size-10"
          />
        </div>
        <code
          v-if="phase !== 'idle'"
          class="bento-cmd"
        >lucity eject</code>
      </div>

      <!-- Right: stacked artifacts with connector lines -->
      <div class="bento-artifacts">
        <div
          v-for="(item, i) in outputs"
          :key="item.label"
          class="bento-artifact-row"
        >
          <!-- Horizontal connector line with dot -->
          <div
            class="bento-connector"
            :class="{ 'bento-connector-active': artifactCount > i }"
          >
            <div class="bento-line" />
            <div class="bento-dot" />
          </div>

          <!-- Artifact card -->
          <div
            class="bento-output"
            :class="{ 'bento-output-visible': artifactCount > i }"
            :style="{ animationDelay: `${i * 60}ms` }"
          >
            <UIcon
              :name="item.icon"
              class="size-4 shrink-0 text-(--bento-accent)"
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
        </div>
      </div>
    </div>

    <!-- Tagline -->
    <div
      v-if="phase === 'done'"
      class="bento-tagline"
    >
      No lock-in. Standard tools.
    </div>
  </div>
</template>

<style scoped>
.bento-eject {
  min-height: 160px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px 20px 16px;
  gap: 14px;
}

.bento-eject-layout {
  display: flex;
  align-items: center;
  gap: 0;
  width: 100%;
  max-width: 460px;
}

/* Left section: door icon + command */
.bento-door-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  width: 100px;
}

@media (min-width: 640px) {
  .bento-door-section {
    width: 120px;
  }
}

.bento-door {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  border: 1.5px solid var(--bento-accent);
  background: var(--ui-bg-elevated);
  color: var(--bento-accent);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transform: scale(0.7);
  transition: all 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

@media (min-width: 640px) {
  .bento-door {
    width: 64px;
    height: 64px;
  }
}

.bento-door-visible {
  opacity: 1;
  transform: scale(1);
}

.bento-cmd {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--bento-accent);
  font-weight: 500;
  animation: bento-fade-in 0.4s ease both;
  white-space: nowrap;
}

/* Right section: stacked artifacts */
.bento-artifacts {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
  min-width: 0;
}

.bento-artifact-row {
  display: flex;
  align-items: center;
  gap: 0;
}

/* Connector line with dot */
.bento-connector {
  display: flex;
  align-items: center;
  width: 28px;
  flex-shrink: 0;
}

@media (min-width: 640px) {
  .bento-connector {
    width: 40px;
  }
}

.bento-line {
  flex: 1;
  height: 1.5px;
  background: var(--ui-border);
  transition: background 0.3s ease;
}

.bento-connector-active .bento-line {
  background: var(--bento-accent);
}

.bento-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--ui-border);
  flex-shrink: 0;
  transition: all 0.3s ease;
}

.bento-connector-active .bento-dot {
  background: var(--bento-accent);
  box-shadow: 0 0 6px var(--bento-accent-glow);
}

/* Artifact cards */
.bento-output {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: 8px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  white-space: nowrap;
  opacity: 0;
  transform: translateX(-6px);
  min-width: 0;
}

.bento-output-visible {
  animation: bento-artifact-in 0.35s cubic-bezier(0.16, 1, 0.3, 1) both;
}

.bento-tagline {
  text-align: center;
  font-size: 11px;
  font-weight: 500;
  color: var(--bento-accent);
  animation: bento-fade-in 0.6s ease both;
}

@keyframes bento-artifact-in {
  from { opacity: 0; transform: translateX(-6px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes bento-fade-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
