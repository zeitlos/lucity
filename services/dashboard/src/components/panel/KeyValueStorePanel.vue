<script setup lang="ts">
import { reactive, computed, ref } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { X, Copy, Eye, EyeOff, DatabaseZap, Trash2, Database, HardDrive } from '@lucide/vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { onKeyStroke } from '@vueuse/core';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const KeyValueStoreCredentialsDocument = graphql(`
  query KeyValueStoreCredentials($keyValueStore: KeyValueStoreID!) {
    keyValueStoreCredentials(keyValueStore: $keyValueStore) {
      type
      host
      port
      password
      uri
    }
  }
`);

const DeleteKeyValueStoreDocument = graphql(`
  mutation DeleteKeyValueStore($keyValueStore: KeyValueStoreID!) {
    deleteKeyValueStore(keyValueStore: $keyValueStore)
  }
`);

const props = defineProps<{
  store: {
    id: string;
    name: string;
    version: string;
    size: string;
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'store-removed'): void;
}>();

onKeyStroke('Escape', () => {
  emit('close');
});

const { result, loading, error } = useQuery(
  KeyValueStoreCredentialsDocument,
  () => ({ keyValueStore: props.store.id }),
  () => ({ enabled: !!props.store.id }),
);

const credentials = computed(() => result.value?.keyValueStoreCredentials ?? []);

const isProvisioning = computed(() => {
  if (!error.value) return false;
  const gqlErrors = (error.value as unknown as { graphQLErrors?: { extensions?: { code?: string } }[] }).graphQLErrors;
  return gqlErrors?.some(e => e.extensions?.code === 'DATABASE_PROVISIONING') ?? false;
});

const endpointLabels: Record<string, string> = {
  INTERNAL: 'Private network',
  PLATFORM: 'Public internet',
  CUSTOM: 'Custom domain',
};

const groups = computed(() =>
  credentials.value.map(cred => ({
    type: cred.type,
    label: endpointLabels[cred.type] ?? cred.type,
    fields: [
      { key: `${cred.type}-uri`, label: 'REDIS_URL', value: cred.uri, sensitive: true },
      { key: `${cred.type}-host`, label: 'Host', value: cred.host, sensitive: false },
      { key: `${cred.type}-port`, label: 'Port', value: cred.port, sensitive: false },
      { key: `${cred.type}-password`, label: 'Password', value: cred.password, sensitive: true },
    ],
  })),
);

const revealed = reactive<Record<string, boolean>>({});

function toggleReveal(key: string) {
  revealed[key] = !revealed[key];
}

function mask(value: string): string {
  if (value.length <= 4) return '*'.repeat(value.length);
  return '*'.repeat(value.length - 2) + value.slice(-2);
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text);
  toast.success('Copied to clipboard');
}

const { mutate: deleteStore, loading: deleting } = useMutation(DeleteKeyValueStoreDocument);
const deleteDialogOpen = ref(false);

async function handleDelete() {
  try {
    await deleteStore({ keyValueStore: props.store.id });
    toast.success('Redis store removed');
    deleteDialogOpen.value = false;
    emit('store-removed');
  } catch (e: unknown) {
    errorToast('Failed to remove Redis store', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="flex h-full flex-col rounded-lg border bg-card shadow-sm">
    <!-- Header -->
    <div class="flex shrink-0 items-center justify-between border-b px-4 py-3">
      <div class="flex items-center gap-3">
        <img
          src="https://devicons.railway.com/i/redis.svg"
          :width="24"
          :height="24"
          class="shrink-0"
          alt=""
        />
        <h2 class="text-lg font-semibold text-foreground">{{ store.name }}</h2>
      </div>

      <Button variant="ghost" size="icon" class="h-7 w-7" @click="emit('close')">
        <X :size="16" />
      </Button>
    </div>

    <!-- Tab Content -->
    <ScrollArea class="flex-1">
      <Tabs default-value="connect" class="h-full">
        <div class="px-4 pt-2">
          <TabsList class="w-full">
            <TabsTrigger value="connect">Connect</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="connect" class="px-4 py-4">
          <div class="space-y-4">
            <div>
              <h3 class="text-sm font-medium text-foreground">Connection Details</h3>
              <p class="text-xs text-muted-foreground">
                Credentials for <strong>{{ store.name }}</strong>. Reachable from your other services over the private network.
              </p>
            </div>

            <!-- Loading -->
            <div v-if="loading" class="space-y-2">
              <Skeleton v-for="i in 4" :key="i" class="h-10 w-full" />
            </div>

            <!-- Provisioning -->
            <div
              v-else-if="isProvisioning"
              class="flex flex-col items-center justify-center gap-3 py-12 text-center"
            >
              <DatabaseZap :size="24" class="text-muted-foreground" />
              <div class="space-y-1">
                <p class="text-sm font-medium">Redis is provisioning</p>
                <p class="text-xs text-muted-foreground">Credentials will appear once the store is ready.</p>
              </div>
              <Spinner :size="16" class="animate-spin text-muted-foreground" />
            </div>

            <!-- Error (non-provisioning) -->
            <div
              v-else-if="error && !isProvisioning"
              class="rounded-lg border border-destructive/30 bg-destructive/5 p-3"
            >
              <p class="font-mono text-xs text-destructive">{{ error.message }}</p>
            </div>

            <!-- Credentials -->
            <div v-else-if="groups.length" class="space-y-4">
              <div v-for="group in groups" :key="group.type" class="space-y-1.5">
                <span class="text-xs font-medium text-muted-foreground">{{ group.label }}</span>
                <div
                  v-for="field in group.fields"
                  :key="field.key"
                  class="group flex items-center gap-2 rounded-md bg-muted/40 px-3 py-2"
                >
                  <span class="w-28 shrink-0 text-xs font-medium text-muted-foreground">{{ field.label }}</span>
                  <span class="flex-1 truncate font-mono text-xs text-foreground">
                    {{ field.sensitive && !revealed[field.key] ? mask(field.value) : field.value }}
                  </span>
                  <Button
                    v-if="field.sensitive"
                    variant="ghost"
                    size="icon"
                    class="h-6 w-6 shrink-0"
                    @click="toggleReveal(field.key)"
                  >
                    <EyeOff v-if="revealed[field.key]" :size="12" />
                    <Eye v-else :size="12" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-6 w-6 shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
                    @click="copyToClipboard(field.value)"
                  >
                    <Copy :size="12" />
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <div class="space-y-6">
            <!-- Configuration -->
            <section class="space-y-2">
              <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Configuration
              </h3>

              <div class="overflow-hidden rounded-lg border">
                <div class="divide-y">
                  <div class="flex items-center gap-3 px-4 py-3">
                    <Database :size="16" class="shrink-0 text-muted-foreground" />
                    <span class="flex-1 text-sm text-muted-foreground">Version</span>
                    <span class="font-mono text-sm font-medium text-foreground">Redis {{ store.version }}</span>
                  </div>
                  <div class="flex items-center gap-3 px-4 py-3">
                    <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
                    <span class="flex-1 text-sm text-muted-foreground">Storage</span>
                    <span class="font-mono text-sm font-medium text-foreground">{{ store.size }}</span>
                  </div>
                </div>
              </div>
            </section>

            <!-- Danger Zone -->
            <section class="mt-8">
              <div class="relative overflow-hidden rounded-lg border border-destructive/20">
                <div class="pattern-crosshatch pointer-events-none absolute inset-0 opacity-[0.04]" />
                <div class="relative border-b border-destructive/15 bg-destructive/[0.03] px-4 py-2.5">
                  <h3 class="text-xs font-semibold uppercase tracking-wider text-destructive/70">
                    Danger Zone
                  </h3>
                </div>
                <div class="relative px-4 py-4">
                  <div class="flex items-center justify-between">
                    <div>
                      <p class="text-sm font-medium text-foreground">Delete Redis</p>
                      <p class="text-xs text-muted-foreground">
                        Permanently delete this store and all its data.
                      </p>
                    </div>
                    <AlertDialog v-model:open="deleteDialogOpen">
                      <AlertDialogTrigger as-child>
                        <Button variant="destructive" size="sm">
                          <Trash2 :size="14" class="mr-1" />
                          Delete
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete Redis store "{{ store.name }}"?</AlertDialogTitle>
                          <AlertDialogDescription>
                            This will permanently delete the store and all its data.
                            This action cannot be undone.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            variant="destructive"
                            :disabled="deleting"
                            @click="handleDelete"
                          >
                            {{ deleting ? 'Deleting...' : 'Delete Redis' }}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                </div>
              </div>
            </section>
          </div>
        </TabsContent>
      </Tabs>
    </ScrollArea>
  </div>
</template>
