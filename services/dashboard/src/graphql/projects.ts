import gql from 'graphql-tag';

export const ProjectsQuery = gql`
  query Projects {
    projects {
      id
      name
      sourceUrl
      createdAt
      environments {
        id
        name
        syncStatus
      }
    }
  }
`;

export const CreateProjectMutation = gql`
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      id
      name
      sourceUrl
      services {
        name
        image
        port
        public
        framework
      }
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

export const ProjectQuery = gql`
  query Project($id: ID!) {
    project(id: $id) {
      id
      name
      sourceUrl
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
          deployment {
            id
            imageTag
            active
          }
        }
      }
      services {
        name
        image
        port
        public
        framework
        instances {
          environment
          imageTag
          ready
          replicas
          deployment {
            id
            imageTag
            active
          }
        }
      }
    }
  }
`;
