<script setup lang="ts">
const captions = [
  'You push. The build starts, and a scan for leaked keys starts with it.',
  'The scan finds an AWS key in your history and asks AWS whether it still works. It does.',
  'Nothing is blocked. The release ships, and the finding waits in the Security tab.',
  'Two hours later the running app is checked again. A vulnerability announced today shows up on what you shipped before.',
];

interface StepState {
  track: string;
  kind: 'wait' | 'run' | 'ok' | 'key';
  text: string;
}

const steps: { name: string; highlight?: boolean; states: StepState[] }[] = [
  {
    name: 'Build',
    states: [
      { track: 'b1', kind: 'run', text: 'Building the image' },
      { track: 'b2', kind: 'ok', text: 'Ready in 1m 12s' },
    ],
  },
  {
    name: 'Secrets',
    highlight: true,
    states: [
      { track: 's1', kind: 'run', text: 'Reading 214 commits' },
      { track: 's2', kind: 'key', text: 'Found a key, asking AWS' },
      { track: 's3', kind: 'key', text: '1 key that still works' },
    ],
  },
  {
    name: 'Deploy',
    states: [
      { track: 'd1', kind: 'wait', text: 'Waiting for the build' },
      { track: 'd2', kind: 'run', text: 'Pointing at the new image' },
      { track: 'd3', kind: 'ok', text: 'Done' },
    ],
  },
  {
    name: 'Rollout',
    states: [
      { track: 'r1', kind: 'wait', text: 'Waiting' },
      { track: 'r2', kind: 'run', text: 'Starting 2 instances' },
      { track: 'r3', kind: 'ok', text: 'Live in 41s' },
    ],
  },
];

const tabs = ['Deployments', 'Variables', 'Metrics', 'Logs', 'Security'];
</script>

<template>
  <div
    class="hs"
    role="img"
    aria-label="One deploy and the two hours after it. You push, and the build and a scan for leaked keys start together. The scan finds an AWS access key in the history of the repository and confirms with AWS that it still works. Nothing is blocked: the release goes live and the finding waits in the Security tab of the service. Two hours later the running app is checked again, and a critical vulnerability in openssl that was announced in the meantime appears at the top of the list."
  >
    <div class="hs-sunset" aria-hidden="true" />
    <img class="hs-bridge" src="/img/landing/guard-rail-bridge.svg" alt="" width="640" height="540">

    <div class="hs-caps stack">
      <p v-for="(caption, index) in captions" :key="caption" :class="`a-c${index + 1}`">{{ caption }}</p>
    </div>

    <div class="hs-card a-card">
      <div class="hs-head">
        <div class="hs-title">
          <UIcon name="i-simple-icons-github" class="size-5 shrink-0" />
          <span class="flex-1">web</span>
          <span class="stack hs-pills">
            <span class="a-p1"><i class="dot-build" />Building</span>
            <span class="a-p2"><i class="dot-deploy" />Deploying</span>
            <span class="a-p3"><i class="dot-on" />Online</span>
          </span>
        </div>
        <div class="hs-ref">Deployment 142 · main · 9b1e4d0</div>
      </div>
      <div v-for="step in steps" :key="step.name" class="hs-step" :class="{ 'a-hl': step.highlight }">
        <b>{{ step.name }}</b>
        <span class="stack hs-states">
          <span v-for="state in step.states" :key="state.track" :class="[`a-${state.track}`, `is-${state.kind}`]">
            <i v-if="state.kind === 'run'" class="hs-spin" />
            <svg v-else-if="state.kind === 'ok'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M4 12.5 L10 18 L20 6" /></svg>
            <UIcon v-else-if="state.kind === 'key'" name="i-lucide-key-round" class="size-3.5 shrink-0" />
            {{ state.text }}
          </span>
        </span>
      </div>
    </div>

    <p class="hs-foot">Build and Secrets run side by side. The scan reports what it finds and never holds a release back.</p>

    <i class="hs-line a-l1" />
    <i class="hs-dot a-dot" />

    <div class="hs-bezel">
      <div class="hs-panel">
        <div class="hs-tabs">
          <span v-for="tab in tabs" :key="tab" :class="{ on: tab === 'Security' }">{{ tab }}</span>
        </div>

        <div class="hs-sec">
          <b class="hs-label">Leaked secrets</b>
          <div class="hs-slot">
            <span class="hs-empty a-empty">Nothing found so far.</span>
            <div class="hs-find a-find">
              <span class="hs-key"><UIcon name="i-lucide-key-round" class="size-4.5" /></span>
              <span class="min-w-0 flex-1">
                <b>AWS access key</b>
                <small>config/settings.py:42 · 3f2a1c9 · 14 months ago</small>
              </span>
              <span class="stack hs-badges">
                <span class="hs-badge is-asking a-k1">Asking AWS</span>
                <span class="hs-badge is-live a-k2">Still works</span>
              </span>
            </div>
          </div>
        </div>

        <div class="hs-sec hs-sec-vul">
          <div class="hs-vhead">
            <b class="hs-label">Vulnerabilities</b>
            <span class="stack hs-checked">
              <span class="a-v0">checked after the first build</span>
              <span class="a-v1">checked 12:06</span>
              <span class="is-again a-v2">checking again</span>
              <span class="is-now a-v3">checked 14:06</span>
            </span>
          </div>
          <div class="hs-counts a-cnt">
            <span class="stack">
              <span class="is-none a-n0"><i />0 critical</span>
              <span class="is-crit a-n1"><i />1 critical</span>
            </span>
            <span><i class="sev-high" />3 high</span>
            <span><i class="sev-med" />12 medium</span>
          </div>
          <div class="hs-rows">
            <div class="hs-row a-va"><b>express</b><span>Dependency · 4.18.2, fixed in 4.19.2</span><em class="is-high">High</em></div>
            <div class="hs-row hs-row-2 a-vb"><b>jsonwebtoken</b><span>Dependency · 8.5.1, fixed in 9.0.0</span><em class="is-high">High</em></div>
            <div class="hs-row a-vn"><b>openssl</b><span>Base image · 3.0.13, fixed in 3.0.15</span><em class="is-crit">Critical</em></div>
          </div>
        </div>

        <i class="hs-sweep a-sweep" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.hs {
  --bright: oklch(0.97 0.01 80);
  --body: oklch(0.88 0.02 80);
  --meta: oklch(0.72 0.03 70);
  --amber: oklch(0.82 0.14 80);
  --green: oklch(0.80 0.16 160);
  --vrow: 46px;
  position: relative;
  width: 100%;
  height: 630px;
  overflow: hidden;
  border-radius: 24px;
  background-color: oklch(0.20 0.03 55);
  background-image: radial-gradient(circle at 1px 1px, oklch(1 0 0 / .07) 1px, transparent 0);
  background-size: 22px 22px;
  box-shadow: 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
  font-size: 14px;
  line-height: 20px;
  text-align: left;
  color: var(--body);
}

.dark .hs {
  box-shadow: inset 0 0 0 1px oklch(1 0 0 / .10);
}

.stack {
  display: grid;
  align-items: center;
}

.stack > * {
  grid-area: 1 / 1;
}

.hs-sunset {
  position: absolute;
  right: -60px;
  top: 21px;
  width: 700px;
  height: 560px;
  filter: blur(30px);
  background:
    radial-gradient(ellipse 30% 30% at 80% 48%, oklch(0.82 0.14 58 / .7), oklch(0.82 0.14 58 / 0) 100%),
    radial-gradient(ellipse 34% 30% at 52% 50%, oklch(0.70 0.21 0 / .55), oklch(0.70 0.21 0 / 0) 100%),
    radial-gradient(ellipse 25% 30% at 30% 53%, oklch(0.46 0.23 282 / .8), oklch(0.46 0.23 282 / 0) 100%);
}

.hs-bridge {
  position: absolute;
  right: 0;
  top: -19px;
  width: 520px;
  height: auto;
  pointer-events: none;
}

.hs-caps {
  position: absolute;
  left: 48px;
  top: 44px;
  width: 520px;
  align-items: start;
  font-family: 'Redaction 20', Georgia, serif;
  font-size: 25px;
  line-height: 34px;
  color: var(--bright);
}

.hs-card {
  position: absolute;
  left: 48px;
  top: 176px;
  width: 500px;
  overflow: hidden;
  border-radius: 16px;
  background: oklch(0.17 0.025 55 / .92);
  box-shadow: 0 0 0 1px rgb(255 255 255 / .16), 0 24px 44px -20px rgb(0 0 0 / .7);
}

.hs-head {
  box-sizing: border-box;
  height: 86px;
  padding: 20px 24px 0;
}

.hs-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 17px;
  font-weight: 600;
  color: var(--bright);
}

.hs-pills {
  justify-items: end;
}

.hs-pills > span {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 3px 10px;
  border-radius: 999px;
  background: rgb(255 255 255 / .12);
  font-size: 12px;
  line-height: 20px;
  font-weight: 500;
  color: oklch(0.92 0.02 80);
}

.hs-pills i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.dot-build {
  background: oklch(0.80 0.14 80);
}

.dot-deploy {
  background: oklch(0.72 0.17 300);
}

.dot-on {
  background: var(--green);
}

.hs-ref {
  margin-top: 14px;
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 12.5px;
  color: var(--meta);
  white-space: nowrap;
}

.hs-step {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 62px;
  padding: 0 24px;
  border-top: 1px solid rgb(255 255 255 / .10);
  font-size: 15px;
}

.hs-step b {
  font-weight: 600;
  color: var(--bright);
}

.hs-states {
  justify-items: end;
  font-size: 14px;
}

.hs-states > span {
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
  color: var(--meta);
}

.hs-states > .is-ok {
  color: var(--green);
}

.hs-states > .is-key {
  color: var(--amber);
}

.hs-spin {
  flex: none;
  box-sizing: border-box;
  width: 11px;
  height: 11px;
  border: 2px solid rgb(255 255 255 / .25);
  border-top-color: oklch(0.90 0.03 80);
  border-radius: 50%;
  animation: hs-spin .9s linear infinite;
}

.hs-foot {
  position: absolute;
  left: 48px;
  top: 529px;
  width: 500px;
  color: var(--meta);
}

.hs-line {
  position: absolute;
  left: 548px;
  top: 354px;
  width: 92px;
  height: 2px;
  background: white;
}

.hs-dot {
  position: absolute;
  left: 0;
  top: 0;
  width: 12px;
  height: 12px;
  border-radius: 6px;
  background: var(--amber);
  box-shadow: 0 0 14px 3px var(--amber);
}

.hs-bezel {
  position: absolute;
  left: 632px;
  right: 0;
  top: 209px;
  box-sizing: border-box;
  height: 480px;
  padding: 8px 0 0 8px;
  border: 1px solid oklch(1 0 0 / .26);
  border-right: 0;
  border-bottom: 0;
  border-radius: 24px 0 0;
  background: oklch(1 0 0 / .08);
}

.hs-panel {
  position: relative;
  height: 100%;
  overflow: hidden;
  border-radius: 16px 0 0;
  background: oklch(0.17 0.02 55);
  box-shadow: 0 0 0 1px oklch(1 0 0 / .14);
  font-family: 'Sora', 'Mona Sans', system-ui, sans-serif;
}

.hs-tabs {
  box-sizing: border-box;
  display: flex;
  align-items: stretch;
  gap: 26px;
  height: 52px;
  padding: 0 28px;
  border-bottom: 1px solid rgb(255 255 255 / .10);
  font-size: 14px;
  line-height: 52px;
  color: var(--meta);
  white-space: nowrap;
}

.hs-tabs .on {
  font-weight: 600;
  color: var(--bright);
  box-shadow: 0 2px 0 var(--green);
}

.hs-sec {
  padding: 18px 36px 0 28px;
}

.hs-sec-vul {
  margin: 18px 36px 0 28px;
  padding: 17px 0 0;
  border-top: 1px solid rgb(255 255 255 / .10);
}

.hs-label {
  display: block;
  font-size: 15px;
  font-weight: 600;
  color: var(--bright);
}

.hs-slot {
  position: relative;
  margin-top: 10px;
  min-height: 76px;
}

.hs-empty {
  position: absolute;
  left: 0;
  top: 22px;
  color: var(--meta);
}

.hs-find {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 14px;
  min-height: 76px;
  padding: 0 16px;
  border-radius: 12px;
  background: rgb(255 255 255 / .06);
  box-shadow: 0 0 0 1px rgb(255 255 255 / .12);
}

.hs-key {
  flex: none;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 9px;
  background: oklch(0.82 0.14 80 / .2);
  color: var(--amber);
}

.hs-find b {
  display: block;
  font-size: 15px;
  font-weight: 600;
  color: var(--bright);
}

.hs-find small {
  display: block;
  margin-top: 2px;
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 12.5px;
  color: var(--meta);
  white-space: nowrap;
}

.hs-badges {
  justify-items: end;
}

.hs-badge {
  box-sizing: border-box;
  height: 24px;
  padding: 0 10px;
  border-radius: 6px;
  font-size: 12.5px;
  font-weight: 600;
  line-height: 24px;
  white-space: nowrap;
}

.hs-badge.is-asking {
  background: rgb(255 255 255 / .12);
  color: var(--body);
}

.hs-badge.is-live {
  background: oklch(0.62 0.20 27 / .28);
  color: oklch(0.86 0.09 27);
}

.hs-vhead {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.hs-checked {
  justify-items: end;
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 12.5px;
  color: var(--meta);
  white-space: nowrap;
}

.hs-checked .is-again {
  color: var(--amber);
}

.hs-checked .is-now {
  color: var(--bright);
}

.hs-counts {
  display: flex;
  gap: 22px;
  margin-top: 12px;
  font-size: 14px;
  color: var(--body);
}

.hs-counts span:not(.stack) {
  display: flex;
  align-items: center;
  gap: 7px;
  white-space: nowrap;
}

.hs-counts i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.hs-counts .is-none {
  color: var(--meta);
}

.hs-counts .is-none i {
  background: oklch(0.45 0.03 55);
}

.hs-counts .is-crit {
  font-weight: 600;
  color: oklch(0.86 0.09 27);
}

.hs-counts .is-crit i {
  background: oklch(0.74 0.16 27);
}

.sev-high {
  background: oklch(0.78 0.14 50);
}

.sev-med {
  background: oklch(0.84 0.12 88);
}

.hs-rows {
  position: relative;
  height: calc(var(--vrow) * 3);
  margin-top: 16px;
}

.hs-row {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 14px;
  height: var(--vrow);
  padding: 0 12px;
  border-radius: 8px;
  font-size: 13.5px;
  white-space: nowrap;
}

.hs-row-2 {
  top: var(--vrow);
}

.hs-row b {
  width: 118px;
  font-family: 'Fira Code', ui-monospace, monospace;
  font-weight: 500;
  color: var(--bright);
}

.hs-row span {
  flex: 1;
  min-width: 0;
  color: var(--meta);
}

.hs-row em {
  font-style: normal;
  font-weight: 600;
}

.hs-row .is-high {
  color: oklch(0.78 0.14 50);
}

.hs-row .is-crit {
  color: oklch(0.86 0.09 27);
}

.hs-sweep {
  position: absolute;
  left: 0;
  top: 200px;
  width: 120px;
  height: 270px;
  background: linear-gradient(to right, transparent, oklch(0.82 0.14 80 / .22), transparent);
}

@keyframes hs-spin { to { transform: rotate(360deg); } }

@keyframes hs-c1 { 0% { opacity: 0; } 0.21% { opacity: 0; } 1.67% { opacity: 1; } 15.42% { opacity: 1; } 16.67% { opacity: 0; } 100% { opacity: 0; } }
.a-c1 { animation: hs-c1 24s linear infinite both; opacity: 0; }
@keyframes hs-c2 { 0% { opacity: 0; } 16.88% { opacity: 0; } 18.33% { opacity: 1; } 36.25% { opacity: 1; } 37.5% { opacity: 0; } 100% { opacity: 0; } }
.a-c2 { animation: hs-c2 24s linear infinite both; opacity: 0; }
@keyframes hs-c3 { 0% { opacity: 0; } 37.71% { opacity: 0; } 39.17% { opacity: 1; } 55% { opacity: 1; } 56.25% { opacity: 0; } 100% { opacity: 0; } }
.a-c3 { animation: hs-c3 24s linear infinite both; opacity: 0; }
@keyframes hs-c4 { 0% { opacity: 0; } 56.46% { opacity: 0; } 57.92% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-c4 { animation: hs-c4 24s linear infinite both; opacity: 1; }
@keyframes hs-card { 0% { opacity: 0; transform: translateY(10px); } 1.25% { opacity: 0; transform: translateY(10px); animation-timing-function: ease-out; } 2.92% { opacity: 1; transform: translateY(0px); } 93.75% { opacity: 1; transform: translateY(0px); } 95.62% { opacity: 0; transform: translateY(10px); } 100% { opacity: 0; transform: translateY(10px); } }
.a-card { animation: hs-card 24s linear infinite both; opacity: 1; transform: translateY(0px); }
@keyframes hs-p1 { 0% { opacity: 0; } 0.42% { opacity: 0; } 1.04% { opacity: 1; } 35.83% { opacity: 1; } 36.46% { opacity: 0; } 100% { opacity: 0; } }
.a-p1 { animation: hs-p1 24s linear infinite both; opacity: 0; }
@keyframes hs-p2 { 0% { opacity: 0; } 35.83% { opacity: 0; } 36.46% { opacity: 1; } 48.33% { opacity: 1; } 48.96% { opacity: 0; } 100% { opacity: 0; } }
.a-p2 { animation: hs-p2 24s linear infinite both; opacity: 0; }
@keyframes hs-p3 { 0% { opacity: 0; } 48.33% { opacity: 0; } 48.96% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-p3 { animation: hs-p3 24s linear infinite both; opacity: 1; }
@keyframes hs-hl { 0% { background-color: oklch(0.82 0.14 80 / 0); } 21.67% { background-color: oklch(0.82 0.14 80 / 0); } 22.5% { background-color: oklch(0.82 0.14 80 / .14); } 93.75% { background-color: oklch(0.82 0.14 80 / .14); } 95.62% { background-color: oklch(0.82 0.14 80 / 0); } 100% { background-color: oklch(0.82 0.14 80 / 0); } }
.a-hl { animation: hs-hl 24s linear infinite both; background-color: oklch(0.82 0.14 80 / .14); }
@keyframes hs-b1 { 0% { opacity: 0; } 0.42% { opacity: 0; } 1.04% { opacity: 1; } 35.83% { opacity: 1; } 36.46% { opacity: 0; } 100% { opacity: 0; } }
.a-b1 { animation: hs-b1 24s linear infinite both; opacity: 0; }
@keyframes hs-b2 { 0% { opacity: 0; } 35.83% { opacity: 0; } 36.46% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-b2 { animation: hs-b2 24s linear infinite both; opacity: 1; }
@keyframes hs-s1 { 0% { opacity: 0; } 0.42% { opacity: 0; } 1.04% { opacity: 1; } 21.67% { opacity: 1; } 22.29% { opacity: 0; } 100% { opacity: 0; } }
.a-s1 { animation: hs-s1 24s linear infinite both; opacity: 0; }
@keyframes hs-s2 { 0% { opacity: 0; } 21.67% { opacity: 0; } 22.29% { opacity: 1; } 30.83% { opacity: 1; } 31.46% { opacity: 0; } 100% { opacity: 0; } }
.a-s2 { animation: hs-s2 24s linear infinite both; opacity: 0; }
@keyframes hs-s3 { 0% { opacity: 0; } 30.83% { opacity: 0; } 31.46% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-s3 { animation: hs-s3 24s linear infinite both; opacity: 1; }
@keyframes hs-d1 { 0% { opacity: 0; } 0.42% { opacity: 0; } 1.04% { opacity: 1; } 35.83% { opacity: 1; } 36.46% { opacity: 0; } 100% { opacity: 0; } }
.a-d1 { animation: hs-d1 24s linear infinite both; opacity: 0; }
@keyframes hs-d2 { 0% { opacity: 0; } 35.83% { opacity: 0; } 36.46% { opacity: 1; } 40.83% { opacity: 1; } 41.46% { opacity: 0; } 100% { opacity: 0; } }
.a-d2 { animation: hs-d2 24s linear infinite both; opacity: 0; }
@keyframes hs-d3 { 0% { opacity: 0; } 40.83% { opacity: 0; } 41.46% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-d3 { animation: hs-d3 24s linear infinite both; opacity: 1; }
@keyframes hs-r1 { 0% { opacity: 0; } 0.42% { opacity: 0; } 1.04% { opacity: 1; } 40.83% { opacity: 1; } 41.46% { opacity: 0; } 100% { opacity: 0; } }
.a-r1 { animation: hs-r1 24s linear infinite both; opacity: 0; }
@keyframes hs-r2 { 0% { opacity: 0; } 40.83% { opacity: 0; } 41.46% { opacity: 1; } 48.33% { opacity: 1; } 48.96% { opacity: 0; } 100% { opacity: 0; } }
.a-r2 { animation: hs-r2 24s linear infinite both; opacity: 0; }
@keyframes hs-r3 { 0% { opacity: 0; } 48.33% { opacity: 0; } 48.96% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-r3 { animation: hs-r3 24s linear infinite both; opacity: 1; }
@keyframes hs-l1 { 0% { opacity: 0.18; } 21.67% { opacity: 0.18; } 22.5% { opacity: 0.6; } 32.5% { opacity: 0.6; } 34.17% { opacity: 0.18; } 100% { opacity: 0.18; } }
.a-l1 { animation: hs-l1 24s linear infinite both; opacity: 0.18; }
@keyframes hs-dot { 0% { opacity: 0; transform: translate(542px,349px); } 22.08% { opacity: 0; transform: translate(542px,349px); } 22.5% { opacity: 1; transform: translate(542px,349px); animation-timing-function: ease-in-out; } 25% { opacity: 1; transform: translate(634px,349px); } 25.63% { opacity: 0; transform: translate(634px,349px); } 93.75% { opacity: 0; transform: translate(634px,349px); } 93.83% { opacity: 0; transform: translate(542px,349px); } 100% { opacity: 0; transform: translate(542px,349px); } }
.a-dot { animation: hs-dot 24s linear infinite both; opacity: 0; transform: translate(634px,349px); }
@keyframes hs-empty { 0% { opacity: 1; } 24.58% { opacity: 1; } 25.21% { opacity: 0; } 93.75% { opacity: 0; } 95.62% { opacity: 1; } 100% { opacity: 1; } }
.a-empty { animation: hs-empty 24s linear infinite both; opacity: 0; }
@keyframes hs-find { 0% { opacity: 0; transform: translateX(-12px); } 25% { opacity: 0; transform: translateX(-12px); animation-timing-function: ease-out; } 26.46% { opacity: 1; transform: translateX(0px); } 93.75% { opacity: 1; transform: translateX(0px); } 95.62% { opacity: 0; transform: translateX(-12px); } 100% { opacity: 0; transform: translateX(-12px); } }
.a-find { animation: hs-find 24s linear infinite both; opacity: 1; transform: translateX(0px); }
@keyframes hs-k1 { 0% { opacity: 0; } 25.42% { opacity: 0; } 26.04% { opacity: 1; } 30.83% { opacity: 1; } 31.46% { opacity: 0; } 100% { opacity: 0; } }
.a-k1 { animation: hs-k1 24s linear infinite both; opacity: 0; }
@keyframes hs-k2 { 0% { opacity: 0; } 30.83% { opacity: 0; } 31.46% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-k2 { animation: hs-k2 24s linear infinite both; opacity: 1; }
@keyframes hs-v0 { 0% { opacity: 1; } 48.75% { opacity: 1; } 49.38% { opacity: 0; } 93.75% { opacity: 0; } 95.62% { opacity: 1; } 100% { opacity: 1; } }
.a-v0 { animation: hs-v0 24s linear infinite both; opacity: 0; }
@keyframes hs-v1 { 0% { opacity: 0; } 49.17% { opacity: 0; } 49.79% { opacity: 1; } 57.92% { opacity: 1; } 58.54% { opacity: 0; } 100% { opacity: 0; } }
.a-v1 { animation: hs-v1 24s linear infinite both; opacity: 0; }
@keyframes hs-v2 { 0% { opacity: 0; } 57.92% { opacity: 0; } 58.54% { opacity: 1; } 64.58% { opacity: 1; } 65.21% { opacity: 0; } 100% { opacity: 0; } }
.a-v2 { animation: hs-v2 24s linear infinite both; opacity: 0; }
@keyframes hs-v3 { 0% { opacity: 0; } 64.58% { opacity: 0; } 65.21% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-v3 { animation: hs-v3 24s linear infinite both; opacity: 1; }
@keyframes hs-cnt { 0% { opacity: 0; } 49.17% { opacity: 0; } 50.42% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-cnt { animation: hs-cnt 24s linear infinite both; opacity: 1; }
@keyframes hs-n0 { 0% { opacity: 1; } 64.79% { opacity: 1; } 65.21% { opacity: 0; } 93.75% { opacity: 0; } 95.62% { opacity: 1; } 100% { opacity: 1; } }
.a-n0 { animation: hs-n0 24s linear infinite both; opacity: 0; }
@keyframes hs-n1 { 0% { opacity: 0; } 65% { opacity: 0; } 65.62% { opacity: 1; } 93.75% { opacity: 1; } 95.62% { opacity: 0; } 100% { opacity: 0; } }
.a-n1 { animation: hs-n1 24s linear infinite both; opacity: 1; }
@keyframes hs-va { 0% { opacity: 0; transform: translateY(0px); } 49.58% { opacity: 0; transform: translateY(0px); } 50.83% { opacity: 1; transform: translateY(0px); } 64.58% { opacity: 1; transform: translateY(0px); animation-timing-function: ease-in-out; } 66.04% { opacity: 1; transform: translateY(var(--vrow)); } 93.75% { opacity: 1; transform: translateY(var(--vrow)); } 95.42% { opacity: 0; transform: translateY(var(--vrow)); } 95.62% { opacity: 0; transform: translateY(var(--vrow)); } 95.71% { opacity: 0; transform: translateY(0px); } 100% { opacity: 0; transform: translateY(0px); } }
.a-va { animation: hs-va 24s linear infinite both; opacity: 1; transform: translateY(var(--vrow)); }
@keyframes hs-vb { 0% { opacity: 0; transform: translateY(0px); } 50.21% { opacity: 0; transform: translateY(0px); } 51.46% { opacity: 1; transform: translateY(0px); } 64.58% { opacity: 1; transform: translateY(0px); animation-timing-function: ease-in-out; } 66.04% { opacity: 1; transform: translateY(var(--vrow)); } 93.75% { opacity: 1; transform: translateY(var(--vrow)); } 95.42% { opacity: 0; transform: translateY(var(--vrow)); } 95.62% { opacity: 0; transform: translateY(var(--vrow)); } 95.71% { opacity: 0; transform: translateY(0px); } 100% { opacity: 0; transform: translateY(0px); } }
.a-vb { animation: hs-vb 24s linear infinite both; opacity: 1; transform: translateY(var(--vrow)); }
@keyframes hs-vn { 0% { opacity: 0; background-color: oklch(0.62 0.20 27 / .3); } 65.62% { opacity: 0; background-color: oklch(0.62 0.20 27 / .3); } 66.88% { opacity: 1; background-color: oklch(0.62 0.20 27 / .3); } 75.83% { opacity: 1; background-color: oklch(0.62 0.20 27 / .3); } 79.17% { opacity: 1; background-color: oklch(0.62 0.20 27 / .14); } 93.75% { opacity: 1; background-color: oklch(0.62 0.20 27 / .14); } 95.42% { opacity: 0; background-color: oklch(0.62 0.20 27 / .14); } 95.62% { opacity: 0; background-color: oklch(0.62 0.20 27 / .14); } 95.71% { opacity: 0; background-color: oklch(0.62 0.20 27 / .3); } 100% { opacity: 0; background-color: oklch(0.62 0.20 27 / .3); } }
.a-vn { animation: hs-vn 24s linear infinite both; opacity: 1; background-color: oklch(0.62 0.20 27 / .14); }
@keyframes hs-sweep { 0% { opacity: 0; transform: translateX(-140px); } 58.12% { opacity: 0; transform: translateX(-140px); } 58.33% { opacity: 1; transform: translateX(-140px); } 64.17% { opacity: 1; transform: translateX(760px); } 64.58% { opacity: 0; transform: translateX(760px); } 64.79% { opacity: 0; transform: translateX(760px); } 64.88% { opacity: 0; transform: translateX(-140px); } 100% { opacity: 0; transform: translateX(-140px); } }
.a-sweep { animation: hs-sweep 24s linear infinite both; opacity: 0; transform: translateX(-140px); }

@media (max-width: 1391px) {
  .hs {
    zoom: .857;
  }
}

@media (max-width: 1199px) {
  .hs {
    zoom: .726;
  }
}

@media (max-width: 1023px) {
  .hs {
    zoom: 1;
    display: flex;
    flex-direction: column;
    height: auto;
    padding: 32px 0 0 28px;
  }

  .hs-sunset {
    top: -40px;
    width: 560px;
    height: 460px;
  }

  .hs-bridge {
    top: -14px;
    width: min(54%, 400px);
  }

  .hs-caps {
    position: relative;
    left: auto;
    top: auto;
    width: calc(46% - 28px);
    min-height: 204px;
  }

  .hs-card {
    position: relative;
    left: auto;
    top: auto;
    width: auto;
    max-width: 560px;
    margin: 28px 28px 0 0;
  }

  .hs-foot {
    position: static;
    width: auto;
    max-width: 560px;
    margin: 14px 28px 0 0;
  }

  .hs-line,
  .hs-dot {
    display: none;
  }

  .hs-bezel {
    position: relative;
    left: auto;
    top: auto;
    height: auto;
    margin-top: 32px;
  }

  .hs-panel {
    padding-bottom: 20px;
  }

  .hs-sweep {
    top: 0;
    height: 100%;
  }
}

@media (max-width: 639px) {
  .hs {
    --vrow: 62px;
    padding: 0 0 0 16px;
  }

  .hs-sunset {
    top: -90px;
    right: -120px;
  }

  .hs-bridge {
    top: -8px;
    width: min(82%, 320px);
  }

  .hs-caps {
    width: auto;
    min-height: 0;
    margin: 168px 16px 0 0;
    font-size: 21px;
    line-height: 29px;
  }

  .hs-card {
    margin: 22px 16px 0 0;
  }

  .hs-head {
    height: 78px;
    padding: 16px 16px 0;
  }

  .hs-ref {
    margin-top: 10px;
    font-size: 12px;
  }

  .hs-step {
    height: 54px;
    padding: 0 16px;
  }

  .hs-states {
    font-size: 13px;
  }

  .hs-foot {
    margin: 12px 16px 0 0;
  }

  .hs-bezel {
    margin-top: 24px;
  }

  .hs-tabs {
    gap: 20px;
    padding: 0 20px;
    font-size: 13.5px;
  }

  .hs-tabs span:first-child {
    display: none;
  }

  .hs-sec {
    padding: 18px 16px 0 20px;
  }

  .hs-sec-vul {
    margin: 18px 16px 0 20px;
    padding: 17px 0 0;
  }

  .hs-find {
    display: grid;
    grid-template-columns: 36px minmax(0, 1fr) auto;
    gap: 6px 10px;
    align-items: center;
    padding: 12px 10px;
  }

  .hs-find > .min-w-0 {
    display: contents;
  }

  .hs-find b {
    font-size: 14px;
  }

  .hs-find small {
    grid-row: 2;
    grid-column: 1 / -1;
    margin-top: 0;
    white-space: normal;
  }

  .hs-badges {
    grid-row: 1;
    grid-column: 3;
  }

  .hs-empty {
    top: 28px;
  }

  .hs-vhead {
    display: block;
  }

  .hs-checked {
    justify-items: start;
    margin-top: 4px;
  }

  .hs-counts {
    gap: 16px;
    margin-top: 10px;
    font-size: 13.5px;
  }

  .hs-rows {
    margin-top: 12px;
  }

  .hs-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-content: center;
    gap: 2px 12px;
    padding: 0 10px;
    white-space: normal;
  }

  .hs-row b {
    width: auto;
  }

  .hs-row span {
    grid-row: 2;
    grid-column: 1 / -1;
    font-size: 12.5px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .hs *,
  .hs-spin {
    animation: none !important;
  }
}
</style>
