<script setup lang="ts">
interface Placement {
  x: number;
  y: number;
  z: number;
}

const windows = reactive<Placement[]>([
  { x: 0, y: 0, z: 1 },
  { x: 0, y: 0, z: 2 },
  { x: 0, y: 0, z: 3 },
]);

let top = 3;

const lineNumbers = Array.from({ length: 18 }, (_, index) => index + 1).join('\n');

function placement(index: number) {
  const item = windows[index]!;
  return { transform: `translate(${item.x}px, ${item.y}px)`, zIndex: item.z };
}

function drag(index: number, event: PointerEvent) {
  const element = event.currentTarget as HTMLElement;
  if (event.button || getComputedStyle(element).position !== 'absolute') {
    return;
  }
  const item = windows[index]!;
  const scene = element.parentElement!.getBoundingClientRect();
  const rect = element.getBoundingClientRect();
  const start = { px: event.clientX, py: event.clientY, x: item.x, y: item.y };
  const clamp = (value: number, low: number, high: number) => Math.max(low, Math.min(high, value));
  top += 1;
  item.z = top;

  const move = (moveEvent: PointerEvent) => {
    const dx = clamp(moveEvent.clientX - start.px, scene.left - rect.right + 120, scene.right - rect.left - 120);
    const dy = clamp(moveEvent.clientY - start.py, scene.top - rect.top, scene.bottom - rect.top - 60);
    item.x = start.x + dx;
    item.y = start.y + dy;
  };
  const stop = () => {
    element.removeEventListener('pointermove', move);
    element.removeEventListener('pointerup', stop);
    element.removeEventListener('pointercancel', stop);
  };
  element.addEventListener('pointermove', move);
  element.addEventListener('pointerup', stop);
  element.addEventListener('pointercancel', stop);
  element.setPointerCapture(event.pointerId);
  event.preventDefault();
}
</script>

<template>
  <div class="scr">
    <div class="scr-note">
      <h3 class="font-display text-2xl/tight text-stone-900 dark:text-stone-100">Script it, or leave it to your pipeline.</h3>
      <p class="mt-2.5 text-base text-[oklch(0.36_0.04_300)] dark:text-[oklch(0.85_0.04_300)]">
        One CLI for your terminal and your CI. GitHub Actions deploys without a stored token, and the API does everything the dashboard does.
      </p>
      <p class="mt-3 flex flex-wrap items-baseline gap-x-5 gap-y-1.5 text-base">
        <NuxtLink to="/docs/cli" class="font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">CLI reference</NuxtLink>
        <NuxtLink to="/docs/api" class="font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">API</NuxtLink>
        <span class="drag-hint text-sm text-[oklch(0.42_0.05_300)] dark:text-[oklch(0.78_0.05_300)]">Go on, drag the windows.</span>
      </p>
    </div>

    <div class="win win-0 ide" role="group" aria-label="An editor with a script that deploys through the API" :style="placement(0)" @pointerdown="drag(0, $event)">
      <div class="ide-bar" aria-hidden="true">
        <i /><i /><i />
        <span class="ide-tab"><svg width="14" height="14" viewBox="0 0 14 14"><rect width="14" height="14" rx="2.5" fill="oklch(0.86 0.15 95)" /></svg>deploy.js</span>
      </div>
      <div class="ide-code">
        <pre aria-hidden="true">{{ lineNumbers }}</pre>
        <pre><span class="k">import</span> { execSync } <span class="k">from</span> <span class="s">'node:child_process'</span>;

<span class="k">const</span> token = <span class="f">execSync</span>(<span class="s">'lucity token'</span>, { encoding: <span class="s">'utf8'</span> }).<span class="f">trim</span>();
<span class="k">const</span> query = <span class="s">`mutation {</span>
<span class="s">  deploy(service: "zeitlos/vouch/production/web") { id status }</span>
<span class="s">}`</span>;

<span class="k">const</span> response = <span class="k">await</span> <span class="f">fetch</span>(<span class="s">'https://lucity.cloud/graphql'</span>, {
  method: <span class="s">'POST'</span>,
  headers: {
    Authorization: <span class="s">`Bearer </span><span class="v">${token}</span><span class="s">`</span>,
    <span class="s">'Content-Type'</span>: <span class="s">'application/json'</span>,
  },
  body: JSON.<span class="f">stringify</span>({ query }),
});

<span class="k">const</span> { data } = <span class="k">await</span> response.<span class="f">json</span>();
console.<span class="f">log</span>(data.deploy.status);</pre>
      </div>
    </div>

    <div class="win win-1 gh" role="group" aria-label="A GitHub Actions workflow that deploys to Lucity without a stored token" :style="placement(1)" @pointerdown="drag(1, $event)">
      <div class="gh-bar" aria-hidden="true"><i /><i /><i /><span class="ml-2.5">.github/workflows/deploy.yml</span></div>
      <pre><span class="yl"><span class="yk">name</span>: <span class="yv">Deploy</span></span><span class="yl"><span class="yk">on</span>:</span><span class="yl">  <span class="yk">push</span>:</span><span class="yl">    <span class="yk">branches</span>: <span class="yv">[main]</span></span><span class="yl"> </span><span class="yl"><span class="yk">jobs</span>:</span><span class="yl">  <span class="yk">deploy</span>:</span><span class="yl">    <span class="yk">runs-on</span>: <span class="yv">ubuntu-latest</span></span><span class="yl">    <span class="yk">permissions</span>:</span><span class="yl yl-hi">      <span class="yk">id-token</span>: <span class="yv">write</span>   <span class="yc"># all it needs, no secret to store</span></span><span class="yl">    <span class="yk">steps</span>:</span><span class="yl">      - <span class="yk">name</span>: <span class="yv">Install the Lucity CLI</span> <span class="yd">⋯</span></span><span class="yl">      - <span class="yk">name</span>: <span class="yv">Deploy</span></span><span class="yl">        <span class="yk">run</span>: <span class="yv">|</span></span><span class="yl">          <span class="yv">lucity deploy vouch/production/web \</span></span><span class="yl">            <span class="yv">--ref "$GITHUB_SHA" --wait</span></span></pre>
    </div>

    <div class="win win-2 terminal" role="group" aria-label="A terminal running the Lucity CLI" :style="placement(2)" @pointerdown="drag(2, $event)">
      <div class="terminal-bar" aria-hidden="true"><i /><i /><i /><span>lucity</span></div>
      <pre><span class="c-dim">$</span> lucity deploy vouch/production/web --wait
Release 8c1d2f7a90 queued for zeitlos/vouch/production/web (main)
<span class="c-dim">→ building</span>
<span class="c-dim">→ deploying</span>
<span class="c-dim">  rollout progressing</span>
<span class="c-dim">→ live</span>
<span class="c-ok">✓ Release 8c1d2f7a90 is live</span>
  https://vouch.example.com

<span class="c-dim">$</span> lucity status vouch/production/web
Service zeitlos/vouch/production/web: <span class="c-ok">HEALTHY</span>
Deployment zeitlos/vouch/production/web/56bc5cdfc5: <span class="c-ok">ACTIVE</span>
Replicas:   3/3 ready
Source:     c41f0aa
Rollout:    <span class="c-ok">READY</span></pre>
    </div>
  </div>
</template>

<style scoped>
.scr {
  position: relative;
  margin: 28px 0 -40px;
  height: 660px;
}

.scr-note {
  position: absolute;
  left: 0;
  bottom: 40px;
  width: 28%;
}

.win {
  position: absolute;
  border-radius: 12px;
  overflow: hidden;
  cursor: grab;
  user-select: none;
  touch-action: none;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .16), 0 26px 50px -20px oklch(0.25 0.04 55 / .5);
}

.win:active {
  cursor: grabbing;
}

.win-0 {
  left: 0;
  top: 0;
  width: 47%;
}

.win-1 {
  left: 51%;
  top: 22px;
  width: 55%;
}

.win-2 {
  left: 30%;
  top: 408px;
  width: 45%;
}

.ide pre,
.gh pre {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.65;
  white-space: pre;
  overflow: hidden;
}

.ide {
  background: oklch(0.23 0.035 290);
  color: oklch(0.91 0.02 290);
}

.ide-bar,
.gh-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 16px;
  font-size: 13px;
}

.ide-bar i,
.gh-bar i {
  flex: none;
  width: 11px;
  height: 11px;
  border-radius: 50%;
}

.ide-bar {
  background: oklch(0.185 0.035 290);
  color: oklch(0.82 0.03 290);
}

.ide-bar i {
  background: oklch(0.38 0.04 290);
}

.ide-tab {
  margin-left: 10px;
  align-self: stretch;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 14px;
  background: oklch(0.23 0.035 290);
  font-family: var(--font-mono);
  font-size: 12.5px;
}

.ide-code {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0 18px;
  padding: 16px 18px 20px;
}

.ide-code pre:first-child {
  text-align: right;
  color: oklch(0.52 0.03 290);
}

.k { color: oklch(0.78 0.12 300); }
.s { color: oklch(0.83 0.13 150); }
.f { color: oklch(0.84 0.09 230); }
.v { color: oklch(0.85 0.11 70); }

.gh {
  background: oklch(0.985 0.006 80);
  color: oklch(0.25 0.03 55);
}

.gh-bar {
  background: oklch(0.955 0.012 78);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
  color: oklch(0.45 0.04 55);
}

.gh-bar i {
  background: oklch(0.80 0.03 70);
}

.gh pre {
  padding: 14px 0;
}

.yl {
  display: block;
  padding: 0 18px;
}

.yl-hi {
  background: oklch(0.75 0.18 160 / .2);
  box-shadow: inset 3px 0 0 oklch(0.60 0.17 160);
}

.yk { color: oklch(0.42 0.15 300); }
.yv { color: oklch(0.38 0.10 160); }
.yc { color: oklch(0.40 0.11 160); font-weight: 500; }
.yd { color: oklch(0.54 0.04 60); }

.terminal {
  background: oklch(0.20 0.03 55);
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

.c-dim { color: oklch(0.70 0.03 70); }
.c-ok { color: oklch(0.80 0.16 160); }

@media (max-width: 1100px) {
  .scr {
    height: auto;
    margin-bottom: 0;
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .scr-note {
    position: static;
    width: auto;
    order: -1;
  }

  .drag-hint {
    display: none;
  }

  .win {
    position: static;
    width: auto;
    transform: none !important;
    cursor: default;
    touch-action: auto;
    user-select: text;
  }

  .ide pre,
  .gh pre,
  .terminal pre {
    overflow-x: auto;
  }
}
</style>
