<script setup lang="ts">
import { reactive, computed, ref } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Copy, Eye, EyeOff, Globe } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

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

const SetBucketPublicDocument = graphql(`
  mutation SetBucketPublic($bucket: BucketID!, $public: Boolean!) {
    setBucketPublic(bucket: $bucket, public: $public) {
      id
      public
      publicEndpoint
    }
  }
`);

const props = defineProps<{
  bucketId: string;
  bucketName: string;
  public: boolean;
  publicEndpoint: string;
}>();

const { result, loading, error } = useQuery(
  BucketCredentialsDocument,
  () => ({ bucket: props.bucketId }),
  () => ({ enabled: !!props.bucketId }),
);

const credentials = computed(() => result.value?.bucketCredentials ?? null);

const { mutate: setBucketPublic } = useMutation(SetBucketPublicDocument);
const publicSaving = ref(false);

async function handleTogglePublic(value: boolean) {
  publicSaving.value = true;

  try {
    await setBucketPublic({ bucket: props.bucketId, public: value });
    toast.success(value ? 'Bucket is now public' : 'Bucket is now private');
  } catch (e: unknown) {
    errorToast('Failed to update public access', { description: errorMessage(e) });
  } finally {
    publicSaving.value = false;
  }
}

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

const SNIPPETS = [
  { key: 'env', label: 'Env' },
  { key: 'cli', label: 'AWS CLI' },
  { key: 'rclone', label: 'rclone' },
  { key: 'node', label: 'Node' },
  { key: 'python', label: 'Python' },
] as const;

type SnippetKey = (typeof SNIPPETS)[number]['key'];

const selectedSnippet = ref<SnippetKey>('env');

const snippets = computed<Record<SnippetKey, string>>(() => {
  const c = credentials.value;
  if (!c) return { env: '', cli: '', rclone: '', node: '', python: '' };
  return {
    env: [
      `S3_ENDPOINT=${c.endpoint}`,
      `S3_REGION=${c.region}`,
      `S3_BUCKET=${c.bucket}`,
      `S3_ACCESS_KEY_ID=${c.accessKeyId}`,
      `S3_SECRET_ACCESS_KEY=${c.secretAccessKey}`,
      'S3_FORCE_PATH_STYLE=true',
    ].join('\n'),
    cli: `aws s3 --endpoint-url ${c.endpoint} ls s3://${c.bucket}`,
    rclone: [
      'rclone config create lucity s3 \\',
      '  provider=Other \\',
      `  endpoint=${c.endpoint} \\`,
      `  region=${c.region} \\`,
      `  access_key_id=${c.accessKeyId} \\`,
      `  secret_access_key=${c.secretAccessKey}`,
      '',
      `rclone ls lucity:${c.bucket}`,
    ].join('\n'),
    node: [
      "import { S3Client, ListObjectsV2Command } from '@aws-sdk/client-s3';",
      '',
      'const s3 = new S3Client({',
      `  endpoint: '${c.endpoint}',`,
      `  region: '${c.region}',`,
      '  forcePathStyle: true,',
      '  credentials: {',
      `    accessKeyId: '${c.accessKeyId}',`,
      `    secretAccessKey: '${c.secretAccessKey}',`,
      '  },',
      '});',
      '',
      `await s3.send(new ListObjectsV2Command({ Bucket: '${c.bucket}' }));`,
    ].join('\n'),
    python: [
      'import boto3',
      '',
      's3 = boto3.client(',
      "    's3',",
      `    endpoint_url='${c.endpoint}',`,
      `    region_name='${c.region}',`,
      `    aws_access_key_id='${c.accessKeyId}',`,
      `    aws_secret_access_key='${c.secretAccessKey}',`,
      ')',
      '',
      `s3.list_objects_v2(Bucket='${c.bucket}')`,
    ].join('\n'),
  };
});

const currentSnippet = computed(() => snippets.value[selectedSnippet.value]);

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
    <!-- Public access -->
    <div class="rounded-lg border">
      <div class="flex items-start gap-3 px-4 py-3">
        <Globe :size="16" class="mt-0.5 shrink-0 text-muted-foreground" />
        <div class="flex-1 space-y-0.5">
          <p class="text-sm font-medium text-foreground">Public read</p>
          <p class="text-xs text-muted-foreground">
            Serve this bucket's objects read-only over a public edge URL.
          </p>
        </div>
        <Switch
          :model-value="props.public"
          :disabled="publicSaving"
          class="data-[state=unchecked]:bg-border"
          @update:model-value="handleTogglePublic"
        />
      </div>
      <div v-if="props.public && props.publicEndpoint" class="space-y-1.5 border-t px-4 py-3">
        <div class="flex items-center justify-between px-1">
          <span class="text-xs font-medium text-muted-foreground">Public URL</span>
          <Button variant="ghost" size="icon" class="h-6 w-6" @click="copyToClipboard(props.publicEndpoint)">
            <Copy :size="12" />
          </Button>
        </div>
        <div class="flex items-center gap-2 rounded-md bg-muted/40 px-3 py-2">
          <span class="flex-1 truncate font-mono text-xs text-foreground">{{ props.publicEndpoint }}</span>
        </div>
        <p class="px-1 text-xs text-muted-foreground">
          Objects are readable anonymously over the edge. Writes still use the S3 credentials below.
        </p>
      </div>
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

      <!-- Snippet switcher -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between px-1">
          <div class="flex gap-0.5 rounded-md bg-muted/40 p-0.5">
            <button
              v-for="snip in SNIPPETS"
              :key="snip.key"
              type="button"
              class="rounded px-2 py-1 text-sm font-medium transition-colors"
              :class="
                selectedSnippet === snip.key
                  ? 'bg-background text-foreground shadow-sm ring-1 ring-border'
                  : 'text-muted-foreground hover:text-foreground'
              "
              @click="selectedSnippet = snip.key"
            >
              {{ snip.label }}
            </button>
          </div>
          <Button variant="ghost" size="icon" class="h-6 w-6" @click="copyToClipboard(currentSnippet)">
            <Copy :size="12" />
          </Button>
        </div>
        <pre class="overflow-x-auto rounded-md bg-muted/40 px-3 py-2 font-mono text-xs text-foreground">{{ currentSnippet }}</pre>
      </div>

      <p class="text-sm text-muted-foreground">
        S3-compatible credentials for <strong>{{ bucketName }}</strong>. Works with any S3 SDK, the AWS CLI, or rclone.
      </p>
    </template>
  </div>
</template>
