import gql from 'graphql-tag';

export const DetectServicesQuery = gql`
  query DetectServices($sourceUrl: String!) {
    detectServices(sourceUrl: $sourceUrl) {
      name
      language
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
      framework
      sourceUrl
      contextPath
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
      rolloutHealth
      rolloutMessage
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
      rolloutHealth
      rolloutMessage
    }
  }
`;

export const RollbackMutation = gql`
  mutation Rollback($input: RollbackInput!) {
    rollback(input: $input)
  }
`;

export const SetServiceDomainMutation = gql`
  mutation SetServiceDomain($input: SetServiceDomainInput!) {
    setServiceDomain(input: $input)
  }
`;

export const DeployLogsSubscription = gql`
  subscription DeployLogs($id: ID!) {
    deployLogs(id: $id)
  }
`;

export const ServiceLogsSubscription = gql`
  subscription ServiceLogs($projectId: ID!, $service: String!, $environment: String!, $tailLines: Int) {
    serviceLogs(projectId: $projectId, service: $service, environment: $environment, tailLines: $tailLines) {
      line
      pod
    }
  }
`;
