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

// --- Card 3D tilt ---
const cardRef = ref<HTMLElement | null>(null);
const cardTransform = ref('');

function onCardMouseMove(e: MouseEvent) {
  if (!cardRef.value) return;
  const rect = cardRef.value.getBoundingClientRect();
  const x = (e.clientX - rect.left) / rect.width - 0.5;
  const y = (e.clientY - rect.top) / rect.height - 0.5;
  cardTransform.value = `perspective(800px) rotateY(${x * 6}deg) rotateX(${-y * 6}deg)`;
}

function onCardMouseLeave() {
  cardTransform.value = '';
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
  <div class="relative h-screen w-screen overflow-hidden">
    <!-- Background images with parallax (two stacked for cross-fade) -->
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

    <!-- Dimming overlay -->
    <div class="absolute inset-0 bg-black/30 dark:bg-black/50" />

    <!-- Content -->
    <div class="relative z-10 flex h-full items-center justify-center p-6 lg:justify-end lg:pr-[12%]">
      <!-- Frosted glass card -->
      <div
        ref="cardRef"
        class="login-card relative w-full max-w-sm overflow-hidden rounded-xl border border-white/20 bg-white/15 p-8 shadow-2xl backdrop-blur-xl dark:border-white/10 dark:bg-black/25"
        :style="{ transform: cardTransform }"
        @mousemove="onCardMouseMove"
        @mouseleave="onCardMouseLeave"
      >
        <!-- Pattern texture -->
        <div class="pattern-crosshatch pointer-events-none absolute inset-0 opacity-[0.04]" />

        <!-- Card content -->
        <div class="relative z-10 space-y-6">
          <!-- Logo -->
          <div class="flex justify-center">
            <BaseLogo
              :size="96"
              :class="['login-logo', { 'logo-breathing': logoLoaded }]"
            />
          </div>

          <!-- Wordmark -->
          <h1 class="text-center font-serif text-4xl tracking-tight text-white drop-shadow-sm">
            Lucity
          </h1>

          <!-- Subtitle -->
          <p class="text-center text-sm text-white/70">
            Connect your GitHub account to start deploying.
          </p>

          <!-- Error message -->
          <div
            v-if="errorMessage"
            class="rounded-lg border border-red-400/30 bg-red-500/15 p-3 text-sm text-white/90 backdrop-blur-sm"
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
/* Card tilt spring-back + entry animation */
.login-card {
  transition: transform 0.4s cubic-bezier(0.23, 1, 0.32, 1), box-shadow 0.3s ease;
  will-change: transform;
  animation: card-enter 0.6s cubic-bezier(0.23, 1, 0.32, 1) both;
}

.login-card:hover {
  box-shadow:
    0 20px 80px -20px rgba(0, 0, 0, 0.3),
    0 8px 40px -10px oklch(0.75 0.12 160 / 0.15);
}

@keyframes card-enter {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.97);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* Logo breathing animation */
.login-logo {
  --primary: oklch(0.95 0.02 160);
  --accent: oklch(0.90 0.03 80);
  --accent-foreground: oklch(0.90 0.03 80);
  transition: transform 0.3s ease, filter 0.3s ease;
}

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
