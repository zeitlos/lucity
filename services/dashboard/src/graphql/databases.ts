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

export const DatabaseTablesQuery = gql`
  query DatabaseTables($projectId: ID!, $environment: String!, $database: String!) {
    databaseTables(projectId: $projectId, environment: $environment, database: $database) {
      name
      schema
      estimatedRows
      columns {
        name
        type
        nullable
        primaryKey
      }
    }
  }
`;

export const DatabaseTableDataQuery = gql`
  query DatabaseTableData(
    $projectId: ID!
    $environment: String!
    $database: String!
    $table: String!
    $schema: String
    $limit: Int
    $offset: Int
  ) {
    databaseTableData(
      projectId: $projectId
      environment: $environment
      database: $database
      table: $table
      schema: $schema
      limit: $limit
      offset: $offset
    ) {
      columns
      rows
      totalEstimatedRows
    }
  }
`;

export const ExecuteQueryMutation = gql`
  mutation ExecuteQuery($input: DatabaseQueryInput!) {
    executeQuery(input: $input) {
      columns
      rows
      affectedRows
    }
  }
`;

export const ConnectDatabaseMutation = gql`
  mutation ConnectDatabase($projectId: ID!, $environment: String!, $database: String!) {
    connectDatabase(projectId: $projectId, environment: $environment, database: $database)
  }
`;
