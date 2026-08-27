<script setup lang="ts">
const props = withDefaults(defineProps<{
  repo?: string;
  label?: string;
}>(), {
  repo: 'zeitlos/vouch',
  label: '',
});

const href = computed(() => `https://github.com/${props.repo}/fork`);
</script>

<template>
  <div class="fork">
    <a
      :href="href"
      target="_blank"
      rel="noopener"
      class="fork-button"
    >
      <UIcon
        name="i-simple-icons-github"
        class="size-5"
      />
      {{ label || `Fork ${repo}` }}
    </a>
    <p class="fork-note">Opens GitHub. You need to be signed in to fork.</p>
  </div>
</template>

<style scoped>
.fork {
  margin: 2rem 0;
}

.fork-button {
  display: inline-flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.875rem 1.5rem;
  border-radius: 0.75rem;
  font-size: 1.0625rem;
  font-weight: 600;
  color: var(--ui-bg);
  background: var(--ui-text-highlighted);
  text-decoration: none;
  transition: transform 0.15s ease, opacity 0.15s ease;
}

.fork-button:hover {
  transform: translateY(-1px);
  opacity: 0.9;
}

.fork-note {
  margin: 0.625rem 0 0;
  font-size: 0.8125rem;
  color: var(--ui-text-muted);
}
</style>
