<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { Github } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';

const route = useRoute();
const { login } = useAuth();

const errorMessage = computed(() => {
  if (route.query.error === 'no_installation') {
    return 'The Lucity GitHub App is not installed on your account. Please install it first, then try signing in again.';
  }
  return null;
});
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div class="w-full max-w-sm space-y-6 p-8">
      <h1 class="text-2xl font-semibold text-gray-900">Sign in to Lucity</h1>
      <p class="text-gray-600">Connect your GitHub account to get started.</p>
      <div
        v-if="errorMessage"
        class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800"
      >
        {{ errorMessage }}
      </div>
      <button
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-gray-900 px-4 py-2 text-white hover:bg-gray-800"
        @click="login"
      >
        <Github :size="18" />
        Sign in with GitHub
      </button>
    </div>
  </div>
</template>
