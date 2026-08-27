<script setup lang="ts">
interface Node {
  label: string;
  sub?: string;
  icon?: string;
  tone?: 'default' | 'primary' | 'muted';
}

withDefaults(defineProps<{
  nodes?: Node[];
  caption?: string;
}>(), {
  nodes: () => [],
  caption: '',
});
</script>

<template>
  <figure class="flow">
    <div class="flow-track">
      <template
        v-for="(node, index) in nodes"
        :key="node.label"
      >
        <div
          class="flow-node"
          :class="`flow-node-${node.tone ?? 'default'}`"
        >
          <UIcon
            v-if="node.icon"
            :name="node.icon"
            class="flow-node-icon size-5"
          />
          <p class="flow-node-label">{{ node.label }}</p>
          <p
            v-if="node.sub"
            class="flow-node-sub"
          >
            {{ node.sub }}
          </p>
        </div>

        <UIcon
          v-if="index < nodes.length - 1"
          name="i-lucide-arrow-right"
          class="flow-arrow size-5"
        />
      </template>
    </div>

    <figcaption v-if="caption">{{ caption }}</figcaption>
  </figure>
</template>

<style scoped>
.flow {
  margin: 2rem 0;
}

.flow-track {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 0.75rem;
}

@media (min-width: 768px) {
  .flow-track {
    flex-direction: row;
    align-items: center;
  }
}

.flow-node {
  flex: 1;
  padding: 1rem 1.125rem;
  border: 1px solid var(--ui-border);
  border-radius: 0.875rem;
  background: var(--ui-bg-elevated);
}

.flow-node-primary {
  border-color: color-mix(in oklab, var(--ui-primary) 40%, transparent);
  background: color-mix(in oklab, var(--ui-primary) 8%, var(--ui-bg-elevated));
}

.flow-node-muted {
  background: transparent;
  border-style: dashed;
}

.flow-node-icon {
  color: var(--ui-text-dimmed);
  margin-bottom: 0.5rem;
}

.flow-node-primary .flow-node-icon {
  color: var(--ui-primary);
}

.flow-node-label {
  margin: 0;
  font-weight: 600;
  font-size: 0.9375rem;
  color: var(--ui-text-highlighted);
}

.flow-node-sub {
  margin: 0.125rem 0 0;
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--ui-text-muted);
}

.flow-arrow {
  flex: none;
  align-self: center;
  color: var(--ui-text-dimmed);
  rotate: 90deg;
}

@media (min-width: 768px) {
  .flow-arrow {
    rotate: none;
  }
}

figcaption {
  margin-top: 0.875rem;
  font-size: 0.875rem;
  color: var(--ui-text-muted);
  text-align: center;
}
</style>
