<script setup lang="ts">
import { useTemplateRef } from 'vue';
import { Check, ChevronDown } from '@lucide/vue';

import { environmentNames, environments } from '@/lib/environment';

defineProps<{
  current: string | null;
}>();

const emit = defineEmits<{
  switch: [name: string];
}>();

const menu = useTemplateRef<HTMLElement>('menu');

const markClass: Record<string, string> = {
  dev: 'bg-environment-dev',
  prod: 'bg-environment-prod',
};

function pick(name: string) {
  menu.value?.hidePopover();
  emit('switch', name);
}
</script>

<template>
  <button
    type="button"
    popovertarget="environment-menu"
    aria-haspopup="menu"
    :disabled="!current"
    class="flex h-8 items-center gap-2 rounded-full border pr-2.5 pl-3 font-medium shadow-[inset_0_1px_0_oklch(1_0_0/0.65)]"
    :class="current === 'prod'
      ? 'border-status-danger/35 bg-linear-to-b from-status-danger-background/40 to-status-danger-background text-status-danger'
      : 'bg-linear-to-b from-card to-background'"
    style="anchor-name: --environment"
  >
    <span class="h-3.5 w-1.5 rounded-[2px]" :class="current ? markClass[current] : 'bg-muted-foreground/40'" aria-hidden="true" />
    {{ current ? environmentNames[current] : 'Connecting' }}
    <ChevronDown class="size-3.5 opacity-70" aria-hidden="true" />
  </button>
  <div
    id="environment-menu"
    ref="menu"
    popover
    role="menu"
    aria-label="Environment"
    class="m-0 mt-2 w-72 rounded-xl border bg-card p-1.5 text-foreground shadow-lg [position-area:bottom_span-left]"
    style="position-anchor: --environment"
  >
    <button
      v-for="name in environments"
      :key="name"
      type="button"
      role="menuitemradio"
      :aria-checked="name === current"
      class="grid w-full grid-cols-[12px_minmax(0,1fr)_16px] items-center gap-3 rounded-lg px-3 py-2 text-left hover:bg-selected"
      :class="name === current ? 'bg-selected' : ''"
      @click="pick(name)"
    >
      <span class="h-4 w-1.5 justify-self-center rounded-[2px]" :class="markClass[name]" aria-hidden="true" />
      <span class="flex min-w-0 flex-col">
        <span class="font-medium">{{ environmentNames[name] }}</span>
        <span class="truncate text-muted-foreground">lucity-{{ name }}</span>
      </span>
      <Check class="size-4" :class="name === current ? '' : 'invisible'" aria-hidden="true" />
    </button>
  </div>
</template>
