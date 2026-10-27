import { ApolloError } from '@apollo/client/core';

export type ErrorKind = 'network' | 'unexpected' | 'explicit';

export interface DescribedError {
  kind: ErrorKind;
  message: string;
  detail: string;
}

const UNEXPECTED_MESSAGE = 'an unexpected error occurred';

function sentence(text: string): string {
  return text ? text.charAt(0).toUpperCase() + text.slice(1) : text;
}

/**
 * Sorts an error into what the person can do about it: wait for the network,
 * report a bug on our side, or act on the message the server gave.
 */
export function describeError(error: unknown): DescribedError {
  if (error instanceof ApolloError) {
    if (error.networkError) {
      return {
        kind: 'network',
        message: "Couldn't reach Lucity. Check your connection and try again.",
        detail: error.networkError.message,
      };
    }

    const first = error.graphQLErrors[0];

    if (!first || first.message === UNEXPECTED_MESSAGE) {
      return {
        kind: 'unexpected',
        message: 'Something went wrong on our side. Try again, or report it if it keeps happening.',
        detail: error.message,
      };
    }

    return { kind: 'explicit', message: sentence(first.message), detail: first.message };
  }

  if (error instanceof TypeError) {
    return {
      kind: 'network',
      message: "Couldn't reach Lucity. Check your connection and try again.",
      detail: error.message,
    };
  }

  const message = error instanceof Error ? error.message : String(error);
  return { kind: 'explicit', message: sentence(message), detail: message };
}
