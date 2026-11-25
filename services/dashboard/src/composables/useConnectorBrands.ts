import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { connectorBrand, resourceName } from '@/lib/agents';

const ConnectorBrandsDocument = graphql(`
  query ConnectorBrands {
    connectors {
      id
      name
      url
    }
  }
`);

export function useConnectorBrands() {
  const { result } = useQuery(ConnectorBrandsDocument);

  const brands = computed(() => {
    const byKey = new Map<string, string | null>();
    for (const connector of result.value?.connectors ?? []) {
      const brand = connectorBrand(connector);
      byKey.set(resourceName(connector.id), brand);
      byKey.set(connector.name, brand);
    }
    return byKey;
  });

  return (connector: string) => brands.value.get(connector) ?? null;
}
