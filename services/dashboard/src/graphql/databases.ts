import gql from 'graphql-tag';

export const CreateDatabaseMutation = gql`
  mutation CreateDatabase($input: CreateDatabaseInput!) {
    createDatabase(input: $input) {
      name
      version
      instances
      size
    }
  }
`;

export const DeleteDatabaseMutation = gql`
  mutation DeleteDatabase($projectId: ID!, $name: String!) {
    deleteDatabase(projectId: $projectId, name: $name)
  }
`;
