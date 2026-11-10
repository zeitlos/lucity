<script setup lang="ts">
type FileKey = 'readme' | 'build' | 'chart' | 'production' | 'development';

interface Entry {
  name: string;
  note: string;
  folder?: 'open' | 'closed';
  nested?: boolean;
  file?: FileKey;
}

const entries: Entry[] = [
  { name: 'README.md', note: 'How to install it', file: 'readme' },
  { name: 'build.sh', note: 'Rebuilds your images', file: 'build' },
  { name: 'chart', note: 'How every service runs', folder: 'closed', file: 'chart' },
  { name: 'values', note: 'One file per environment', folder: 'open' },
  { name: 'production.yaml', note: 'One environment, as it runs now', nested: true, file: 'production' },
  { name: 'development.yaml', note: 'And every other environment', nested: true, file: 'development' },
];

const files: { key: FileKey; title: string; text: string }[] = [
  { key: 'readme', title: 'README.md', text: 'The install command for each environment, the list of what your cluster has to provide, and a link to the full guide.' },
  { key: 'build', title: 'build.sh', text: 'Images Lucity built from your code sit in a registry only Lucity can reach. The script clones each repository at the deployed commit, builds it the way Lucity did and pushes it to a registry of yours. It needs Docker, git and yq.' },
  { key: 'chart', title: 'chart/', text: 'The Helm chart every Lucity environment is deployed with. Not a simplified export: the same chart we apply ourselves, templates included.' },
  { key: 'production', title: 'values/production.yaml', text: 'One environment as it runs right now: every service with its image, variables, instances, resources and domains, plus its databases, key-value stores and volumes. The marked lines name things that only exist on Lucity, so you change them.' },
  { key: 'development', title: 'values/development.yaml', text: 'Every environment that has something in it gets its own file. Install the ones you want to keep, each in its own namespace.' },
];

const selected = ref<FileKey>('production');

const folderPath = 'M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.7-.9l-.8-1.2A2 2 0 0 0 7.9 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z';
</script>

<template>
  <div class="explorer">
    <div class="finder" role="group" aria-label="What is in the zip">
      <div class="finder-bar" aria-hidden="true">
        <i /><i /><i />
        <b><svg width="18" height="18" viewBox="0 0 24 24" fill="oklch(0.90 0.06 0)" stroke="oklch(0.58 0.20 0)" stroke-width="1.8" stroke-linejoin="round"><path :d="folderPath" /></svg>shop-ejected</b>
      </div>
      <div class="finder-head" aria-hidden="true"><span>Name</span><span>What it is for</span></div>
      <component
        :is="entry.file ? 'button' : 'div'"
        v-for="entry in entries"
        :key="entry.name"
        class="finder-row"
        :class="{ on: entry.file === selected, static: !entry.file }"
        :type="entry.file ? 'button' : undefined"
        :aria-pressed="entry.file ? entry.file === selected : undefined"
        :aria-controls="entry.file ? `zip-pane-${entry.file}` : undefined"
        @click="entry.file && (selected = entry.file)"
      >
        <span :class="{ 'pl-5.5': entry.nested }">
          <svg v-if="entry.folder" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="oklch(0.50 0.03 55)" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path :d="entry.folder === 'open' ? 'm6 9 6 6 6-6' : 'm9 6 6 6-6 6'" />
          </svg>
          <span v-else class="w-3 shrink-0" />
          <svg v-if="entry.folder" width="18" height="18" viewBox="0 0 24 24" fill="oklch(0.90 0.06 0)" stroke="oklch(0.58 0.20 0)" stroke-width="1.8" stroke-linejoin="round" aria-hidden="true"><path :d="folderPath" /></svg>
          <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="oklch(0.99 0.004 80)" stroke="oklch(0.45 0.04 55)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z" /><path d="M14 2v4a2 2 0 0 0 2 2h4" /></svg>
          {{ entry.name }}
        </span>
        <span>{{ entry.note }}</span>
      </component>
    </div>

    <div class="panes">
      <div
        v-for="file in files"
        :id="`zip-pane-${file.key}`"
        :key="file.key"
        class="pane"
        :class="{ on: file.key === selected }"
        :aria-hidden="file.key !== selected"
      >
        <h3>{{ file.title }}</h3>
        <p>{{ file.text }}</p>
        <div class="code">
          <pre v-if="file.key === 'readme'" :tabindex="file.key === selected ? 0 : -1"><span class="c-acc"># shop</span>

This is a self-contained export of your Lucity project.
It is a standard Helm chart plus one values file per
environment, with no dependency on the Lucity control plane.

<span class="c-acc">## Deploy an environment</span>

helm upgrade --install lucity-app ./chart \
  -f values/production.yaml \
  --namespace shop-production --create-namespace

<span class="c-acc">## What you need to provide</span>

Container images, an image pull secret, a gateway,
the database operator, bucket credentials.</pre>
          <pre v-else-if="file.key === 'build'" :tabindex="file.key === selected ? 0 : -1"><span class="c-dim">$</span> docker login registry.example.com
<span class="c-dim">$</span> ./build.sh production registry.example.com/shop

<span class="c-dim">web      cloned at 9b1e4d0, built, pushed</span>
<span class="c-dim">worker   cloned at 9b1e4d0, built, pushed</span>
<span class="c-ok">values/production.yaml now points at your registry</span></pre>
          <pre v-else-if="file.key === 'chart'" :tabindex="file.key === selected ? 0 : -1">chart/
  Chart.yaml
  values.yaml
  templates/

<span class="c-dim"># the release has to be called lucity-app,</span>
<span class="c-dim"># every resource name derives from it</span>
<span class="c-dim">$</span> helm upgrade --install lucity-app ./chart \
    -f values/production.yaml -n shop-production</pre>
          <pre v-else-if="file.key === 'production'" :tabindex="file.key === selected ? 0 : -1"><span class="c-warm">gateway:</span>                         <span class="c-dim"># yours to change</span>
<span class="c-warm">  name: lucity-gateway</span>
<span class="c-warm">  namespace: lucity-system</span>
<span class="c-warm">imagePullSecrets: [ ... ]</span>           <span class="c-dim"># yours to change</span>
services:
  web:
    image:
      repository: registry.example.com/shop/web
      tag: 9b1e4d0
    replicas: 2
    domains:
      - host: shop.example.com
<span class="c-warm">        issuerRef: { kind: ClusterIssuer, name: ... }</span>  <span class="c-dim"># yours to change</span>
databases:
  postgres:
    main: { instances: 2, size: 10Gi }</pre>
          <pre v-else :tabindex="file.key === selected ? 0 : -1"><span class="c-dim"># each environment is its own release, in its own namespace</span>
<span class="c-dim">$</span> helm upgrade --install lucity-app ./chart \
    -f values/development.yaml \
    -n shop-development --create-namespace</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.explorer {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 32px 48px;
}

.finder {
  flex: 1 1 380px;
  min-width: 0;
  max-width: 520px;
  overflow: hidden;
  border-radius: 12px;
  background: oklch(0.988 0.005 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .14), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
  font-size: 15px;
  color: oklch(0.25 0.03 55);
}

.finder-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 16px;
  background: oklch(0.955 0.012 60);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
}

.finder-bar i {
  flex: none;
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: oklch(0.80 0.03 40);
}

.finder-bar b {
  margin-left: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
}

.finder-head,
.finder-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.2fr);
  gap: 16px;
  align-items: center;
  width: 100%;
  box-sizing: border-box;
  padding: 0 18px;
}

.finder-head {
  height: 32px;
  font-size: 13px;
  color: oklch(0.47 0.04 55);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
}

.finder-row {
  min-height: 46px;
  border: 0;
  border-top: 1px solid oklch(0.25 0.04 55 / .06);
  background: none;
  font: inherit;
  font-size: 15px;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.finder-row.static {
  cursor: default;
}

.finder-row.on {
  background-color: oklch(0.93 0.045 8);
  box-shadow: inset 3px 0 0 oklch(0.58 0.20 0);
}

.finder-row > span:first-child {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  white-space: nowrap;
}

.finder-row > span:last-child {
  overflow: hidden;
  font-size: 14px;
  color: oklch(0.45 0.04 55);
  white-space: nowrap;
  text-overflow: ellipsis;
}

.finder-row svg {
  flex: none;
}

.finder-row:focus-visible {
  outline: 2px solid oklch(0.58 0.20 0);
  outline-offset: -2px;
}

.panes {
  flex: 1.5 1 520px;
  min-width: 0;
  display: grid;
}

.pane {
  grid-area: 1 / 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  visibility: hidden;
}

.pane.on {
  visibility: visible;
}

.pane h3 {
  font-family: var(--font-mono);
  font-size: 19px;
  line-height: 1.2;
  color: oklch(0.25 0.03 55);
}

.pane p {
  max-width: 640px;
  margin-top: 10px;
  font-size: 16.5px;
  line-height: 1.5;
  color: oklch(0.36 0.05 0);
}

.code {
  flex: 1;
  margin-top: 22px;
  overflow: hidden;
  border-radius: 14px;
  background: oklch(0.20 0.03 55);
  box-shadow: 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.code pre {
  margin: 0;
  padding: 20px 24px 24px;
  overflow-x: auto;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.7;
  color: oklch(0.90 0.03 80);
  white-space: pre;
}

.code pre:focus-visible {
  outline: 2px solid oklch(0.58 0.20 0);
  outline-offset: -2px;
}

.c-dim {
  color: oklch(0.70 0.03 70);
}

.c-ok {
  color: oklch(0.80 0.16 160);
}

.c-acc {
  color: oklch(0.80 0.11 300);
}

.c-warm {
  color: oklch(0.84 0.11 70);
}

@media (max-width: 639px) {
  .code pre {
    padding: 16px 18px 18px;
    font-size: 12.5px;
  }
}

@media (max-width: 520px) {
  .finder-head,
  .finder-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .finder-head > span:last-child,
  .finder-row > span:last-child {
    display: none;
  }
}
</style>
