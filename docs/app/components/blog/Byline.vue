<script setup lang="ts">
defineProps<{
  name?: string;
}>();
</script>

<template>
  <div class="flex items-center gap-3 text-base text-stone-900 dark:text-stone-100">
    <BlogAvatar :name="name" />
    <span class="min-w-0">
      {{ name }}
      <small class="block text-sm text-stone-600 dark:text-stone-400"><slot /></small>
    </span>
  </div>
</template>
