<script setup lang="ts">
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { logto } from '@/lib/logto';
import { useAuth } from '@/composables/useAuth';

const router = useRouter();
const { fetchUser } = useAuth();

onMounted(async () => {
  try {
    await logto.handleSignInCallback(window.location.href);
    await fetchUser();
    router.replace('/');
  } catch {
    router.replace('/login');
  }
});
</script>

<template>
  <div class="flex h-screen items-center justify-center text-sm text-muted-foreground">
    Signing you in…
  </div>
</template>
