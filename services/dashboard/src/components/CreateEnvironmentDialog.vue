<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { CreateEnvironmentMutation, ProjectQuery } from '@/graphql/projects';
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  open: boolean;
  projectId: string;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const { environments } = useEnvironment();

const name = ref('');
const mode = ref<'duplicate' | 'empty'>('duplicate');
const fromEnvironment = ref<string>('');
const tier = ref<string>('ECO');

const { mutate, loading } = useMutation(CreateEnvironmentMutation, {
  refetchQueries: () => [{ query: ProjectQuery, variables: { id: props.projectId } }],
});

const nonEphemeralEnvs = computed(() =>
  environments.value.filter(e => !e.ephemeral),
);

// Default to 'duplicate' when environments exist, 'empty' when none
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    if (nonEphemeralEnvs.value.length > 0) {
      mode.value = 'duplicate';
      fromEnvironment.value = nonEphemeralEnvs.value[0].name;
    } else {
      mode.value = 'empty';
    }
  }
});

async function handleCreate() {
  if (!name.value.trim()) return;

  try {
    const input: Record<string, string> = {
      projectId: props.projectId,
      name: name.value.trim(),
      tier: tier.value,
    };
    if (mode.value === 'duplicate' && fromEnvironment.value) {
      input.fromEnvironment = fromEnvironment.value;
    }

    const res = await mutate({ input });

    if (res?.errors?.length) {
      toast.error('Failed to create environment', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success(`Environment "${name.value.trim()}" created`);
    name.value = '';
    fromEnvironment.value = '';
    tier.value = 'ECO';
    emit('update:open', false);
  } catch (e: unknown) {
    toast.error('Failed to create environment', { description: errorMessage(e) });
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>New Environment</DialogTitle>
        <DialogDescription>
          All changes will be isolated from other environments.
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="handleCreate">
        <div class="space-y-2">
          <Label for="env-name">Environment name</Label>
          <Input
            id="env-name"
            v-model="name"
            placeholder="e.g. staging, preview"
            :disabled="loading"
          />
        </div>

        <div
          v-if="nonEphemeralEnvs.length > 0"
          class="space-y-2"
        >
          <RadioGroup v-model="mode" class="space-y-3">
            <label
              class="flex cursor-pointer flex-col gap-2 rounded-lg border p-3 transition-colors"
              :class="mode === 'duplicate' ? 'border-primary bg-primary/5' : 'border-border'"
            >
              <div class="flex items-center gap-2">
                <RadioGroupItem value="duplicate" />
                <span class="text-sm font-medium">Duplicate Environment</span>
              </div>
              <p class="text-xs text-muted-foreground pl-6">
                Copy all the services, variables, and configuration from an existing environment.
              </p>
              <div v-if="mode === 'duplicate'" class="pl-6 pt-1">
                <Select v-model="fromEnvironment">
                  <SelectTrigger>
                    <SelectValue placeholder="Select environment" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem
                      v-for="env in nonEphemeralEnvs"
                      :key="env.id"
                      :value="env.name"
                    >
                      {{ env.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </label>
            <label
              class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
              :class="mode === 'empty' ? 'border-primary bg-primary/5' : 'border-border'"
            >
              <div class="flex items-center gap-2">
                <RadioGroupItem value="empty" />
                <span class="text-sm font-medium">Empty Environment</span>
              </div>
              <p class="text-xs text-muted-foreground pl-6">
                An empty environment with no services or variables included.
              </p>
            </label>
          </RadioGroup>
        </div>

        <div class="space-y-2">
          <Label>Tier</Label>
          <RadioGroup v-model="tier" class="grid grid-cols-2 gap-3">
            <label
              class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
              :class="tier === 'ECO' ? 'border-primary bg-primary/5' : 'border-border'"
            >
              <div class="flex items-center gap-2">
                <RadioGroupItem value="ECO" />
                <span class="text-sm font-medium">Eco</span>
              </div>
              <p class="text-xs text-muted-foreground">
                Pay for what you use. Best for development, staging, and side projects.
              </p>
            </label>
            <label
              class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
              :class="tier === 'PRODUCTION' ? 'border-primary bg-primary/5' : 'border-border'"
            >
              <div class="flex items-center gap-2">
                <RadioGroupItem value="PRODUCTION" />
                <span class="text-sm font-medium">Production</span>
              </div>
              <p class="text-xs text-muted-foreground">
                Reserved resources. Best for production workloads with predictable load.
              </p>
            </label>
          </RadioGroup>
        </div>

        <DialogFooter>
          <Button
            type="submit"
            :disabled="!name.trim() || loading"
          >
            {{ loading ? 'Creating...' : 'Create Environment' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
