<script setup lang="ts">
import { Slash } from 'lucide-vue-next';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import EnvironmentSwitcher from '@/components/EnvironmentSwitcher.vue';

defineProps<{
  projectName: string;
}>();
</script>

<template>
  <Breadcrumb>
    <BreadcrumbList>
      <BreadcrumbItem>
        <BreadcrumbLink href="/">Projects</BreadcrumbLink>
      </BreadcrumbItem>
      <BreadcrumbSeparator>
        <Slash :size="14" />
      </BreadcrumbSeparator>
      <BreadcrumbItem>
        <span class="text-sm font-medium text-foreground">{{ projectName }}</span>
      </BreadcrumbItem>
      <BreadcrumbSeparator>
        <Slash :size="14" />
      </BreadcrumbSeparator>
      <BreadcrumbItem>
        <EnvironmentSwitcher />
      </BreadcrumbItem>
    </BreadcrumbList>
  </Breadcrumb>
</template>
