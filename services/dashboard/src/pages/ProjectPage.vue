<script setup lang="ts">
import { useRoute } from 'vue-router';

const route = useRoute();
</script>

<template>
  <div class="p-8">
    <h1 class="text-2xl font-semibold text-gray-900">Project {{ route.params.id }}</h1>
    <p class="mt-2 text-gray-600">Environments and services.</p>
  </div>
</template>
