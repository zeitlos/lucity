<script setup lang="ts">
import { ref } from 'vue';
import BentoDeploy from './BentoDeploy.vue';
import BentoEnvironments from './BentoEnvironments.vue';
import BentoBatteries from './BentoBatteries.vue';
import BentoEject from './BentoEject.vue';
import BentoGitOps from './BentoGitOps.vue';
import BentoOpenSource from './BentoOpenSource.vue';

const cards = [
  {
    id: 'deploy',
    component: BentoDeploy,
    span: 'bento-span-full',
    title: 'Push to deploy',
    description: 'Connect your GitHub repo. Push your code, watch it flow through the pipeline and land on <span class="bento-hl">Kubernetes</span>. Zero Dockerfiles required.',
  },
  {
    id: 'envs',
    component: BentoEnvironments,
    span: 'bento-span-half',
    title: 'Multi-environment',
    description: 'Dev, staging, production, and <span class="bento-hl">PR previews</span>. Clone environments in seconds. Promote images without rebuilding.',
  },
  {
    id: 'batteries',
    component: BentoBatteries,
    span: 'bento-span-half',
    title: 'Batteries included',
    description: '<span class="bento-hl">PostgreSQL</span> via CloudNativePG, <span class="bento-hl">Redis</span>, cron jobs, and HTTP routing via Gateway API. Everything your app needs.',
  },
  {
    id: 'eject',
    component: BentoEject,
    span: 'bento-span-full',
    title: 'Eject anytime',
    description: 'One command. Standard <span class="bento-hl">Helm charts</span>, <span class="bento-hl">ArgoCD configs</span>, environment values, and a README. No lock-in, no strings attached.',
  },
  {
    id: 'gitops',
    component: BentoGitOps,
    span: 'bento-span-half',
    title: 'GitOps native',
    description: 'How the big players do it, just cleverly automated. Every deploy is a <span class="bento-hl">Git commit</span>. ArgoCD syncs your workloads.',
  },
  {
    id: 'oss',
    component: BentoOpenSource,
    span: 'bento-span-full',
    title: 'Open source',
    description: '<span class="bento-hl">AGPL-3.0</span> licensed. Self-host on your own Kubernetes cluster. Built on ArgoCD, Helm, CloudNativePG, and friends.',
  },
];

/* Spotlight cursor-follow effect */
const spotlightCard = ref<string | null>(null);
const spotlightX = ref(0);
const spotlightY = ref(0);

function onMouseMove(e: MouseEvent, cardId: string) {
  const el = (e.currentTarget as HTMLElement);
  const rect = el.getBoundingClientRect();
  spotlightX.value = e.clientX - rect.left;
  spotlightY.value = e.clientY - rect.top;
  spotlightCard.value = cardId;
}

function onMouseLeave() {
  spotlightCard.value = null;
}
</script>

<template>
  <div class="bento-grid">
    <div
      v-for="card in cards"
      :key="card.id"
      :class="['bento-card', `bento-card-${card.id}`, card.span]"
      @mousemove="(e) => onMouseMove(e, card.id)"
      @mouseleave="onMouseLeave"
    >
      <!-- Spotlight glow -->
      <div
        v-if="spotlightCard === card.id"
        class="bento-spotlight"
        :style="{
          background: `radial-gradient(400px circle at ${spotlightX}px ${spotlightY}px, var(--bento-accent-glow), transparent 70%)`,
        }"
      />

      <!-- Visual area (overflow hidden happens on .bento-card) -->
      <div class="bento-visual">
        <component :is="card.component" />
      </div>

      <!-- Text content -->
      <div class="bento-text">
        <h3 class="bento-title">
          {{ card.title }}
        </h3>
        <!-- eslint-disable-next-line vue/no-v-html -->
        <p
          class="bento-desc"
          v-html="card.description"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.bento-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
  width: 100%;
}

@media (min-width: 1024px) {
  .bento-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.bento-span-full {
  grid-column: 1 / -1;
}

.bento-span-half {
  grid-column: 1 / -1;
}

@media (min-width: 1024px) {
  .bento-span-half {
    grid-column: span 1;
  }
}

.bento-card {
  position: relative;
  border-radius: 16px;
  border: 1px solid var(--ui-border);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* Spotlight overlay */
.bento-spotlight {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  opacity: 0.6;
  transition: opacity 0.2s ease;
}

.bento-visual {
  position: relative;
  z-index: 1;
  flex: 1;
  min-height: 0;
}

.bento-text {
  position: relative;
  z-index: 1;
  padding: 24px 28px 28px;
}

@media (min-width: 640px) {
  .bento-text {
    padding: 28px 36px 36px;
  }
}

.bento-title {
  font-family: var(--font-serif);
  font-size: 1.5rem;
  font-weight: normal;
  color: var(--ui-text);
  line-height: 1.2;
}

@media (min-width: 640px) {
  .bento-title {
    font-size: 1.875rem;
  }
}

.bento-desc {
  font-family: var(--font-sans);
  font-size: 1rem;
  color: var(--ui-text-muted);
  line-height: 1.6;
  margin-top: 12px;
  text-wrap: pretty;
}

@media (min-width: 640px) {
  .bento-desc {
    font-size: 1.125rem;
  }
}
</style>
