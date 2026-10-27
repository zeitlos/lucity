<script setup lang="ts">
import CopyButton from '@/components/CopyButton.vue';
import QueryError from '@/components/QueryError.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { reactive, computed, ref } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Eye, EyeOff, Globe } from '@lucide/vue';
import { graphql } from '@/gql';
import IconButton from '@/components/IconButton.vue';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { failureToast } from '@/components/ui/sonner';

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

const { result, loading, error, refetch } = useQuery(
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
  } catch (e: unknown) {
    failureToast('Failed to update public access', e);
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

function buildSnippets(secretAccessKey: string): Record<SnippetKey, string> {
  const c = credentials.value;
  if (!c) return { env: '', cli: '', rclone: '', node: '', python: '' };
  return {
    env: [
      `S3_ENDPOINT=${c.endpoint}`,
      `S3_REGION=${c.region}`,
      `S3_BUCKET=${c.bucket}`,
      `S3_ACCESS_KEY_ID=${c.accessKeyId}`,
      `S3_SECRET_ACCESS_KEY=${secretAccessKey}`,
      'S3_FORCE_PATH_STYLE=true',
    ].join('\n'),
    cli: `aws s3 --endpoint-url ${c.endpoint} ls s3://${c.bucket}`,
    rclone: [
      'rclone config create lucity s3 \\',
      '  provider=Other \\',
      `  endpoint=${c.endpoint} \\`,
      `  region=${c.region} \\`,
      `  access_key_id=${c.accessKeyId} \\`,
      `  secret_access_key=${secretAccessKey}`,
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
      `    secretAccessKey: '${secretAccessKey}',`,
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
      `    aws_secret_access_key='${secretAccessKey}',`,
      ')',
      '',
      `s3.list_objects_v2(Bucket='${c.bucket}')`,
    ].join('\n'),
  };
}

const currentSnippet = computed(() => buildSnippets(credentials.value?.secretAccessKey ?? '')[selectedSnippet.value]);

const revealed = reactive<Record<string, boolean>>({});

const displayedSnippet = computed(() => {
  const secretAccessKey = credentials.value?.secretAccessKey ?? '';
  return buildSnippets(revealed.secretAccessKey ? secretAccessKey : mask(secretAccessKey))[selectedSnippet.value];
});

function toggleReveal(key: string) {
  revealed[key] = !revealed[key];
}

function mask(value: string): string {
  if (value.length <= 4) return '*'.repeat(value.length);
  return '*'.repeat(value.length - 2) + value.slice(-2);
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
        <Spinner v-if="publicSaving" :size="14" class="text-muted-foreground" />
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
          <CopyButton label="Copy public URL" size="icon-xs" :value="props.publicEndpoint" />
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
    <QueryError
      v-else-if="error"
      title="Couldn't load the connection details"
      :error="error"
      @retry="refetch()"
    />

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
          <IconButton
            :label="`${revealed[field.key] ? 'Hide' : 'Show'} ${field.label}`"
            v-if="field.sensitive"
            size="icon-xs"
            class="shrink-0"
            @click="toggleReveal(field.key)"
          >
            <EyeOff v-if="revealed[field.key]" />
            <Eye v-else />
          </IconButton>
          <CopyButton
            :label="`Copy ${field.label}`"
            size="icon-xs"
            class="shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
            :value="field.value"
          />
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
          <CopyButton label="Copy snippet" size="icon-xs" :value="currentSnippet" />
        </div>
        <pre class="overflow-x-auto rounded-md bg-muted/40 px-3 py-2 font-mono text-xs text-foreground">{{ displayedSnippet }}</pre>
      </div>

      <p class="text-sm text-muted-foreground">
        S3-compatible credentials for <strong>{{ bucketName }}</strong>. Works with any S3 SDK, the AWS CLI, or rclone.
      </p>
    </template>
  </div>
</template>
