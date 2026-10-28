<script setup lang="ts">
import { useGuardedTab } from '@/composables/useUnsavedChanges';
import PanelShell from '@/components/PanelShell.vue';

import BucketIcon from '@/components/BucketIcon.vue';
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

const { tab: activeTab, selectTab } = useGuardedTab('files');

</script>

<template>
  <PanelShell :title="bucket.name" @close="emit('close')">
    <template #icon>
      <BucketIcon />
    </template>

    <!-- Tab Content -->
    <Tabs :model-value="activeTab" @update:model-value="selectTab" class="flex min-h-0 flex-1 flex-col">
      <div class="shrink-0 border-b px-4 pt-2">
        <TabsList class="w-full border-b-0">
          <TabsTrigger value="files">Files</TabsTrigger>
          <TabsTrigger value="connect">Connect</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>
      </div>

      <ScrollArea class="min-h-0 flex-1 [&>[data-reka-scroll-area-viewport]>div]:h-full">
        <div class="flex h-full flex-col">
          <TabsContent value="files" class="flex-1 px-4 py-4">
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
        </div>
      </ScrollArea>
    </Tabs>
  </PanelShell>
</template>
