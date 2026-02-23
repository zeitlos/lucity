<script setup lang="ts">
import { RouterLink } from 'vue-router';
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
        <BreadcrumbLink as-child>
          <RouterLink to="/">Projects</RouterLink>
        </BreadcrumbLink>
      </BreadcrumbItem>
      <BreadcrumbSeparator>
        <span class="text-border">/</span>
      </BreadcrumbSeparator>
      <BreadcrumbItem>
        <BreadcrumbLink as-child class="font-medium">
          <RouterLink :to="`/projects/${projectName}`">
            {{ projectName }}
          </RouterLink>
        </BreadcrumbLink>
      </BreadcrumbItem>
      <BreadcrumbSeparator>
        <span class="text-border">/</span>
      </BreadcrumbSeparator>
      <BreadcrumbItem>
        <EnvironmentSwitcher />
      </BreadcrumbItem>
    </BreadcrumbList>
  </Breadcrumb>
</template>
