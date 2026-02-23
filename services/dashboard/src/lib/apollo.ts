import { ApolloClient, InMemoryCache, createHttpLink, from, split } from '@apollo/client/core';
import { onError } from '@apollo/client/link/error';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import router from '@/router';
import { toast } from '@/components/ui/sonner';

const httpLink = createHttpLink({
  uri: '/graphql',
  credentials: 'include',
});

const errorLink = onError(({ graphQLErrors, networkError, operation }) => {
  if (graphQLErrors) {
    for (const err of graphQLErrors) {
      if (err.message === 'unauthorized') {
        router.push('/login');
        return;
      }
      if (err.extensions?.code === 'GITHUB_TOKEN_EXPIRED') {
        window.location.href = '/auth/github';
        return;
      }
    }

    // Toast query errors globally (mutations handle errors at component level)
    const def = getMainDefinition(operation.query);
    if (def.kind === 'OperationDefinition' && def.operation === 'query') {
      toast.error(graphQLErrors.map(e => e.message).join(', '));
    }
  }

  if (networkError) {
    toast.error('Network error', { description: networkError.message });
  }
});

function getToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)lucity_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}

const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

const wsLink = new GraphQLWsLink(createClient({
  url: `${wsProtocol}//${window.location.host}/graphql`,
  connectionParams: () => {
    const token = getToken();
    return token ? { Authorization: `Bearer ${token}` } : {};
  },
  lazy: true,
  retryAttempts: 3,
}));

const splitLink = split(
  ({ query }) => {
    const def = getMainDefinition(query);
    return def.kind === 'OperationDefinition' && def.operation === 'subscription';
  },
  wsLink,
  from([errorLink, httpLink]),
);

export const apolloClient = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache(),
});
