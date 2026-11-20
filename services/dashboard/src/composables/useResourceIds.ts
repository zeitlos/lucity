import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { lastSegment } from '@/lib/id';

const EnvironmentResourceIdsDocument = graphql(`
  query EnvironmentResourceIds($environment: EnvironmentID!) {
    environment(environment: $environment) {
      id
      services {
        id
      }
      databases {
        id
      }
      volumes {
        id
      }
      keyValueStores {
        id
      }
      buckets {
        id
      }
    }
  }
`);

export function useResourceIds(environmentId: MaybeRefOrGetter<string | undefined>, enabled: MaybeRefOrGetter<boolean> = true) {
  const { result } = useQuery(EnvironmentResourceIdsDocument, () => ({
    environment: toValue(environmentId) ?? '',
  }), () => ({
    enabled: toValue(enabled) && !!toValue(environmentId),
    fetchPolicy: 'cache-and-network',
  }));

  return computed(() => {
    const environment = result.value?.environment;
    if (!environment) return new Set<string>();
    return new Set([
      ...environment.services,
      ...environment.databases,
      ...environment.volumes,
      ...environment.keyValueStores,
      ...environment.buckets,
    ].map(resource => lastSegment(resource.id)));
  });
}
