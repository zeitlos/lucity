import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { useAuth } from '@/composables/useAuth';
import { useEnvironment } from '@/composables/useEnvironment';
import { graphql } from '@/gql';
import { apolloClient } from '@/lib/apollo';
import { preferredEnvironment } from '@/lib/lastEnvironment';

const WorkspacesDocument = graphql(`
  query Workspaces {
    workspaces {
      id
      name
      personal
    }
  }
`);

const ProjectsForNavDocument = graphql(`
  query ProjectsForNav {
    projects {
      id
      name
      environments {
        id
        name
      }
    }
  }
`);

/**
 * The workspace, project and environment the person is in, and the moves
 * between them that the header breadcrumb and the mobile context bar share.
 */
export function useContextSwitch() {
  const route = useRoute();
  const router = useRouter();
  const { activeWorkspace, setActiveWorkspace } = useAuth();
  const { activeEnvironment, environments } = useEnvironment();

  const { result: wsResult } = useQuery(WorkspacesDocument);
  const workspaces = computed(() => wsResult.value?.workspaces ?? []);
  const activeWs = computed(() => workspaces.value.find(w => w.id === activeWorkspace.value));

  const { result: projResult } = useQuery(ProjectsForNavDocument);
  const projects = computed(() => projResult.value?.projects ?? []);

  const isEnvironmentRoute = computed(() => route.name === 'environment');

  const activeProjectId = computed<string | undefined>(() => {
    const id = route.params.projectId;
    return Array.isArray(id) ? id[0] : (id as string | undefined);
  });

  const activeProject = computed(() => projects.value.find(p => p.id === activeProjectId.value));
  const activeProjectName = computed(() => activeProject.value?.name ?? activeProjectId.value ?? '');

  const projectEnvironments = computed(() => {
    if (environments.value.length > 0) return environments.value;
    return activeProject.value?.environments ?? [];
  });

  const switchingWorkspace = ref(false);
  const switchingWorkspaceName = ref('');

  async function handleWorkspaceSwitch(id: string) {
    if (id === activeWorkspace.value) return;
    const ws = workspaces.value.find(w => w.id === id);
    switchingWorkspaceName.value = ws?.name ?? id;
    switchingWorkspace.value = true;
    setActiveWorkspace(id);
    await router.push('/');
    await apolloClient.resetStore();
    switchingWorkspace.value = false;
  }

  function handleProjectSwitch(id: string) {
    if (id === activeProjectId.value) return;
    const project = projects.value.find(p => p.id === id);
    const target = preferredEnvironment(id, project?.environments ?? []);
    if (target) {
      router.push({ name: 'environment', params: { projectId: id, environmentId: target.id } });
    } else {
      router.push({ name: 'project', params: { projectId: id } });
    }
  }

  function handleEnvironmentSwitch(envId: string) {
    if (envId === activeEnvironment.value?.id) return;
    if (!activeProjectId.value) return;
    router.push({
      name: 'environment',
      params: { projectId: activeProjectId.value, environmentId: envId, drawer: route.params.drawer },
    });
  }

  function handleEnvironmentCreated(environmentId: string) {
    if (!activeProjectId.value) return;
    router.push({
      name: 'environment',
      params: { projectId: activeProjectId.value, environmentId },
    });
  }

  function handleNewProject() {
    router.push({ name: 'projects', query: { new: 'project' } });
  }

  return {
    activeWorkspace,
    activeEnvironment,
    workspaces,
    activeWs,
    projects,
    isEnvironmentRoute,
    activeProjectId,
    activeProject,
    activeProjectName,
    projectEnvironments,
    switchingWorkspace,
    switchingWorkspaceName,
    handleWorkspaceSwitch,
    handleProjectSwitch,
    handleEnvironmentSwitch,
    handleEnvironmentCreated,
    handleNewProject,
  };
}
