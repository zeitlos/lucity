import gql from 'graphql-tag';

export const SharedVariablesQuery = gql`
  query SharedVariables($projectId: ID!, $environment: String!) {
    sharedVariables(projectId: $projectId, environment: $environment) {
      key
      value
    }
  }
`;

export const SetSharedVariablesMutation = gql`
  mutation SetSharedVariables($projectId: ID!, $environment: String!, $variables: [VariableInput!]!) {
    setSharedVariables(projectId: $projectId, environment: $environment, variables: $variables)
  }
`;

export const ServiceVariablesQuery = gql`
  query ServiceVariables($projectId: ID!, $environment: String!, $service: String!) {
    serviceVariables(projectId: $projectId, environment: $environment, service: $service) {
      key
      value
      fromShared
    }
  }
`;

export const SetServiceVariablesMutation = gql`
  mutation SetServiceVariables($projectId: ID!, $environment: String!, $service: String!, $variables: [ServiceVariableInput!]!) {
    setServiceVariables(projectId: $projectId, environment: $environment, service: $service, variables: $variables)
  }
`;
