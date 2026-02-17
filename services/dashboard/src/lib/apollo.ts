import { ApolloClient, InMemoryCache, createHttpLink, from } from '@apollo/client/core';
import { onError } from '@apollo/client/link/error';
import router from '@/router';

const httpLink = createHttpLink({
  uri: '/graphql',
  credentials: 'include',
});

const errorLink = onError(({ graphQLErrors }) => {
  if (graphQLErrors) {
    for (const err of graphQLErrors) {
      if (err.message === 'unauthorized') {
        router.push('/login');
        return;
      }
    }
  }
});

export const apolloClient = new ApolloClient({
  link: from([errorLink, httpLink]),
  cache: new InMemoryCache(),
});
