<script setup lang="ts">
import { PaginationListItem, type PaginationListItemProps } from 'reka-ui';
import { Button } from '@/components/ui/button';

const props = defineProps<PaginationListItemProps>();
</script>

<template>
  <PaginationListItem v-bind="props" as-child>
    <Button class="h-10 w-10 p-0" :variant="props.value === props.page ? 'outline' : 'ghost'">
      {{ props.value }}
    </Button>
  </PaginationListItem>
</template>
