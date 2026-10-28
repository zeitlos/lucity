<script setup lang="ts">
import { computed, ref, type Component } from 'vue';
import type { RouteLocationRaw } from 'vue-router';
import { CornerLeftUp, Download } from '@lucide/vue';
import ProjectEjectDialog from '@/components/ProjectEjectDialog.vue';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { darkTooltipClass } from '@/lib/tooltip';

interface RailItem {
  key: string;
  label: string;
  icon: Component;
  to: RouteLocationRaw;
}

const props = defineProps<{
  items: RailItem[];
  activeKey: string;
  upLabel?: string;
  upTo?: RouteLocationRaw;
  projectId?: string;
  projectName?: string;
}>();

const ITEM_PITCH = 52;

const activeIndex = computed(() => props.items.findIndex(item => item.key === props.activeKey));
const tabTop = computed(() => `${activeIndex.value * ITEM_PITCH - 12}px`);

const ejectOpen = ref(false);

</script>

<template>
  <div class="relative z-20 w-13 shrink-0">
    <svg
      v-if="activeIndex === 0"
      class="pointer-events-none absolute left-0 top-0 overflow-visible"
      width="67"
      height="60"
      viewBox="0 0 67 60"
      aria-hidden="true"
    >
      <path
        class="fill-card"
        d="M12.5 0.5 L67 0.5 L67 15 L53 15 L53 59.5 L52.5 59.5 A12 12 0 0 0 40.5 47.5 L12.5 47.5 A12 12 0 0 1 0.5 35.5 L0.5 12.5 A12 12 0 0 1 12.5 0.5 Z"
      />
      <path
        class="fill-none stroke-border"
        d="M66.5 0.5 L12.5 0.5 A12 12 0 0 0 0.5 12.5 L0.5 35.5 A12 12 0 0 0 12.5 47.5 L40.5 47.5 A12 12 0 0 1 52.5 59.5"
      />
    </svg>
    <svg
      v-else-if="activeIndex > 0"
      class="pointer-events-none absolute left-0 overflow-visible"
      :style="{ top: tabTop }"
      width="53"
      height="72"
      viewBox="0 0 53 72"
      aria-hidden="true"
    >
      <path
        class="fill-card"
        d="M52.5 0.5 A12 12 0 0 1 40.5 12.5 L12.5 12.5 A12 12 0 0 0 0.5 24.5 L0.5 47.5 A12 12 0 0 0 12.5 59.5 L40.5 59.5 A12 12 0 0 1 52.5 71.5 L53 71.5 L53 0.5 Z"
      />
      <path
        class="fill-none stroke-border"
        d="M52.5 0.5 A12 12 0 0 1 40.5 12.5 L12.5 12.5 A12 12 0 0 0 0.5 24.5 L0.5 47.5 A12 12 0 0 0 12.5 59.5 L40.5 59.5 A12 12 0 0 1 52.5 71.5"
      />
    </svg>

      <nav class="relative flex h-full w-12 flex-col gap-1">
        <Tooltip v-for="item in items" :key="item.key">
          <TooltipTrigger as-child>
            <RouterLink
              :to="item.to"
              :aria-label="item.label"
              :aria-current="item.key === activeKey ? 'page' : undefined"
              class="flex size-12 items-center justify-center rounded-md transition-colors"
              :class="item.key === activeKey
                ? 'text-foreground'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
            >
              <component :is="item.icon" :size="18" />
            </RouterLink>
          </TooltipTrigger>
          <TooltipContent side="right" :class="darkTooltipClass">{{ item.label }}</TooltipContent>
        </Tooltip>

        <div v-if="projectId" class="flex h-3 items-center justify-center" aria-hidden="true">
          <div class="h-px w-6 bg-border" />
        </div>

        <Tooltip v-if="projectId">
          <TooltipTrigger as-child>
            <button
              type="button"
              aria-label="Eject"
              class="flex size-12 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
              @click="ejectOpen = true"
            >
              <Download :size="18" />
            </button>
          </TooltipTrigger>
          <TooltipContent side="right" :class="darkTooltipClass">Eject</TooltipContent>
        </Tooltip>

        <div class="flex-1" />

        <Tooltip v-if="upTo">
          <TooltipTrigger as-child>
            <RouterLink
              :to="upTo"
              :aria-label="upLabel"
              class="flex size-12 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            >
              <CornerLeftUp :size="18" />
            </RouterLink>
          </TooltipTrigger>
          <TooltipContent side="right" :class="darkTooltipClass">{{ upLabel }}</TooltipContent>
        </Tooltip>
      </nav>

    <ProjectEjectDialog
      v-if="projectId"
      v-model:open="ejectOpen"
      :project-id="projectId"
      :project-name="projectName ?? projectId"
    />
  </div>
</template>
