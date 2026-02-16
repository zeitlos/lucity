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
          imageTag
          ready
          replicas
        }
      }
      services {
        name
        image
        port
        public
      }
    }
  }
`;
