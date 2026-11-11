<script setup lang="ts">
import { mobileSheetClass } from '@/lib/dialog';
import { ref, watch } from 'vue';
import { KeyRound } from '@lucide/vue';
import DialogIcon from '@/components/DialogIcon.vue';
import SecretInput from '@/components/SecretInput.vue';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';

const open = defineModel<boolean>('open', { default: false });

defineProps<{
  title: string;
  description: string;
  label: string;
  confirmLabel: string;
  pending: boolean;
}>();

const emit = defineEmits<{
  (e: 'submit', value: string): void;
}>();

const value = ref('');

watch(open, (isOpen) => {
  if (isOpen) value.value = '';
});
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-md']">
      <form class="space-y-4" @submit.prevent="value && emit('submit', value)">
        <DialogHeader>
          <DialogIcon :icon="KeyRound" />
          <DialogTitle>{{ title }}</DialogTitle>
          <DialogDescription>{{ description }}</DialogDescription>
        </DialogHeader>

        <div class="space-y-2">
          <Label for="replace-secret">{{ label }}</Label>
          <SecretInput id="replace-secret" v-model="value" :disabled="pending" />
        </div>

        <DialogFooter>
          <Button type="button" variant="ghost" :disabled="pending" @click="open = false">Cancel</Button>
          <Button type="submit" :pending="pending" :disabled="!value">
            {{ pending ? 'Checking…' : confirmLabel }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
