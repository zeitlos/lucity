<script setup lang="ts">
import { reactive, computed, ref } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Copy, Eye, EyeOff, Loader2, DatabaseZap, Globe, ShieldAlert } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const DatabasePublicDocument = graphql(`
  query DatabasePublic($database: DatabaseID!) {
    database(id: $database) {
      id
      public
    }
  }
`);

const DatabaseCredentialsDocument = graphql(`
  query DatabaseCredentials($database: DatabaseID!) {
    databaseCredentials(database: $database) {
      type
      host
      port
      dbname
      user
      password
      uri
    }
  }
`);

const ExposeDatabaseDocument = graphql(`
  mutation ExposeDatabase($database: DatabaseID!) {
    exposeDatabase(database: $database) {
      id
      public
    }
  }
`);

const UnexposeDatabaseDocument = graphql(`
  mutation UnexposeDatabase($database: DatabaseID!) {
    unexposeDatabase(database: $database) {
      id
      public
    }
  }
`);

const props = defineProps<{
  databaseId: string;
  databaseName: string;
}>();

const { result: dbResult, refetch: refetchDatabase } = useQuery(
  DatabasePublicDocument,
  () => ({ database: props.databaseId }),
  () => ({ enabled: !!props.databaseId }),
);

const isPublic = computed(() => dbResult.value?.database?.public ?? false);

const { result, loading, error, refetch: refetchCredentials } = useQuery(
  DatabaseCredentialsDocument,
  () => ({ database: props.databaseId }),
  () => ({ enabled: !!props.databaseId }),
);

const credentials = computed(() => result.value?.databaseCredentials ?? []);

const isProvisioning = computed(() => {
  if (!error.value) return false;
  const gqlErrors = (error.value as unknown as { graphQLErrors?: { extensions?: { code?: string } }[] }).graphQLErrors;
  return gqlErrors?.some(e => e.extensions?.code === 'DATABASE_PROVISIONING') ?? false;
});

const { mutate: exposeDatabase, loading: exposing } = useMutation(ExposeDatabaseDocument);
const { mutate: unexposeDatabase, loading: unexposing } = useMutation(UnexposeDatabaseDocument);
const toggling = computed(() => exposing.value || unexposing.value);

const exposeDialogOpen = ref(false);

function onToggle(next: boolean) {
  if (next) {
    exposeDialogOpen.value = true;
  } else {
    void setExposed(false);
  }
}

async function setExposed(next: boolean) {
  try {
    if (next) {
      await exposeDatabase({ database: props.databaseId });
    } else {
      await unexposeDatabase({ database: props.databaseId });
    }
    exposeDialogOpen.value = false;
    await Promise.all([refetchDatabase(), refetchCredentials()]);
    toast.success(next ? 'Database exposed to the internet' : 'Database is now private');
  } catch (e: unknown) {
    errorToast(next ? 'Failed to expose database' : 'Failed to make database private', {
      description: errorMessage(e),
    });
  }
}

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

const endpointLabels: Record<string, string> = {
  INTERNAL: 'Internal (in-cluster)',
  PLATFORM: 'Public (internet)',
  CUSTOM: 'Custom domain',
};

const groups = computed(() =>
  credentials.value.map(cred => ({
    type: cred.type,
    label: endpointLabels[cred.type] ?? cred.type,
    fields: [
      { key: `${cred.type}-uri`, label: 'DATABASE_URL', value: cred.uri, sensitive: true },
      { key: `${cred.type}-host`, label: 'Host', value: cred.host, sensitive: false },
      { key: `${cred.type}-port`, label: 'Port', value: cred.port, sensitive: false },
      { key: `${cred.type}-dbname`, label: 'Database', value: cred.dbname, sensitive: false },
      { key: `${cred.type}-user`, label: 'User', value: cred.user, sensitive: false },
      { key: `${cred.type}-password`, label: 'Password', value: cred.password, sensitive: true },
    ],
  })),
);
</script>

<template>
  <div class="space-y-4">
    <div>
      <h3 class="text-sm font-medium text-foreground">Connection Details</h3>
      <p class="text-xs text-muted-foreground">
        Credentials for <strong>{{ databaseName }}</strong>.
      </p>
    </div>

    <!-- Public access toggle -->
    <div class="flex items-start gap-3 rounded-lg border px-4 py-3">
      <Globe :size="16" class="mt-0.5 shrink-0 text-muted-foreground" />
      <div class="flex-1 space-y-0.5">
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium text-foreground">Internet access</span>
          <Badge v-if="isPublic" variant="secondary" class="text-[10px]">Public</Badge>
        </div>
        <p class="text-xs text-muted-foreground">
          {{ isPublic
            ? 'Reachable from the internet over TLS on port 5432.'
            : 'Only reachable from inside the cluster.' }}
        </p>
      </div>
      <Switch
        :model-value="isPublic"
        :disabled="toggling"
        @update:model-value="onToggle"
      />
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-2">
      <Skeleton v-for="i in 6" :key="i" class="h-10 w-full" />
    </div>

    <!-- Provisioning -->
    <div
      v-else-if="isProvisioning"
      class="flex flex-col items-center justify-center gap-3 py-12 text-center"
    >
      <DatabaseZap :size="24" class="text-muted-foreground" />
      <div class="space-y-1">
        <p class="text-sm font-medium">Database is provisioning</p>
        <p class="text-xs text-muted-foreground">Credentials will appear once PostgreSQL is ready.</p>
      </div>
      <Loader2 :size="16" class="animate-spin text-muted-foreground" />
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

    <!-- Expose confirmation -->
    <AlertDialog v-model:open="exposeDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle class="flex items-center gap-2">
            <ShieldAlert :size="18" class="text-destructive" />
            Expose "{{ databaseName }}" to the internet?
          </AlertDialogTitle>
          <AlertDialogDescription>
            This database will be reachable from anywhere on the internet on port 5432.
            Connections stay encrypted end-to-end (TLS terminates at the database), but
            anyone with the credentials can connect. Make sure a strong password is in place.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction :disabled="exposing" @click="setExposed(true)">
            {{ exposing ? 'Exposing...' : 'Expose database' }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
