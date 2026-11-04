<script setup lang="ts" generic="Item extends { id: string; kind?: ResourceKind; name: string; detail?: string; to?: RouteLocationRaw }">
import { RouterLink, type RouteLocationRaw } from 'vue-router';
import { ChevronRight } from '@lucide/vue';
import ResourceIcon from '@/components/ResourceIcon.vue';
import type { ResourceKind } from '@/lib/resources';

defineProps<{
  items: Item[];
}>();

const emit = defineEmits<{
  (e: 'select', item: Item): void;
}>();
</script>

<template>
  <ul class="max-h-60 divide-y overflow-y-auto rounded-lg border">
    <li v-for="item in items" :key="item.id">
      <component
        :is="item.to ? RouterLink : 'button'"
        :to="item.to"
        :type="item.to ? undefined : 'button'"
        class="flex w-full items-center gap-2.5 px-3 py-2.5 text-left text-sm transition-colors hover:bg-muted/30"
        @click="emit('select', item)"
      >
        <span v-if="item.kind" class="text-muted-foreground"><ResourceIcon :kind="item.kind" /></span>
        <span class="min-w-0 flex-1 truncate font-medium text-foreground">{{ item.name }}</span>
        <span v-if="item.detail" class="shrink-0 text-xs text-muted-foreground">{{ item.detail }}</span>
        <ChevronRight :size="14" class="shrink-0 text-muted-foreground" />
      </component>
    </li>
  </ul>
</template>
