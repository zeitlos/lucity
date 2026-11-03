import { computed, reactive, ref, watch } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { graphql } from '@/gql';
import { activeBuild, type Service } from '@/composables/useEnvironment';
import { failureToast, infoToast } from '@/components/ui/sonner';

const DeployDocument = graphql(`
  mutation Deploy($service: ServiceID!, $gitRef: String) {
    deploy(service: $service, gitRef: $gitRef) {
      id
      status
      build {
        id
        status
        startedAt
        finishedAt
      }
    }
  }
`);

export interface DeployState {
  isDeploying: boolean;
  startDeploy: (gitRef?: string) => Promise<void>;
}

export function useDeploy(service: () => Pick<Service, 'id' | 'name' | 'builds'>): DeployState {
  const starting = ref(false);
  const unseenBuildId = ref<string | null>(null);

  watch(() => service().builds, (builds) => {
    if (builds?.some(b => b.id === unseenBuildId.value)) {
      unseenBuildId.value = null;
    }
  });

  return reactive({
    isDeploying: computed(() => starting.value || unseenBuildId.value !== null || activeBuild(service()) !== null),

    async startDeploy(gitRef?: string) {
      starting.value = true;

      try {
        const res = await apolloClient.mutate({
          mutation: DeployDocument,
          variables: {
            service: service().id,
            gitRef: gitRef || undefined,
          },
        });

        const build = res?.data?.deploy?.build;

        if (!build) {
          throw new Error('Failed to start deploy');
        }

        unseenBuildId.value = build.id;
        infoToast('Deploy started', { description: `Building ${service().name}…` });
      } catch (e: unknown) {
        failureToast('Failed to start deploy', e);
      } finally {
        starting.value = false;
      }
    },
  });
}
