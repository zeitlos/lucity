import { ApolloClient, InMemoryCache, createHttpLink, from, split } from '@apollo/client/core';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import { useAuth } from '@/composables/useAuth';
import { gateFromError, useGate } from '@/composables/useGate';

const { activeWorkspace, login } = useAuth();

const httpLink = createHttpLink({ uri: '/graphql', credentials: 'include' });

const workspaceLink = setContext((_, { headers }) => ({
  headers: {
    ...headers,
    'X-Lucity-Workspace': activeWorkspace.value,
  },
}));

const { open: openGate } = useGate();

const errorLink = onError(({ graphQLErrors, networkError, operation }) => {
  const authFailed =
    graphQLErrors?.some(e => e.message === 'unauthenticated' || e.message === 'unauthorized') ||
    (!!networkError && 'statusCode' in networkError && (networkError.statusCode === 401 || networkError.statusCode === 403));

  if (authFailed) login();

  const gate = gateFromError({ graphQLErrors });
  if (gate && operation.operationName !== 'CompletePaymentMethodCheckout') openGate(gate);
});

const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

const wsLink = new GraphQLWsLink(createClient({
  url: `${wsProtocol}//${window.location.host}/graphql`,
  connectionParams: () => ({
    'X-Lucity-Workspace': activeWorkspace.value,
  }),
  lazy: true,
  retryAttempts: 3,
}));

const splitLink = split(
  ({ query }) => {
    const def = getMainDefinition(query);
    return def.kind === 'OperationDefinition' && def.operation === 'subscription';
  },
  wsLink,
  from([errorLink, workspaceLink, httpLink]),
);

export const apolloClient = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache({
    typePolicies: {
      VulnerabilityReport: {
        merge: true,
      },
    },
  }),
});
