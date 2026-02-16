<script setup lang="ts">
import { useRoute } from 'vue-router';

const route = useRoute();
</script>

<template>
  <div class="p-8">
    <h1 class="text-2xl font-semibold text-gray-900">{{ route.params.env }}</h1>
    <p class="mt-2 text-gray-600">Deployed services and sync status.</p>
  </div>
</template>
