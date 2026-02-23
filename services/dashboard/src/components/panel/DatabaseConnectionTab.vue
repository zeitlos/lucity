<script setup lang="ts">
import { reactive, computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { Copy, Eye, EyeOff, Loader2, DatabaseZap } from 'lucide-vue-next';
import { DatabaseCredentialsQuery } from '@/graphql/databases';
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from '@/components/ui/sonner';

const props = defineProps<{
  projectId: string;
  database: {
    name: string;
  };
}>();

const { activeEnvironment } = useEnvironment();

const queryEnabled = computed(() => !!activeEnvironment.value);
const queryVars = computed(() => ({
  projectId: props.projectId,
  environment: activeEnvironment.value?.name ?? '',
  database: props.database.name,
}));

const { result, loading, error } = useQuery(
  DatabaseCredentialsQuery,
  queryVars,
  () => ({ enabled: queryEnabled.value }),
);

const creds = computed(() => result.value?.databaseCredentials ?? null);

const isProvisioning = computed(() => {
  if (!error.value) return false;
  const gqlErrors = (error.value as { graphQLErrors?: { extensions?: { code?: string } }[] }).graphQLErrors;
  return gqlErrors?.some(e => e.extensions?.code === 'DATABASE_PROVISIONING') ?? false;
});

// Track which fields are revealed
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

const fields = computed(() => {
  if (!creds.value) return [];
  return [
    { key: 'uri', label: 'DATABASE_URL', value: creds.value.uri, sensitive: true },
    { key: 'host', label: 'Host', value: creds.value.host, sensitive: false },
    { key: 'port', label: 'Port', value: creds.value.port, sensitive: false },
    { key: 'dbname', label: 'Database', value: creds.value.dbname, sensitive: false },
    { key: 'user', label: 'User', value: creds.value.user, sensitive: false },
    { key: 'password', label: 'Password', value: creds.value.password, sensitive: true },
  ];
});
</script>

<template>
  <div class="space-y-4">
    <div>
      <h3 class="text-sm font-medium text-foreground">Connection Details</h3>
      <p class="text-xs text-muted-foreground">
        Credentials for <strong>{{ database.name }}</strong> in {{ activeEnvironment?.name ?? 'this environment' }}.
      </p>
    </div>

    <!-- No environment selected -->
    <div
      v-if="!activeEnvironment"
      class="flex flex-col items-center justify-center gap-2 py-12 text-center"
    >
      <p class="text-sm text-muted-foreground">Select an environment to view connection details.</p>
    </div>

    <!-- Loading -->
    <div v-else-if="loading" class="space-y-2">
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
    <div v-else-if="creds" class="space-y-1.5">
      <div
        v-for="field in fields"
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
</template>
