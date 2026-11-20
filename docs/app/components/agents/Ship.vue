<script setup lang="ts">
const features = [
  { name: 'Right next to your app', text: 'An agent reaches the services of its environment by name, when you let it. No tunnels, no public endpoints.' },
  { name: 'Every change is a version', text: 'Instructions, grants and triggers are versioned together, with who changed what.' },
  { name: 'One click to production', text: 'Promote the agent you tried in staging, exactly as it was.' },
];

const versions = [
  { version: 'v14', stage: 'staging', tone: 'ask', change: 'Also watch the worker queue', by: 'Noa · 2 hours ago', action: 'Promote', link: true },
  { version: 'v13', stage: 'production', tone: 'ok', change: 'Ask Ops before any rollback', by: 'Christian · Tuesday', action: 'live', link: false },
  { version: 'v12', stage: 'earlier', tone: 'off', change: 'Report to Shop on call', by: 'Noa · 2 November', action: 'Roll back', link: true },
];
</script>

<template>
  <section class="mx-auto flex max-w-content flex-wrap items-center gap-x-18 gap-y-14 px-6" aria-labelledby="ship">
    <div class="min-w-0 flex-[1_1_460px]">
      <h2 id="ship" class="font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">Agents ship the way your apps do.</h2>
      <p class="mt-6 max-w-[560px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">
        An agent lives in an environment, next to the services it looks after. Try a change in staging, promote it to production, roll it back when it misbehaves.
      </p>
      <div class="mt-10 border-b border-stone-300 dark:border-stone-800">
        <div v-for="feature in features" :key="feature.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
          <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ feature.name }}</b>
          <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ feature.text }}</span>
        </div>
      </div>
    </div>

    <div class="cv min-w-0 flex-[1.25_1_624px]">
      <div class="cvf" aria-hidden="true">
        <div class="cv-top">
          <svg width="20" height="20" viewBox="0 0 20 20" class="shrink-0"><rect width="20" height="20" rx="5" fill="#8843db" /></svg>
          <span>shop</span><em>/</em><span>production</span>
        </div>
        <div class="cv-node cv-agent">
          <b><AgentsIcon name="bot" :size="20" class="text-[#8843db]" />Night watch</b>
          <small><AgentsIcon name="bell" :size="13" /><span>wakes on alerts for web</span></small>
          <small class="second"><AgentsIcon name="terminal" :size="13" /><span>OpenCode · v13</span></small>
          <div class="cv-foot"><span class="pill"><i class="dot-asleep" />Asleep</span></div>
        </div>
        <div class="cv-link" />
        <div class="cv-node cv-web">
          <b><AgentsIcon name="github" :size="20" />web</b>
          <small><AgentsIcon name="globe" :size="13" /><span>shop.example</span></small>
          <small class="second"><AgentsIcon name="git-branch" :size="13" /><span>zeitlos/shop</span></small>
          <div class="cv-foot"><span class="pill"><i class="dot-on" />Online</span></div>
        </div>
        <div class="cv-node cv-db">
          <b><AgentsIcon name="database" :size="20" class="text-[oklch(0.50_0.13_250)]" />orders</b>
          <small><span>PostgreSQL</span></small>
          <small class="second"><span>2 instances</span></small>
          <div class="cv-foot"><span class="pill"><i class="dot-on" />Online</span></div>
        </div>
      </div>

      <div class="cv-ver">
        <p class="ver-title">Night watch · versions</p>
        <div v-for="row in versions" :key="row.version" class="ver">
          <span class="v">{{ row.version }}</span>
          <span class="e" :class="row.tone">{{ row.stage }}</span>
          <span class="m">{{ row.change }}<small>{{ row.by }}</small></span>
          <span class="a" :class="{ act: row.link }">{{ row.action }}</span>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.cv {
  position: relative;
  padding: 0 0 110px 64px;
}

.cvf {
  position: relative;
  height: 500px;
  border-radius: 18px;
  overflow: hidden;
  background-color: oklch(0.955 0.018 80);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .16) 1px, transparent 0);
  background-size: 20px 20px;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
  font-family: 'Sora', 'Mona Sans', system-ui, sans-serif;
  color: #301c0e;
}

.cv-top {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 54px;
  padding: 0 22px;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
  font-size: 14.5px;
  font-weight: 500;
}

.cv-top em {
  font-style: normal;
  color: oklch(0.62 0.03 60);
}

.cv-node {
  position: absolute;
  top: 104px;
  width: 228px;
  box-sizing: border-box;
  padding: 16px 18px;
  border-radius: 16px;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .14), 0 10px 24px -14px oklch(0.25 0.04 55 / .4);
}

.cv-agent {
  left: 36px;
  outline: 1.5px solid oklch(0.55 0.15 300 / .7);
  outline-offset: -1px;
}

.cv-web {
  left: 308px;
}

.cv-db {
  left: 556px;
}

.cv-link {
  position: absolute;
  left: 264px;
  top: 168px;
  width: 44px;
  border-top: 1.5px dashed oklch(0.55 0.15 300 / .8);
}

.cv-node b {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
}

.cv-node small {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-top: 8px;
  font-size: 12px;
  color: oklch(0.42 0.03 55);
  white-space: nowrap;
}

.cv-node small.second {
  margin-top: 4px;
}

.cv-foot {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid oklch(0.90 0.015 78);
}

.pill {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 3px 10px;
  border-radius: 999px;
  background: oklch(0.90 0.015 75);
  font-size: 12px;
  font-weight: 500;
  color: oklch(0.35 0.03 55);
}

.pill i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.dot-asleep {
  background: oklch(0.62 0.18 300);
}

.dot-on {
  background: oklch(0.70 0.19 150);
}

.cv-ver {
  position: absolute;
  left: 0;
  right: 12%;
  bottom: 0;
  z-index: 2;
  box-sizing: border-box;
  padding: 22px 28px 8px;
  border-radius: 14px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.ver-title {
  margin: 0 0 8px;
  font-size: 17px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.ver {
  display: grid;
  grid-template-columns: 44px 104px minmax(0, 1fr) auto;
  gap: 6px 18px;
  align-items: baseline;
  padding: 14px 0;
  border-top: 1px solid oklch(0.91 0.015 80);
  font-size: 16px;
}

.ver .v {
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 14px;
  color: oklch(0.47 0.04 55);
}

.ver .e {
  font-size: 15px;
  font-weight: 600;
}

.ver .ok {
  color: oklch(0.43 0.12 160);
}

.ver .ask {
  color: oklch(0.48 0.11 70);
}

.ver .off {
  color: oklch(0.47 0.04 55);
}

.ver .m {
  color: oklch(0.25 0.03 55);
}

.ver .m small {
  display: block;
  font-size: 14px;
  color: oklch(0.47 0.04 55);
}

.ver .a {
  font-size: 15px;
  color: oklch(0.47 0.04 55);
}

.ver .a.act {
  font-weight: 600;
  color: oklch(0.25 0.03 55);
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 4px;
}

@media (max-width: 640px) {
  .cv {
    padding: 0;
  }

  .cvf {
    height: 300px;
  }

  .cv-node {
    top: 82px;
  }

  .cv-agent {
    left: 16px;
  }

  .cv-link {
    left: 244px;
    top: 146px;
    width: 20px;
  }

  .cv-web {
    left: 264px;
  }

  .cv-db {
    display: none;
  }

  .cv-ver {
    position: relative;
    right: auto;
    margin: -44px 12px 0;
    padding: 18px 20px 6px;
  }

  .ver {
    grid-template-columns: 44px minmax(0, 1fr) auto;
    gap: 4px 14px;
  }

  .ver .e {
    grid-area: 1 / 2;
  }

  .ver .a {
    grid-area: 1 / 3;
  }

  .ver .m {
    grid-area: 2 / 2 / auto / -1;
  }
}
</style>
