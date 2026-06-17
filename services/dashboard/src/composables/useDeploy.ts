import { reactive } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { graphql } from '@/gql';
import { BuildStatus } from '@/gql/graphql';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const DeployDocument = graphql(`
  mutation Deploy($service: ServiceID!, $gitRef: String) {
    deploy(service: $service, gitRef: $gitRef) {
      id
      status
      startedAt
      finishedAt
    }
  }
`);

const BuildStatusDocument = graphql(`
  query BuildStatus($id: String!) {
    build(id: $id) {
      id
      status
      startedAt
      finishedAt
    }
  }
`);

export interface DeployState {
  buildId: string | null;
  status: BuildStatus | null;
  isDeploying: boolean;
  error: string | null;
  startDeploy: (serviceId: string, serviceName: string, gitRef?: string) => Promise<void>;
  pollBuild: (buildId: string) => void;
  reset: () => void;
}

const TERMINAL_STATUSES = new Set<BuildStatus>([
  BuildStatus.Succeeded,
  BuildStatus.Failed,
  BuildStatus.Cancelled,
]);

export function useDeploy(): DeployState {
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  function startPolling() {
    pollTimer = setInterval(async () => {
      if (!state.buildId) return;

      try {
        const { data } = await apolloClient.query({
          query: BuildStatusDocument,
          variables: { id: state.buildId },
          fetchPolicy: 'network-only',
        });

        const build = data?.build;
        if (!build) return;

        state.status = build.status;

        if (build.status === BuildStatus.Succeeded) {
          stopPolling();
          state.isDeploying = false;
          toast.success('Build complete');
        } else if (build.status === BuildStatus.Failed) {
          stopPolling();
          state.isDeploying = false;
          state.error = 'Build failed';
          errorToast('Build failed');
        } else if (build.status === BuildStatus.Cancelled) {
          stopPolling();
          state.isDeploying = false;
          state.error = 'Build cancelled';
        }
      } catch (e: unknown) {
        stopPolling();
        state.isDeploying = false;
        state.error = errorMessage(e);
        errorToast('Build status check failed', { description: state.error });
      }
    }, 2000);
  }

  const state: DeployState = reactive({
    buildId: null,
    status: null,
    isDeploying: false,
    error: null,

    async startDeploy(serviceId: string, serviceName: string, gitRef?: string) {
      state.error = null;
      state.status = BuildStatus.Queued;
      state.isDeploying = true;

      try {
        const res = await apolloClient.mutate({
          mutation: DeployDocument,
          variables: {
            service: serviceId,
            gitRef: gitRef || undefined,
          },
        });

        if (!res?.data?.deploy) {
          throw new Error('Failed to start deploy');
        }

        state.buildId = res.data.deploy.id;
        state.status = res.data.deploy.status;
        startPolling();
        toast.info('Deploy started', { description: `Building ${serviceName}...` });
      } catch (e: unknown) {
        state.isDeploying = false;
        state.error = errorMessage(e);
        errorToast('Failed to start deploy', { description: state.error });
      }
    },

    pollBuild(buildId: string) {
      state.buildId = buildId;
      state.status = BuildStatus.Running;
      state.isDeploying = true;
      startPolling();
    },

    reset() {
      stopPolling();
      state.buildId = null;
      state.status = null;
      state.isDeploying = false;
      state.error = null;
    },
  });

  return state;
}

export { TERMINAL_STATUSES };
