<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const shown = ref(0);

const badges = [
  { icon: 'i-lucide-map-pin', label: 'Hosted in the EU' },
  { icon: 'i-lucide-shield-check', label: 'GDPR compliant' },
  { icon: 'i-lucide-building-2', label: 'Swiss owned' },
];

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { shown.value = 1; }, 200);
  badges.forEach((_, i) => setTimeout(() => { shown.value = i + 2; }, 600 + i * 180));
});
</script>

<template>
  <div
    ref="root"
    class="bento-swiss"
  >
    <div
      class="bento-swiss-flag"
      :class="{ 'bento-swiss-flag-in': shown >= 1 }"
    >
      <svg
        viewBox="0 0 32 32"
        class="size-full"
      >
        <rect
          width="32"
          height="32"
          rx="7"
          fill="var(--bento-accent)"
        />
        <path
          d="M13 6 h6 v7 h7 v6 h-7 v7 h-6 v-7 h-7 v-6 h7 z"
          fill="white"
        />
      </svg>
    </div>

    <div class="bento-swiss-badges">
      <span
        v-for="(b, i) in badges"
        :key="b.label"
        class="bento-swiss-badge"
        :class="{ 'bento-swiss-badge-in': shown >= i + 2 }"
      >
        <UIcon
          :name="b.icon"
          class="size-3.5 text-(--bento-accent)"
        />
        {{ b.label }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.bento-swiss {
  min-height: 130px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 16px 22px;
  padding: 16px 24px 28px;
}

.bento-swiss-flag {
  width: 56px;
  height: 56px;
  flex-shrink: 0;
  opacity: 0;
  transform: scale(0.6) rotate(-8deg);
  transition: all 0.6s cubic-bezier(0.16, 1, 0.3, 1);
  filter: drop-shadow(0 4px 12px var(--bento-accent-glow));
}

.bento-swiss-flag-in {
  opacity: 1;
  transform: scale(1) rotate(0);
}

.bento-swiss-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
}

.bento-swiss-badge {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 8px 14px;
  border-radius: 20px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  color: var(--ui-text);
  font-size: 13px;
  font-weight: 500;
  opacity: 0;
  transform: translateY(6px);
  transition: opacity 0.4s ease, transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.bento-swiss-badge-in {
  opacity: 1;
  transform: translateY(0);
}
</style>
