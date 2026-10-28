<script setup lang="ts">
import CredentialField from '@/components/CredentialField.vue';
import EmptyState from '@/components/EmptyState.vue';
import QueryError from '@/components/QueryError.vue';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { computed, ref } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { DatabaseZap, Globe, Lock, ShieldAlert } from '@lucide/vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { graphql } from '@/gql';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import DatabaseAllowRules from '@/components/panel/DatabaseAllowRules.vue';
import { failureToast } from '@/components/ui/sonner';

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
  } catch (e: unknown) {
    failureToast(next ? 'Failed to expose database' : 'Failed to make database private', e);
  }
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
    <Skeleton v-if="!dbResult" class="h-16 w-full" />
    <div v-else class="rounded-lg border">
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
        <Spinner v-if="toggling" :size="14" class="text-muted-foreground" />
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
        <CredentialField
          v-for="field in publicGroup.fields"
          :key="field.key"
          :label="field.label"
          :value="field.value"
          :sensitive="field.sensitive"
        />
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-2">
      <Skeleton v-for="i in 6" :key="i" class="h-10 w-full" />
    </div>

    <!-- Provisioning -->
    <EmptyState
      v-else-if="isProvisioning"
      size="compact"
      :icon="DatabaseZap"
      title="Database is provisioning"
      description="Credentials appear once PostgreSQL is ready."
    >
      <template #action>
        <Spinner :size="16" class="text-muted-foreground" />
      </template>
    </EmptyState>

    <!-- Error (non-provisioning) -->
    <QueryError
      v-else-if="error && !isProvisioning"
      title="Couldn't load the connection details"
      :error="error"
      @retry="refetchCredentials()"
    />

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
          <CredentialField
            v-for="field in group.fields"
            :key="field.key"
            :label="field.label"
            :value="field.value"
            :sensitive="field.sensitive"
          />
        </div>
      </div>
    </template>

    <!-- Expose confirmation -->
    <ConfirmDialog
      v-model:open="exposeDialogOpen"
      :title="`Expose &quot;${databaseName}&quot; to the internet?`"
      description="It gets a public hostname on port 5432, and only the addresses you allow can connect. Connections stay encrypted end to end. You can change who is allowed at any time."
      confirm-label="Expose database"
      pending-label="Exposing…"
      :pending="toggling"
      :icon="ShieldAlert"
      @confirm="setExposed(true)"
    >
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
    </ConfirmDialog>
  </div>
</template>
