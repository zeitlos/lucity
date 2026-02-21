import gql from 'graphql-tag';

export const DetectServicesQuery = gql`
  query DetectServices($projectId: ID!) {
    detectServices(projectId: $projectId) {
      name
      provider
      framework
      startCommand
      suggestedPort
    }
  }
`;

export const AddServiceMutation = gql`
  mutation AddService($input: AddServiceInput!) {
    addService(input: $input) {
      name
      image
      port
      public
      framework
    }
  }
`;

export const RemoveServiceMutation = gql`
  mutation RemoveService($projectId: ID!, $service: String!) {
    removeService(projectId: $projectId, service: $service)
  }
`;

export const BuildServiceMutation = gql`
  mutation BuildService($input: BuildServiceInput!) {
    buildService(input: $input) {
      id
      phase
      imageRef
      digest
      error
    }
  }
`;

export const BuildStatusQuery = gql`
  query BuildStatus($id: ID!) {
    buildStatus(id: $id) {
      id
      phase
      imageRef
      digest
      error
    }
  }
`;

export const DeployBuildMutation = gql`
  mutation DeployBuild($input: DeployBuildInput!) {
    deployBuild(input: $input)
  }
`;

export const DeployMutation = gql`
  mutation Deploy($input: DeployInput!) {
    deploy(input: $input) {
      id
      phase
      buildId
      imageRef
      digest
      error
    }
  }
`;

export const DeployStatusQuery = gql`
  query DeployStatus($id: ID!) {
    deployStatus(id: $id) {
      id
      phase
      buildId
      imageRef
      digest
      error
      argoHealth
      argoMessage
    }
  }
`;

export const ActiveDeploymentQuery = gql`
  query ActiveDeployment($projectId: ID!, $service: String!, $environment: String!) {
    activeDeployment(projectId: $projectId, service: $service, environment: $environment) {
      id
      phase
      buildId
      imageRef
      digest
      error
      argoHealth
      argoMessage
    }
  }
`;
