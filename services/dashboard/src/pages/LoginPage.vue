<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';
import { Github } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';
import { toast } from '@/components/ui/sonner';
import { Button } from '@/components/ui/button';
import BaseLogo from '@/components/BaseLogo.vue';
import ThemeToggle from '@/components/ThemeToggle.vue';
import alpsHarborImg from '../../assets/img/alps_harbor.webp';
import gopherShipImg from '../../assets/img/gopher_ship.webp';

const route = useRoute();
const { login } = useAuth();

const errorMessage = computed(() => {
  if (route.query.error === 'no_installation') {
    return 'The Lucity GitHub App is not installed on your account. Please install it first, then try signing in again.';
  }
  return null;
});

// --- Mouse parallax ---
const mouse = reactive({ x: 0.5, y: 0.5 });
const parallaxStyle = computed(() => {
  const moveX = (mouse.x - 0.5) * -30;
  const moveY = (mouse.y - 0.5) * -15;
  return { transform: `translate(${moveX}px, ${moveY}px) scale(1.05)` };
});

function onMouseMove(e: MouseEvent) {
  mouse.x = e.clientX / window.innerWidth;
  mouse.y = e.clientY / window.innerHeight;
}

// --- Konami code easter egg ---
const KONAMI = ['ArrowUp', 'ArrowUp', 'ArrowDown', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'ArrowLeft', 'ArrowRight', 'b', 'a'];
const konamiProgress = ref(0);
const easterEggActive = ref(false);

function onKeyDown(e: KeyboardEvent) {
  if (e.key === KONAMI[konamiProgress.value]) {
    konamiProgress.value++;
    if (konamiProgress.value === KONAMI.length) {
      triggerEasterEgg();
      konamiProgress.value = 0;
    }
  } else {
    konamiProgress.value = e.key === KONAMI[0] ? 1 : 0;
  }
}

function triggerEasterEgg() {
  easterEggActive.value = true;
  toast('You found the gopher!', {
    description: 'Captain Gopher is navigating the deployment seas.',
  });
  setTimeout(() => { easterEggActive.value = false; }, 4000);
}

// --- Logo breathing delay ---
const logoLoaded = ref(false);

onMounted(() => {
  window.addEventListener('mousemove', onMouseMove);
  window.addEventListener('keydown', onKeyDown);
  setTimeout(() => { logoLoaded.value = true; }, 600);
});

onUnmounted(() => {
  window.removeEventListener('mousemove', onMouseMove);
  window.removeEventListener('keydown', onKeyDown);
});
</script>

<template>
  <div class="flex min-h-screen items-center justify-center p-6 lg:p-8">
    <!-- Main card wrapper -->
    <div class="login-wrapper flex w-full max-w-5xl flex-col overflow-hidden rounded-2xl border border-border shadow-lg lg:flex-row">
      <!-- Left panel: image -->
      <div class="image-panel relative h-[36vh] overflow-hidden lg:h-auto lg:flex-[3]">
        <div
          class="absolute inset-0"
          :style="parallaxStyle"
        >
          <img
            :src="alpsHarborImg"
            alt=""
            class="absolute inset-0 h-full w-full object-cover transition-opacity duration-1000"
            :class="easterEggActive ? 'opacity-0' : 'opacity-100'"
          >
          <img
            :src="gopherShipImg"
            alt=""
            class="absolute inset-0 h-full w-full object-cover transition-opacity duration-1000"
            :class="easterEggActive ? 'opacity-100' : 'opacity-0'"
          >
        </div>
        <!-- Inset shadow overlay -->
        <div class="image-inset pointer-events-none absolute inset-0" />
      </div>

      <!-- Right panel: login form -->
      <div class="relative flex flex-[2] flex-col items-center justify-center bg-card px-8 py-12 lg:px-12">
        <!-- Pattern texture -->
        <div class="pattern-crosshatch pointer-events-none absolute inset-0 opacity-[0.03]" />

        <div class="relative z-10 w-full max-w-xs space-y-6">
          <!-- Logo -->
          <div class="flex justify-center">
            <BaseLogo
              :size="96"
              :class="{ 'logo-breathing': logoLoaded }"
            />
          </div>

          <!-- Wordmark -->
          <h1 class="text-center font-serif text-4xl tracking-tight text-foreground">
            Lucity
          </h1>

          <!-- Subtitle -->
          <p class="text-center text-sm text-muted-foreground">
            Connect your GitHub account to start deploying.
          </p>

          <!-- Error message -->
          <div
            v-if="errorMessage"
            class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-foreground"
          >
            {{ errorMessage }}
          </div>

          <!-- Sign-in button -->
          <Button
            class="w-full gap-2"
            @click="login"
          >
            <Github :size="18" />
            Sign in with GitHub
          </Button>

          <!-- Theme toggle -->
          <div class="flex justify-center">
            <ThemeToggle />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Wrapper entry animation */
.login-wrapper {
  animation: wrapper-enter 0.5s cubic-bezier(0.23, 1, 0.32, 1) both;
}

@keyframes wrapper-enter {
  from {
    opacity: 0;
    transform: scale(0.98) translateY(8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* Inset shadow over the image — vignette / recessed look */
.image-inset {
  box-shadow:
    inset 0 4px 30px oklch(0 0 0 / 0.15),
    inset 0 0 80px oklch(0 0 0 / 0.06);
}

:global(.dark) .image-inset {
  box-shadow:
    inset 0 4px 30px oklch(0 0 0 / 0.3),
    inset 0 0 80px oklch(0 0 0 / 0.12);
}

/* Logo breathing animation */
.logo-breathing {
  animation: breathe 4s ease-in-out infinite;
}

@keyframes breathe {
  0%, 100% {
    transform: translateY(0);
    filter: drop-shadow(0 4px 12px oklch(0.75 0.18 160 / 0.3));
  }
  50% {
    transform: translateY(-4px);
    filter: drop-shadow(0 8px 20px oklch(0.75 0.18 160 / 0.5));
  }
}
</style>
