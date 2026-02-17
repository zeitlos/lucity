<script setup lang="ts">
import { Github } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';

const { login } = useAuth();
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div class="w-full max-w-sm space-y-6 p-8">
      <h1 class="text-2xl font-semibold text-gray-900">Sign in to Lucity</h1>
      <p class="text-gray-600">Connect your GitHub account to get started.</p>
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
