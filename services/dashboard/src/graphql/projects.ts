import gql from 'graphql-tag';

export const ProjectsQuery = gql`
  query Projects {
    projects {
      id
      name
      createdAt
      environments {
        id
        name
        syncStatus
        resourceTier
        services {
          name
          sourceUrl
        }
      }
      databases {
        name
        version
      }
    }
  }
`;

export const CreateProjectMutation = gql`
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      id
      name
    }
  }
`;

export const DeleteProjectMutation = gql`
  mutation DeleteProject($id: ID!) {
    deleteProject(id: $id)
  }
`;

export const CreateEnvironmentMutation = gql`
  mutation CreateEnvironment($input: CreateEnvironmentInput!) {
    createEnvironment(input: $input) {
      id
      name
      namespace
      ephemeral
      syncStatus
      resourceTier
    }
  }
`;

export const DeleteEnvironmentMutation = gql`
  mutation DeleteEnvironment($projectId: ID!, $environment: String!) {
    deleteEnvironment(projectId: $projectId, environment: $environment)
  }
`;

export const SetServiceScalingMutation = gql`
  mutation SetServiceScaling($input: SetServiceScalingInput!) {
    setServiceScaling(input: $input) {
      replicas
      autoscaling {
        enabled
        minReplicas
        maxReplicas
        targetCPU
      }
    }
  }
`;

export const ProjectQuery = gql`
  query Project($id: ID!) {
    project(id: $id) {
      id
      name
      createdAt
      environments {
        id
        name
        namespace
        ephemeral
        syncStatus
        resourceTier
        services {
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
          ready
          replicas
          scaling {
            replicas
            autoscaling {
              enabled
              minReplicas
              maxReplicas
              targetCPU
            }
          }
          resources {
            cpuMillicores
            memoryMB
            cpuLimitMillicores
            memoryLimitMB
          }
          domains {
            hostname
            type
            dnsStatus
            tlsStatus
          }
          deployments {
            id
            imageTag
            active
            timestamp
            revision
            message
            sourceCommitMessage
            sourceUrl
          }
        }
        databases {
          name
          environment
          ready
          instances
          version
          size
          volume {
            name
            size
            requestedSize
            usedBytes
            capacityBytes
          }
        }
      }
      databases {
        name
        version
        instances
        size
      }
    }
  }
`;
