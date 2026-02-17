import { ref } from 'vue';
import { useMutation, useLazyQuery } from '@vue/apollo-composable';
import { BuildServiceMutation, BuildStatusQuery, DeployBuildMutation } from '@/graphql/services';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

export function useBuild() {
  const buildId = ref<string | null>(null);
  const phase = ref<string | null>(null);
  const error = ref<string | null>(null);
  const isBuilding = ref(false);
  const imageRef = ref<string | null>(null);
  const digest = ref<string | null>(null);

  const { mutate: buildServiceMutate } = useMutation(BuildServiceMutation);
  const { mutate: deployBuildMutate } = useMutation(DeployBuildMutation);
  const { load: loadBuildStatus, result: statusResult, onResult: onStatusResult, onError: onStatusError } = useLazyQuery(BuildStatusQuery, () => ({
    id: buildId.value!,
  }), {
    fetchPolicy: 'network-only',
  });

  let pollTimer: ReturnType<typeof setInterval> | null = null;

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  onStatusError((err) => {
    stopPolling();
    isBuilding.value = false;
    error.value = err.message;
    toast.error('Build status check failed', { description: err.message });
  });

  onStatusResult((result) => {
    const status = result.data?.buildStatus;
    if (!status) return;

    phase.value = status.phase;

    if (status.phase === 'SUCCEEDED') {
      stopPolling();
      isBuilding.value = false;
      imageRef.value = status.imageRef ?? null;
      digest.value = status.digest ?? null;
      toast.success('Build succeeded');
    } else if (status.phase === 'FAILED') {
      stopPolling();
      isBuilding.value = false;
      error.value = status.error ?? 'Build failed';
      toast.error('Build failed', { description: status.error });
    }
  });

  async function startBuild(projectId: string, service: string, gitRef?: string, contextPath?: string) {
    error.value = null;
    phase.value = 'QUEUED';
    isBuilding.value = true;
    imageRef.value = null;
    digest.value = null;

    try {
      const res = await buildServiceMutate({
        input: {
          projectId,
          service,
          gitRef: gitRef || undefined,
          contextPath: contextPath || undefined,
        },
      });

      if (!res?.data?.buildService) {
        throw new Error('Failed to start build');
      }

      buildId.value = res.data.buildService.id;
      phase.value = res.data.buildService.phase;

      // Start polling every 2 seconds
      pollTimer = setInterval(() => {
        loadBuildStatus();
      }, 2000);

      toast.info('Build started', { description: `Building ${service}...` });
    } catch (e: unknown) {
      isBuilding.value = false;
      error.value = errorMessage(e);
      toast.error('Failed to start build', { description: error.value });
    }
  }

  async function deploy(projectId: string, service: string, environment: string, tag: string, buildDigest?: string) {
    try {
      await deployBuildMutate({
        input: {
          projectId,
          service,
          environment,
          tag,
          digest: buildDigest || undefined,
        },
      });
      toast.success('Deployed', { description: `${service} deployed to ${environment}` });
      return true;
    } catch (e: unknown) {
      toast.error('Deploy failed', { description: errorMessage(e) });
      return false;
    }
  }

  function reset() {
    stopPolling();
    buildId.value = null;
    phase.value = null;
    error.value = null;
    isBuilding.value = false;
    imageRef.value = null;
    digest.value = null;
  }

  return {
    buildId,
    phase,
    error,
    isBuilding,
    imageRef,
    digest,
    startBuild,
    deploy,
    reset,
  };
}
