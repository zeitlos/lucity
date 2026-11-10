<script setup lang="ts">
const access = [
  { what: 'Model: Mistral Large', tone: 'ok', state: 'allowed' },
  { what: 'Lucity: logs, metrics, deploys', tone: 'ok', state: 'allowed' },
  { what: 'Lucity: memory, instances, rollbacks', tone: 'ask', state: 'asks Ops first' },
  { what: 'Lucity: delete anything', tone: 'off', state: 'hidden' },
  { what: 'Sentry: read issues', tone: 'ok', state: 'allowed' },
  { what: 'Linear: create issues', tone: 'ask', state: 'asks first' },
  { what: 'Internet: status.mistral.ai', tone: 'ok', state: 'allowed' },
  { what: 'Network: other services in production', tone: 'off', state: 'off' },
  { what: 'Model spend per day', tone: 'ink', state: 'CHF 20, then it stops' },
];

const record = [
  { time: '03:12:41', kind: 'trigger', what: 'restart loop · web', tone: 'ok', state: 'woke' },
  { time: '03:12:44', kind: 'model', what: 'mistral-large · 4.1k in, 380 out', tone: 'ok', state: 'ok' },
  { time: '03:12:52', kind: 'tool', what: 'lucity · logs, metrics', tone: 'ok', state: 'ok' },
  { time: '03:13:11', kind: 'tool', what: 'sentry · read issues', tone: 'ok', state: 'ok' },
  { time: '03:13:20', kind: 'tool', what: 'lucity · set memory to 1 GB', tone: 'ask', state: 'held' },
  { time: '03:15:02', kind: 'approval', what: 'Noa, on Telegram', tone: 'ok', state: 'allowed' },
  { time: '03:15:03', kind: 'tool', what: 'lucity · set memory to 1 GB', tone: 'ok', state: 'ok' },
  { time: '03:15:44', kind: 'network', what: 'paste.example.org', tone: 'no', state: 'blocked' },
];

const features = [
  { name: 'Grants, one by one', text: 'Allowed, asks first or hidden, down to a single tool of a single connector.' },
  { name: 'Keys stay outside', text: 'Lucity adds the key as a call leaves. The agent never holds one, so it cannot leak one.' },
  { name: 'A yes from the right person', text: 'Risky calls wait. You decide who may answer, and they answer from wherever they are.' },
  { name: 'A record of every call', text: 'Model calls, tool calls and approvals, kept for a year, searchable and exportable.' },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="ctl">
    <LandingChapterHeading id="ctl" beat="grow" promise="Under your control" rest="down to a single tool.">
      Every agent starts with nothing. You hand it each model, tool and host on purpose, the keys never go inside it, and you can read back everything it did.
    </LandingChapterHeading>

    <LandingStage tint="grow" class="mt-14 px-4 pt-9 pb-10 sm:px-9">
      <div class="ct">
        <div class="ct-acc">
          <p class="ct-title">Night watch · access</p>
          <ul class="list">
            <li v-for="row in access" :key="row.what">
              <span>{{ row.what }}</span>
              <em :class="row.tone">{{ row.state }}</em>
            </li>
          </ul>
        </div>

        <div class="ct-rec" role="img" aria-label="The record of Night watch on one night: a restart loop wakes it, a model call, tool calls that read, a memory change that is held until Noa allows it on Telegram, and a host that is blocked.">
          <div class="ct-bar"><span>Record</span><small>Night watch · last night</small></div>
          <div class="log">
            <div v-for="row in record" :key="row.time" class="row">
              <span class="t">{{ row.time }}</span>
              <span class="kd">{{ row.kind }}</span>
              <span class="w">{{ row.what }}</span>
              <span class="st" :class="row.tone">{{ row.state }}</span>
            </div>
          </div>
        </div>

        <div class="ct-ask">
          <p class="ask-meta">Shop on call · 03:13</p>
          <p class="ask-title">Night watch wants to set the memory of web to 1 GB.</p>
          <p class="ask-text">Changes to production need a yes from someone in Ops. The agent cannot answer for itself.</p>
          <div class="ask-acts"><span class="bb bb-primary">Approve</span><span class="bb">Deny</span></div>
        </div>
      </div>
    </LandingStage>

    <div class="mt-10 grid grid-cols-[repeat(auto-fit,minmax(min(250px,100%),1fr))] gap-x-10 border-b border-stone-300 dark:border-stone-800">
      <div v-for="feature in features" :key="feature.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
        <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ feature.name }}</b>
        <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ feature.text }}</span>
      </div>
    </div>

    <p class="mt-8 text-base text-stone-700 dark:text-stone-400">
      Also included: a list of hosts each agent may reach, a daily spend limit, step by step replay of any run, and export of the record to your own bucket.
      <NuxtLink to="/security#agents" class="font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">How agent security works</NuxtLink>
    </p>
  </section>
</template>

<style scoped>
.ct {
  position: relative;
  height: 606px;
}

.ct-acc {
  position: absolute;
  left: 0;
  top: 40px;
  width: 40%;
  z-index: 2;
  box-sizing: border-box;
  padding: 24px 28px 12px;
  border-radius: 14px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.ct-title {
  margin: 0 0 12px;
  font-size: 17px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.list li {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 0;
  border-top: 1px solid oklch(0.91 0.015 80);
  font-size: 16px;
  color: oklch(0.25 0.03 55);
}

.list em {
  font-size: 14.5px;
  font-style: normal;
  font-weight: 600;
  white-space: nowrap;
}

.list .ok {
  color: oklch(0.43 0.12 160);
}

.list .ask {
  color: oklch(0.48 0.11 70);
}

.list .off {
  color: oklch(0.47 0.04 55);
}

.list .ink {
  color: oklch(0.25 0.03 55);
}

.ct-rec {
  position: absolute;
  left: 35%;
  right: -60px;
  top: 0;
  height: 432px;
  border-radius: 18px 0 0 18px;
  overflow: hidden;
  background: oklch(0.20 0.03 55);
  box-shadow: 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
  color: oklch(0.88 0.02 80);
}

.ct-bar {
  display: flex;
  align-items: center;
  gap: 14px;
  height: 56px;
  padding: 0 96px 0 calc(5% + 56px);
  border-bottom: 1px solid oklch(0.30 0.03 55);
  font-size: 15.5px;
  font-weight: 600;
  color: white;
}

.ct-bar small {
  font-size: 14px;
  font-weight: 400;
  color: oklch(0.72 0.03 70);
}

.log {
  padding: 6px 96px 0 calc(5% + 56px);
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 13.5px;
  line-height: 1.5;
}

.log .row {
  display: grid;
  grid-template-columns: 76px 78px minmax(0, 1fr) auto;
  gap: 4px 14px;
  padding: 9px 0;
  border-top: 1px solid oklch(0.30 0.03 55);
  white-space: nowrap;
}

.log .row:first-child {
  border-top: 0;
}

.log .t {
  color: oklch(0.72 0.03 70);
}

.log .kd {
  color: oklch(0.80 0.10 300);
}

.log .ok {
  color: oklch(0.80 0.14 160);
}

.log .ask {
  color: oklch(0.82 0.12 80);
}

.log .no {
  color: oklch(0.76 0.15 27);
}

.ct-ask {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 420px;
  z-index: 3;
  box-sizing: border-box;
  padding: 20px 22px;
  border-radius: 18px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 30px 50px -20px oklch(0.15 0.04 55 / .6);
}

.ask-meta {
  margin: 0;
  font-size: 14px;
  color: oklch(0.47 0.04 55);
}

.ask-title {
  margin: 6px 0 0;
  font-size: 17px;
  font-weight: 600;
  line-height: 1.4;
  color: oklch(0.25 0.03 55);
  text-wrap: balance;
}

.ask-text {
  margin: 6px 0 0;
  font-size: 15px;
  line-height: 1.45;
  color: oklch(0.45 0.04 55);
}

.ask-acts {
  display: flex;
  gap: 10px;
  margin-top: 14px;
}

.bb {
  display: inline-flex;
  align-items: center;
  min-height: 42px;
  padding: 0 18px;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
  white-space: nowrap;
  box-shadow: inset 0 0 0 1px oklch(0.85 0.03 75);
}

.bb-primary {
  background: oklch(0.25 0.03 55);
  color: oklch(0.95 0.02 80);
  box-shadow: none;
}

@media (max-width: 1000px) {
  .ct {
    display: flex;
    flex-direction: column;
    gap: 20px;
    height: auto;
  }

  .ct-acc,
  .ct-rec,
  .ct-ask {
    position: static;
    width: auto;
  }

  .ct-rec {
    height: auto;
    margin: 0 -16px;
    border-radius: 18px;
  }

  .ct-bar,
  .log {
    padding-right: 20px;
    padding-left: 20px;
  }

  .log {
    padding-bottom: 10px;
    overflow-x: auto;
  }
}

@media (max-width: 640px) {
  .ct-acc {
    padding: 20px 20px 8px;
  }

  .log {
    font-size: 13px;
  }

  .log .row {
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 2px 12px;
    padding: 10px 0;
    white-space: normal;
  }

  .log .t {
    grid-area: 1 / 1;
  }

  .log .kd {
    grid-area: 1 / 2;
  }

  .log .st {
    grid-area: 1 / 3;
  }

  .log .w {
    grid-area: 2 / 1 / auto / -1;
  }
}
</style>
