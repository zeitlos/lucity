import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { Availability } from '@/gql/graphql';

export const CapabilitiesDocument = graphql(`
  query Capabilities {
    workspace {
      id
      capabilities {
        services {
          availability
          plan
          cpu { min max upgrades { plan max } }
          memory { min max upgrades { plan max } }
          replicas
        }
        databases {
          availability
          plan
          cpu { min max upgrades { plan max } }
          memory { min max upgrades { plan max } }
          storage { min max upgrades { plan max } }
        }
        keyValueStores {
          availability
          plan
          size { min max upgrades { plan max } }
        }
        volumes {
          availability
          plan
          size { min max upgrades { plan max } }
        }
        buckets { availability plan }
        agents { availability plan }
        members { availability plan limit upgrades { plan limit } }
        customDomains { availability plan limit upgrades { plan limit } }
      }
    }
  }
`);

export function useCapabilities() {
  const { result, refetch } = useQuery(CapabilitiesDocument);
  const capabilities = computed(() => result.value?.workspace.capabilities);

  return {
    capabilities,
    agents: computed(() => !!capabilities.value && capabilities.value.agents.availability !== Availability.Unavailable),
    cardRequired: computed(() => capabilities.value?.services.availability === Availability.PaymentMethodRequired),
    refetch,
  };
}
