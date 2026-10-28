<script setup lang="ts">
import { computed } from 'vue';
import { useIntervalFn, useNow } from '@vueuse/core';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { darkTooltipClass } from '@/lib/tooltip';
import { formatExactTime, formatRelative } from '@/lib/format';

const props = defineProps<{ value: string | number | Date }>();

const now = useNow({ scheduler: cb => useIntervalFn(cb, 30_000) });
const date = computed(() => new Date(props.value));
const relative = computed(() => formatRelative(date.value, now.value.getTime()));
</script>

<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <time
        :datetime="date.toISOString()"
        class="underline decoration-muted-foreground/40 decoration-dotted underline-offset-2"
      >{{ relative }}</time>
    </TooltipTrigger>
    <TooltipContent :class="darkTooltipClass">
      {{ formatExactTime(date) }}
    </TooltipContent>
  </Tooltip>
</template>
