<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { inject, type Ref } from 'vue';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
const id = inject<Ref<string>>('form-item-id');
</script>

<template>
  <Label :for="id" :class="cn(props.class)">
    <slot />
  </Label>
</template>
