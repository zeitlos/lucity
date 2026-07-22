<script setup lang="ts">
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
import { Skeleton } from '@/components/ui/skeleton';
import { Progress } from '@/components/ui/progress';
import { Input } from '@/components/ui/input';
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { errorToast } from '@/components/ui/sonner';
import { errorMessage, formatBytes } from '@/lib/utils';

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

function formatDate(value: string): string {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) return '';

  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
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
    errorToast('Upload failed', { description: errorMessage(e) });
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
const newFolderName = ref('');

function startFolder() {
  creatingFolder.value = true;
  newFolderName.value = '';
}

async function createFolder() {
  const name = newFolderName.value.trim().replace(/^\/+|\/+$/g, '');

  if (!name) {
    creatingFolder.value = false;
    return;
  }

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
    errorToast('Could not create folder', { description: errorMessage(e) });
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
    errorToast('Download failed', { description: errorMessage(e) });
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
    errorToast('Delete failed', { description: errorMessage(e) });
  } finally {
    deleting.value = false;
  }
}
</script>

<template>
  <div class="space-y-3">
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
        <Button variant="outline" size="sm" :disabled="uploading" @click="startFolder">
          <FolderPlus :size="14" class="mr-1" />
          New folder
        </Button>
        <Button variant="outline" size="sm" :disabled="uploading" @click="pickFiles">
          <Upload :size="14" class="mr-1" />
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
      class="relative rounded-lg transition-colors"
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
      <div v-else-if="error" class="rounded-lg border border-destructive/30 bg-destructive/5 p-3">
        <p class="font-mono text-xs text-destructive">{{ error.message }}</p>
      </div>

      <!-- Content -->
      <template v-else>
        <!-- New folder row -->
        <div v-if="creatingFolder" class="flex items-center gap-2 px-3 py-1.5">
          <Folder :size="16" class="shrink-0 text-muted-foreground" />
          <Input
            v-model="newFolderName"
            placeholder="Folder name"
            class="h-7 flex-1"
            autofocus
            @keyup.enter="createFolder"
            @keyup.esc="creatingFolder = false"
          />
          <Button variant="ghost" size="sm" class="h-7" @click="createFolder">Create</Button>
          <Button variant="ghost" size="sm" class="h-7" @click="creatingFolder = false">Cancel</Button>
        </div>

        <!-- Empty -->
        <div
          v-if="isEmpty && !creatingFolder"
          class="flex flex-col items-center justify-center gap-2 py-14 text-center"
        >
          <Folder :size="24" class="text-muted-foreground" />
          <p class="text-sm text-muted-foreground">This folder is empty.</p>
          <p class="text-xs text-muted-foreground">Drop files here or use the Upload button.</p>
        </div>

        <!-- Rows -->
        <div v-else class="divide-y divide-border/60">
          <!-- Folders -->
          <button
            v-for="folder in folders"
            :key="folder.prefix"
            type="button"
            class="group flex w-full items-center gap-3 px-3 py-1.5 text-left transition-colors hover:bg-accent/50"
            @click="openFolder(folder.prefix)"
          >
            <Folder :size="16" class="shrink-0 text-muted-foreground" />
            <span class="min-w-0 flex-1 truncate text-sm font-medium text-foreground">
              {{ folderName(folder.prefix) }}
            </span>
            <span class="w-16 shrink-0" />
            <span class="w-24 shrink-0" />
            <span class="flex w-14 shrink-0 justify-end">
              <ChevronRight :size="14" class="text-muted-foreground" />
            </span>
          </button>

          <!-- Files -->
          <div
            v-for="object in objects"
            :key="object.key"
            class="group flex items-center gap-3 px-3 py-1.5 transition-colors hover:bg-accent/50"
          >
            <FileIcon :size="16" class="shrink-0 text-muted-foreground" />
            <span class="min-w-0 flex-1 truncate text-sm text-foreground">{{ baseName(object.key) }}</span>
            <span class="w-16 shrink-0 text-right text-xs tabular-nums text-muted-foreground">
              {{ formatBytes(object.size) }}
            </span>
            <span class="w-24 shrink-0 truncate text-right text-xs text-muted-foreground">
              {{ formatDate(object.lastModified) }}
            </span>
            <div class="flex w-14 shrink-0 items-center justify-end gap-0.5">
              <Button
                variant="ghost"
                size="icon"
                class="h-6 w-6 opacity-0 transition-opacity group-hover:opacity-100"
                title="Download"
                @click="download(object.key)"
              >
                <Download :size="13" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                class="h-6 w-6 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
                title="Delete"
                @click="askDelete(object.key)"
              >
                <Trash2 :size="13" />
              </Button>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Delete confirmation -->
    <AlertDialog v-model:open="deleteOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete file?</AlertDialogTitle>
          <AlertDialogDescription>
            <span class="font-mono">{{ deleteKey ? baseName(deleteKey) : '' }}</span> will be permanently
            deleted. This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="deleting">Cancel</AlertDialogCancel>
          <Button variant="destructive" :disabled="deleting" @click="confirmDelete">
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
