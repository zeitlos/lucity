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
  buildAndDeploy: (projectId: string, service: string, environment: string, gitRef?: string, contextPath?: string) => Promise<void>;
  deploy: (projectId: string, service: string, environment: string, tag: string, buildDigest?: string) => Promise<boolean>;
  reset: () => void;
}

function extractTag(imageRef: string): string {
  const parts = imageRef.split(':');
  return parts.length > 1 ? parts[parts.length - 1] : imageRef;
}

export function useBuild(): BuildState {
  let pollTimer: ReturnType<typeof setInterval> | null = null;
  let onBuildSucceeded: (() => void) | null = null;

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
          query: BuildStatusQuery,
          variables: { id: state.buildId },
          fetchPolicy: 'network-only',
        });

        const status = data?.buildStatus;
        if (!status) return;

        state.phase = status.phase;

        if (status.phase === 'SUCCEEDED') {
          stopPolling();
          state.imageRef = status.imageRef ?? null;
          state.digest = status.digest ?? null;

          if (onBuildSucceeded) {
            onBuildSucceeded();
            onBuildSucceeded = null;
          } else {
            state.isBuilding = false;
            toast.success('Build succeeded');
          }
        } else if (status.phase === 'FAILED') {
          stopPolling();
          onBuildSucceeded = null;
          state.isBuilding = false;
          state.error = status.error ?? 'Build failed';
          toast.error('Build failed', { description: status.error });
        }
      } catch (e: unknown) {
        stopPolling();
        onBuildSucceeded = null;
        state.isBuilding = false;
        state.error = errorMessage(e);
        toast.error('Build status check failed', { description: state.error });
      }
    }, 2000);
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
      onBuildSucceeded = null;

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
        startPolling();
        toast.info('Build started', { description: `Building ${service}...` });
      } catch (e: unknown) {
        state.isBuilding = false;
        state.error = errorMessage(e);
        toast.error('Failed to start build', { description: state.error });
      }
    },

    async buildAndDeploy(projectId: string, service: string, environment: string, gitRef?: string, contextPath?: string) {
      state.error = null;
      state.phase = 'QUEUED';
      state.isBuilding = true;
      state.imageRef = null;
      state.digest = null;

      // Set up the callback that fires when build succeeds
      onBuildSucceeded = async () => {
        if (!state.imageRef) {
          state.isBuilding = false;
          state.error = 'Build succeeded but no image reference returned';
          toast.error('Deploy failed', { description: state.error });
          return;
        }

        state.phase = 'DEPLOYING';
        const tag = extractTag(state.imageRef);
        const ok = await state.deploy(projectId, service, environment, tag, state.digest ?? undefined);
        state.phase = ok ? 'DEPLOYED' : 'FAILED';
        state.isBuilding = false;
      };

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
        startPolling();
        toast.info('Build & deploy started', { description: `Building ${service}...` });
      } catch (e: unknown) {
        onBuildSucceeded = null;
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
      onBuildSucceeded = null;
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
