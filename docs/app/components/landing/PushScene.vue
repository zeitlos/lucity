<script setup lang="ts">
const github = 'M12 .3a12 12 0 0 0-3.8 23.4c.6.1.8-.3.8-.6v-2c-3.3.7-4-1.6-4-1.6-.6-1.4-1.4-1.8-1.4-1.8-1-.7.1-.7.1-.7 1.2.1 1.8 1.2 1.8 1.2 1 1.8 2.8 1.3 3.5 1 0-.8.4-1.3.7-1.6-2.7-.3-5.5-1.3-5.5-6 0-1.2.5-2.3 1.3-3.1-.2-.4-.6-1.6 0-3.2 0 0 1-.3 3.4 1.2a11.5 11.5 0 0 1 6 0c2.3-1.5 3.3-1.2 3.3-1.2.6 1.6.2 2.8.1 3.2.8.8 1.3 1.9 1.3 3.2 0 4.6-2.8 5.6-5.5 5.9.5.4.9 1 .9 2.2v3.3c0 .3.1.7.8.6A12 12 0 0 0 12 .3';

const steps = [
  { name: 'Build', time: '1m 42s', states: ['b-run', 'b-done'] },
  { name: 'Secrets', time: '17s', states: ['s-run', 's-done'] },
  { name: 'Deploy', time: '38s', states: ['d-wait', 'd-run', 'd-done'] },
  { name: 'Rollout', time: '', states: ['r-wait', 'r-run', 'r-done'] },
];

const history = [
  { title: 'Attach images to a request', when: '2d ago' },
  { title: 'Filter requests by status', when: '5d ago' },
  { title: 'Weekly digest email', when: '6d ago' },
];
</script>

<template>
  <div class="push" role="img" aria-label="A git push in a terminal starts a deployment in the dashboard. The build and the secret scan run, then the deploy and the rollout, and the service goes from building back to online.">
    <div class="push-term terminal" aria-hidden="true">
      <div class="terminal-bar"><i /><i /><i /><span>~/vouch</span></div>
      <pre><span class="c-dim">~/vouch</span> <span class="c-acc">main</span> $ <span class="typed">git push</span>
<span class="pa-out">To github.com:zeitlos/vouch.git
   7b32d9e..c41f0aa  main -&gt; main</span></pre>
    </div>

    <div class="push-note">
      <h3 class="font-display text-2xl/tight text-stone-900">Push. That was the deploy.</h3>
      <p class="mt-2.5 text-base text-[oklch(0.36_0.04_300)]">
        Lucity builds it, scans it for leaked keys and rolls it out with no downtime. Every deployment can be inspected, redeployed or rolled back.
      </p>
    </div>

    <div class="push-ui" aria-hidden="true">
      <div>
        <div class="pu-node">
          <b><svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" class="shrink-0"><path :d="github" /></svg> vouch</b>
          <small>vouch.example.com</small>
          <small style="margin-top: 3px">zeitlos/vouch</small>
          <div class="stack pu-node-foot">
            <span class="pill pa-n-on"><i class="dot-on" />Online</span>
            <span class="pill pa-n-build"><i class="dot-build" />Building</span>
            <span class="pill pa-n-deploy"><i class="dot-deploy" />Deploying</span>
          </div>
        </div>
      </div>

      <div class="pu-panel">
        <div class="pu-head"><svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" class="shrink-0"><path :d="github" /></svg> vouch</div>
        <div class="pu-tabs"><span class="on">Deployments</span><span>Logs</span><span>Metrics</span><span>Variables</span><span>Security</span><span>Settings</span></div>
        <div class="pu-line"><span>Tracks the default branch. Every push deploys.</span><span class="pu-btn">Deploy</span></div>

        <div class="pu-card pa-card">
          <div class="pu-rel">
            <span class="stack shrink-0">
              <span class="badge b-active pa-bg-active">Active</span>
              <span class="badge b-build pa-bg-build">Building</span>
              <span class="badge b-deploy pa-bg-deploy">Deploying</span>
            </span>
            <div class="min-w-0"><b>Sort the board by votes</b><small>just now · via GitHub · by Christian Blättler</small></div>
          </div>
          <div v-for="step in steps" :key="step.name" class="pu-step">
            <span>{{ step.name }}</span>
            <span class="stack">
              <template v-for="state in step.states" :key="state">
                <span v-if="state.endsWith('wait')" class="pu-wait" :class="`pa-${state}`">Waiting</span>
                <span v-else-if="state.endsWith('run')" class="pu-run" :class="`pa-${state}`"><i />Running</span>
                <span v-else class="pu-ok" :class="`pa-${state}`">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg>Succeeded<span v-if="step.time" class="pu-time"> · {{ step.time }}</span>
                </span>
              </template>
            </span>
            <span class="pu-logs">Logs</span>
          </div>
        </div>

        <div class="pu-hist">History</div>
        <div v-for="item in history" :key="item.title" class="pu-card">
          <div class="pu-rel">
            <span class="badge b-old shrink-0">Superseded</span>
            <div class="min-w-0"><b>{{ item.title }}</b><small>{{ item.when }} · via GitHub · by Christian Blättler</small></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.push {
  position: relative;
  margin-top: 28px;
  height: 540px;
}

.push-term {
  position: absolute;
  left: 0;
  top: 0;
  width: 42%;
  z-index: 2;
}

.push-note {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 26%;
}

.terminal {
  border-radius: 14px;
  overflow: hidden;
  background: oklch(0.20 0.03 55);
  box-shadow: 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.terminal-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 13px 18px;
  border-bottom: 1px solid oklch(1 0 0 / .08);
  font-size: 13px;
  color: oklch(0.74 0.03 70);
}

.terminal-bar i {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: oklch(0.40 0.03 55);
}

.terminal-bar span {
  margin-left: 8px;
}

.terminal pre {
  margin: 0;
  padding: 20px 24px 24px;
  font-family: var(--font-mono);
  font-size: 13.5px;
  line-height: 1.72;
  color: oklch(0.90 0.03 80);
  white-space: pre;
  overflow-x: auto;
}

.c-dim {
  color: oklch(0.70 0.03 70);
}

.c-acc {
  color: oklch(0.80 0.11 300);
}

.push-ui {
  position: absolute;
  left: 30%;
  right: -60px;
  top: 64px;
  bottom: -40px;
  display: grid;
  grid-template-columns: 30% minmax(0, 1fr);
  border-radius: 18px 0 0 0;
  overflow: hidden;
  background-color: oklch(0.955 0.018 80);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .16) 1px, transparent 0);
  background-size: 20px 20px;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
  font-family: 'Sora', 'Mona Sans', system-ui, sans-serif;
  color: #301c0e;
}

.pu-node {
  position: absolute;
  left: 28px;
  top: 238px;
  width: min(210px, calc(30% - 44px));
  box-sizing: border-box;
  padding: 16px 18px;
  border-radius: 16px;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .14), 0 10px 24px -14px oklch(0.25 0.04 55 / .4);
}

.pu-node b {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
}

.pu-node small {
  display: block;
  margin-top: 8px;
  font-size: 12px;
  color: oklch(0.42 0.03 55);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pu-node-foot {
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

.dot-on {
  background: oklch(0.70 0.19 150);
}

.dot-build {
  background: oklch(0.75 0.15 75);
}

.dot-deploy {
  background: oklch(0.62 0.18 300);
}

.stack {
  display: grid;
  justify-items: start;
  align-items: center;
}

.stack > * {
  grid-area: 1 / 1;
}

.pu-panel {
  margin: 14px 0 0;
  padding: 18px 84px 0 22px;
  border-radius: 16px 0 0 0;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12);
  font-size: 14px;
}

.pu-head {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 17px;
  font-weight: 600;
}

.pu-tabs {
  margin-top: 16px;
  display: flex;
  gap: 22px;
  border-bottom: 1px solid oklch(0.88 0.02 75);
  font-size: 14px;
  color: oklch(0.42 0.03 55);
  white-space: nowrap;
}

.pu-tabs span {
  padding-bottom: 10px;
}

.pu-tabs .on {
  color: #301c0e;
  font-weight: 600;
  box-shadow: 0 2px 0 oklch(0.70 0.17 160);
}

.pu-line {
  margin-top: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  font-size: 12.5px;
  color: oklch(0.42 0.03 55);
}

.pu-btn {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: oklch(0.75 0.18 160);
  color: oklch(0.22 0.05 160);
  font-size: 13px;
  font-weight: 600;
}

.pu-card {
  margin-top: 12px;
  border-radius: 12px;
  background: oklch(0.985 0.008 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .13);
}

.pu-rel {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 13px 16px;
}

.pu-rel b {
  display: block;
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pu-rel small {
  display: block;
  margin-top: 2px;
  font-size: 12px;
  color: oklch(0.42 0.03 55);
}

.badge {
  display: inline-block;
  padding: 3px 9px;
  border-radius: 5px;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: .04em;
  text-transform: uppercase;
  white-space: nowrap;
}

.b-active {
  background: oklch(0.92 0.07 160);
  color: oklch(0.40 0.11 160);
}

.b-build {
  background: oklch(0.93 0.08 85);
  color: oklch(0.42 0.10 70);
}

.b-deploy {
  background: oklch(0.93 0.045 300);
  color: oklch(0.42 0.15 300);
}

.b-old {
  background: oklch(0.91 0.015 75);
  color: oklch(0.42 0.03 55);
}

.pu-step {
  display: grid;
  grid-template-columns: 104px minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  padding: 9px 16px;
  border-top: 1px solid oklch(0.91 0.015 78);
  font-size: 13.5px;
}

.pu-step > span:first-child {
  font-weight: 600;
}

.pu-ok {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: oklch(0.48 0.14 160);
  font-weight: 500;
  white-space: nowrap;
}

.pu-time {
  color: oklch(0.42 0.03 55);
  font-weight: 400;
}

.pu-run {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: oklch(0.42 0.10 70);
  font-weight: 500;
}

.pu-run i {
  width: 11px;
  height: 11px;
  box-sizing: border-box;
  border-radius: 50%;
  border: 2px solid oklch(0.80 0.10 80);
  border-top-color: oklch(0.50 0.12 70);
  animation: pa-spin .9s linear infinite;
}

.pu-wait {
  color: oklch(0.50 0.03 60);
}

.pu-logs {
  padding: 4px 10px;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .16);
  font-size: 12px;
  font-weight: 500;
}

.pu-hist {
  margin-top: 18px;
  font-size: 13px;
  font-weight: 600;
  color: oklch(0.42 0.03 55);
}

.typed {
  display: inline-block;
  overflow: hidden;
  white-space: pre;
  vertical-align: bottom;
  width: 8ch;
  animation: pa-typed 16s steps(8) infinite;
}

.pa-out { animation: pa-out 16s linear infinite; }
.pa-card { animation: pa-card 16s ease-out infinite; }
.pa-b-run { animation: pa-b-run 16s linear infinite; opacity: 0; }
.pa-b-done { animation: pa-b-done 16s linear infinite; }
.pa-s-run { animation: pa-s-run 16s linear infinite; opacity: 0; }
.pa-s-done { animation: pa-s-done 16s linear infinite; }
.pa-d-wait { animation: pa-d-wait 16s linear infinite; opacity: 0; }
.pa-d-run { animation: pa-d-run 16s linear infinite; opacity: 0; }
.pa-d-done { animation: pa-d-done 16s linear infinite; }
.pa-r-wait { animation: pa-r-wait 16s linear infinite; opacity: 0; }
.pa-r-run { animation: pa-r-run 16s linear infinite; opacity: 0; }
.pa-r-done { animation: pa-r-done 16s linear infinite; }
.pa-n-on { animation: pa-n-on 16s linear infinite; }
.pa-n-build { animation: pa-n-build 16s linear infinite; opacity: 0; }
.pa-n-deploy { animation: pa-n-deploy 16s linear infinite; opacity: 0; }
.pa-bg-build { animation: pa-b-run 16s linear infinite; opacity: 0; }
.pa-bg-deploy { animation: pa-n-deploy 16s linear infinite; opacity: 0; }
.pa-bg-active { animation: pa-r-done 16s linear infinite; }

@keyframes pa-typed { 0%, 3% { width: 0; } 9%, 100% { width: 8ch; } }
@keyframes pa-out { 0% { opacity: 0; } 10.5% { opacity: 0; } 11.25% { opacity: 1; } 100% { opacity: 1; } }
@keyframes pa-card { 0%, 12.5% { opacity: 0; transform: translateY(-10px); } 15%, 96% { opacity: 1; transform: none; } 100% { opacity: 0; transform: none; } }
@keyframes pa-b-run { 0% { opacity: 1; } 40.62% { opacity: 1; } 41.38% { opacity: 0; } 100% { opacity: 0; } }
@keyframes pa-b-done { 0% { opacity: 0; } 39.88% { opacity: 0; } 40.62% { opacity: 1; } 100% { opacity: 1; } }
@keyframes pa-s-run { 0% { opacity: 1; } 24.38% { opacity: 1; } 25.12% { opacity: 0; } 100% { opacity: 0; } }
@keyframes pa-s-done { 0% { opacity: 0; } 23.62% { opacity: 0; } 24.38% { opacity: 1; } 100% { opacity: 1; } }
@keyframes pa-d-wait { 0% { opacity: 1; } 40.62% { opacity: 1; } 41.38% { opacity: 0; } 100% { opacity: 0; } }
@keyframes pa-d-run { 0% { opacity: 0; } 39.88% { opacity: 0; } 40.62% { opacity: 1; } 53.12% { opacity: 1; } 53.87% { opacity: 0; } 100% { opacity: 0; } }
@keyframes pa-d-done { 0% { opacity: 0; } 52.38% { opacity: 0; } 53.12% { opacity: 1; } 100% { opacity: 1; } }
@keyframes pa-r-wait { 0% { opacity: 1; } 53.12% { opacity: 1; } 53.87% { opacity: 0; } 100% { opacity: 0; } }
@keyframes pa-r-run { 0% { opacity: 0; } 52.38% { opacity: 0; } 53.12% { opacity: 1; } 68.75% { opacity: 1; } 69.5% { opacity: 0; } 100% { opacity: 0; } }
@keyframes pa-r-done { 0% { opacity: 0; } 68% { opacity: 0; } 68.75% { opacity: 1; } 100% { opacity: 1; } }
@keyframes pa-n-on { 0% { opacity: 1; } 15% { opacity: 1; } 15.75% { opacity: 0; } 68% { opacity: 0; } 68.75% { opacity: 1; } 100% { opacity: 1; } }
@keyframes pa-n-build { 0% { opacity: 0; } 14.25% { opacity: 0; } 15% { opacity: 1; } 40.62% { opacity: 1; } 41.38% { opacity: 0; } 100% { opacity: 0; } }
@keyframes pa-n-deploy { 0% { opacity: 0; } 39.88% { opacity: 0; } 40.62% { opacity: 1; } 68.75% { opacity: 1; } 69.5% { opacity: 0; } 100% { opacity: 0; } }
@keyframes pa-spin { to { transform: rotate(360deg); } }

@media (max-width: 860px) {
  .push {
    height: auto;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .push-term,
  .push-note,
  .push-ui {
    position: static;
    width: auto;
  }

  .push-ui {
    margin: 0 -16px -40px;
    grid-template-columns: minmax(0, 1fr);
    border-radius: 18px 18px 0 0;
  }

  .pu-node {
    display: none;
  }

  .pu-panel {
    margin: 0;
    padding: 18px 18px 0;
    height: 420px;
    overflow: hidden;
  }
}

@media (max-width: 520px) {
  .pu-step {
    grid-template-columns: 64px minmax(0, 1fr) auto;
    padding: 9px 12px;
  }

  .pu-time {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .push * {
    animation: none !important;
  }
}
</style>
