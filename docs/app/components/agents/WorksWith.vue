<script setup lang="ts">
const works = [
  { icon: 'claudecode', label: 'Claude Code' },
  { icon: 'openai', label: 'Codex' },
  { icon: 'opencode', label: 'OpenCode' },
  { icon: 'pi', label: 'Pi' },
  { icon: 'anthropic', label: 'Anthropic' },
  { icon: 'mistralai', label: 'Mistral' },
  { icon: 'infomaniak', label: 'Infomaniak' },
  { icon: 'scaleway', label: 'Scaleway' },
  { icon: 'github', label: 'GitHub' },
  { icon: 'linear', label: 'Linear' },
  { icon: 'jira', label: 'Jira' },
  { icon: 'sentry', label: 'Sentry' },
  { icon: 'grafana', label: 'Grafana' },
  { icon: 'notion', label: 'Notion' },
  { icon: 'telegram', label: 'Telegram' },
  { icon: 'mattermost', label: 'Mattermost' },
  { icon: 'slack', label: 'Slack' },
  { icon: 'modelcontextprotocol', label: 'Any MCP server' },
];

const label = works.map(item => item.label).join(', ');
</script>

<template>
  <section aria-label="Works with">
    <p class="px-6 text-center text-lg text-stone-700 dark:text-stone-400">
      Works with the agents, models and tools you already use. No framework to adopt.
    </p>
    <div class="marquee mt-9" role="img" :aria-label="label">
      <div class="marquee-track">
        <div v-for="copy in 2" :key="copy" class="marquee-group" :aria-hidden="copy === 2 ? 'true' : undefined">
          <span
            v-for="item in works"
            :key="item.label"
            class="flex items-center gap-3 text-xl font-semibold whitespace-nowrap text-stone-800 dark:text-stone-200"
          >
            <AgentsIcon :name="item.icon" :size="28" />
            {{ item.label }}
          </span>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.marquee {
  overflow: hidden;
  mask-image: linear-gradient(to right, transparent, #000 7%, #000 93%, transparent);
}

.marquee-track {
  display: flex;
  width: max-content;
  animation: marquee 80s linear infinite;
}

.marquee:hover .marquee-track {
  animation-play-state: paused;
}

.marquee-group {
  display: flex;
  gap: 60px;
  padding-right: 60px;
}

@keyframes marquee {
  to {
    transform: translateX(-50%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .marquee-track {
    animation: none;
  }
}
</style>
