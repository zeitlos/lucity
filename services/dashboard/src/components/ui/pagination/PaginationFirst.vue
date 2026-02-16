<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { PaginationFirst, type PaginationFirstProps } from 'reka-ui';
import { ChevronsLeft } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const props = defineProps<PaginationFirstProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <PaginationFirst v-bind="props" as-child>
    <Button :class="cn('h-10 w-10 p-0', props.class)" variant="outline">
      <ChevronsLeft class="h-4 w-4" />
      <span class="sr-only">First page</span>
    </Button>
  </PaginationFirst>
</template>
