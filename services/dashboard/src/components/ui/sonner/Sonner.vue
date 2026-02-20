<script setup lang="ts">
import { Toaster as Sonner } from 'vue-sonner';

const props = defineProps<{
  position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'top-center' | 'bottom-center';
  expand?: boolean;
  richColors?: boolean;
  theme?: 'light' | 'dark' | 'system';
}>();
</script>

<template>
  <Sonner
    class="toaster group"
    v-bind="props"
    :toast-options="{
      classes: {
        toast: 'group toast group-[.toaster]:bg-card/90 group-[.toaster]:text-card-foreground group-[.toaster]:border-border group-[.toaster]:shadow-lg group-[.toaster]:backdrop-blur-md group-[.toaster]:[background-image:var(--gradient-card)]',
        description: 'group-[.toast]:text-muted-foreground',
        actionButton: 'group-[.toast]:bg-primary group-[.toast]:text-primary-foreground',
        cancelButton: 'group-[.toast]:bg-muted group-[.toast]:text-muted-foreground',
      },
    }"
  />
</template>

<style>
/* Override sonner's rich-color backgrounds — we use our own card style */
[data-sonner-toaster][data-theme='light'] [data-sonner-toast][data-type='success'],
[data-sonner-toaster][data-theme='dark'] [data-sonner-toast][data-type='success'],
[data-sonner-toaster][data-theme='light'] [data-sonner-toast][data-type='error'],
[data-sonner-toaster][data-theme='dark'] [data-sonner-toast][data-type='error'],
[data-sonner-toaster][data-theme='light'] [data-sonner-toast][data-type='info'],
[data-sonner-toaster][data-theme='dark'] [data-sonner-toast][data-type='info'],
[data-sonner-toaster][data-theme='light'] [data-sonner-toast][data-type='warning'],
[data-sonner-toaster][data-theme='dark'] [data-sonner-toast][data-type='warning'] {
  --normal-bg: var(--card) !important;
  --normal-text: var(--card-foreground) !important;
  --normal-border: var(--border) !important;
  background: var(--gradient-card), var(--card) !important;
  backdrop-filter: blur(12px) !important;
  color: var(--card-foreground) !important;
  border-color: var(--border) !important;
}

/* Colored left-border accents per toast type */
[data-sonner-toast][data-type='success'] {
  border-left: 3px solid var(--status-ok) !important;
}

[data-sonner-toast][data-type='error'] {
  border-left: 3px solid var(--status-danger) !important;
}

[data-sonner-toast][data-type='info'] {
  border-left: 3px solid var(--primary) !important;
}

[data-sonner-toast][data-type='warning'] {
  border-left: 3px solid var(--status-warn) !important;
}

/* Icon color overrides — sonner rich colors paint the icon too */
[data-sonner-toast][data-type='success'] [data-icon] {
  color: var(--status-ok) !important;
}

[data-sonner-toast][data-type='error'] [data-icon] {
  color: var(--status-danger) !important;
}

[data-sonner-toast][data-type='info'] [data-icon] {
  color: var(--primary) !important;
}

[data-sonner-toast][data-type='warning'] [data-icon] {
  color: var(--status-warn) !important;
}

/* Description text stays muted regardless of toast type */
[data-sonner-toast] [data-description] {
  color: var(--muted-foreground) !important;
}

/* Close button styling */
[data-sonner-toast] [data-close-button] {
  border-color: var(--border) !important;
  background: var(--card) !important;
  color: var(--muted-foreground) !important;
}

[data-sonner-toast] [data-close-button]:hover {
  background: var(--accent) !important;
  color: var(--foreground) !important;
}
</style>
