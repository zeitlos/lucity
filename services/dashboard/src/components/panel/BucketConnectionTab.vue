<script setup lang="ts">
import { reactive, computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { Copy, Eye, EyeOff } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from '@/components/ui/sonner';

const BucketCredentialsDocument = graphql(`
  query BucketCredentials($bucket: BucketID!) {
    bucketCredentials(bucket: $bucket) {
      endpoint
      region
      bucket
      accessKeyId
      secretAccessKey
    }
  }
`);

const props = defineProps<{
  bucketId: string;
  bucketName: string;
}>();

const { result, loading, error } = useQuery(
  BucketCredentialsDocument,
  () => ({ bucket: props.bucketId }),
  () => ({ enabled: !!props.bucketId }),
);

const credentials = computed(() => result.value?.bucketCredentials ?? null);

const fields = computed(() => {
  const c = credentials.value;
  if (!c) return [];
  return [
    { key: 'endpoint', label: 'Endpoint', value: c.endpoint, sensitive: false },
    { key: 'region', label: 'Region', value: c.region, sensitive: false },
    { key: 'bucket', label: 'Bucket', value: c.bucket, sensitive: false },
    { key: 'accessKeyId', label: 'Access Key ID', value: c.accessKeyId, sensitive: false },
    { key: 'secretAccessKey', label: 'Secret Access Key', value: c.secretAccessKey, sensitive: true },
  ];
});

const envBlock = computed(() => {
  const c = credentials.value;
  if (!c) return '';
  return [
    `S3_ENDPOINT=${c.endpoint}`,
    `S3_REGION=${c.region}`,
    `S3_BUCKET=${c.bucket}`,
    `S3_ACCESS_KEY_ID=${c.accessKeyId}`,
    `S3_SECRET_ACCESS_KEY=${c.secretAccessKey}`,
    'S3_FORCE_PATH_STYLE=true',
  ].join('\n');
});

const cliExample = computed(() => {
  const c = credentials.value;
  if (!c) return '';
  return `aws s3 --endpoint-url ${c.endpoint} ls s3://${c.bucket}`;
});

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
</script>

<template>
  <div class="space-y-4">
    <div>
      <h3 class="text-sm font-medium text-foreground">Connection Details</h3>
      <p class="text-xs text-muted-foreground">
        S3-compatible credentials for <strong>{{ bucketName }}</strong>. Works with any S3 SDK, the AWS CLI, or rclone.
      </p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-2">
      <Skeleton v-for="i in 5" :key="i" class="h-10 w-full" />
    </div>

    <!-- Error -->
    <div
      v-else-if="error"
      class="rounded-lg border border-destructive/30 bg-destructive/5 p-3"
    >
      <p class="font-mono text-xs text-destructive">{{ error.message }}</p>
    </div>

    <!-- Credentials -->
    <template v-else-if="credentials">
      <div class="space-y-1.5">
        <div
          v-for="field in fields"
          :key="field.key"
          class="group flex items-center gap-2 rounded-md bg-muted/40 px-3 py-2"
        >
          <span class="w-32 shrink-0 text-xs font-medium text-muted-foreground">{{ field.label }}</span>
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

      <!-- Environment variables -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between px-1">
          <span class="text-xs font-medium text-muted-foreground">Environment variables</span>
          <Button variant="ghost" size="icon" class="h-6 w-6" @click="copyToClipboard(envBlock)">
            <Copy :size="12" />
          </Button>
        </div>
        <pre class="overflow-x-auto rounded-md bg-muted/40 px-3 py-2 font-mono text-xs text-foreground">{{ envBlock }}</pre>
      </div>

      <!-- AWS CLI -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between px-1">
          <span class="text-xs font-medium text-muted-foreground">AWS CLI</span>
          <Button variant="ghost" size="icon" class="h-6 w-6" @click="copyToClipboard(cliExample)">
            <Copy :size="12" />
          </Button>
        </div>
        <pre class="overflow-x-auto rounded-md bg-muted/40 px-3 py-2 font-mono text-xs text-foreground">{{ cliExample }}</pre>
      </div>
    </template>
  </div>
</template>
