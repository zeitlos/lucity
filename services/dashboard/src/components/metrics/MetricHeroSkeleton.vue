<script setup lang="ts">
import { Skeleton } from '@/components/ui/skeleton';
</script>

<template>
  <div class="flex min-w-0 flex-col gap-3.5 bg-card px-4 pb-3 pt-4">
    <div class="flex items-start justify-between gap-3">
      <div>
        <div class="flex h-4 items-center">
          <Skeleton class="h-3 w-14" />
        </div>
        <div class="mt-0.5 flex h-9 items-center gap-2">
          <Skeleton class="h-6 w-20" />
          <Skeleton class="h-3 w-12" />
        </div>
      </div>
      <div class="flex h-5 items-center">
        <Skeleton class="h-2.5 w-14" />
      </div>
    </div>

    <div class="relative h-37.5">
      <div class="absolute inset-x-0 top-0 border-t border-dashed border-border" />
      <div class="absolute inset-x-0 top-1/2 border-t border-border/50" />
      <div class="absolute inset-x-0 bottom-0 border-t border-border" />
    </div>

    <div class="-mt-1.5 flex h-4 items-center justify-between">
      <Skeleton v-for="tick in 5" :key="tick" class="h-2.5 w-8" />
    </div>
  </div>
</template>
