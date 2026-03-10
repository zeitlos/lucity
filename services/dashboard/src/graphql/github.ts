import gql from 'graphql-tag';

export const GitHubConnectedQuery = gql`
  query GitHubConnected {
    githubConnected
  }
`;

export const GitHubSourcesQuery = gql`
  query GitHubSources {
    githubSources {
      id
      accountLogin
      accountAvatarUrl
      accountType
    }
  }
`;

export const GitHubRepositoriesQuery = gql`
  query GitHubRepositories($installationId: ID!) {
    githubRepositories(installationId: $installationId) {
      id
      name
      fullName
      htmlUrl
      defaultBranch
      private
    }
  }
`;
