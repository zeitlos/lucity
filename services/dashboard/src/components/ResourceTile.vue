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
      <span class="flex size-6 items-center justify-center rounded-sm border bg-card text-muted-foreground shadow-2xs">
        <ResourceIcon :kind="kind" :size="12" />
      </span>
    </TooltipTrigger>
    <TooltipContent side="top" :class="darkTooltipClass">{{ name }}</TooltipContent>
  </Tooltip>
</template>
