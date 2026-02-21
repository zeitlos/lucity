import { reactive } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { DeployMutation, DeployStatusQuery } from '@/graphql/services';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

export interface DeployState {
  deployId: string | null;
  phase: string | null;
  error: string | null;
  isDeploying: boolean;
  imageRef: string | null;
  digest: string | null;
  argoHealth: string | null;
  argoMessage: string | null;
  startDeploy: (projectId: string, service: string, environment: string, gitRef?: string, contextPath?: string) => Promise<void>;
  pollDeploy: (deployId: string) => void;
  reset: () => void;
}

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
      if (!state.deployId) return;

      try {
        const { data } = await apolloClient.query({
          query: DeployStatusQuery,
          variables: { id: state.deployId },
          fetchPolicy: 'network-only',
        });

        const status = data?.deployStatus;
        if (!status) return;

        state.phase = status.phase;
        state.argoHealth = status.argoHealth ?? null;
        state.argoMessage = status.argoMessage ?? null;

        if (status.phase === 'SUCCEEDED') {
          stopPolling();
          state.imageRef = status.imageRef ?? null;
          state.digest = status.digest ?? null;
          state.isDeploying = false;
          toast.success('Deployed');
        } else if (status.phase === 'FAILED') {
          stopPolling();
          state.isDeploying = false;
          state.error = status.error ?? 'Deploy failed';
          toast.error('Deploy failed', { description: status.error });
        }
      } catch (e: unknown) {
        stopPolling();
        state.isDeploying = false;
        state.error = errorMessage(e);
        toast.error('Deploy status check failed', { description: state.error });
      }
    }, 2000);
  }

  const state: DeployState = reactive({
    deployId: null,
    phase: null,
    error: null,
    isDeploying: false,
    imageRef: null,
    digest: null,
    argoHealth: null,
    argoMessage: null,

    async startDeploy(projectId: string, service: string, environment: string, gitRef?: string, contextPath?: string) {
      state.error = null;
      state.phase = 'QUEUED';
      state.isDeploying = true;
      state.imageRef = null;
      state.digest = null;
      state.argoHealth = null;
      state.argoMessage = null;

      try {
        const res = await apolloClient.mutate({
          mutation: DeployMutation,
          variables: {
            input: {
              projectId,
              service,
              environment,
              gitRef: gitRef || undefined,
              contextPath: contextPath || undefined,
            },
          },
        });

        if (!res?.data?.deploy) {
          throw new Error('Failed to start deploy');
        }

        state.deployId = res.data.deploy.id;
        state.phase = res.data.deploy.phase;
        startPolling();
        toast.info('Deploy started', { description: `Deploying ${service}...` });
      } catch (e: unknown) {
        state.isDeploying = false;
        state.error = errorMessage(e);
        toast.error('Failed to start deploy', { description: state.error });
      }
    },

    pollDeploy(deployId: string) {
      state.deployId = deployId;
      state.phase = 'QUEUED';
      state.isDeploying = true;
      startPolling();
    },

    reset() {
      stopPolling();
      state.deployId = null;
      state.phase = null;
      state.error = null;
      state.isDeploying = false;
      state.imageRef = null;
      state.digest = null;
      state.argoHealth = null;
      state.argoMessage = null;
    },
  });

  return state;
}
