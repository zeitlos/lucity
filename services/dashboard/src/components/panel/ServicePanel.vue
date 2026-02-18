<script setup lang="ts">
import { computed } from 'vue';
import { X } from 'lucide-vue-next';
import { onKeyStroke } from '@vueuse/core';
import { usePanel } from '@/composables/usePanel';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import DeploymentsTab from './DeploymentsTab.vue';
import ServiceSettingsTab from './ServiceSettingsTab.vue';

defineProps<{
  projectId: string;
  service: {
    name: string;
    image: string;
    port: number;
    public: boolean;
    framework?: string;
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'service-removed'): void;
}>();

const { panelStack, currentPanel, popPanel } = usePanel();

const isNestedView = computed(() => panelStack.value.length > 1);

onKeyStroke('Escape', () => {
  if (isNestedView.value) {
    popPanel();
  } else {
    emit('close');
  }
});
</script>

<template>
  <div class="flex h-full flex-col border-l bg-background">
    <!-- Header -->
    <div class="flex shrink-0 items-center justify-between border-b px-4 py-3">
      <div class="flex items-center gap-3">
        <!-- Breadcrumb for nested views -->
        <template v-if="isNestedView">
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink
                  class="cursor-pointer"
                  @click="popPanel"
                >
                  {{ service.name }}
                </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <span class="text-sm text-foreground">{{ currentPanel?.label }}</span>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </template>

        <!-- Normal header -->
        <template v-else>
          <FrameworkIcon :framework="service.framework" :size="24" />
          <h2 class="text-lg font-semibold text-foreground">{{ service.name }}</h2>
        </template>
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
      <Tabs default-value="deployments" class="h-full">
        <div class="px-4 pt-2">
          <TabsList class="w-full">
            <TabsTrigger value="deployments">Deployments</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="deployments" class="px-4 py-4">
          <DeploymentsTab
            :project-id="projectId"
            :service="service"
          />
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <ServiceSettingsTab
            :project-id="projectId"
            :service="service"
            @removed="emit('service-removed')"
          />
        </TabsContent>
      </Tabs>
    </ScrollArea>
  </div>
</template>
