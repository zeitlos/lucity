<script setup lang="ts">
import ResourceIcon from '@/components/ResourceIcon.vue';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { ResourceKind } from '@/lib/resources';
import { darkTooltipClass } from '@/lib/tooltip';

defineProps<{
  kind: ResourceKind;
  name: string;
}>();
</script>

<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <span class="flex size-6 items-center justify-center rounded-[6px] border bg-card text-muted-foreground shadow-[0_1px_2px_oklch(0.5_0.02_55/0.06)]">
        <ResourceIcon :kind="kind" :size="12" />
      </span>
    </TooltipTrigger>
    <TooltipContent side="top" :class="darkTooltipClass">{{ name }}</TooltipContent>
  </Tooltip>
</template>
