import { ApolloClient, InMemoryCache, createHttpLink, from } from '@apollo/client/core';
import { setContext } from '@apollo/client/link/context';

import { environment } from '@/lib/environment';

const environmentLink = setContext((_, { headers }) => ({
  headers: environment.value ? { ...headers, 'X-Cockpit-Environment': environment.value } : headers,
}));

export const apolloClient = new ApolloClient({
  link: from([environmentLink, createHttpLink({ uri: '/graphql' })]),
  cache: new InMemoryCache(),
});
