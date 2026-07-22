import { ApolloClient, InMemoryCache, createHttpLink, from, split, Observable } from '@apollo/client/core';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import { errorToast } from '@/components/ui/sonner';
import { useAuth } from '@/composables/useAuth';
import { openBugReport } from '@/composables/useReportBug';
import { logto } from '@/lib/logto';

const { activeWorkspace, login, refreshToken, bearerToken } = useAuth();

const httpLink = createHttpLink({ uri: '/graphql' });

const authLink = setContext(async (_, { headers }) => {
  const token = await bearerToken().catch(() => undefined);
  const accountToken = await logto.getAccessToken().catch(() => undefined);
  return {
    headers: {
      ...headers,
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(accountToken ? { 'X-Lucity-Account-Token': accountToken } : {}),
      'X-Lucity-Workspace': activeWorkspace.value,
    },
  };
});

const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
  const authFailed =
    graphQLErrors?.some(e => e.message === 'unauthenticated' || e.message === 'unauthorized') ||
    (!!networkError && 'statusCode' in networkError && (networkError.statusCode === 401 || networkError.statusCode === 403));

  if (authFailed) {
    if (operation.getContext().retried) {
      login();
      return;
    }
    operation.setContext({ retried: true });
    return new Observable(observer => {
      logto.clearAccessToken()
        .then(() => refreshToken())
        .then(() => forward(operation).subscribe(observer))
        .catch(() => {
          login();
          observer.error(new Error('unauthenticated'));
        });
    });
  }

  if (graphQLErrors) {
    const def = getMainDefinition(operation.query);
    if (def.kind === 'OperationDefinition' && def.operation === 'query') {
      const msg = graphQLErrors.map(e => e.message).join(', ');
      errorToast(msg, {
        action: { label: 'Report', onClick: () => openBugReport({ error: msg }) },
      });
    }
  }

  if (networkError) {
    errorToast('Network error', {
      description: networkError.message,
      action: { label: 'Report', onClick: () => openBugReport({ error: networkError.message }) },
    });
  }
});

const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

const wsLink = new GraphQLWsLink(createClient({
  url: `${wsProtocol}//${window.location.host}/graphql`,
  connectionParams: async () => {
    const token = await bearerToken().catch(() => undefined);
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
  from([errorLink, authLink, httpLink]),
);

export const apolloClient = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache(),
});
