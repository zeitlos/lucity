<script setup lang="ts">
const rails = [
  { name: 'Names, never values', text: 'The tools return which variables exist and what they point at. They never return a value or a database password.' },
  { name: 'Logs come back masked', text: 'Values Lucity knows, of eight characters or more, are replaced before the agent reads a line.' },
  { name: 'It asks, and it cannot delete', text: 'It asks before it creates anything paid. Removing a project, service or database takes a person.' },
];

const variables = [
  { name: 'DATABASE_URL', state: 'linked to', target: 'main' },
  { name: 'STRIPE_SECRET', state: 'set' },
  { name: 'SESSION_SECRET', state: 'generated' },
];
</script>

<template>
  <section class="mx-auto flex max-w-content flex-wrap items-center gap-x-18 gap-y-12 px-6" aria-labelledby="agents-title">
    <div class="min-w-0 flex-[1_1_460px]">
      <h2 id="agents-title" class="font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">Agents work inside the same rails.</h2>
      <p class="mt-6 max-w-[560px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">
        Let a coding agent deploy, debug and read logs for you. The Lucity tools it gets are built so that it can do the work without holding your secrets.
      </p>
      <div class="mt-8 border-b border-stone-300 dark:border-stone-800">
        <div v-for="rail in rails" :key="rail.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
          <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ rail.name }}</b>
          <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ rail.text }}</span>
        </div>
      </div>
      <p class="mt-5 max-w-[560px] text-[15px] text-stone-600 dark:text-stone-400">
        This covers what the Lucity tools return. An agent that reads a .env file on your laptop sees what is in it.
        <NuxtLink to="/docs/claude-plugin" class="font-medium text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">How the plugin handles secrets</NuxtLink>
      </p>
    </div>

    <LandingStage tint="joy" class="min-w-0 flex-[1_1_520px]">
      <div class="scene">
        <div class="scene-row" aria-hidden="true">
          <img class="day" src="/img/landing/robot.svg" alt="" width="300" height="280">
          <img class="night" src="/img/landing/robot-night.svg" alt="" width="300" height="280">
          <div class="bubble"><p>STRIPE_SECRET is set. I can check that it is there. I cannot read it, and I do not need to.</p></div>
        </div>
        <SecurityTerminal
          class="scene-term"
          role="img"
          aria-label="An agent lists the variables of the web service in production. It sees that DATABASE_URL points at the main database, that STRIPE_SECRET is set and that SESSION_SECRET was generated. All three values are hidden."
        >
          <div><span class="c-acc">agent</span> list variables for web · production</div>
          <div class="variables">
            <template v-for="variable in variables" :key="variable.name">
              <span>{{ variable.name }}</span>
              <span>{{ variable.state }}<template v-if="variable.target">&nbsp;<span class="c-ok">{{ variable.target }}</span></template></span>
              <span class="c-dim">value hidden</span>
            </template>
          </div>
        </SecurityTerminal>
      </div>
    </LandingStage>
  </section>
</template>

<style scoped>
.scene {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 20px;
  min-height: 400px;
  padding: 28px 36px 0;
}

.scene-row {
  display: flex;
  align-items: flex-end;
  gap: 2px;
}

.scene-row img.night,
.dark .scene-row img.day {
  display: none;
}

.dark .scene-row img.night {
  display: block;
}

.scene-row img {
  flex: none;
  display: block;
  width: 180px;
  height: auto;
}

.bubble {
  position: relative;
  box-sizing: border-box;
  max-width: 360px;
  margin-bottom: 44px;
  padding: 18px 20px 16px;
  border-radius: 18px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.bubble::after {
  content: "";
  position: absolute;
  left: -7px;
  bottom: 30px;
  width: 15px;
  height: 15px;
  background: oklch(0.99 0.006 80);
  box-shadow: -1px 1px 0 oklch(0.25 0.04 55 / .10);
  transform: rotate(45deg);
}

.bubble p {
  font-size: 17px;
  line-height: 1.45;
  color: oklch(0.25 0.03 55);
}

.scene-term {
  border-radius: 14px 14px 0 0;
}

.scene-term :deep(.terminal-body) {
  padding-bottom: 28px;
}

.variables {
  display: grid;
  grid-template-columns: repeat(3, max-content);
  gap: 0 3ch;
  margin-top: 1.72em;
  white-space: nowrap;
}

@media (max-width: 520px) {
  .scene {
    min-height: 0;
    padding: 20px 16px 0;
  }

  .scene-row img {
    width: 84px;
  }

  .bubble {
    margin-bottom: 20px;
    padding: 14px 16px 13px;
  }

  .bubble p {
    font-size: 16px;
  }

  .variables {
    grid-template-columns: max-content minmax(0, 1fr);
    gap: 0 2ch;
    white-space: normal;
  }

  .variables .c-dim {
    grid-column: 2;
    margin-bottom: .5em;
  }
}
</style>
