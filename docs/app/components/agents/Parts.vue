<script setup lang="ts">
interface Part {
  name: string;
  text: string;
  ticks: string[];
  link: { to: string; label: string };
  tiles: { icon: string; label: string }[];
  lines: number[];
}

const parts: Part[] = [
  {
    name: 'Harness and instructions',
    text: 'Pick the agent you already run in your terminal and paste its instructions file. That is the whole move.',
    ticks: ['Open harnesses, pinned and kept up to date for you', 'One plain text file of instructions, the one you already have', 'Skills and subagents come along', 'Export an agent and run it anywhere a container runs'],
    link: { to: '/docs/agents', label: 'How an agent is put together' },
    tiles: [
      { icon: 'claudecode', label: 'Claude Code' },
      { icon: 'openai', label: 'Codex' },
      { icon: 'opencode', label: 'OpenCode' },
      { icon: 'pi', label: 'Pi' },
      { icon: 'plus', label: 'Yours next' },
    ],
    lines: [0, 1],
  },
  {
    name: 'Triggers',
    text: 'An event, a schedule or a message starts a conversation. Events about the same thing land in the same one, so a flapping service is one thread, not forty.',
    ticks: ['From Lucity: failed deploys, crash loops, full disks, new vulnerabilities', 'From your tools: alerts, new issues, new errors', 'Any schedule, in your time zone', "A signed webhook, or an email to the agent's own address"],
    link: { to: '/docs/agents/triggers', label: 'Every trigger' },
    tiles: [
      { icon: 'lucity', label: 'Lucity events' },
      { icon: 'clock', label: 'Schedules' },
      { icon: 'webhook', label: 'Webhooks' },
      { icon: 'github', label: 'GitHub' },
      { icon: 'sentry', label: 'Sentry' },
      { icon: 'grafana', label: 'Grafana' },
      { icon: 'mail', label: 'Email' },
    ],
    lines: [2],
  },
  {
    name: 'Connectors',
    text: 'Any remote MCP server becomes a connector, and so does your own API. The agent calls a tool, and Lucity adds the key on the way out.',
    ticks: ['Lucity itself: logs, metrics, deploys, rollbacks, variables', 'GitHub, Linear, Jira, Sentry, Grafana and Notion in one click', 'Your own API, from its OpenAPI file', 'Allowed, asks first or hidden, tool by tool'],
    link: { to: '/docs/agents/connectors', label: 'Every connector' },
    tiles: [
      { icon: 'lucity', label: 'Lucity' },
      { icon: 'github', label: 'GitHub' },
      { icon: 'linear', label: 'Linear' },
      { icon: 'jira', label: 'Jira' },
      { icon: 'sentry', label: 'Sentry' },
      { icon: 'grafana', label: 'Grafana' },
      { icon: 'notion', label: 'Notion' },
      { icon: 'modelcontextprotocol', label: 'Any MCP server' },
    ],
    lines: [3],
  },
  {
    name: 'Model providers',
    text: 'Connect the providers you already pay for. Each conversation picks its model and how hard it thinks, and the bill stays with your provider.',
    ticks: ['No markup on tokens, ever', 'European providers for work that has to stay in Europe', 'Any OpenAI-compatible API, including models you host yourself', 'A daily spend limit per agent'],
    link: { to: '/docs/agents/models', label: 'Every model provider' },
    tiles: [
      { icon: 'anthropic', label: 'Anthropic' },
      { icon: 'openai', label: 'OpenAI' },
      { icon: 'mistralai', label: 'Mistral' },
      { icon: 'infomaniak', label: 'Infomaniak' },
      { icon: 'scaleway', label: 'Scaleway' },
      { icon: 'server', label: 'Your own' },
    ],
    lines: [4],
  },
  {
    name: 'Channels',
    text: 'An agent answers in the chats you add it to. Teammates link their account once, so it knows who is asking, and who said yes.',
    ticks: ['A thread or topic per conversation', 'Approvals as buttons, right in the chat', 'The dashboard and the API for everything else'],
    link: { to: '/docs/agents/channels', label: 'Every channel' },
    tiles: [
      { icon: 'telegram', label: 'Telegram' },
      { icon: 'mattermost', label: 'Mattermost' },
      { icon: 'slack', label: 'Slack' },
      { icon: 'microsoftteams', label: 'Teams' },
      { icon: 'mail', label: 'Email' },
      { icon: 'layout-dashboard', label: 'Dashboard' },
    ],
    lines: [5],
  },
];

const sheet = [
  { key: 'harness', value: 'OpenCode' },
  { key: 'instructions', value: 'AGENTS.md, 14 lines' },
  { key: 'wakes on', value: 'Restart loops, out of memory' },
  { key: 'may use', value: 'Lucity, Sentry, Linear' },
  { key: 'thinks with', value: 'Mistral Large, on your key' },
  { key: 'talks in', value: 'Shop on call, on Telegram' },
  { key: 'lives in', value: 'shop · production' },
];

const open = ref(0);
const current = computed(() => parts[open.value]!);
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="parts">
    <LandingChapterHeading id="parts" beat="keep" promise="Built from parts" rest="not from a framework.">
      An agent is a harness you know, instructions in plain words and four kinds of parts. Connect a part once, and every agent in the workspace can use it when you allow it.
    </LandingChapterHeading>

    <div class="mt-14 flex flex-wrap items-start gap-x-16 gap-y-10">
      <div class="min-w-0 flex-[1_1_400px] border-b border-stone-300 dark:border-stone-800">
        <div v-for="(part, index) in parts" :key="part.name" class="border-t border-stone-300 dark:border-stone-800">
          <h3>
            <button
              :id="`agent-part-${index}`"
              type="button"
              class="row-button flex min-h-11 w-full cursor-pointer items-center justify-between gap-6 py-4.75 text-left text-xl font-semibold transition-colors hover:text-stone-900 dark:hover:text-stone-100"
              :class="open === index ? 'text-stone-900 dark:text-stone-100' : 'text-stone-700 dark:text-stone-400'"
              :aria-expanded="open === index"
              :aria-controls="`agent-part-panel-${index}`"
              @click="open = index"
            >
              <span>{{ part.name }}</span>
              <UIcon name="i-lucide-chevron-down" class="size-5 shrink-0 transition-transform duration-300 ease-out motion-reduce:transition-none" :class="{ 'rotate-180': open === index }" />
            </button>
          </h3>
          <div :id="`agent-part-panel-${index}`" role="region" :aria-labelledby="`agent-part-${index}`" :inert="open !== index" class="part-panel" :class="{ 'is-open': open === index }">
            <div class="min-h-0 overflow-hidden">
              <div class="part-body pb-6 sm:pr-10">
                <p class="text-lg/relaxed text-stone-700 dark:text-stone-400">{{ part.text }}</p>
                <ul class="ticks mt-3.5 flex flex-col gap-1.75 text-base text-stone-800 dark:text-stone-300">
                  <li v-for="tick in part.ticks" :key="tick" class="flex items-baseline gap-2.5">{{ tick }}</li>
                </ul>
                <p class="mt-4 text-base">
                  <NuxtLink :to="part.link.to" class="font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">{{ part.link.label }}</NuxtLink>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <LandingStage tint="keep" class="min-w-0 flex-[1.55_1_560px] px-4 sm:px-8">
        <div class="scene">
          <div class="tiles">
            <div v-for="tile in current.tiles" :key="tile.label" class="tile">
              <span :class="{ violet: tile.icon === 'lucity' }"><AgentsIcon :name="tile.icon" :size="30" /></span>
              <small>{{ tile.label }}</small>
            </div>
          </div>
          <div class="rise" role="img" aria-label="Everything the Night watch agent is, in seven lines: the OpenCode harness, an AGENTS.md of 14 lines, wakes on restart loops and out of memory, may use Lucity, Sentry and Linear, thinks with Mistral Large on your key, talks in Shop on call on Telegram, lives in shop production.">
            <p class="font-display text-2xl/tight text-stone-900">Night watch</p>
            <p class="rise-sub">Everything it is, in seven lines</p>
            <div v-for="(line, index) in sheet" :key="line.key" class="ln" :class="{ lit: current.lines.includes(index) }">
              <span>{{ line.key }}</span>
              <span>{{ line.value }}</span>
            </div>
          </div>
        </div>
      </LandingStage>
    </div>
  </section>
</template>

<style scoped>
.part-panel {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.part-panel.is-open {
  grid-template-rows: 1fr;
}

.part-body {
  opacity: 0;
  translate: 0 -0.375rem;
  transition:
    opacity 0.2s ease,
    translate 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.is-open .part-body {
  opacity: 1;
  translate: none;
  transition:
    opacity 0.3s ease 0.08s,
    translate 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

@media (prefers-reduced-motion: reduce) {
  .part-panel,
  .part-body,
  .is-open .part-body {
    transition: none;
  }
}

.row-button:focus-visible {
  outline: 2px solid oklch(0.55 0.15 300);
  outline-offset: 2px;
}

.ticks li::before {
  content: "";
  flex: none;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: oklch(0.58 0.20 0);
  transform: translateY(-2px);
}

.scene {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  gap: 36px;
  min-height: 600px;
}

.tiles {
  display: flex;
  flex-wrap: wrap;
  align-content: center;
  justify-content: center;
  gap: 26px 22px;
  max-width: 640px;
  min-height: 226px;
  padding-top: 36px;
}

.tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  width: 104px;
  text-align: center;
}

.tile span {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  border-radius: 18px;
  background: oklch(0.99 0.006 80);
  color: oklch(0.27 0.03 55);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 12px 22px -14px oklch(0.25 0.04 55 / .45);
}

.tile span.violet {
  color: #8843db;
}

.tile small {
  font-size: 14.5px;
  font-weight: 600;
  line-height: 1.25;
  color: oklch(0.25 0.03 55);
}

.dark .tile small {
  color: oklch(0.95 0.02 80);
}

.rise {
  box-sizing: border-box;
  width: 100%;
  max-width: 640px;
  padding: 26px 32px 20px;
  border-radius: 14px 14px 0 0;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.rise-sub {
  margin: 4px 0 16px;
  font-size: 15px;
  color: oklch(0.47 0.04 55);
}

.ln {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 4px 18px;
  margin: 0 -32px;
  padding: 11px 32px;
  border-top: 1px solid oklch(0.91 0.015 80);
  font-size: 16px;
  transition: background-color .25s;
}

.ln.lit {
  background-color: oklch(0.93 0.045 8);
}

.ln span:first-child {
  padding-top: 2px;
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 13.5px;
  color: oklch(0.47 0.04 55);
}

.ln span:last-child {
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

@media (max-width: 767px) {
  .scene {
    gap: 28px;
    min-height: 0;
  }

  .tiles {
    min-height: 0;
    padding-top: 28px;
  }
}

@media (max-width: 560px) {
  .tiles {
    gap: 18px 8px;
  }

  .tile {
    width: 86px;
  }

  .tile span {
    width: 60px;
    height: 60px;
    border-radius: 16px;
  }

  .tile small {
    font-size: 13.5px;
  }

  .rise {
    padding: 22px 20px 14px;
  }

  .ln {
    grid-template-columns: minmax(0, 1fr);
    margin: 0 -20px;
    padding: 10px 20px;
  }
}
</style>
