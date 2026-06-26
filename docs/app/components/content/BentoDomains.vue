<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const step = ref(0);

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { step.value = 1; }, 300);
  setTimeout(() => { step.value = 2; }, 1100);
  setTimeout(() => { step.value = 3; }, 1700);
});
</script>

<template>
  <div
    ref="root"
    class="bento-domains"
  >
    <div class="flex w-full max-w-[280px] flex-col gap-3">
      <!-- Built-in platform domain — live instantly -->
      <div
        class="bento-domain"
        :class="{ 'bento-domain-active': step >= 1 }"
      >
        <UIcon
          name="i-lucide-globe"
          class="size-4 shrink-0 text-(--bento-accent)"
        />
        <span class="bento-domain-url">brave-otter.lucity.app</span>
        <span
          v-if="step >= 1"
          class="bento-domain-badge"
        >
          <UIcon
            name="i-lucide-check"
            class="size-2.5"
          />
          live
        </span>
      </div>

      <!-- Bring your own domain -->
      <div
        class="bento-domain bento-domain-custom"
        :class="{ 'bento-domain-active': step >= 2 }"
      >
        <UIcon
          name="i-lucide-lock"
          class="size-4 shrink-0 text-(--bento-accent)"
        />
        <span class="bento-domain-url">app.yourcompany.com</span>
        <span
          v-if="step >= 3"
          class="bento-domain-badge"
        >
          <UIcon
            name="i-lucide-shield-check"
            class="size-2.5"
          />
          TLS
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bento-domains {
  height: 210px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 24px;
}

.bento-domain {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 14px;
  border-radius: 12px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  color: var(--ui-text);
  opacity: 0;
  transform: translateY(6px);
  transition: opacity 0.4s ease, transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.bento-domain-custom {
  border-style: dashed;
}

.bento-domain-active {
  opacity: 1;
  transform: translateY(0);
}

.bento-domain-url {
  font-family: var(--font-mono);
  font-size: 12px;
  white-space: nowrap;
}

.bento-domain-badge {
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
  white-space: nowrap;
  animation: bento-fade-in 0.3s ease both;
}

@keyframes bento-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
