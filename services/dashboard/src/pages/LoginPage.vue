<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { Github } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';
import { toast } from '@/components/ui/sonner';
import { Button } from '@/components/ui/button';
import { Chip } from '@/components/ui/chip';
import { Separator } from '@/components/ui/separator';
import BaseLogo from '@/components/BaseLogo.vue';
import ThemeToggle from '@/components/ThemeToggle.vue';
import { useTheme } from '@/composables/useTheme';
import alpsHarborImg from '../../assets/img/alps_harbor.webp';
import mountainCityImg from '../../assets/img/mountain_city.webp';
import gopherShipImg from '../../assets/img/gopher_ship.webp';

const route = useRoute();
const { login } = useAuth();
const { theme } = useTheme();

const isDark = computed(() => theme.value === 'dark');

const errorMessage = computed(() => {
  if (route.query.error === 'no_installation') {
    return 'The Lucity GitHub App is not installed on your account. Please install it first, then try signing in again.';
  }
  if (route.query.error === 'no_workspace') {
    return 'You are not a member of any workspace. Contact your administrator.';
  }
  return null;
});

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

// --- Progressive image loading ---
const imageLoaded = ref(false);

function onImageLoad() {
  imageLoaded.value = true;
}

// --- Logo breathing delay ---
const logoLoaded = ref(false);

onMounted(() => {
  window.addEventListener('keydown', onKeyDown);
  setTimeout(() => { logoLoaded.value = true; }, 600);
});

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown);
});
</script>

<template>
  <div class="h-screen w-screen px-4 pt-4 pb-5">
    <!-- Image frame: inset from viewport edges, rounded, with inner shadow -->
    <div class="image-frame relative flex h-full items-center justify-center overflow-hidden rounded-2xl bg-muted">
      <!-- Image layer: both images stacked, crossfade on theme change -->
      <div
        class="absolute inset-0 transition-opacity duration-700"
        :class="imageLoaded ? 'opacity-100' : 'opacity-0'"
      >
        <!-- Light mode image (always at bottom) -->
        <img
          :src="alpsHarborImg"
          alt=""
          class="bg-image absolute inset-0 h-full w-full object-cover"
          :class="[
            easterEggActive ? 'opacity-0' : 'opacity-100',
            isDark ? 'scale-105' : 'scale-100',
          ]"
          @load="onImageLoad"
        >
        <!-- Dark mode image (fades in on top) -->
        <img
          :src="mountainCityImg"
          alt=""
          class="bg-image absolute inset-0 h-full w-full object-cover"
          :class="[
            easterEggActive ? 'opacity-0' : (isDark ? 'opacity-100' : 'opacity-0'),
            isDark ? 'scale-100' : 'scale-105',
          ]"
        >
        <!-- Easter egg gopher -->
        <img
          :src="gopherShipImg"
          alt=""
          class="absolute inset-0 h-full w-full object-cover transition-opacity duration-1000"
          :class="easterEggActive ? 'opacity-100' : 'opacity-0'"
        >
      </div>

      <!-- Inset shadow overlay -->
      <div class="image-inset pointer-events-none absolute inset-0 z-10 rounded-2xl" />

      <!-- Login card floating over the image -->
      <div class="login-card relative z-20 w-full max-w-sm overflow-hidden rounded-xl border border-border p-8">
        <!-- Pattern texture -->
        <div class="pattern-crosshatch pointer-events-none absolute inset-0 opacity-[0.03]" />

        <div class="relative z-10 space-y-6">
          <!-- Badge -->
          <div class="flex justify-center">
            <Chip>Fully Ejectable</Chip>
          </div>

          <!-- Logo -->
          <div class="flex justify-center">
            <BaseLogo
              :size="128"
              :class="logoLoaded ? 'logo-breathing' : ''"
            />
          </div>

          <!-- Wordmark -->
          <h1 class="text-center font-serif text-4xl tracking-tight text-foreground">
            Lucity
          </h1>

          <!-- Subtitle -->
          <p class="text-center text-sm text-muted-foreground">
            Start deploying in under 5 minutes.
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
            Continue with GitHub
          </Button>

          <!-- Footer -->
          <Separator />
          <p class="text-center text-xs text-muted-foreground">
            Built on Kubernetes, ArgoCD &amp; Helm
          </p>
        </div>
      </div>

      <!-- Credit: bottom-left -->
      <div class="theme-toggle-corner absolute bottom-4 left-4 z-20">
        <p class="text-xs text-white/60">
          cooked with care at
          <a
            href="https://zeitlos.software"
            target="_blank"
            class="underline decoration-white/30 underline-offset-2 transition-colors hover:text-white/80"
          >zeitlos.software</a>
          in Switzerland 🇨🇭
        </p>
      </div>

      <!-- Theme toggle: bottom-right -->
      <div class="theme-toggle-corner absolute bottom-4 right-4 z-20">
        <ThemeToggle />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Inset shadow — vignette effect over the image */
.image-inset {
  box-shadow:
    inset 0 0 80px oklch(0 0 0 / 0.2),
    inset 0 2px 20px oklch(0 0 0 / 0.15);
}

:global(.dark) .image-inset {
  box-shadow:
    inset 0 0 100px oklch(0 0 0 / 0.35),
    inset 0 2px 30px oklch(0 0 0 / 0.25);
}

/* Login card */
.login-card {
  background: var(--gradient-card);
  box-shadow: var(--shadow-lg);
  animation: card-enter 0.5s cubic-bezier(0.23, 1, 0.32, 1) both;
}

@keyframes card-enter {
  from {
    opacity: 0;
    transform: translateY(12px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* Theme toggle in corner — smooth entry */
.theme-toggle-corner {
  animation: fade-in 0.6s cubic-bezier(0.23, 1, 0.32, 1) 0.3s both;
}

@keyframes fade-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Background image crossfade — slow dissolve with subtle zoom */
.bg-image {
  transition:
    opacity 1.2s cubic-bezier(0.4, 0, 0.2, 1),
    transform 1.8s cubic-bezier(0.4, 0, 0.2, 1);
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
