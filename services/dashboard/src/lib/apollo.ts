import { ApolloClient, InMemoryCache, createHttpLink, from, split } from '@apollo/client/core';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import router from '@/router';
import { toast } from '@/components/ui/sonner';
import { useAuth } from '@/composables/useAuth';

const { activeWorkspace, login } = useAuth();

const httpLink = createHttpLink({
  uri: '/graphql',
  credentials: 'include',
});

const workspaceLink = setContext((_, { headers }) => ({
  headers: {
    ...headers,
    'X-Lucity-Workspace': activeWorkspace.value,
  },
}));

const errorLink = onError(({ graphQLErrors, networkError, operation }) => {
  if (graphQLErrors) {
    for (const err of graphQLErrors) {
      if (err.message === 'unauthorized') {
        router.push('/login');
        return;
      }
      if (err.extensions?.code === 'SESSION_EXPIRED') {
        login();
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
    // 403 from workspace authorization — JWT has stale workspace claims, re-login
    if ('statusCode' in networkError && networkError.statusCode === 403) {
      login();
      return;
    }
    toast.error('Network error', { description: networkError.message });
  }
});

function getToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)lucity_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]!) : '';
}

const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

const wsLink = new GraphQLWsLink(createClient({
  url: `${wsProtocol}//${window.location.host}/graphql`,
  connectionParams: () => {
    const token = getToken();
    return {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      'X-Lucity-Workspace': activeWorkspace.value,
    };
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
  from([errorLink, workspaceLink, httpLink]),
);

export const apolloClient = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache(),
});
