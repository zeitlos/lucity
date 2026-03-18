import gql from 'graphql-tag';

export const DetectServicesQuery = gql`
  query DetectServices($sourceUrl: String!, $installationId: ID) {
    detectServices(sourceUrl: $sourceUrl, installationId: $installationId) {
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
      id
      name
      environment
      image
      port
      framework
      startCommand
      sourceUrl
      contextPath
      customStartCommand
      imageTag
      initialDeploy {
        id
        phase
      }
    }
  }
`;

export const SetCustomStartCommandMutation = gql`
  mutation SetCustomStartCommand($projectId: ID!, $environment: String!, $service: String!, $command: String!) {
    setCustomStartCommand(projectId: $projectId, environment: $environment, service: $service, command: $command)
  }
`;

export const RemoveServiceMutation = gql`
  mutation RemoveService($projectId: ID!, $environment: String!, $service: String!) {
    removeService(projectId: $projectId, environment: $environment, service: $service)
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
      startedAt
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
      startedAt
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
      startedAt
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

export const GenerateDomainMutation = gql`
  mutation GenerateDomain($input: GenerateDomainInput!) {
    generateDomain(input: $input) {
      hostname
      type
      dnsStatus
      tlsStatus
    }
  }
`;

export const AddCustomDomainMutation = gql`
  mutation AddCustomDomain($input: AddCustomDomainInput!) {
    addCustomDomain(input: $input) {
      hostname
      type
      dnsStatus
      tlsStatus
    }
  }
`;

export const RemoveDomainMutation = gql`
  mutation RemoveDomain($input: RemoveDomainInput!) {
    removeDomain(input: $input)
  }
`;

export const CheckDnsStatusQuery = gql`
  query CheckDnsStatus($hostname: String!) {
    checkDnsStatus(hostname: $hostname) {
      hostname
      status
      cnameTarget
      expectedTarget
      message
      tlsStatus
    }
  }
`;

export const PlatformConfigQuery = gql`
  query PlatformConfig {
    platformConfig {
      workloadDomain
      domainTarget
      ipAddress
    }
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
