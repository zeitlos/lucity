<script setup lang="ts">
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import QueryError from '@/components/QueryError.vue';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { ref, computed, useTemplateRef } from 'vue';
import { useQuery, useMutation, useApolloClient } from '@vue/apollo-composable';
import { useDropZone } from '@vueuse/core';
import {
  Folder,
  File as FileIcon,
  Upload,
  Download,
  Trash2,
  FolderPlus,
  ChevronRight,
} from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import IconButton from '@/components/IconButton.vue';
import { Skeleton } from '@/components/ui/skeleton';
import { Progress } from '@/components/ui/progress';
import { Input } from '@/components/ui/input';
import { failureToast } from '@/components/ui/sonner';
import { formatBytes, formatDateTime } from '@/lib/format';
import EmptyState from '@/components/EmptyState.vue';

const BucketObjectsDocument = graphql(`
  query BucketObjects($bucket: BucketID!, $prefix: String) {
    bucketObjects(bucket: $bucket, prefix: $prefix) {
      prefix
      folders {
        prefix
      }
      objects {
        key
        size
        lastModified
      }
    }
  }
`);

const BucketObjectUploadUrlDocument = graphql(`
  mutation BucketObjectUploadUrl($bucket: BucketID!, $key: String!) {
    bucketObjectUploadUrl(bucket: $bucket, key: $key)
  }
`);

const BucketObjectDownloadUrlDocument = graphql(`
  query BucketObjectDownloadUrl($bucket: BucketID!, $key: String!) {
    bucketObjectDownloadUrl(bucket: $bucket, key: $key)
  }
`);

const DeleteBucketObjectDocument = graphql(`
  mutation DeleteBucketObject($bucket: BucketID!, $key: String!) {
    deleteBucketObject(bucket: $bucket, key: $key)
  }
`);

const props = defineProps<{
  bucketId: string;
}>();

const prefix = ref('');

const { result, loading, error, refetch } = useQuery(
  BucketObjectsDocument,
  () => ({ bucket: props.bucketId, prefix: prefix.value }),
  () => ({ enabled: !!props.bucketId, fetchPolicy: 'network-only' }),
);

const folders = computed(() => result.value?.bucketObjects.folders ?? []);
const objects = computed(() => result.value?.bucketObjects.objects ?? []);
const isEmpty = computed(() => folders.value.length === 0 && objects.value.length === 0);

const segments = computed(() => prefix.value.split('/').filter(Boolean));

function navigate(to: string) {
  prefix.value = to;
}

function openFolder(folderPrefix: string) {
  prefix.value = folderPrefix;
}

function crumbTarget(index: number): string {
  return segments.value.slice(0, index + 1).join('/') + '/';
}

function baseName(key: string): string {
  return key.split('/').filter(Boolean).pop() ?? key;
}

function folderName(folderPrefix: string): string {
  return baseName(folderPrefix);
}

// --- Upload -------------------------------------------------------------

const { mutate: mintUploadUrl } = useMutation(BucketObjectUploadUrlDocument);

const uploading = ref(false);
const uploadTotal = ref(0);
const uploadDone = ref(0);
const uploadPercent = ref(0);

const fileInput = useTemplateRef<HTMLInputElement>('fileInput');

function pickFiles() {
  fileInput.value?.click();
}

function onFilesPicked(event: Event) {
  const input = event.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  input.value = '';
  uploadFiles(files);
}

function putWithProgress(url: string, file: File, onProgress: (fraction: number) => void): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open('PUT', url);
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) onProgress(e.loaded / e.total);
    };
    xhr.onload = () =>
      xhr.status >= 200 && xhr.status < 300
        ? resolve()
        : reject(new Error(`upload failed (${xhr.status})`));
    xhr.onerror = () => reject(new Error('network error during upload'));
    xhr.send(file);
  });
}

async function uploadFiles(files: File[]) {
  if (files.length === 0 || uploading.value) return;

  uploading.value = true;
  uploadTotal.value = files.length;
  uploadDone.value = 0;
  uploadPercent.value = 0;

  try {
    for (const file of files) {
      const key = prefix.value + file.name;
      const response = await mintUploadUrl({ bucket: props.bucketId, key });
      const url = response?.data?.bucketObjectUploadUrl;

      if (!url) throw new Error('could not obtain an upload URL');

      await putWithProgress(url, file, (fraction) => {
        uploadPercent.value = Math.round(((uploadDone.value + fraction) / uploadTotal.value) * 100);
      });

      uploadDone.value += 1;
      uploadPercent.value = Math.round((uploadDone.value / uploadTotal.value) * 100);
    }

    await refetch();
  } catch (e: unknown) {
    failureToast('Failed to upload files', e);
  } finally {
    uploading.value = false;
  }
}

const dropZone = useTemplateRef<HTMLElement>('dropZone');
const { isOverDropZone } = useDropZone(dropZone, {
  onDrop: (files) => uploadFiles(files ?? []),
});

// --- New folder ---------------------------------------------------------

const creatingFolder = ref(false);
const savingFolder = ref(false);
const newFolderName = ref('');

function startFolder() {
  creatingFolder.value = true;
  newFolderName.value = '';
}

async function createFolder() {
  if (savingFolder.value) return;

  const name = newFolderName.value.trim().replace(/^\/+|\/+$/g, '');

  if (!name) {
    creatingFolder.value = false;
    return;
  }

  savingFolder.value = true;

  try {
    const key = `${prefix.value}${name}/`;
    const response = await mintUploadUrl({ bucket: props.bucketId, key });
    const url = response?.data?.bucketObjectUploadUrl;

    if (!url) throw new Error('could not obtain an upload URL');

    const put = await fetch(url, { method: 'PUT', body: '' });

    if (!put.ok) throw new Error(`create folder failed (${put.status})`);

    creatingFolder.value = false;
    newFolderName.value = '';
    await refetch();
  } catch (e: unknown) {
    failureToast('Failed to create folder', e);
  } finally {
    savingFolder.value = false;
  }
}

// --- Download -----------------------------------------------------------

const { resolveClient } = useApolloClient();

async function download(key: string) {
  try {
    const { data } = await resolveClient().query({
      query: BucketObjectDownloadUrlDocument,
      variables: { bucket: props.bucketId, key },
      fetchPolicy: 'network-only',
    });

    const anchor = document.createElement('a');
    anchor.href = data.bucketObjectDownloadUrl;
    anchor.rel = 'noopener';
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
  } catch (e: unknown) {
    failureToast('Failed to download file', e);
  }
}

// --- Delete -------------------------------------------------------------

const { mutate: deleteObject } = useMutation(DeleteBucketObjectDocument);
const deleteOpen = ref(false);
const deleteKey = ref<string | null>(null);
const deleting = ref(false);

function askDelete(key: string) {
  deleteKey.value = key;
  deleteOpen.value = true;
}

async function confirmDelete() {
  const key = deleteKey.value;

  if (!key) return;

  deleting.value = true;

  try {
    await deleteObject({ bucket: props.bucketId, key });
    deleteOpen.value = false;
    await refetch();
  } catch (e: unknown) {
    failureToast('Failed to delete file', e);
  } finally {
    deleting.value = false;
  }
}
</script>

<template>
  <div class="flex h-full flex-col space-y-3">
    <!-- Toolbar -->
    <div class="flex items-center justify-between gap-2">
      <nav class="flex min-w-0 items-center gap-1 text-sm">
        <button
          type="button"
          class="shrink-0 rounded px-1.5 py-0.5 font-medium text-muted-foreground transition-colors hover:bg-accent/50 hover:text-foreground"
          :class="{ 'text-foreground': segments.length === 0 }"
          @click="navigate('')"
        >
          /
        </button>
        <template v-for="(segment, index) in segments" :key="index">
          <ChevronRight :size="14" class="shrink-0 text-muted-foreground" />
          <button
            type="button"
            class="truncate rounded px-1.5 py-0.5 font-medium text-muted-foreground transition-colors hover:bg-accent/50 hover:text-foreground"
            :class="{ 'text-foreground': index === segments.length - 1 }"
            @click="navigate(crumbTarget(index))"
          >
            {{ segment }}
          </button>
        </template>
      </nav>

      <div class="flex shrink-0 items-center gap-1.5">
        <Button variant="outline" :disabled="uploading" @click="startFolder">
          <FolderPlus />
          New folder
        </Button>
        <Button variant="outline" :disabled="uploading" @click="pickFiles">
          <Upload />
          Upload
        </Button>
        <input ref="fileInput" type="file" multiple class="hidden" @change="onFilesPicked" />
      </div>
    </div>

    <!-- Upload progress -->
    <div v-if="uploading" class="space-y-1.5 rounded-lg border bg-muted/30 px-3 py-2.5">
      <div class="flex items-center justify-between text-xs text-muted-foreground">
        <span>Uploading {{ uploadDone + 1 > uploadTotal ? uploadTotal : uploadDone + 1 }} of {{ uploadTotal }}</span>
        <span>{{ uploadPercent }}%</span>
      </div>
      <Progress :model-value="uploadPercent" class="h-1.5" />
    </div>

    <!-- Drop zone -->
    <div
      ref="dropZone"
      class="relative flex-1 rounded-lg transition-colors"
      :class="{ 'bg-primary/5 ring-2 ring-inset ring-primary/40': isOverDropZone }"
    >
      <div
        v-if="isOverDropZone"
        class="pointer-events-none absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-background/80 backdrop-blur-sm"
      >
        <span
          class="flex items-center gap-2 rounded-md border border-primary/40 bg-background px-3 py-1.5 text-sm font-medium text-primary shadow-sm"
        >
          <Upload :size="14" />
          Drop to upload
        </span>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="space-y-1.5 py-1">
        <Skeleton v-for="i in 6" :key="i" class="h-7 w-full" />
      </div>

      <!-- Error -->
      <QueryError
        v-else-if="error"
        title="Couldn't load files"
        :error="error"
        @retry="refetch()"
      />

      <!-- Content -->
      <template v-else>
        <EmptyState
          v-if="isEmpty && !creatingFolder"
          size="compact"
          :icon="Folder"
          title="This folder is empty"
          description="Drop files here or use Upload."
        />

        <div v-else class="-mx-4 border-y">
          <Table class="table-fixed">
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead class="w-24 text-right">Size</TableHead>
                <TableHead class="w-32 text-right">Modified</TableHead>
                <TableHead class="w-20"><span class="sr-only">Actions</span></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="creatingFolder" class="hover:bg-transparent">
                <TableCell :colspan="4" class="py-1.5">
                  <div class="flex items-center gap-2">
                    <Folder :size="16" class="shrink-0 text-muted-foreground" />
                    <Input
                      v-model="newFolderName"
                      placeholder="Folder name"
                      class="h-7 flex-1"
                      autofocus
                      :disabled="savingFolder"
                      @keyup.enter="createFolder"
                      @keyup.esc="creatingFolder = false"
                    />
                    <Button variant="ghost" size="xs" :pending="savingFolder" @click="createFolder">
                      {{ savingFolder ? 'Creating…' : 'Create' }}
                    </Button>
                    <Button variant="ghost" size="xs" :disabled="savingFolder" @click="creatingFolder = false">Cancel</Button>
                  </div>
                </TableCell>
              </TableRow>

              <TableRow
                v-for="folder in folders"
                :key="folder.prefix"
                class="cursor-pointer"
                @click="openFolder(folder.prefix)"
              >
                <TableCell>
                  <button
                    type="button"
                    class="flex w-full min-w-0 items-center gap-3 text-left font-medium text-foreground focus-visible:outline-none"
                    @click.stop="openFolder(folder.prefix)"
                  >
                    <Folder :size="16" class="shrink-0 text-muted-foreground" />
                    <span class="truncate">{{ folderName(folder.prefix) }}</span>
                  </button>
                </TableCell>
                <TableCell />
                <TableCell />
                <TableCell>
                  <ChevronRight :size="14" class="ml-auto text-muted-foreground" />
                </TableCell>
              </TableRow>

              <TableRow v-for="object in objects" :key="object.key">
                <TableCell>
                  <div class="flex min-w-0 items-center gap-3">
                    <FileIcon :size="16" class="shrink-0 text-muted-foreground" />
                    <span class="truncate text-foreground">{{ baseName(object.key) }}</span>
                  </div>
                </TableCell>
                <TableCell class="text-right text-xs text-muted-foreground tabular-nums">
                  {{ formatBytes(object.size) }}
                </TableCell>
                <TableCell class="truncate text-right text-xs text-muted-foreground tabular-nums">
                  {{ formatDateTime(object.lastModified) }}
                </TableCell>
                <TableCell class="py-1">
                  <div class="flex items-center justify-end gap-0.5">
                    <IconButton label="Download" size="icon-xs" @click="download(object.key)">
                      <Download />
                    </IconButton>
                    <IconButton
                      label="Delete"
                      size="icon-xs"
                      class="text-muted-foreground hover:text-destructive"
                      @click="askDelete(object.key)"
                    >
                      <Trash2 />
                    </IconButton>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </template>
    </div>

    <!-- Delete confirmation -->
    <ConfirmDialog
      v-model:open="deleteOpen"
      :title="`Delete file &quot;${deleteKey ? baseName(deleteKey) : ''}&quot;?`"
      description="This can't be undone."
      confirm-label="Delete file"
      pending-label="Deleting…"
      :pending="deleting"
      destructive
      @confirm="confirmDelete"
    />
  </div>
</template>
