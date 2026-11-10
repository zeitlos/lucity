<script setup lang="ts">
interface FinderEntry {
  name: string;
  kind: 'zip' | 'folder' | 'file';
  depth: number;
  twist?: 'open' | 'closed';
  note: string;
}

const contents = [
  { icon: 'i-lucide-ship', label: 'Helm chart with all templates' },
  { icon: 'i-lucide-git-branch', label: 'Config per environment' },
  { icon: 'i-lucide-shield', label: 'Zero Lucity dependency after export' },
];

const entries: FinderEntry[] = [
  { name: 'shop-ejected.zip', kind: 'zip', depth: 0, note: 'The download' },
  { name: 'shop-ejected', kind: 'folder', depth: 0, twist: 'open', note: '' },
  { name: 'README.md', kind: 'file', depth: 1, note: 'How to install it' },
  { name: 'build.sh', kind: 'file', depth: 1, note: 'Rebuilds your images' },
  { name: 'chart', kind: 'folder', depth: 1, twist: 'closed', note: 'How every service runs' },
  { name: 'values', kind: 'folder', depth: 1, twist: 'open', note: 'One file per environment' },
  { name: 'development.yaml', kind: 'file', depth: 2, note: '' },
  { name: 'production.yaml', kind: 'file', depth: 2, note: 'Services, variables, databases' },
];

const commands = [
  { fade: 't1', lead: '$ ', typed: 'ty1', text: './build.sh production registry.example.com/shop' },
  { fade: 't2', lead: '  web      ', typed: '', text: 'built and pushed' },
  { fade: 't3', lead: '  worker   ', typed: '', text: 'built and pushed' },
  { fade: 't4', lead: '$ ', typed: 'ty2', text: 'helm upgrade --install lucity-app ./chart \\' },
  { fade: 't5', lead: '    ', typed: 'ty3', text: '-f values/production.yaml -n shop-production' },
];

const nodes = [
  { name: 'web', icon: 'i-simple-icons-github', note: '2 instances' },
  { name: 'worker', icon: 'i-simple-icons-github', note: '1 instance' },
  { name: 'main', icon: 'i-lucide-database', note: 'PostgreSQL, empty' },
];

const steps = [
  { name: 'Click Eject', note: 'In the dashboard, on any project. It keeps running on Lucity as before.' },
  { name: 'Open the zip', note: 'The chart, one values file per environment and a build script.' },
  { name: 'Install it', note: 'Rebuild your images, change a few values, run one helm command.' },
  { name: 'It runs', note: 'The same services, on a cluster of your own. Nothing in them knows Lucity.' },
];

const folderPath = 'M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.7-.9l-.8-1.2A2 2 0 0 0 7.9 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z';
</script>

<template>
  <div>
    <LandingStage
      tint="keep"
      class="scene-fit"
      role="img"
      aria-label="Leaving, in four steps. You click Download in the Eject dialog of a project, which keeps running on Lucity. A zip arrives with a Helm chart, one values file per environment, a build script and a readme. In a terminal you rebuild your images and run one helm command against a cluster of your own. The same services come up there: web, worker and an empty PostgreSQL database that is ready for your data."
    >
      <div class="scene" aria-hidden="true">
        <div class="side side-left">
          <div class="dialog">
            <div class="dialog-icon"><UIcon name="i-lucide-file-archive" class="size-6" /></div>
            <div class="dialog-title">Eject project</div>
            <div class="dialog-text">Export this project as a self-contained zip you can deploy anywhere.</div>
            <div class="dialog-list">
              <div v-for="item in contents" :key="item.label">
                <span class="dialog-mark"><UIcon :name="item.icon" class="size-3.5" /></span>{{ item.label }}
              </div>
            </div>
            <div class="dialog-note">Non-destructive. Your project keeps running on Lucity.</div>
            <div class="dialog-foot">
              <span class="dialog-button dialog-cancel">Cancel</span>
              <span class="dialog-button dialog-download nlh-press">
                <span class="nlh-dla"><UIcon name="i-lucide-download" class="size-[15px]" />Download .zip</span>
                <span class="nlh-dlb"><UIcon name="i-lucide-check" class="size-[15px]" />Downloaded</span>
              </span>
            </div>
          </div>

          <div class="finder">
            <div class="finder-bar"><i /><i /><i /><b>Downloads</b></div>
            <div class="finder-head"><span>Name</span><span>What it is for</span></div>
            <div v-for="(entry, index) in entries" :key="entry.name" class="finder-row" :class="`nlh-f${index}`">
              <span :style="{ paddingLeft: `${entry.depth * 14}px` }">
                <svg v-if="entry.twist" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="oklch(0.50 0.03 55)" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round">
                  <path :d="entry.twist === 'open' ? 'm6 9 6 6 6-6' : 'm9 6 6 6-6 6'" />
                </svg>
                <span v-else class="finder-gap" />
                <UIcon v-if="entry.kind === 'zip'" name="i-lucide-file-archive" class="zip-icon size-[18px]" />
                <svg v-else-if="entry.kind === 'folder'" width="18" height="18" viewBox="0 0 24 24" fill="oklch(0.90 0.06 0)" stroke="oklch(0.58 0.20 0)" stroke-width="1.8" stroke-linejoin="round"><path :d="folderPath" /></svg>
                <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="oklch(0.99 0.004 80)" stroke="oklch(0.45 0.04 55)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z" /><path d="M14 2v4a2 2 0 0 0 2 2h4" /></svg>
                {{ entry.name }}
              </span>
              <span>{{ entry.note }}</span>
            </div>
          </div>

          <div class="chip nlh-zip"><UIcon name="i-lucide-file-archive" class="zip-icon size-[18px]" />shop-ejected.zip</div>

          <div class="pointer nlh-cur">
            <svg width="22" height="22" viewBox="0 0 24 24"><path d="M4 3l15 7.5-6.4 1.9L10.2 19z" fill="oklch(0.25 0.03 55)" stroke="oklch(0.99 0.006 80)" stroke-width="1.6" stroke-linejoin="round" /></svg>
          </div>
        </div>

        <div class="side side-right">
          <div class="terminal">
            <div class="terminal-bar"><i /><i /><i /><span>shop-ejected</span></div>
            <div class="terminal-body">
              <div class="c-dim">{{ '# in the unzipped folder, with access to your cluster' }}</div>
              <div v-for="command in commands" :key="command.fade" :class="`nlh-${command.fade}`"><span :class="{ 'c-dim': command.lead.trim() }">{{ command.lead }}</span><span :class="command.typed ? `typed nlh-${command.typed}` : undefined">{{ command.text }}</span></div>
              <div class="nlh-t6"><span>{{ '  STATUS: ' }}</span><span class="c-ok">deployed</span></div>
            </div>
          </div>

          <div class="cluster">
            <div class="cluster-head">
              <UIcon name="i-lucide-server" class="size-4" /><span>your cluster</span><em>/</em><span>shop-production</span>
            </div>
            <div class="cluster-empty nlh-empty">Nothing here yet</div>
            <div v-for="(node, index) in nodes" :key="node.name" class="node" :class="`nlh-n${index}`" :style="{ left: `${16 + index * 144}px` }">
              <b><UIcon :name="node.icon" class="size-4" :class="{ 'node-database': node.name === 'main' }" />{{ node.name }}</b>
              <small>{{ node.note }}</small>
              <div class="node-foot">
                <span class="pill" :class="`nlh-ns${index}`"><i class="dot-starting" />Starting</span>
                <span class="pill" :class="`nlh-nr${index}`"><i class="dot-running" />Running</span>
              </div>
            </div>
          </div>

          <div class="honest nlh-honest">Databases start empty. Moving your data is the next step, and the guide walks you through it.</div>
        </div>

        <div class="strip">
          <div v-for="(step, index) in steps" :key="step.name">
            <div class="strip-rail"><div :class="`nlh-bar${index}`" /></div>
            <div class="strip-name" :class="`nlh-tt${index}`">{{ step.name }}</div>
            <div class="strip-note" :class="`nlh-nt${index}`">{{ step.note }}</div>
          </div>
        </div>
      </div>
    </LandingStage>

    <ol class="scene-list">
      <li v-for="(step, index) in steps" :key="step.name">
        <span aria-hidden="true">{{ index + 1 }}</span>
        <div><b>{{ step.name }}</b>{{ step.note }}</div>
      </li>
    </ol>
  </div>
</template>

<style scoped>
.scene-list {
  display: none;
  margin: 0;
  padding: 6px 22px;
  list-style: none;
  border-radius: 20px;
  background-color: oklch(0.928 0.026 8);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .14) 1px, transparent 0);
  background-size: 22px 22px;
  box-shadow: inset 0 0 0 1px oklch(0.25 0.04 55 / .07);
  font-size: 16px;
  line-height: 1.5;
  color: oklch(0.36 0.05 0);
}

.scene-list li {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr);
  gap: 0 10px;
  align-items: baseline;
  padding: 16px 0;
  border-top: 1px solid oklch(0.25 0.04 55 / .12);
}

.scene-list li:first-child {
  border-top: 0;
}

.scene-list li > span {
  font-family: var(--font-display);
  font-size: 30px;
  line-height: 1;
  color: oklch(0.48 0.17 0);
}

.scene-list b {
  display: block;
  font-size: 18px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.dark .scene-list {
  background-color: oklch(0.285 0.05 8);
  background-image: radial-gradient(circle at 1px 1px, oklch(1 0 0 / .10) 1px, transparent 0);
  box-shadow: inset 0 0 0 1px oklch(1 0 0 / .09);
  color: oklch(0.85 0.03 75);
}

.dark .scene-list li {
  border-top-color: oklch(1 0 0 / .14);
}

.dark .scene-list li > span {
  color: oklch(0.78 0.13 0);
}

.dark .scene-list b {
  color: oklch(0.95 0.02 80);
}

.scene {
  position: relative;
  height: 696px;
  font-size: 14px;
  line-height: 20px;
  text-align: left;
  color: oklch(0.25 0.03 55);
}

.scene svg,
.scene .iconify {
  flex: none;
}

.side {
  position: absolute;
  top: 0;
  width: 0;
  height: 0;
}

.side-left {
  left: 0;
}

.side-right {
  left: calc(100% - 1344px);
}

.dialog,
.finder,
.terminal,
.cluster {
  position: absolute;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.dialog {
  left: 48px;
  top: 56px;
  z-index: 3;
  width: 380px;
  height: 408px;
  box-sizing: border-box;
  padding: 26px 28px 24px;
  border-radius: 16px;
  background: oklch(0.985 0.008 80);
  font-family: 'Sora', 'Mona Sans', system-ui, sans-serif;
  color: #301c0e;
  text-align: center;
}

.dialog-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  margin: 0 auto;
  border-radius: 14px;
  background: white;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .14), 0 2px 4px oklch(0.25 0.04 55 / .08);
  color: oklch(0.50 0.14 160);
}

.dialog-title {
  margin-top: 14px;
  font-family: var(--font-display);
  font-size: 32px;
  line-height: 38px;
}

.dialog-text {
  margin-top: 6px;
  font-size: 13.5px;
  line-height: 20px;
  color: oklch(0.42 0.03 55);
}

.dialog-list {
  margin-top: 12px;
  text-align: left;
}

.dialog-list > div {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
  font-size: 13.5px;
}

.dialog-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
  width: 26px;
  height: 26px;
  border-radius: 8px;
  background: oklch(0.75 0.18 160 / .16);
  color: oklch(0.43 0.12 160);
}

.dialog-note {
  margin-top: 12px;
  padding: 7px 10px;
  border: 1px dashed oklch(0.25 0.04 55 / .22);
  border-radius: 8px;
  background: oklch(0.25 0.04 55 / .04);
  font-size: 12px;
  line-height: 16px;
  color: oklch(0.42 0.03 55);
  text-wrap: balance;
}

.dialog-foot {
  position: absolute;
  left: 28px;
  right: 28px;
  bottom: 24px;
  height: 38px;
}

.dialog-button {
  position: absolute;
  top: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 156px;
  height: 38px;
  box-sizing: border-box;
  font-size: 13.5px;
  font-weight: 600;
}

.dialog-cancel {
  left: 0;
  box-shadow: inset 0 0 0 1px oklch(0.25 0.04 55 / .18);
}

.dialog-download {
  left: 168px;
  background: oklch(0.75 0.18 160);
  color: oklch(0.22 0.05 160);
}

.dialog-download > span {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.finder {
  left: 396px;
  top: 150px;
  z-index: 2;
  width: 460px;
  height: 368px;
  overflow: hidden;
  border-radius: 12px;
  background: oklch(0.988 0.005 80);
  font-size: 14.5px;
  line-height: 20px;
}

.finder-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 16px 0 48px;
  background: oklch(0.955 0.012 60);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
}

.finder-bar i,
.terminal-bar i {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: oklch(0.80 0.03 40);
}

.finder-bar b {
  margin-left: 10px;
  font-size: 14px;
  font-weight: 600;
}

.finder-head,
.finder-row {
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr);
  gap: 14px;
  align-items: center;
  padding: 0 18px 0 48px;
}

.finder-head {
  height: 30px;
  font-size: 13px;
  color: oklch(0.45 0.04 55);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
}

.finder-row {
  height: 36px;
}

.finder-row:nth-child(even) {
  background: oklch(0.25 0.04 55 / .04);
}

.finder-row > span:first-child {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  white-space: nowrap;
}

.finder-row > span:last-child {
  font-size: 13.5px;
  color: oklch(0.45 0.04 55);
  white-space: nowrap;
}

.finder-gap {
  flex: none;
  width: 12px;
}

.zip-icon {
  color: oklch(0.58 0.20 0);
}

.chip {
  position: absolute;
  left: 0;
  top: 0;
  z-index: 7;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  box-sizing: border-box;
  padding: 0 12px 0 10px;
  border-radius: 9px;
  background: white;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .14), 0 14px 24px -10px oklch(0.25 0.04 55 / .5);
  font-size: 13.5px;
  font-weight: 600;
  white-space: nowrap;
  transform-origin: 0 0;
  transform: translate(444px, 223px);
  opacity: 0;
}

.pointer {
  position: absolute;
  left: 0;
  top: 0;
  z-index: 8;
  width: 22px;
  height: 22px;
  transform-origin: 4px 3px;
  transform: translate(322px, 418px);
  opacity: 0;
}

.terminal {
  left: 796px;
  top: 56px;
  z-index: 4;
  width: 520px;
  overflow: hidden;
  border-radius: 14px;
  background: oklch(0.20 0.03 55);
}

.terminal-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 18px;
  border-bottom: 1px solid oklch(1 0 0 / .08);
  font-size: 13px;
  color: oklch(0.74 0.03 70);
}

.terminal-bar i {
  background: oklch(0.40 0.03 55);
}

.terminal-bar span {
  margin-left: 8px;
}

.terminal-body {
  padding: 16px 22px 22px;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 23px;
  color: oklch(0.90 0.03 80);
}

.terminal-body > div {
  white-space: pre;
}

.typed {
  display: inline-block;
  overflow: hidden;
  white-space: pre;
  vertical-align: bottom;
}

.c-dim {
  color: oklch(0.70 0.03 70);
}

.c-ok {
  color: oklch(0.80 0.16 160);
}

.cluster {
  left: 888px;
  top: 286px;
  z-index: 5;
  width: 470px;
  height: 184px;
  overflow: hidden;
  border-radius: 16px;
  background-color: oklch(0.955 0.018 80);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .16) 1px, transparent 0);
  background-size: 20px 20px;
  font-family: 'Sora', 'Mona Sans', system-ui, sans-serif;
  color: #301c0e;
}

.cluster-head {
  display: flex;
  align-items: center;
  gap: 9px;
  height: 44px;
  padding: 0 20px;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
  font-size: 13.5px;
  font-weight: 500;
}

.cluster-head .iconify {
  color: oklch(0.42 0.03 55);
}

.cluster-head em {
  font-style: normal;
  color: oklch(0.62 0.03 60);
}

.cluster-empty {
  position: absolute;
  left: 0;
  right: 14px;
  top: 98px;
  text-align: center;
  font-size: 13.5px;
  color: oklch(0.42 0.03 55);
  opacity: 0;
}

.node {
  position: absolute;
  top: 64px;
  width: 136px;
  box-sizing: border-box;
  padding: 12px 12px 11px;
  border-radius: 13px;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .14), 0 10px 24px -14px oklch(0.25 0.04 55 / .4);
}

.node b {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
}

.node-database {
  color: oklch(0.50 0.13 250);
}

.node small {
  display: block;
  margin-top: 5px;
  font-size: 11px;
  color: oklch(0.42 0.03 55);
  white-space: nowrap;
}

.node-foot {
  display: grid;
  justify-items: start;
  margin-top: 9px;
  padding-top: 9px;
  border-top: 1px solid oklch(0.90 0.015 78);
}

.pill {
  grid-area: 1 / 1;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 9px;
  border-radius: 999px;
  background: oklch(0.90 0.015 75);
  font-size: 11.5px;
  font-weight: 500;
  color: oklch(0.35 0.03 55);
}

.pill i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.dot-starting {
  background: oklch(0.75 0.15 75);
}

.dot-running {
  background: oklch(0.70 0.19 150);
}

.honest {
  position: absolute;
  left: 888px;
  top: 492px;
  width: 400px;
  font-size: 15px;
  line-height: 22px;
  color: oklch(0.42 0.14 0);
}

.strip {
  position: absolute;
  left: 48px;
  right: 48px;
  top: 580px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.strip-rail {
  height: 3px;
  overflow: hidden;
  border-radius: 2px;
  background: oklch(0.25 0.04 55 / .14);
}

.strip-rail div {
  height: 3px;
  background: oklch(0.58 0.20 0);
  transform-origin: left center;
}

.strip-name {
  margin-top: 11px;
  font-size: 16px;
  font-weight: 600;
  line-height: 22px;
  white-space: nowrap;
}

.strip-note {
  margin-top: 3px;
  max-width: 290px;
  font-size: 14.5px;
  line-height: 21px;
  color: oklch(0.38 0.04 20);
}

.dark .honest {
  color: oklch(0.80 0.11 0);
}

.dark .strip-rail {
  background: oklch(1 0 0 / .14);
}

.dark .strip-rail div {
  background: oklch(0.72 0.19 0);
}

.dark .strip-name {
  color: oklch(0.95 0.02 80);
}

.dark .strip-note {
  color: oklch(0.85 0.03 75);
}

.nlh-dla,
.nlh-ns0,
.nlh-ns1,
.nlh-ns2 {
  opacity: 0;
}

.nlh-ty1 {
  width: 47ch;
}

.nlh-ty2 {
  width: 43ch;
}

.nlh-ty3 {
  width: 44ch;
}

.nlh-press { animation: nlh-press 26s linear infinite both; }
.nlh-dla { animation: nlh-dla 26s linear infinite both; }
.nlh-dlb { animation: nlh-dlb 26s linear infinite both; }
.nlh-f0 { animation: nlh-f0 26s linear infinite both; }
.nlh-f1 { animation: nlh-f1 26s linear infinite both; }
.nlh-f2 { animation: nlh-f2 26s linear infinite both; }
.nlh-f3 { animation: nlh-f3 26s linear infinite both; }
.nlh-f4 { animation: nlh-f4 26s linear infinite both; }
.nlh-f5 { animation: nlh-f5 26s linear infinite both; }
.nlh-f6 { animation: nlh-f6 26s linear infinite both; }
.nlh-f7 { animation: nlh-f7 26s linear infinite both; }
.nlh-zip { animation: nlh-zip 26s linear infinite both; }
.nlh-cur { animation: nlh-cur 26s linear infinite both; }
.nlh-ty1 { animation: nlh-ty1 26s linear infinite both; }
.nlh-t1 { animation: nlh-t1 26s linear infinite both; }
.nlh-t2 { animation: nlh-t2 26s linear infinite both; }
.nlh-t3 { animation: nlh-t3 26s linear infinite both; }
.nlh-ty2 { animation: nlh-ty2 26s linear infinite both; }
.nlh-t4 { animation: nlh-t4 26s linear infinite both; }
.nlh-ty3 { animation: nlh-ty3 26s linear infinite both; }
.nlh-t5 { animation: nlh-t5 26s linear infinite both; }
.nlh-t6 { animation: nlh-t6 26s linear infinite both; }
.nlh-n0 { animation: nlh-n0 26s linear infinite both; }
.nlh-ns0 { animation: nlh-ns0 26s linear infinite both; }
.nlh-nr0 { animation: nlh-nr0 26s linear infinite both; }
.nlh-n1 { animation: nlh-n1 26s linear infinite both; }
.nlh-ns1 { animation: nlh-ns1 26s linear infinite both; }
.nlh-nr1 { animation: nlh-nr1 26s linear infinite both; }
.nlh-n2 { animation: nlh-n2 26s linear infinite both; }
.nlh-ns2 { animation: nlh-ns2 26s linear infinite both; }
.nlh-nr2 { animation: nlh-nr2 26s linear infinite both; }
.nlh-empty { animation: nlh-empty 26s linear infinite both; }
.nlh-honest { animation: nlh-honest 26s linear infinite both; }
.nlh-bar0 { animation: nlh-bar0 26s linear infinite both; }
.nlh-tt0 { animation: nlh-tt0 26s linear infinite both; }
.nlh-nt0 { animation: nlh-nt0 26s linear infinite both; }
.nlh-bar1 { animation: nlh-bar1 26s linear infinite both; }
.nlh-tt1 { animation: nlh-tt1 26s linear infinite both; }
.nlh-nt1 { animation: nlh-nt1 26s linear infinite both; }
.nlh-bar2 { animation: nlh-bar2 26s linear infinite both; }
.nlh-tt2 { animation: nlh-tt2 26s linear infinite both; }
.nlh-nt2 { animation: nlh-nt2 26s linear infinite both; }
.nlh-bar3 { animation: nlh-bar3 26s linear infinite both; }
.nlh-tt3 { animation: nlh-tt3 26s linear infinite both; }
.nlh-nt3 { animation: nlh-nt3 26s linear infinite both; }
.dark .nlh-tt0 { animation-name: nlh-ttd0; }
.dark .nlh-tt1 { animation-name: nlh-ttd1; }
.dark .nlh-tt2 { animation-name: nlh-ttd2; }
.dark .nlh-tt3 { animation-name: nlh-ttd3; }

@keyframes nlh-press { 0% { transform: scale(1); } 9.62% { transform: scale(1); } 10% { transform: scale(.95); } 10.38% { transform: scale(.95); } 10.96% { transform: scale(1); } 100% { transform: scale(1); } }
@keyframes nlh-dla { 0% { opacity: 1; } 11.15% { opacity: 1; } 11.73% { opacity: 0; } 95.77% { opacity: 0; } 95.85% { opacity: 1; } 100% { opacity: 1; } }
@keyframes nlh-dlb { 0% { opacity: 0; } 11.35% { opacity: 0; } 11.92% { opacity: 1; } 95.77% { opacity: 1; } 95.85% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-f0 { 0% { opacity: 0; } 20.96% { opacity: 0; } 21.92% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-f1 { 0% { opacity: 0; } 24.23% { opacity: 0; } 25.19% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-f2 { 0% { opacity: 0; } 25.77% { opacity: 0; } 26.73% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-f3 { 0% { opacity: 0; } 26.92% { opacity: 0; } 27.88% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-f4 { 0% { opacity: 0; } 28.08% { opacity: 0; } 29.04% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-f5 { 0% { opacity: 0; } 29.23% { opacity: 0; } 30.19% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-f6 { 0% { opacity: 0; } 30.38% { opacity: 0; } 31.35% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-f7 { 0% { opacity: 0; } 31.54% { opacity: 0; } 32.5% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-zip { 0% { opacity: 0; transform: translate(250px,388px) scale(.6); } 11.35% { opacity: 0; transform: translate(250px,388px) scale(.6); animation-timing-function: ease-out; } 12.31% { opacity: 1; transform: translate(250px,356px) scale(1); } 17.69% { opacity: 1; transform: translate(250px,356px) scale(1); animation-timing-function: ease-in-out; } 20.96% { opacity: 1; transform: translate(444px,223px) scale(1); } 21.15% { opacity: 1; transform: translate(444px,223px) scale(1); } 21.92% { opacity: 0; transform: translate(444px,223px) scale(1); } 95.77% { opacity: 0; transform: translate(444px,223px) scale(1); } 95.85% { opacity: 0; transform: translate(250px,388px) scale(.6); } 100% { opacity: 0; transform: translate(250px,388px) scale(.6); } }
@keyframes nlh-cur { 0% { opacity: 0; transform: translate(210px,548px); } 3.08% { opacity: 0; transform: translate(210px,548px); } 3.85% { opacity: 1; transform: translate(210px,548px); } 4.62% { opacity: 1; transform: translate(210px,548px); animation-timing-function: ease-in-out; } 9.23% { opacity: 1; transform: translate(322px,418px); } 9.62% { opacity: 1; transform: translate(322px,418px); } 10% { opacity: 1; transform: translate(322px,418px) scale(.84); } 10.38% { opacity: 1; transform: translate(322px,418px) scale(.84); } 10.96% { opacity: 1; transform: translate(322px,418px); } 13.46% { opacity: 1; transform: translate(322px,418px); } 14.42% { opacity: 0; transform: translate(322px,418px); } 95.77% { opacity: 0; transform: translate(322px,418px); } 95.85% { opacity: 0; transform: translate(210px,548px); } 100% { opacity: 0; transform: translate(210px,548px); } }
@keyframes nlh-ty1 { 0% { width: 0ch; } 36.15% { width: 0ch; animation-timing-function: steps(47); } 41.54% { width: 47ch; } 96.54% { width: 47ch; } 96.62% { width: 0ch; } 100% { width: 0ch; } }
@keyframes nlh-t1 { 0% { opacity: 0; } 35.38% { opacity: 0; } 35.96% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-t2 { 0% { opacity: 0; } 43.46% { opacity: 0; } 44.04% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-t3 { 0% { opacity: 0; } 45.38% { opacity: 0; } 45.96% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-ty2 { 0% { width: 0ch; } 48.85% { width: 0ch; animation-timing-function: steps(43); } 53.85% { width: 43ch; } 96.54% { width: 43ch; } 96.62% { width: 0ch; } 100% { width: 0ch; } }
@keyframes nlh-t4 { 0% { opacity: 0; } 48.08% { opacity: 0; } 48.65% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-ty3 { 0% { width: 0ch; } 54.04% { width: 0ch; animation-timing-function: steps(44); } 58.85% { width: 44ch; } 96.54% { width: 44ch; } 96.62% { width: 0ch; } 100% { width: 0ch; } }
@keyframes nlh-t5 { 0% { opacity: 0; } 53.85% { opacity: 0; } 54.42% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-t6 { 0% { opacity: 0; } 61.15% { opacity: 0; } 61.73% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-n0 { 0% { opacity: 0; transform: translateY(8px) scale(.95); } 65.38% { opacity: 0; transform: translateY(8px) scale(.95); animation-timing-function: ease-out; } 66.54% { opacity: 1; transform: translateY(0) scale(1); } 94.62% { opacity: 1; transform: translateY(0) scale(1); } 96.35% { opacity: 0; transform: translateY(8px) scale(.95); } 100% { opacity: 0; transform: translateY(8px) scale(.95); } }
@keyframes nlh-ns0 { 0% { opacity: 1; } 73.08% { opacity: 1; } 73.65% { opacity: 0; } 96.54% { opacity: 0; } 96.62% { opacity: 1; } 100% { opacity: 1; } }
@keyframes nlh-nr0 { 0% { opacity: 0; } 73.08% { opacity: 0; } 73.65% { opacity: 1; } 96.54% { opacity: 1; } 96.62% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-n1 { 0% { opacity: 0; transform: translateY(8px) scale(.95); } 67.31% { opacity: 0; transform: translateY(8px) scale(.95); animation-timing-function: ease-out; } 68.46% { opacity: 1; transform: translateY(0) scale(1); } 94.62% { opacity: 1; transform: translateY(0) scale(1); } 96.35% { opacity: 0; transform: translateY(8px) scale(.95); } 100% { opacity: 0; transform: translateY(8px) scale(.95); } }
@keyframes nlh-ns1 { 0% { opacity: 1; } 75% { opacity: 1; } 75.58% { opacity: 0; } 96.54% { opacity: 0; } 96.62% { opacity: 1; } 100% { opacity: 1; } }
@keyframes nlh-nr1 { 0% { opacity: 0; } 75% { opacity: 0; } 75.58% { opacity: 1; } 96.54% { opacity: 1; } 96.62% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-n2 { 0% { opacity: 0; transform: translateY(8px) scale(.95); } 69.23% { opacity: 0; transform: translateY(8px) scale(.95); animation-timing-function: ease-out; } 70.38% { opacity: 1; transform: translateY(0) scale(1); } 94.62% { opacity: 1; transform: translateY(0) scale(1); } 96.35% { opacity: 0; transform: translateY(8px) scale(.95); } 100% { opacity: 0; transform: translateY(8px) scale(.95); } }
@keyframes nlh-ns2 { 0% { opacity: 1; } 76.92% { opacity: 1; } 77.5% { opacity: 0; } 96.54% { opacity: 0; } 96.62% { opacity: 1; } 100% { opacity: 1; } }
@keyframes nlh-nr2 { 0% { opacity: 0; } 76.92% { opacity: 0; } 77.5% { opacity: 1; } 96.54% { opacity: 1; } 96.62% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-empty { 0% { opacity: 1; } 63.85% { opacity: 1; } 65% { opacity: 0; } 96.35% { opacity: 0; } 97.5% { opacity: 1; } 100% { opacity: 1; } }
@keyframes nlh-honest { 0% { opacity: 0; } 79.23% { opacity: 0; } 80.77% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0; } 100% { opacity: 0; } }
@keyframes nlh-bar0 { 0% { transform: scaleX(0); opacity: 1; } 17.31% { transform: scaleX(1); opacity: 1; } 94.62% { transform: scaleX(1); opacity: 1; } 95.96% { transform: scaleX(1); opacity: 0; } 96.15% { transform: scaleX(1); opacity: 0; } 96.23% { transform: scaleX(0); opacity: 0; } 96.54% { transform: scaleX(0); opacity: 0; } 97.31% { transform: scaleX(0); opacity: 1; } 100% { transform: scaleX(0); opacity: 1; } }
@keyframes nlh-tt0 { 0% { color: oklch(0.50 0.04 20); } 0.77% { color: oklch(0.25 0.03 55); } 94.62% { color: oklch(0.25 0.03 55); } 96.35% { color: oklch(0.50 0.04 20); } 100% { color: oklch(0.50 0.04 20); } }
@keyframes nlh-nt0 { 0% { opacity: 0.45; } 0.38% { opacity: 0.45; } 1.54% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0.45; } 100% { opacity: 0.45; } }
@keyframes nlh-bar1 { 0% { transform: scaleX(0); opacity: 1; } 17.31% { transform: scaleX(0); opacity: 1; } 34.62% { transform: scaleX(1); opacity: 1; } 94.62% { transform: scaleX(1); opacity: 1; } 95.96% { transform: scaleX(1); opacity: 0; } 96.15% { transform: scaleX(1); opacity: 0; } 96.23% { transform: scaleX(0); opacity: 0; } 96.54% { transform: scaleX(0); opacity: 0; } 97.31% { transform: scaleX(0); opacity: 1; } 100% { transform: scaleX(0); opacity: 1; } }
@keyframes nlh-tt1 { 0% { color: oklch(0.50 0.04 20); } 17.31% { color: oklch(0.50 0.04 20); } 18.08% { color: oklch(0.25 0.03 55); } 94.62% { color: oklch(0.25 0.03 55); } 96.35% { color: oklch(0.50 0.04 20); } 100% { color: oklch(0.50 0.04 20); } }
@keyframes nlh-nt1 { 0% { opacity: 0.45; } 17.69% { opacity: 0.45; } 18.85% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0.45; } 100% { opacity: 0.45; } }
@keyframes nlh-bar2 { 0% { transform: scaleX(0); opacity: 1; } 34.62% { transform: scaleX(0); opacity: 1; } 63.46% { transform: scaleX(1); opacity: 1; } 94.62% { transform: scaleX(1); opacity: 1; } 95.96% { transform: scaleX(1); opacity: 0; } 96.15% { transform: scaleX(1); opacity: 0; } 96.23% { transform: scaleX(0); opacity: 0; } 96.54% { transform: scaleX(0); opacity: 0; } 97.31% { transform: scaleX(0); opacity: 1; } 100% { transform: scaleX(0); opacity: 1; } }
@keyframes nlh-tt2 { 0% { color: oklch(0.50 0.04 20); } 34.62% { color: oklch(0.50 0.04 20); } 35.38% { color: oklch(0.25 0.03 55); } 94.62% { color: oklch(0.25 0.03 55); } 96.35% { color: oklch(0.50 0.04 20); } 100% { color: oklch(0.50 0.04 20); } }
@keyframes nlh-nt2 { 0% { opacity: 0.45; } 35% { opacity: 0.45; } 36.15% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0.45; } 100% { opacity: 0.45; } }
@keyframes nlh-bar3 { 0% { transform: scaleX(0); opacity: 1; } 63.46% { transform: scaleX(0); opacity: 1; } 88.46% { transform: scaleX(1); opacity: 1; } 94.62% { transform: scaleX(1); opacity: 1; } 95.96% { transform: scaleX(1); opacity: 0; } 96.15% { transform: scaleX(1); opacity: 0; } 96.23% { transform: scaleX(0); opacity: 0; } 96.54% { transform: scaleX(0); opacity: 0; } 97.31% { transform: scaleX(0); opacity: 1; } 100% { transform: scaleX(0); opacity: 1; } }
@keyframes nlh-tt3 { 0% { color: oklch(0.50 0.04 20); } 63.46% { color: oklch(0.50 0.04 20); } 64.23% { color: oklch(0.25 0.03 55); } 94.62% { color: oklch(0.25 0.03 55); } 96.35% { color: oklch(0.50 0.04 20); } 100% { color: oklch(0.50 0.04 20); } }
@keyframes nlh-nt3 { 0% { opacity: 0.45; } 63.85% { opacity: 0.45; } 65% { opacity: 1; } 94.62% { opacity: 1; } 96.35% { opacity: 0.45; } 100% { opacity: 0.45; } }
@keyframes nlh-ttd0 { 0% { color: oklch(0.70 0.03 65); } 0.77% { color: oklch(0.95 0.02 80); } 94.62% { color: oklch(0.95 0.02 80); } 96.35% { color: oklch(0.70 0.03 65); } 100% { color: oklch(0.70 0.03 65); } }
@keyframes nlh-ttd1 { 0% { color: oklch(0.70 0.03 65); } 17.31% { color: oklch(0.70 0.03 65); } 18.08% { color: oklch(0.95 0.02 80); } 94.62% { color: oklch(0.95 0.02 80); } 96.35% { color: oklch(0.70 0.03 65); } 100% { color: oklch(0.70 0.03 65); } }
@keyframes nlh-ttd2 { 0% { color: oklch(0.70 0.03 65); } 34.62% { color: oklch(0.70 0.03 65); } 35.38% { color: oklch(0.95 0.02 80); } 94.62% { color: oklch(0.95 0.02 80); } 96.35% { color: oklch(0.70 0.03 65); } 100% { color: oklch(0.70 0.03 65); } }
@keyframes nlh-ttd3 { 0% { color: oklch(0.70 0.03 65); } 63.46% { color: oklch(0.70 0.03 65); } 64.23% { color: oklch(0.95 0.02 80); } 94.62% { color: oklch(0.95 0.02 80); } 96.35% { color: oklch(0.70 0.03 65); } 100% { color: oklch(0.70 0.03 65); } }

@media (max-width: 1391px) {
  .scene {
    zoom: .946;
  }
}

@media (max-width: 1319px) {
  .scene {
    zoom: .894;
  }
}

@media (max-width: 1249px) {
  .scene {
    zoom: .842;
  }
}

@media (max-width: 1179px) {
  .scene {
    zoom: .79;
  }
}

@media (max-width: 1109px) {
  .scene {
    zoom: .726;
  }
}

@media (max-width: 1023px) {
  .scene-fit {
    display: none;
  }

  .scene-list {
    display: block;
  }
}

@media (prefers-reduced-motion: reduce) {
  .scene * {
    animation: none !important;
  }
}
</style>
