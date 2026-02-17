import { reactive } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { BuildServiceMutation, BuildStatusQuery, DeployBuildMutation } from '@/graphql/services';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

export interface BuildState {
  buildId: string | null;
  phase: string | null;
  error: string | null;
  isBuilding: boolean;
  imageRef: string | null;
  digest: string | null;
  startBuild: (projectId: string, service: string, gitRef?: string, contextPath?: string) => Promise<void>;
  deploy: (projectId: string, service: string, environment: string, tag: string, buildDigest?: string) => Promise<boolean>;
  reset: () => void;
}

export function useBuild(): BuildState {
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  const state: BuildState = reactive({
    buildId: null,
    phase: null,
    error: null,
    isBuilding: false,
    imageRef: null,
    digest: null,

    async startBuild(projectId: string, service: string, gitRef?: string, contextPath?: string) {
      state.error = null;
      state.phase = 'QUEUED';
      state.isBuilding = true;
      state.imageRef = null;
      state.digest = null;

      try {
        const res = await apolloClient.mutate({
          mutation: BuildServiceMutation,
          variables: {
            input: {
              projectId,
              service,
              gitRef: gitRef || undefined,
              contextPath: contextPath || undefined,
            },
          },
        });

        if (!res?.data?.buildService) {
          throw new Error('Failed to start build');
        }

        state.buildId = res.data.buildService.id;
        state.phase = res.data.buildService.phase;

        // Start polling every 2 seconds
        pollTimer = setInterval(async () => {
          if (!state.buildId) return;

          try {
            const { data } = await apolloClient.query({
              query: BuildStatusQuery,
              variables: { id: state.buildId },
              fetchPolicy: 'network-only',
            });

            const status = data?.buildStatus;
            if (!status) return;

            state.phase = status.phase;

            if (status.phase === 'SUCCEEDED') {
              stopPolling();
              state.isBuilding = false;
              state.imageRef = status.imageRef ?? null;
              state.digest = status.digest ?? null;
              toast.success('Build succeeded');
            } else if (status.phase === 'FAILED') {
              stopPolling();
              state.isBuilding = false;
              state.error = status.error ?? 'Build failed';
              toast.error('Build failed', { description: status.error });
            }
          } catch (e: unknown) {
            stopPolling();
            state.isBuilding = false;
            state.error = errorMessage(e);
            toast.error('Build status check failed', { description: state.error });
          }
        }, 2000);

        toast.info('Build started', { description: `Building ${service}...` });
      } catch (e: unknown) {
        state.isBuilding = false;
        state.error = errorMessage(e);
        toast.error('Failed to start build', { description: state.error });
      }
    },

    async deploy(projectId: string, service: string, environment: string, tag: string, buildDigest?: string) {
      try {
        await apolloClient.mutate({
          mutation: DeployBuildMutation,
          variables: {
            input: {
              projectId,
              service,
              environment,
              tag,
              digest: buildDigest || undefined,
            },
          },
        });
        toast.success('Deployed', { description: `${service} deployed to ${environment}` });
        return true;
      } catch (e: unknown) {
        toast.error('Deploy failed', { description: errorMessage(e) });
        return false;
      }
    },

    reset() {
      stopPolling();
      state.buildId = null;
      state.phase = null;
      state.error = null;
      state.isBuilding = false;
      state.imageRef = null;
      state.digest = null;
    },
  });

  return state;
}
