import { environmentRoute, projectRoute, routeDrawer, routeProjectId } from '@/lib/routes';
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

export interface WorkspaceSwitch {
  from: { id: string; name: string };
  to: { id: string; name: string };
  arrived: boolean;
}

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

  const activeProjectId = computed(() => routeProjectId(route));

  const activeProject = computed(() => projects.value.find(p => p.id === activeProjectId.value));
  const activeProjectName = computed(() => activeProject.value?.name ?? activeProjectId.value ?? '');

  const projectEnvironments = computed(() => {
    if (environments.value.length > 0) return environments.value;
    return activeProject.value?.environments ?? [];
  });

  const workspaceSwitch = ref<WorkspaceSwitch | null>(null);

  async function handleWorkspaceSwitch(id: string) {
    if (id === activeWorkspace.value) return;
    const ws = workspaces.value.find(w => w.id === id);
    workspaceSwitch.value = {
      from: { id: activeWorkspace.value, name: activeWs.value?.name ?? activeWorkspace.value },
      to: { id, name: ws?.name ?? id },
      arrived: false,
    };
    setActiveWorkspace(id);
    try {
      await router.push({ name: 'projects', params: { workspace: id } });
      await apolloClient.resetStore();
    } finally {
      workspaceSwitch.value.arrived = true;
    }
  }

  function finishWorkspaceSwitch() {
    workspaceSwitch.value = null;
  }

  function handleProjectSwitch(id: string) {
    if (id === activeProjectId.value) return;
    const project = projects.value.find(p => p.id === id);
    const target = preferredEnvironment(id, project?.environments ?? []);
    if (target) {
      router.push(environmentRoute(target.id));
    } else {
      router.push(projectRoute(id));
    }
  }

  function handleEnvironmentSwitch(envId: string) {
    if (envId === activeEnvironment.value?.id) return;
    if (!activeProjectId.value) return;
    router.push(environmentRoute(envId, routeDrawer(route)));
  }

  function handleEnvironmentCreated(environmentId: string) {
    if (!activeProjectId.value) return;
    router.push(environmentRoute(environmentId));
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
    workspaceSwitch,
    handleWorkspaceSwitch,
    finishWorkspaceSwitch,
    handleProjectSwitch,
    handleEnvironmentSwitch,
    handleEnvironmentCreated,
    handleNewProject,
  };
}
