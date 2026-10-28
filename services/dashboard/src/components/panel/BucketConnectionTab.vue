<script setup lang="ts">
import CodeBlock from '@/components/CodeBlock.vue';
import CredentialField from '@/components/CredentialField.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
import CopyButton from '@/components/CopyButton.vue';
import QueryError from '@/components/QueryError.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { computed, ref } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Globe } from '@lucide/vue';
import { graphql } from '@/gql';
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
    { key: 'accessKeyId', label: 'Access key ID', value: c.accessKeyId, sensitive: false },
    { key: 'secretAccessKey', label: 'Secret access key', value: c.secretAccessKey, sensitive: true },
  ];
});

const SNIPPETS = [
  { value: 'env' as const, label: 'Env' },
  { value: 'cli' as const, label: 'AWS CLI' },
  { value: 'rclone' as const, label: 'rclone' },
  { value: 'node' as const, label: 'Node' },
  { value: 'python' as const, label: 'Python' },
];

type SnippetKey = (typeof SNIPPETS)[number]['value'];

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
        <CredentialField
          v-for="field in fields"
          :key="field.key"
          :label="field.label"
          :value="field.value"
          :sensitive="field.sensitive"
        />
      </div>

      <CodeBlock :code="currentSnippet" :secrets="[credentials.secretAccessKey]">
        <template #header>
          <SegmentedControl v-model="selectedSnippet" label="Snippet" :options="SNIPPETS" />
        </template>
      </CodeBlock>

      <p class="text-sm text-muted-foreground">
        S3-compatible credentials for <strong>{{ bucketName }}</strong>. Works with any S3 SDK, the AWS CLI, or rclone.
      </p>
    </template>
  </div>
</template>
