<script setup lang="ts">
import { ref } from 'vue';
import BentoDeploy from './BentoDeploy.vue';
import BentoDomains from './BentoDomains.vue';
import BentoBatteries from './BentoBatteries.vue';
import BentoEject from './BentoEject.vue';
import BentoScale from './BentoScale.vue';
import BentoSwiss from './BentoSwiss.vue';
import BentoPricing from './BentoPricing.vue';
import BentoOpenSource from './BentoOpenSource.vue';

const cards = [
  /* Row 1: 50 / 50 */
  {
    id: 'deploy',
    component: BentoDeploy,
    span: 'bento-span-3',
    corner: 'bento-corner-tl',
    title: 'Push to deploy',
    description: 'Connect your GitHub repo, push, and watch it build and go live. No Dockerfile or YAML required.',
  },
  {
    id: 'batteries',
    component: BentoBatteries,
    span: 'bento-span-3',
    corner: 'bento-corner-tr',
    textFirst: true,
    title: 'Batteries included',
    description: 'Provision replicated, highly-available <span class="bento-hl">PostgreSQL</span>, <span class="bento-hl">Redis</span> and <span class="bento-hl">S3-compatible Object Storage</span> in one click.',
  },
  /* Row 2: 66 / 33 */
  {
    id: 'eject',
    component: BentoEject,
    span: 'bento-span-4',
    title: 'Eject anytime',
    description: 'Deliberately designed to be lock-in free. Want to leave? Download the full <span class="bento-hl">Helm chart</span>, <span class="bento-hl">build config</span>, and <span class="bento-hl">values</span> for your app and run it anywhere.',
  },
  {
    id: 'scale',
    component: BentoScale,
    span: 'bento-span-2',
    title: 'Scale without thinking about it',
    description: 'Spin up more replicas or give them more power. Enable <span class="bento-hl">auto-scaling</span> for turbulent workloads.',
  },
  /* Row 3: 33 / 66 */
  {
    id: 'envs',
    component: BentoDomains,
    span: 'bento-span-2',
    title: 'Free public domain + custom domains',
    description: 'Every deploy gets a live URL on a built-in <span class="bento-hl">platform domain</span>, instantly. Bring your own <span class="bento-hl">custom domain</span> when you\'re ready to go live.',
  },
  {
    id: 'oss',
    component: BentoOpenSource,
    span: 'bento-span-4',
    textFirst: true,
    title: 'Open source',
    description: '<span class="bento-hl">AGPL-3.0</span> licensed. Self-host on your own Kubernetes cluster. Built on Helm, CloudNativePG, Valkey, VictoriaMetrics, and friends.',
  },
  /* Row 4: 50 / 50 */
  {
    id: 'swiss',
    component: BentoSwiss,
    span: 'bento-span-3',
    corner: 'bento-corner-bl',
    textFirst: true,
    title: 'Backed by a Swiss company',
    description: 'There’s a US law called the <span class="bento-hl">CLOUD Act</span>. This cloud isn’t subject to it. Your data stays under Swiss and EU jurisdiction, where US subpoenas don’t reach.',
  },
  {
    id: 'pricing',
    component: BentoPricing,
    span: 'bento-span-3',
    corner: 'bento-corner-br',
    textFirst: true,
    title: 'Pricing so simple, you’d think our sales team is lazy.',
    description: 'Just kidding, we don’t have a sales team… we only bill you for <span class="bento-hl">what you actually use</span>.',
    link: { label: 'Learn more', to: '/pricing' },
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
  <section class="px-6">
    <div class="mx-auto max-w-content">

      <h2 class="font-display text-5xl leading-tight text-neutral-800 md:text-6xl dark:text-neutral-100">
        Everything you need to ship
      </h2>

      <p class="mt-8 text-2xl leading-relaxed mb-22">
        All the building blocks for deploying and running your apps. Built on standard, open tools, so you can eject whenever you want.
      </p>


      <div class="bento-grid">
        <div
          v-for="card in cards"
          :key="card.id"
          :class="[
            'bento-card-wrap',
            `bento-card-${card.id}`,
            card.span,
            card.corner,
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
              <NuxtLink
                v-if="card.link"
                :to="card.link.to"
                class="bento-link"
              >
                {{ card.link.label }}
                <UIcon
                  name="i-lucide-arrow-right"
                  class="size-4"
                />
              </NuxtLink>
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
              <NuxtLink
                v-if="card.link"
                :to="card.link.to"
                class="bento-link"
              >
                {{ card.link.label }}
                <UIcon
                  name="i-lucide-arrow-right"
                  class="size-4"
                />
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
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
    grid-template-columns: repeat(6, 1fr);
  }
}

/* All spans are full-width on mobile */
.bento-span-2,
.bento-span-3,
.bento-span-4,
.bento-span-6 {
  grid-column: 1 / -1;
}

@media (min-width: 1024px) {
  .bento-span-2 { grid-column: span 2; }
  .bento-span-3 { grid-column: span 3; }
  .bento-span-4 { grid-column: span 4; }
  .bento-span-6 { grid-column: span 6; }
}

/* Outer wrapper — 1px padding acts as the "border".
   Default fill is --ui-border (subtle gray line).
   On hover, bento-border-glow overlays an accent gradient. */
.bento-card-wrap {
  position: relative;
  min-width: 0;
  border-radius: 17px;
  padding: 1px;
  background: var(--ui-border);
}

/* Large corner radius on grid edges (desktop only) */
@media (min-width: 1024px) {
  .bento-corner-tl { border-top-left-radius: 49px; }
  .bento-corner-tl .bento-card { border-top-left-radius: 48px; }

  .bento-corner-tr { border-top-right-radius: 49px; }
  .bento-corner-tr .bento-card { border-top-right-radius: 48px; }

  .bento-corner-bl { border-bottom-left-radius: 49px; }
  .bento-corner-bl .bento-card { border-bottom-left-radius: 48px; }

  .bento-corner-br { border-bottom-right-radius: 49px; }
  .bento-corner-br .bento-card { border-bottom-right-radius: 48px; }
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

/* Inner card — solid bg, fits snugly inside the 1px border gap.
   height: 100% ensures the card fills the wrapper so the 1px
   border (padding) doesn't show as a thick gap at the bottom. */
.bento-card {
  position: relative;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(
    180deg,
    var(--bento-card-bg, var(--ui-bg-elevated)) 0%,
    var(--ui-bg-elevated) 100%
  );
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

/* GitOps card — full-bleed background image.
   Visual is taken out of flow; text pushed to bottom via auto margin. */
.bento-card-gitops .bento-visual {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.bento-card-gitops .bento-text {
  position: relative;
  z-index: 3;
  margin-top: auto;
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
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: normal;
  color: var(--ui-text);
  line-height: 1.2;
}

@media (min-width: 640px) {
  .bento-title {
    font-size: 2.125rem;
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

.bento-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 16px;
  font-family: var(--font-sans);
  font-size: 1rem;
  font-weight: 600;
  color: var(--bento-accent);
  transition: gap 0.2s ease;
}

.bento-link:hover {
  gap: 10px;
}
</style>
