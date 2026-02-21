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
      }
      services {
        name
        sourceUrl
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
      initialDeploys {
        id
        phase
      }
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
        services {
          name
          environment
          imageTag
          ready
          replicas
          host
          deployments {
            id
            imageTag
            active
            timestamp
            revision
            message
          }
        }
      }
      services {
        name
        image
        port
        framework
        sourceUrl
        contextPath
        instances {
          environment
          imageTag
          ready
          replicas
          host
          deployments {
            id
            imageTag
            active
            timestamp
            revision
            message
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
