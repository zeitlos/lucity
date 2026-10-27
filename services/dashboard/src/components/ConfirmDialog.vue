<script setup lang="ts">
import { computed, ref, useId, watch, type Component } from 'vue';
import { Trash2 } from '@lucide/vue';
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import DialogIcon from '@/components/DialogIcon.vue';

const open = defineModel<boolean>('open', { default: false });

const props = withDefaults(defineProps<{
  title: string;
  description?: string;
  confirmLabel: string;
  pendingLabel?: string;
  pending?: boolean;
  destructive?: boolean;
  icon?: Component;
  confirmText?: string;
  cancelLabel?: string;
}>(), {
  description: undefined,
  pendingLabel: undefined,
  pending: false,
  destructive: false,
  icon: undefined,
  confirmText: undefined,
  cancelLabel: 'Cancel',
});

const emit = defineEmits<{
  (e: 'confirm'): void;
}>();

const inputId = useId();
const typed = ref('');

watch(open, (value) => {
  if (value) typed.value = '';
});

const tileIcon = computed(() => props.icon ?? Trash2);
const canConfirm = computed(() => !props.pending && (!props.confirmText || typed.value.trim() === props.confirmText));

function confirm() {
  if (canConfirm.value) emit('confirm');
}
</script>

<template>
  <AlertDialog v-model:open="open">
    <AlertDialogTrigger v-if="$slots.trigger" as-child>
      <slot name="trigger" />
    </AlertDialogTrigger>
    <AlertDialogContent class="flex min-h-68 max-w-md flex-col">
      <AlertDialogHeader>
        <DialogIcon :icon="tileIcon" :destructive="destructive" />
        <AlertDialogTitle>{{ title }}</AlertDialogTitle>
        <AlertDialogDescription v-if="description || $slots.description">
          <slot name="description">{{ description }}</slot>
        </AlertDialogDescription>
      </AlertDialogHeader>

      <slot />

      <form v-if="confirmText" class="space-y-1.5" @submit.prevent="confirm">
        <label :for="inputId" class="text-sm text-muted-foreground">
          Type "{{ confirmText }}" to confirm
        </label>
        <Input
          :id="inputId"
          v-model="typed"
          autocomplete="off"
          spellcheck="false"
          data-1p-ignore="true"
        />
      </form>

      <AlertDialogFooter class="mt-auto">
        <AlertDialogCancel :disabled="pending">{{ cancelLabel }}</AlertDialogCancel>
        <Button
          :variant="destructive ? 'destructive' : 'default'"
          :pending="pending"
          :disabled="!canConfirm"
          @click="confirm"
        >
          {{ pending && pendingLabel ? pendingLabel : confirmLabel }}
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
