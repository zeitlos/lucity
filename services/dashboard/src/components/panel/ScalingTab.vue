<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { SetServiceScalingMutation } from '@/graphql/projects';
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Slider } from '@/components/ui/slider';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  projectId: string;
  service: {
    name: string;
    image: string;
    port: number;
    framework?: string;
  };
}>();

const { activeEnvironment } = useEnvironment();

const envService = computed(() =>
  activeEnvironment.value?.services.find(s => s.name === props.service.name),
);

// Scaling state
const mode = ref<'manual' | 'autoscaling'>('manual');
const replicas = ref(1);
const minReplicas = ref(1);
const maxReplicas = ref(10);
const targetCPU = ref(70);
const saving = ref(false);

const { mutate: setScalingMutate } = useMutation(SetServiceScalingMutation);

// Sync state from current service instance whenever it changes
function syncFromService() {
  const svc = envService.value;
  if (!svc) return;

  if (svc.scaling?.autoscaling?.enabled) {
    mode.value = 'autoscaling';
    replicas.value = svc.scaling.replicas || svc.replicas || 1;
    minReplicas.value = svc.scaling.autoscaling.minReplicas;
    maxReplicas.value = svc.scaling.autoscaling.maxReplicas;
    targetCPU.value = svc.scaling.autoscaling.targetCPU;
  } else {
    mode.value = 'manual';
    replicas.value = svc.scaling?.replicas || svc.replicas || 1;
    minReplicas.value = 1;
    maxReplicas.value = 10;
    targetCPU.value = 70;
  }
}

watch(envService, syncFromService, { immediate: true });

async function handleSave() {
  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  saving.value = true;
  try {
    const input: Record<string, unknown> = {
      projectId: props.projectId,
      environment: envName,
      service: props.service.name,
      replicas: replicas.value,
    };

    if (mode.value === 'autoscaling') {
      input.autoscaling = {
        enabled: true,
        minReplicas: minReplicas.value,
        maxReplicas: maxReplicas.value,
        targetCPU: targetCPU.value,
      };
    } else {
      input.autoscaling = {
        enabled: false,
        minReplicas: 1,
        maxReplicas: 1,
        targetCPU: 70,
      };
    }

    await setScalingMutate({ input });
    toast.success('Scaling updated');
  } catch (e: unknown) {
    toast.error('Failed to update scaling', { description: errorMessage(e) });
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Current status -->
    <div
      v-if="envService"
      class="rounded-lg border border-border/60 bg-muted/30 px-4 py-3"
    >
      <p class="text-xs text-muted-foreground">
        <template v-if="envService.scaling?.autoscaling?.enabled">
          Currently <strong class="text-foreground">{{ envService.replicas }}</strong>
          replica{{ envService.replicas !== 1 ? 's' : '' }}
          &middot; autoscaling {{ envService.scaling.autoscaling.minReplicas }}&ndash;{{ envService.scaling.autoscaling.maxReplicas }}
          &middot; target {{ envService.scaling.autoscaling.targetCPU }}% CPU
        </template>
        <template v-else>
          Currently <strong class="text-foreground">{{ envService.replicas }}</strong>
          replica{{ envService.replicas !== 1 ? 's' : '' }}
          &middot; manual scaling
        </template>
      </p>
    </div>

    <!-- Mode -->
    <div class="space-y-2">
      <Label>Scaling mode</Label>
      <Select v-model="mode">
        <SelectTrigger class="w-48">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="manual">Manual</SelectItem>
          <SelectItem value="autoscaling">Autoscaling</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <!-- Manual: replicas -->
    <div v-if="mode === 'manual'" class="space-y-2">
      <div class="flex items-center justify-between">
        <Label>Replicas</Label>
        <Input
          type="number"
          v-model.number="replicas"
          class="h-7 w-20 text-right text-xs"
          :min="1"
          :max="20"
        />
      </div>
      <Slider
        :model-value="[replicas]"
        :min="1"
        :max="20"
        :step="1"
        @update:model-value="replicas = $event?.[0] ?? replicas"
      />
    </div>

    <!-- Autoscaling -->
    <template v-if="mode === 'autoscaling'">
      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <Label>Min replicas</Label>
          <Input
            type="number"
            v-model.number="minReplicas"
            class="h-7 w-20 text-right text-xs"
            :min="1"
            :max="20"
          />
        </div>
        <Slider
          :model-value="[minReplicas]"
          :min="1"
          :max="20"
          :step="1"
          @update:model-value="minReplicas = $event?.[0] ?? minReplicas"
        />
      </div>

      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <Label>Max replicas</Label>
          <Input
            type="number"
            v-model.number="maxReplicas"
            class="h-7 w-20 text-right text-xs"
            :min="1"
            :max="20"
          />
        </div>
        <Slider
          :model-value="[maxReplicas]"
          :min="1"
          :max="20"
          :step="1"
          @update:model-value="maxReplicas = $event?.[0] ?? maxReplicas"
        />
      </div>

      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <Label>Target CPU</Label>
          <div class="flex items-center gap-1.5">
            <Input
              type="number"
              v-model.number="targetCPU"
              class="h-7 w-20 text-right text-xs"
              :min="10"
              :max="95"
            />
            <span class="text-xs text-muted-foreground">%</span>
          </div>
        </div>
        <Slider
          :model-value="[targetCPU]"
          :min="10"
          :max="95"
          :step="5"
          @update:model-value="targetCPU = $event?.[0] ?? targetCPU"
        />
        <p class="text-[11px] text-muted-foreground">
          Scale up when average CPU exceeds {{ targetCPU }}%.
        </p>
      </div>
    </template>

    <!-- Save -->
    <div class="flex justify-end">
      <Button
        size="sm"
        :disabled="saving"
        @click="handleSave"
      >
        {{ saving ? 'Saving...' : 'Save scaling' }}
      </Button>
    </div>
  </div>
</template>
