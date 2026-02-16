<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { PaginationLast, type PaginationLastProps } from 'reka-ui';
import { ChevronsRight } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const props = defineProps<PaginationLastProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <PaginationLast v-bind="props" as-child>
    <Button :class="cn('h-10 w-10 p-0', props.class)" variant="outline">
      <ChevronsRight class="h-4 w-4" />
      <span class="sr-only">Last page</span>
    </Button>
  </PaginationLast>
</template>
