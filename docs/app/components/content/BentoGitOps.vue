<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const commitCount = ref(0);

const commits = [
  { msg: 'deploy(dev): api a1b2c3d', icon: 'i-lucide-arrow-up-circle' },
  { msg: 'promote(staging): api a1b2c3d', icon: 'i-lucide-arrow-right-circle' },
  { msg: 'config(prod): replicas \u2192 3', icon: 'i-lucide-settings' },
];

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { commitCount.value = 1; }, 300);
  setTimeout(() => { commitCount.value = 2; }, 800);
  setTimeout(() => { commitCount.value = 3; }, 1300);
});
</script>

<template>
  <div
    ref="root"
    class="bento-gitops"
  >
    <!-- River wallpaper background (full opacity, covers entire card) -->
    <div class="bento-river-bg" />
    <!-- Vignette overlay for readability -->
    <div class="bento-vignette" />

    <!-- Git log -->
    <div class="bento-log">
      <div
        v-for="(commit, i) in commits"
        :key="i"
        class="bento-commit"
        :class="{ 'bento-commit-visible': commitCount > i }"
        :style="{ animationDelay: `${i * 100}ms` }"
      >
        <div class="bento-commit-dot" />
        <div
          v-if="i < commits.length - 1"
          class="bento-commit-line"
        />
        <div class="bento-commit-msg">
          <UIcon
            :name="commit.icon"
            class="size-3.5 shrink-0 text-(--ui-text-muted)"
          />
          <span class="font-mono text-[11px]">{{ commit.msg }}</span>
        </div>
      </div>

      <!-- Synced badge -->
      <div
        v-if="commitCount >= 3"
        class="bento-synced"
      >
        <UIcon
          name="i-lucide-check-circle"
          class="size-3.5"
        />
        Synced
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Fill the entire card area (visual is position:absolute in the grid) */
.bento-gitops {
  height: 100%;
  position: relative;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 48px 16px 0;
  overflow: hidden;
}

/* River wallpaper — full opacity, no tint, covers entire visual area */
.bento-river-bg {
  position: absolute;
  inset: 0;
  background: url('/img/branching_river.webp') center / cover no-repeat;
  pointer-events: none;
  z-index: 0;
}

/* Gradient overlay — transparent at top (river visible),
   fades to --bento-card-bg at bottom (text readability).
   Uses color-mix for broad browser support. */
.bento-vignette {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    to bottom,
    color-mix(in oklch, var(--bento-card-bg) 15%, transparent) 0%,
    color-mix(in oklch, var(--bento-card-bg) 30%, transparent) 40%,
    color-mix(in oklch, var(--bento-card-bg) 75%, transparent) 70%,
    var(--bento-card-bg) 100%
  );
  pointer-events: none;
  z-index: 1;
}

.bento-log {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: 0;
  width: 100%;
  max-width: 300px;
}

.bento-commit {
  display: flex;
  align-items: center;
  gap: 10px;
  position: relative;
  padding: 4px 0;
  opacity: 0;
}

.bento-commit-visible {
  animation: bento-commit-in 0.4s ease both;
}

.bento-commit-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--bento-accent);
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.bento-commit-line {
  position: absolute;
  left: 3.5px;
  top: 16px;
  width: 2px;
  height: calc(100% + 4px);
  background: var(--bento-accent);
  opacity: 0.5;
}

.bento-commit-msg {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border-radius: 6px;
  background: var(--ui-bg-elevated);
  border: 1px solid var(--ui-border);
  color: var(--ui-text);
}

.bento-synced {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 18px;
  margin-top: 6px;
  padding: 3px 10px;
  border-radius: 10px;
  background: var(--ui-bg-elevated);
  border: 1px solid var(--ui-border);
  color: var(--bento-accent);
  font-size: 11px;
  font-weight: 500;
  width: fit-content;
  animation: bento-fade-in 0.4s ease both;
  animation-delay: 0.3s;
}

@keyframes bento-commit-in {
  from { opacity: 0; transform: translateX(-6px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes bento-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
