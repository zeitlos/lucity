<script setup lang="ts">
const questions = [
  { q: 'Do I have to rewrite my agent?', a: 'No. Pick the harness you use locally, paste its instructions and connect the same tools. The agent behaves the way it did in your terminal, minus the terminal.' },
  { q: 'Which agents can I bring?', a: 'Claude Code, Codex, OpenCode and Pi today, and the list keeps growing. If it runs in a terminal and reads an instructions file, it is a candidate. Tell us which one you are missing.' },
  { q: 'Can an agent approve its own request?', a: 'No. A held call waits outside the sandbox, and only the people your rule names can answer it, from the chat or the dashboard. The agent sees the answer, never the button.' },
  { q: 'What happens when nothing is going on?', a: 'The agent sleeps and costs next to nothing. An event, a schedule or a message wakes it in seconds, with its files and conversations where it left them.' },
  { q: 'Can the agent on my laptop still use Lucity?', a: 'Yes. It drives Lucity through the plugin as before. Hosted agents take the work that should not wait until you are back at your desk.' },
  { q: 'Which models can I use?', a: 'Any provider your workspace connects, from Anthropic and OpenAI to Mistral, Infomaniak or a model you host yourself. Each conversation can pick its own model and effort.' },
  { q: 'Where does it run?', a: 'In the EU, inside your workspace, operated by a Swiss company. Pair it with a European model provider and nothing the agent does leaves Europe.' },
];
</script>

<template>
  <FaqSection
    id="agents-faq"
    title="Before you move an agent over"
    :items="questions"
    ask-label="Ask about agents"
    ask-href="mailto:christian@zeitlos.software?subject=Agents"
  />
</template>
