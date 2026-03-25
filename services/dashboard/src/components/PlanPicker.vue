<script setup lang="ts">
import { Check } from 'lucide-vue-next';

defineProps<{
  modelValue: 'HOBBY' | 'PRO';
  disabled?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: 'HOBBY' | 'PRO'): void;
}>();
</script>

<template>
  <div class="space-y-2">
    <div class="grid grid-cols-2 gap-3">
      <button
        type="button"
        class="rounded-lg border p-4 text-left transition-colors"
        :class="modelValue === 'HOBBY'
          ? 'border-primary bg-primary/5'
          : 'hover:border-muted-foreground/50'"
        :disabled="disabled"
        @click="emit('update:modelValue', 'HOBBY')"
      >
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium">Hobby</p>
          <div
            class="flex size-5 items-center justify-center rounded-full border transition-colors"
            :class="modelValue === 'HOBBY'
              ? 'border-primary bg-primary'
              : 'border-muted-foreground/30'"
          >
            <Check
              v-if="modelValue === 'HOBBY'"
              :size="12"
              class="text-primary-foreground"
            />
          </div>
        </div>
        <p class="mt-1 text-lg font-semibold">
          &euro;5<span class="text-sm font-normal text-muted-foreground">/mo</span>
        </p>
        <p class="mt-0.5 text-xs text-muted-foreground">
          &euro;5 resource credit included
        </p>
        <ul class="mt-3 space-y-1.5">
          <li class="flex items-center gap-2 text-xs text-muted-foreground">
            <Check :size="12" class="shrink-0 text-primary" />
            1 user
          </li>
          <li class="flex items-center gap-2 text-xs text-muted-foreground">
            <Check :size="12" class="shrink-0 text-primary" />
            1 custom domain
          </li>
          <li class="flex items-center gap-2 text-xs text-muted-foreground">
            <Check :size="12" class="shrink-0 text-primary" />
            Community support
          </li>
        </ul>
      </button>
      <button
        type="button"
        class="rounded-lg border p-4 text-left transition-colors"
        :class="modelValue === 'PRO'
          ? 'border-primary bg-primary/5'
          : 'hover:border-muted-foreground/50'"
        :disabled="disabled"
        @click="emit('update:modelValue', 'PRO')"
      >
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium">Pro</p>
          <div
            class="flex size-5 items-center justify-center rounded-full border transition-colors"
            :class="modelValue === 'PRO'
              ? 'border-primary bg-primary'
              : 'border-muted-foreground/30'"
          >
            <Check
              v-if="modelValue === 'PRO'"
              :size="12"
              class="text-primary-foreground"
            />
          </div>
        </div>
        <p class="mt-1 text-lg font-semibold">
          &euro;25<span class="text-sm font-normal text-muted-foreground">/mo</span>
        </p>
        <p class="mt-0.5 text-xs text-muted-foreground">
          &euro;25 resource credit included
        </p>
        <ul class="mt-3 space-y-1.5">
          <li class="flex items-center gap-2 text-xs text-muted-foreground">
            <Check :size="12" class="shrink-0 text-primary" />
            Unlimited collaborators
          </li>
          <li class="flex items-center gap-2 text-xs text-muted-foreground">
            <Check :size="12" class="shrink-0 text-primary" />
            Unlimited custom domains
          </li>
          <li class="flex items-center gap-2 text-xs text-muted-foreground">
            <Check :size="12" class="shrink-0 text-primary" />
            Priority support
          </li>
        </ul>
      </button>
    </div>
    <p class="text-xs text-muted-foreground">
      You can change your plan anytime.
    </p>
  </div>
</template>
