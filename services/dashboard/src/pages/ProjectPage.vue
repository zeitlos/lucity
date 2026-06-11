<script setup lang="ts">
import { graphql } from '@/gql';
import { useQuery } from '@vue/apollo-composable';
import { Loader } from 'lucide-vue-next';
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

const ProjectDocument = graphql(`
  query ProjectPage($id: ProjectID!) {
    project(id: $id) {
      id
      name
      environments {
        id
        name
        resourceTier
      }
    }
  }
`);

const route = useRoute();
const router = useRouter();
const projectId = computed(() => {
  const id = route.params.projectId;
  return Array.isArray(id) ? id[0]! : (id as string);
});

const { onResult } = useQuery(ProjectDocument, () => ({
  id: projectId.value,
}));


onResult((result) => {
  if (!result.data || result.data.project.environments.length <= 0) {
    return;
  }

  const firstEnv = result.data.project.environments[0]

  if (!firstEnv) {
    return;
  }

  router.replace({ name: 'environment', params: {
    projectId: projectId.value,
    environmentId:  firstEnv?.id,
  }})
});
</script>

<template>
  <Loader></Loader>
</template>