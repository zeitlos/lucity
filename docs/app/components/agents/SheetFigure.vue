<script setup lang="ts">
defineProps<{
  slug: string;
}>();

const restarts = [
  { time: '06:02', text: 'Restarted, out of memory at 512 MiB' },
  { time: '06:04', text: 'Restarted, out of memory' },
  { time: '06:09', text: 'Restarted, out of memory' },
];

const tools = [
  { name: 'list_issues', setting: 'Allowed', tone: 'allowed' },
  { name: 'get_issue_details', setting: 'Allowed', tone: 'allowed' },
  { name: 'update_issue', setting: 'Asks first', tone: 'asks' },
  { name: 'delete_release', setting: 'Hidden', tone: 'hidden' },
];

const model = [
  { key: 'provider', value: 'Mistral, on your key' },
  { key: 'model', value: 'Mistral Large' },
  { key: 'thinking', value: 'Medium' },
  { key: 'daily limit', value: '€20' },
];
</script>

<template>
  <div
    v-if="slug === 'triggers'"
    class="card"
    role="img"
    aria-label="A restart loop on web in production: 40 events, three of them listed and 37 more in the same thread, and the Night watch agent is on it."
  >
    <div class="flex items-center justify-between gap-3 px-5 pt-4 pb-1">
      <span class="mono flex items-center gap-2"><span class="violet"><AgentsIcon name="lucity" :size="16" /></span>web · production</span>
      <span class="mono">40 events</span>
    </div>
    <p class="px-5 pt-0.5 pb-3.5 text-lg font-semibold">Restart loop</p>
    <div v-for="restart in restarts" :key="restart.time" class="line mono-row">
      <span class="muted">{{ restart.time }}</span>
      <span>{{ restart.text }}</span>
    </div>
    <div class="line mono-row muted">and 37 more, in the same thread</div>
    <div class="line flex items-center gap-2.5 bg-[oklch(0.96_0.02_300)] py-3.25">
      <span class="size-2 rounded-full bg-[oklch(0.55_0.15_300)]" />
      <span class="text-[15px] font-semibold text-[oklch(0.42_0.14_300)]">Night watch is on it</span>
    </div>
  </div>

  <div
    v-else-if="slug === 'connectors'"
    class="card"
    role="img"
    aria-label="The Sentry connector, its key held by Lucity: list_issues and get_issue_details are allowed, update_issue asks first, delete_release is hidden."
  >
    <div class="flex items-center justify-between gap-3 px-5 py-4">
      <span class="flex items-center gap-2.5 text-lg font-semibold"><AgentsIcon name="sentry" :size="22" />Sentry</span>
      <span class="text-sm text-[oklch(0.47_0.04_55)]">Key held by Lucity</span>
    </div>
    <div v-for="tool in tools" :key="tool.name" class="line flex items-center justify-between gap-3 py-2.75">
      <span class="font-mono text-sm">{{ tool.name }}</span>
      <span class="flex items-center gap-1.5 text-[15px] font-semibold" :class="`tone-${tool.tone}`">
        {{ tool.setting }}
        <UIcon name="i-lucide-chevron-down" class="size-4 text-[oklch(0.47_0.04_55)]" />
      </span>
    </div>
  </div>

  <div
    v-else-if="slug === 'models'"
    class="card"
    role="img"
    aria-label="The Night watch agent in shop production thinks with Mistral Large on your Mistral key, at medium thinking, with a daily limit of €20, of which €3.40 is spent today."
  >
    <div class="px-6 pt-5 pb-3.5">
      <p class="font-display text-2xl/tight">Night watch</p>
      <p class="mt-1 text-sm text-[oklch(0.47_0.04_55)]">shop · production</p>
    </div>
    <div v-for="line in model" :key="line.key" class="line kv">
      <span class="mono">{{ line.key }}</span>
      <span class="font-semibold">{{ line.value }}</span>
    </div>
    <div class="line spent">
      <div class="kv">
        <span class="mono">spent today</span>
        <span class="font-semibold">€3.40</span>
      </div>
      <div class="mt-2.5 h-1.5 rounded-full bg-[oklch(0.92_0.015_78)]">
        <div class="h-1.5 w-[17%] rounded-full bg-[oklch(0.58_0.17_35)]" />
      </div>
    </div>
  </div>

  <div
    v-else-if="slug === 'channels'"
    class="card"
    role="img"
    aria-label="In the Shop on call chat on Telegram, Night watch asks to raise the memory of web in production to 1 GiB and redeploy, with Approve and Deny buttons. Mia approves at 06:14, and at 06:16 Night watch reports that web runs on 1 GiB and both replicas are ready."
  >
    <div class="flex items-center gap-2.5 border-b border-[oklch(0.88_0.02_78)] px-5 py-3.5">
      <span class="text-[oklch(0.55_0.12_240)]"><AgentsIcon name="telegram" :size="20" /></span>
      <span class="font-semibold">Shop on call</span>
      <span class="text-sm text-[oklch(0.47_0.04_55)]">Restart loop</span>
    </div>
    <div class="flex gap-3 px-5 pt-4.5 pb-1.5">
      <span class="avatar"><AgentsIcon name="lucity" :size="18" /></span>
      <div class="min-w-0">
        <p class="text-[15px]"><b class="font-semibold">Night watch</b> <span class="text-[oklch(0.47_0.04_55)]">06:12</span></p>
        <p class="mt-1 text-[15.5px]/normal text-[oklch(0.40_0.03_55)]">web in production restarted 40 times in ten minutes, out of memory at 512 MiB each time. Raise its memory to 1 GiB and redeploy?</p>
        <div class="mt-3 flex flex-wrap gap-2">
          <span class="flex h-10 items-center rounded-lg bg-[oklch(0.24_0.03_55)] px-4.5 text-[15px] font-semibold text-[oklch(0.99_0.006_80)]">Approve</span>
          <span class="flex h-10 items-center rounded-lg border border-[oklch(0.85_0.03_75)] px-4.5 text-[15px] font-semibold">Deny</span>
        </div>
        <p class="mt-2.5 text-sm text-[oklch(0.47_0.04_55)]">Mia approved at 06:14</p>
      </div>
    </div>
    <div class="flex gap-3 px-5 pt-3 pb-5">
      <span class="avatar"><AgentsIcon name="lucity" :size="18" /></span>
      <div class="min-w-0">
        <p class="text-[15px]"><b class="font-semibold">Night watch</b> <span class="text-[oklch(0.47_0.04_55)]">06:16</span></p>
        <p class="mt-1 text-[15.5px]/normal text-[oklch(0.40_0.03_55)]">Done. web runs on 1 GiB and both replicas are ready.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card {
  box-sizing: border-box;
  width: 100%;
  max-width: 460px;
  overflow: hidden;
  border-radius: 12px;
  background: oklch(0.99 0.006 80);
  color: oklch(0.24 0.03 55);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 18px 36px -22px oklch(0.25 0.04 55 / .45);
}

.line {
  padding-left: 20px;
  padding-right: 20px;
  border-top: 1px solid oklch(0.88 0.02 78);
}

.mono,
.mono-row {
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 13px;
  color: oklch(0.47 0.04 55);
}

.mono-row {
  display: flex;
  gap: 14px;
  padding-top: 9px;
  padding-bottom: 9px;
  font-size: 13.5px;
  color: oklch(0.40 0.03 55);
}

.muted {
  color: oklch(0.47 0.04 55);
}

.violet {
  color: #8843db;
}

.kv {
  display: flex;
  flex-wrap: wrap;
  gap: 2px 18px;
  padding: 10px 24px;
  font-size: 16px;
}

.spent {
  padding: 12px 24px 18px;
}

.spent .kv {
  padding: 0;
}

.kv .mono {
  flex: 0 0 110px;
  padding-top: 2px;
  font-size: 13.5px;
}

.tone-allowed {
  color: oklch(0.55 0.15 160);
}

.tone-asks {
  color: oklch(0.58 0.17 35);
}

.tone-hidden {
  color: oklch(0.47 0.04 55);
}

.avatar {
  display: flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: oklch(0.93 0.04 300);
  color: #8843db;
}
</style>
