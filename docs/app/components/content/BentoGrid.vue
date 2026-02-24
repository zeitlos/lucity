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
    textFirst: true,
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
    textFirst: true,
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
      :class="[
        'bento-card-wrap',
        `bento-card-${card.id}`,
        card.span,
      ]"
      @mousemove="(e) => onMouseMove(e, card.id)"
      @mouseleave="onMouseLeave"
    >
      <!-- Gradient border glow — cursor-following accent edge.
           The 1px padding on the wrapper creates a "border" gap.
           This gradient overlays it with the accent color at the cursor. -->
      <div
        v-if="spotlightCard === card.id"
        class="bento-border-glow"
        :style="{
          background: `radial-gradient(400px circle at ${spotlightX}px ${spotlightY}px, var(--bento-accent), transparent 60%)`,
        }"
      />

      <!-- Inner card shell -->
      <div class="bento-card">
        <!-- Surface spotlight glow -->
        <div
          v-if="spotlightCard === card.id"
          class="bento-spotlight"
          :style="{
            background: `radial-gradient(400px circle at ${spotlightX}px ${spotlightY}px, var(--bento-accent-glow), transparent 70%)`,
          }"
        />

        <!-- Depth gradient — slight highlight at top, shadow at bottom -->
        <div class="bento-depth" />

        <!-- Text content (shows first when textFirst) -->
        <div
          v-if="card.textFirst"
          class="bento-text"
        >
          <h3 class="bento-title">
            {{ card.title }}
          </h3>
          <!-- eslint-disable-next-line vue/no-v-html -->
          <p
            class="bento-desc"
            v-html="card.description"
          />
        </div>

        <!-- Visual area -->
        <div class="bento-visual">
          <component :is="card.component" />
        </div>

        <!-- Text content (shows after visual when not textFirst) -->
        <div
          v-if="!card.textFirst"
          class="bento-text"
        >
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

/* Outer wrapper — 1px padding acts as the "border".
   Default fill is --ui-border (subtle gray line).
   On hover, bento-border-glow overlays an accent gradient. */
.bento-card-wrap {
  position: relative;
  border-radius: 17px;
  padding: 1px;
  background: var(--ui-border);
}

/* Gradient border glow — radial gradient at cursor position */
.bento-border-glow {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
  z-index: 0;
  animation: bento-glow-in 0.25s ease both;
}

@keyframes bento-glow-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Inner card — solid bg, fits snugly inside the 1px border gap */
.bento-card {
  position: relative;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--bento-card-bg, var(--ui-bg-elevated));
  z-index: 1;
}

/* Surface spotlight glow */
.bento-spotlight {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 2;
  opacity: 0.5;
}

/* Subtle depth gradient — works in both light and dark modes */
.bento-depth {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 2;
  background: linear-gradient(
    180deg,
    oklch(1 0 0 / 0.04) 0%,
    transparent 40%,
    oklch(0 0 0 / 0.03) 100%
  );
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
  max-width: 640px;
  text-wrap: pretty;
}

@media (min-width: 640px) {
  .bento-desc {
    font-size: 1.125rem;
  }
}
</style>
