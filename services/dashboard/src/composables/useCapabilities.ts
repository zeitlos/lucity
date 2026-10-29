import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';

const CapabilitiesDocument = graphql(`
  query Capabilities {
    workspace {
      id
      capabilities {
        agents
      }
    }
  }
`);

export function useCapabilities() {
  const { result } = useQuery(CapabilitiesDocument);

  return {
    agents: computed(() => result.value?.workspace.capabilities.agents ?? false),
  };
}
