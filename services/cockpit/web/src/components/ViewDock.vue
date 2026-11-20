<script setup lang="ts">
import { Layers, Server } from '@lucide/vue';

import { dock, views } from '@/lib/dock';

const icons = {
  workspaces: Layers,
  platform: Server,
};
</script>

<template>
  <Transition
    enter-from-class="translate-y-4 opacity-0"
    leave-to-class="translate-y-4 opacity-0"
  >
    <nav
      v-if="!dock.hidden"
      aria-label="Views"
      class="fixed bottom-11 z-40 flex -translate-x-1/2 items-center gap-1 rounded-3xl border border-white/8 bg-linear-to-b from-[oklch(0.31_0.025_55/0.95)] to-ink/95 p-1.5 shadow-[inset_0_1px_0_oklch(1_0_0/0.14),0_2px_6px_oklch(0.2_0.03_55/0.3),0_22px_44px_-14px_oklch(0.2_0.03_55/0.65)] backdrop-blur-md transition-[left,translate,opacity] duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] motion-reduce:transition-none"
      :style="{ left: `calc(50% - ${dock.right / 2}px)` }"
    >
      <button
        v-for="(view, index) in views"
        :key="view.key"
        type="button"
        :aria-label="view.name"
        :aria-current="dock.view === view.key ? 'page' : undefined"
        class="group relative flex h-11 min-w-11 items-center justify-center gap-2 rounded-2xl font-semibold transition-colors"
        :class="dock.view === view.key
          ? 'bg-linear-to-b from-white/16 to-white/8 pr-4 pl-3.5 text-ink-foreground shadow-[inset_0_1px_0_oklch(1_0_0/0.18)]'
          : 'text-ink-muted-foreground hover:bg-white/7 hover:text-ink-foreground'"
        @click="dock.view = view.key"
      >
        <component :is="icons[view.key]" class="size-5" :stroke-width="1.8" aria-hidden="true" />
        <span v-if="dock.view === view.key">{{ view.name }}</span>
        <span
          v-else
          class="pointer-events-none absolute bottom-full mb-3 rounded-lg bg-ink px-2.5 py-1 font-normal whitespace-nowrap text-ink-foreground opacity-0 shadow-lg transition-opacity group-hover:opacity-100"
          aria-hidden="true"
        >
          {{ view.name }} <span class="text-ink-muted-foreground">⌘{{ index + 1 }}</span>
        </span>
      </button>
    </nav>
  </Transition>
</template>
