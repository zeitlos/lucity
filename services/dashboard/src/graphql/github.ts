import gql from 'graphql-tag';

export const GitHubRepositoriesQuery = gql`
  query GitHubRepositories {
    githubRepositories {
      id
      name
      fullName
      htmlUrl
      defaultBranch
      private
    }
  }
`;

export const GitHubInstallationsQuery = gql`
  query GitHubInstallations {
    githubInstallations {
      id
      accountLogin
      accountAvatarUrl
      accountType
    }
  }
`;
