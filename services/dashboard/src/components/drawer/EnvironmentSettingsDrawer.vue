<script setup lang="ts">
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { ResourceTier } from '@/gql/graphql';

const SetEnvironmentResourcesDocument = graphql(`
  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {
    setEnvironmentResources(input: $input) {
      id
      resourceTier
    }
  }
`);

const DeleteEnvironmentDocument = graphql(`
  mutation DeleteEnvironment($environment: EnvironmentID!) {
    deleteEnvironment(environment: $environment)
  }
`);
import { apolloClient } from '@/lib/apollo';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import { Button } from '@/components/ui/button';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  projectId: string;
  environment: {
    id: string;
    name: string;
    resourceTier: ResourceTier;
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const router = useRouter();

const tier = ref<ResourceTier>(props.environment.resourceTier);

watch(
  () => [props.environment.id, props.environment.resourceTier] as const,
  ([, current]) => {
    tier.value = current;
  },
);

const { mutate: setResourcesMutate, loading: savingTier } = useMutation(SetEnvironmentResourcesDocument);

async function handleSaveTier() {
  try {
    await setResourcesMutate({
      input: {
        environment: props.environment.id,
        tier: tier.value,
      },
    });
    toast.success('Resource tier updated');
  } catch (e: unknown) {
    errorToast('Failed to update resource tier', { description: errorMessage(e) });
  }
}

const { mutate: deleteEnvironmentMutate, loading: deleting } = useMutation(DeleteEnvironmentDocument);
const confirmOpen = ref(false);

async function handleDelete() {
  const { id, name } = props.environment;

  try {
    const res = await deleteEnvironmentMutate({ environment: id });

    if (res?.errors?.length) {
      errorToast('Failed to delete environment', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    apolloClient.cache.evict({ id: apolloClient.cache.identify({ __typename: 'Environment', id }) });
    apolloClient.cache.gc();

    toast.success(`Environment "${name}" deleted`);
    confirmOpen.value = false;
    router.push({ name: 'project', params: { projectId: props.projectId } });
  } catch (e: unknown) {
    errorToast('Failed to delete environment', { description: errorMessage(e) });
  }
}
</script>

<template>
  <ContextDrawer
    title="Environment settings"
    @close="emit('close')"
  >
    <div class="flex min-h-full flex-col gap-6">
      <section class="space-y-2.5">
        <h3 class="text-sm font-medium text-muted-foreground">Resource tier</h3>
        <RadioGroup
          :model-value="tier"
          class="grid gap-2.5"
          @update:model-value="tier = $event as ResourceTier"
        >
          <label
            class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
            :class="tier === ResourceTier.Eco ? 'border-primary bg-primary/5' : 'border-border'"
          >
            <div class="flex items-center gap-2">
              <RadioGroupItem :value="ResourceTier.Eco" />
              <span class="text-sm font-medium">Eco</span>
            </div>
            <p class="text-xs text-muted-foreground">
              Pay for what you use. Best for development, staging, and side projects.
            </p>
          </label>
          <label
            class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
            :class="tier === ResourceTier.Production ? 'border-primary bg-primary/5' : 'border-border'"
          >
            <div class="flex items-center gap-2">
              <RadioGroupItem :value="ResourceTier.Production" />
              <span class="text-sm font-medium">Production</span>
            </div>
            <p class="text-xs text-muted-foreground">
              Reserved resources. Best for production workloads with predictable load.
            </p>
          </label>
        </RadioGroup>
        <p class="text-xs text-muted-foreground">
          {{ tier === ResourceTier.Eco
            ? 'Services run burstable and are billed for what they actually consume. Performance can vary under load.'
            : 'Services reserve their full CPU and memory limit and are billed for the reservation. Each service picks up the change on its next deploy.' }}
        </p>
        <div class="flex justify-end">
          <Button
            :disabled="savingTier || tier === environment.resourceTier"
            @click="handleSaveTier"
          >
            {{ savingTier ? 'Saving...' : 'Save tier' }}
          </Button>
        </div>
      </section>

      <section class="space-y-2.5">
        <h3 class="text-sm font-medium text-muted-foreground">General</h3>
        <div class="space-y-2.5 rounded-lg border p-3">
          <div class="flex items-center justify-between gap-4">
            <span class="text-sm text-muted-foreground">Name</span>
            <span class="text-sm font-medium text-foreground">{{ environment.name }}</span>
          </div>
          <div class="flex items-center justify-between gap-4">
            <span class="text-sm text-muted-foreground">ID</span>
            <span class="truncate font-mono text-xs text-foreground">{{ environment.id }}</span>
          </div>
        </div>
      </section>

      <div class="mt-auto flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
        <div>
          <p class="text-sm font-medium text-foreground">Delete environment</p>
          <p class="text-xs text-muted-foreground">
            It has to be empty, and the last environment of a project can't be deleted.
          </p>
        </div>
        <ConfirmDialog
          v-model:open="confirmOpen"
          :title="`Delete environment &quot;${environment.name}&quot;?`"
          description="Its deployment history goes with it. This can't be undone."
          confirm-label="Delete environment"
          pending-label="Deleting…"
          :pending="deleting"
          destructive
          @confirm="handleDelete"
        >
          <template #trigger>
            <Button variant="destructive" class="shrink-0">
              <Trash2 />
              Delete
            </Button>
          </template>
        </ConfirmDialog>
      </div>
    </div>
  </ContextDrawer>
</template>
