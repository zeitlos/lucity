<script setup lang="ts">
import { reactive, computed, ref } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Copy, Eye, EyeOff, DatabaseZap, Globe, Lock, ShieldAlert } from '@lucide/vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import DatabaseAllowRules from '@/components/panel/DatabaseAllowRules.vue';
import {
  AlertDialog,
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
      allowRules {
        range
        description
      }
    }
    clientAddress
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

const AllowOnExposeDocument = graphql(`
  mutation AllowOnExpose($database: DatabaseID!, $rule: DatabaseAllowRuleInput!) {
    addDatabaseAllowRule(database: $database, rule: $rule) {
      id
      allowRules {
        range
        description
      }
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
const allowRules = computed(() => dbResult.value?.database?.allowRules ?? []);
const clientAddress = computed(() => dbResult.value?.clientAddress ?? null);

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
const { mutate: allowOnExpose, loading: allowingOnExpose } = useMutation(AllowOnExposeDocument);
const toggling = computed(() => exposing.value || unexposing.value || allowingOnExpose.value);

const exposeDialogOpen = ref(false);
const exposeAccess = ref<'client' | 'any' | 'none'>('none');

function onToggle(next: boolean) {
  if (next) {
    exposeAccess.value = clientAddress.value ? 'client' : 'none';
    exposeDialogOpen.value = true;
  } else {
    void setExposed(false);
  }
}

async function setExposed(next: boolean) {
  try {
    if (next) {
      await exposeDatabase({ database: props.databaseId });
      if (exposeAccess.value === 'client' && clientAddress.value) {
        await allowOnExpose({
          database: props.databaseId,
          rule: { range: clientAddress.value, description: 'Added from the dashboard' },
        });
      } else if (exposeAccess.value === 'any') {
        await allowOnExpose({
          database: props.databaseId,
          rule: { range: '0.0.0.0/0', description: 'Any address' },
        });
      }
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
  INTERNAL: 'Private network',
  PLATFORM: 'Public internet',
  CUSTOM: 'Custom domain',
};

const endpointDescriptions: Record<string, string> = {
  INTERNAL: 'Reachable only from your other services over the private network.',
  CUSTOM: 'Reachable over your custom domain.',
};

const groups = computed(() =>
  credentials.value.map(cred => ({
    type: cred.type,
    label: endpointLabels[cred.type] ?? cred.type,
    description: endpointDescriptions[cred.type] ?? '',
    fields: [
      { key: `${cred.type}-uri`, label: 'Connection URL', value: cred.uri, sensitive: true },
      { key: `${cred.type}-host`, label: 'Host', value: cred.host, sensitive: false },
      { key: `${cred.type}-port`, label: 'Port', value: cred.port, sensitive: false },
      { key: `${cred.type}-dbname`, label: 'Database', value: cred.dbname, sensitive: false },
      { key: `${cred.type}-user`, label: 'User', value: cred.user, sensitive: false },
      { key: `${cred.type}-password`, label: 'Password', value: cred.password, sensitive: true },
    ],
  })),
);

const publicGroup = computed(() => groups.value.find(g => g.type === 'PLATFORM'));
const privateGroups = computed(() => groups.value.filter(g => g.type !== 'PLATFORM'));
</script>

<template>
  <div class="space-y-4">
    <!-- Internet access card -->
    <div class="rounded-lg border">
      <div class="flex items-start gap-3 px-4 py-3">
        <Globe :size="16" class="mt-0.5 shrink-0 text-muted-foreground" />
        <div class="flex-1 space-y-0.5">
          <span class="text-sm font-medium text-foreground">Internet access</span>
          <p class="text-xs text-muted-foreground">
            {{ isPublic
              ? 'Reachable over TLS on port 5432 from the addresses you allow.'
              : 'Reachable only from your other services over the private network.' }}
          </p>
        </div>
        <Switch
          :model-value="isPublic"
          :disabled="toggling"
          @update:model-value="onToggle"
        />
      </div>
      <div v-if="isPublic" class="border-t px-4 py-3">
        <DatabaseAllowRules
          :database-id="databaseId"
          :rules="allowRules"
          :client-address="clientAddress"
        />
      </div>
      <div v-if="isPublic && publicGroup" class="space-y-1.5 border-t px-4 py-3">
        <div
          v-for="field in publicGroup.fields"
          :key="field.key"
          class="group flex items-start gap-2 rounded-md bg-muted/40 px-3 py-2"
        >
          <span class="w-28 shrink-0 pt-1 text-xs font-medium text-muted-foreground">{{ field.label }}</span>
          <span class="min-w-0 flex-1 break-all pt-0.5 font-mono text-xs text-foreground">
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
      <Spinner :size="16" class="animate-spin text-muted-foreground" />
    </div>

    <!-- Error (non-provisioning) -->
    <div
      v-else-if="error && !isProvisioning"
      class="rounded-lg border border-destructive/30 bg-destructive/5 p-3"
    >
      <p class="font-mono text-xs text-destructive">{{ error.message }}</p>
    </div>

    <!-- Private network / custom domain cards -->
    <template v-else>
      <div v-for="group in privateGroups" :key="group.type" class="rounded-lg border">
        <div class="flex items-start gap-3 px-4 py-3">
          <Lock :size="16" class="mt-0.5 shrink-0 text-muted-foreground" />
          <div class="flex-1 space-y-0.5">
            <span class="text-sm font-medium text-foreground">{{ group.label }}</span>
            <p v-if="group.description" class="text-xs text-muted-foreground">{{ group.description }}</p>
          </div>
        </div>
        <div class="space-y-1.5 border-t px-4 py-3">
          <div
            v-for="field in group.fields"
            :key="field.key"
            class="group flex items-start gap-2 rounded-md bg-muted/40 px-3 py-2"
          >
            <span class="w-28 shrink-0 pt-1 text-xs font-medium text-muted-foreground">{{ field.label }}</span>
            <span class="min-w-0 flex-1 break-all pt-0.5 font-mono text-xs text-foreground">
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
    </template>

    <!-- Expose confirmation -->
    <AlertDialog v-model:open="exposeDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle class="flex items-center gap-2">
            <ShieldAlert :size="18" class="text-destructive" />
            Expose "{{ databaseName }}" to the internet?
          </AlertDialogTitle>
          <AlertDialogDescription>
            This database gets a public hostname on port 5432, and only the addresses you
            allow can connect. Connections stay encrypted end-to-end (TLS terminates at the
            database). You can change who is allowed at any time.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <RadioGroup v-model="exposeAccess" class="space-y-2">
          <label
            v-if="clientAddress"
            class="flex cursor-pointer items-start gap-2 rounded-lg border p-3 transition-colors"
            :class="exposeAccess === 'client' ? 'border-primary bg-primary/5' : 'border-border'"
          >
            <RadioGroupItem value="client" class="mt-0.5" />
            <div class="space-y-0.5">
              <span class="text-sm font-medium">Only this computer</span>
              <p class="text-xs text-muted-foreground">Allow {{ clientAddress }}, the address you're using right now.</p>
            </div>
          </label>
          <label
            class="flex cursor-pointer items-start gap-2 rounded-lg border p-3 transition-colors"
            :class="exposeAccess === 'any' ? 'border-primary bg-primary/5' : 'border-border'"
          >
            <RadioGroupItem value="any" class="mt-0.5" />
            <div class="space-y-0.5">
              <span class="text-sm font-medium">Anyone on the internet</span>
              <p class="text-xs text-muted-foreground">Only the database password keeps others out.</p>
            </div>
          </label>
          <label
            class="flex cursor-pointer items-start gap-2 rounded-lg border p-3 transition-colors"
            :class="exposeAccess === 'none' ? 'border-primary bg-primary/5' : 'border-border'"
          >
            <RadioGroupItem value="none" class="mt-0.5" />
            <div class="space-y-0.5">
              <span class="text-sm font-medium">Nobody yet</span>
              <p class="text-xs text-muted-foreground">Add the addresses that may connect afterwards.</p>
            </div>
          </label>
        </RadioGroup>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="toggling">Cancel</AlertDialogCancel>
          <Button :disabled="toggling" @click="setExposed(true)">
            {{ toggling ? 'Exposing...' : 'Expose database' }}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
