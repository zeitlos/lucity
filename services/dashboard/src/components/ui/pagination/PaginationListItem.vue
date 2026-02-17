<script setup lang="ts">
import { PaginationListItem, type PaginationListItemProps, injectPaginationRootContext } from 'reka-ui';
import { Button } from '@/components/ui/button';

const props = defineProps<PaginationListItemProps>();
const rootContext = injectPaginationRootContext();
</script>

<template>
  <PaginationListItem v-bind="props" as-child>
    <Button class="h-10 w-10 p-0" :variant="props.value === rootContext.page.value ? 'outline' : 'ghost'">
      {{ props.value }}
    </Button>
  </PaginationListItem>
</template>
