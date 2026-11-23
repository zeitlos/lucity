import { reactive, ref } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { graphql } from '@/gql';
import type { Service } from '@/composables/useEnvironment';
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
  starting: boolean;
  startDeploy: (gitRef?: string) => Promise<void>;
}

export function useDeploy(service: () => Pick<Service, 'id' | 'name'>): DeployState {
  const starting = ref(false);

  return reactive({
    starting,

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

        infoToast('Deploy started', { description: `Building ${service().name}…` });
      } catch (e: unknown) {
        failureToast('Failed to start deploy', e);
      } finally {
        starting.value = false;
      }
    },
  });
}
