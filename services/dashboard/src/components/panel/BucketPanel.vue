<script setup lang="ts">
import { X } from '@lucide/vue';
import BucketIcon from '@/components/BucketIcon.vue';
import { onKeyStroke } from '@vueuse/core';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import BucketConnectionTab from './BucketConnectionTab.vue';
import BucketFilesTab from './BucketFilesTab.vue';
import BucketSettingsTab from './BucketSettingsTab.vue';

const props = defineProps<{
  bucket: {
    id: string;
    name: string;
    region: string;
    sizeBytes: number;
    objectCount: number;
    public: boolean;
    publicEndpoint: string;
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'bucket-removed'): void;
}>();

onKeyStroke('Escape', () => {
  emit('close');
});
</script>

<template>
  <div class="flex h-full flex-col rounded-lg border bg-card shadow-sm">
    <!-- Header -->
    <div class="flex shrink-0 items-center justify-between border-b px-4 py-3">
      <div class="flex items-center gap-3">
        <BucketIcon :size="22" class="shrink-0 text-muted-foreground" />
        <h2 class="text-lg font-semibold text-foreground">{{ bucket.name }}</h2>
      </div>

      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        @click="emit('close')"
      >
        <X :size="16" />
      </Button>
    </div>

    <!-- Tab Content -->
    <ScrollArea class="flex-1">
      <Tabs default-value="files" class="h-full">
        <div class="px-4 pt-2">
          <TabsList class="w-full">
            <TabsTrigger value="files">Files</TabsTrigger>
            <TabsTrigger value="connect">Connect</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="files" class="px-4 py-4">
          <BucketFilesTab :bucket-id="props.bucket.id" />
        </TabsContent>

        <TabsContent value="connect" class="px-4 py-4">
          <BucketConnectionTab
            :bucket-id="props.bucket.id"
            :bucket-name="props.bucket.name"
            :public="props.bucket.public"
            :public-endpoint="props.bucket.publicEndpoint"
          />
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <BucketSettingsTab
            :bucket-id="props.bucket.id"
            :bucket="props.bucket"
            @bucket-removed="emit('bucket-removed')"
          />
        </TabsContent>
      </Tabs>
    </ScrollArea>
  </div>
</template>
