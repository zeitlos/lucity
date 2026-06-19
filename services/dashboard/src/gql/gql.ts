/* eslint-disable */
import * as types from './graphql';
import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 * Learn more about it here: https://the-guild.dev/graphql/codegen/plugins/presets/preset-client#reducing-bundle-size
 */
type Documents = {
    "\n  query Workspaces {\n    workspaces {\n      id\n      name\n      personal\n    }\n  }\n": typeof types.WorkspacesDocument,
    "\n  query ProjectsForNav {\n    projects {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n": typeof types.ProjectsForNavDocument,
    "\n  query GitHubConnected {\n    githubConnected\n  }\n": typeof types.GitHubConnectedDocument,
    "\n  query GitHubSources {\n    githubSources {\n      accountLogin\n      accountAvatarUrl\n      accountType\n    }\n  }\n": typeof types.GitHubSourcesDocument,
    "\n  query GitHubRepositories($account: String!) {\n    githubRepositories(account: $account) {\n      id\n      name\n      fullName\n      htmlUrl\n      defaultBranch\n      private\n    }\n  }\n": typeof types.GitHubRepositoriesDocument,
    "\n  mutation CreateProject($input: CreateProjectInput!) {\n    createProject(input: $input) {\n      id\n      name\n      environments {\n        id\n        name\n      }\n    }\n  }\n": typeof types.CreateProjectDocument,
    "\n  mutation AddService($environmentId: EnvironmentID!, $input: AddServiceInput!) {\n    addService(environment: $environmentId, input: $input) {\n      id\n      name\n    }\n  }\n": typeof types.AddServiceDocument,
    "\n  query DetectServices($repositoryUrl: String!) {\n    detectServices(repositoryUrl: $repositoryUrl) {\n      name\n      language\n      framework\n      startCommand\n      suggestedPort\n    }\n  }\n": typeof types.DetectServicesDocument,
    "\n  query SearchImages($query: String!) {\n    searchImages(query: $query) {\n      name\n      description\n      starCount\n      pullCount\n      official\n    }\n  }\n": typeof types.SearchImagesDocument,
    "\n  mutation CreateDatabase($input: CreateDatabaseInput!) {\n    createDatabase(input: $input) {\n      id\n      name\n      version\n      instances\n      size\n    }\n  }\n": typeof types.CreateDatabaseDocument,
    "\n  mutation CreateEnvironment($input: CreateEnvironmentInput!) {\n    createEnvironment(input: $input) {\n      id\n      name\n      resourceTier\n    }\n  }\n": typeof types.CreateEnvironmentDocument,
    "\n  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {\n    createWorkspaceCheckout(input: $input) {\n      url\n    }\n  }\n": typeof types.CreateWorkspaceCheckoutDocument,
    "\n  query SharedVariables($environment: EnvironmentID!) {\n    sharedVariables(environment: $environment) {\n      key\n      value\n    }\n  }\n": typeof types.SharedVariablesDocument,
    "\n  mutation SetSharedVariables($environment: EnvironmentID!, $variables: [VariableInput!]!) {\n    setSharedVariables(environment: $environment, variables: $variables)\n  }\n": typeof types.SetSharedVariablesDocument,
    "\n  query Subscription {\n    subscription {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n      hasPaymentMethod\n    }\n  }\n": typeof types.SubscriptionDocument,
    "\n  query UsageSummary {\n    usageSummary {\n      resourceCostCents\n      creditsCents\n      estimatedTotalCents\n    }\n  }\n": typeof types.UsageSummaryDocument,
    "\n  mutation CreatePlanCheckout($plan: Plan!) {\n    createPlanCheckout(plan: $plan) {\n      url\n    }\n  }\n": typeof types.CreatePlanCheckoutDocument,
    "\n  query DatabasePublic($database: DatabaseID!) {\n    database(id: $database) {\n      id\n      public\n    }\n  }\n": typeof types.DatabasePublicDocument,
    "\n  query DatabaseCredentials($database: DatabaseID!) {\n    databaseCredentials(database: $database) {\n      type\n      host\n      port\n      dbname\n      user\n      password\n      uri\n    }\n  }\n": typeof types.DatabaseCredentialsDocument,
    "\n  mutation ExposeDatabase($database: DatabaseID!) {\n    exposeDatabase(database: $database) {\n      id\n      public\n    }\n  }\n": typeof types.ExposeDatabaseDocument,
    "\n  mutation UnexposeDatabase($database: DatabaseID!) {\n    unexposeDatabase(database: $database) {\n      id\n      public\n    }\n  }\n": typeof types.UnexposeDatabaseDocument,
    "\n  mutation ExecuteQuery($database: DatabaseID!, $query: String!) {\n    executeQuery(database: $database, query: $query) {\n      columns\n      rows\n      affectedRows\n    }\n  }\n": typeof types.ExecuteQueryDocument,
    "\n  mutation DeleteDatabase($database: DatabaseID!) {\n    deleteDatabase(database: $database)\n  }\n": typeof types.DeleteDatabaseDocument,
    "\n  query DatabaseTables($database: DatabaseID!) {\n    databaseTables(database: $database) {\n      name\n      schema\n      estimatedRows\n      columns {\n        name\n        type\n        nullable\n        primaryKey\n      }\n    }\n  }\n": typeof types.DatabaseTablesDocument,
    "\n  query DatabaseTableData(\n    $database: DatabaseID!\n    $table: String!\n    $schema: String\n    $limit: Int\n    $offset: Int\n  ) {\n    databaseTableData(\n      database: $database\n      table: $table\n      schema: $schema\n      limit: $limit\n      offset: $offset\n    ) {\n      columns\n      rows\n      totalEstimatedRows\n    }\n  }\n": typeof types.DatabaseTableDataDocument,
    "\n  mutation RemoveService($service: ServiceID!) {\n    removeService(service: $service)\n  }\n": typeof types.RemoveServiceDocument,
    "\n  mutation SetCustomStartCommand($service: ServiceID!, $command: String!) {\n    setCustomStartCommand(service: $service, command: $command) {\n      id\n    }\n  }\n": typeof types.SetCustomStartCommandDocument,
    "\n  mutation GenerateDomain($service: ServiceID!) {\n    generateDomain(service: $service) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n": typeof types.GenerateDomainDocument,
    "\n  mutation AddCustomDomain($service: ServiceID!, $hostname: String!) {\n    addCustomDomain(service: $service, hostname: $hostname) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n": typeof types.AddCustomDomainDocument,
    "\n  mutation RemoveDomain($service: ServiceID!, $hostname: String!) {\n    removeDomain(service: $service, hostname: $hostname) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n": typeof types.RemoveDomainDocument,
    "\n  mutation SetServiceScaling($input: SetServiceScalingInput!) {\n    setServiceScaling(input: $input) {\n      id\n      replicas {\n        desired\n        ready\n      }\n      autoscaling {\n        minReplicas\n        maxReplicas\n        targetCpu\n      }\n    }\n  }\n": typeof types.SetServiceScalingDocument,
    "\n  mutation SetServicePort($service: ServiceID!, $port: Int) {\n    setServicePort(service: $service, port: $port) {\n      id\n      port\n    }\n  }\n": typeof types.SetServicePortDocument,
    "\n  mutation SetServiceResources($service: ServiceID!, $resources: ResourcesInput!) {\n    setServiceResources(service: $service, resources: $resources) {\n      id\n      resources {\n        cpu\n        memory\n      }\n    }\n  }\n": typeof types.SetServiceResourcesDocument,
    "\n  query ServiceVariables($service: ServiceID!) {\n    serviceVariables(service: $service) {\n      key\n      value\n      fromShared\n      databaseRef {\n        database\n        key\n      }\n    }\n  }\n": typeof types.ServiceVariablesDocument,
    "\n  mutation SetServiceVariables($service: ServiceID!, $variables: [ServiceVariableInput!]!) {\n    setServiceVariables(service: $service, variables: $variables)\n  }\n": typeof types.SetServiceVariablesDocument,
    "\n  subscription BuildLogs($id: BuildID!) {\n    buildLogs(id: $id)\n  }\n": typeof types.BuildLogsDocument,
    "\n  query CanvasServiceBuilds($id: ServiceID!) {\n    service(id: $id) {\n      id\n      builds {\n        id\n        status\n        startedAt\n        finishedAt\n      }\n    }\n  }\n": typeof types.CanvasServiceBuildsDocument,
    "\n  mutation Deploy($service: ServiceID!, $gitRef: String) {\n    deploy(service: $service, gitRef: $gitRef) {\n      id\n      status\n      startedAt\n      finishedAt\n    }\n  }\n": typeof types.DeployDocument,
    "\n  query BuildStatus($id: BuildID!) {\n    build(id: $id) {\n      id\n      status\n      startedAt\n      finishedAt\n    }\n  }\n": typeof types.BuildStatusDocument,
    "\n  subscription ServiceLogs($service: ServiceID!, $tailLines: Int) {\n    serviceLogs(service: $service, tailLines: $tailLines) {\n      line\n      pod\n    }\n  }\n": typeof types.ServiceLogsDocument,
    "\n  query Workspace {\n    workspace {\n      id\n      name\n      personal\n      suspended\n      members {\n        id\n        email\n        name\n        role\n      }\n    }\n  }\n": typeof types.WorkspaceDocument,
    "\n  mutation CompleteWorkspaceCheckout($sessionId: String!) {\n    completeWorkspaceCheckout(sessionId: $sessionId) {\n      id\n      name\n      personal\n    }\n  }\n": typeof types.CompleteWorkspaceCheckoutDocument,
    "\n  query Environment($environment: EnvironmentID!) {\n    environment(environment: $environment) {\n      id\n      name\n      resourceTier\n      services {\n        id\n        name\n        status\n        replicas {\n          desired\n          ready\n        }\n        autoscaling {\n          minReplicas\n          maxReplicas\n          targetCpu\n        }\n        port\n        endpoints {\n          host\n          port\n          protocol\n          type\n          dns {\n            status\n            requiredRecords {\n              type\n              host\n              value\n            }\n          }\n          tls\n        }\n        sourceUrl\n        contextPath\n        resources {\n          cpu\n          memory\n        }\n        command\n        defaultCommand\n        activeDeployment {\n          id\n          image\n          imageDigest\n          commit\n          commitMessage\n          ref\n          status\n          createdAt\n        }\n        deployments {\n          id\n          image\n          imageDigest\n          commit\n          commitMessage\n          ref\n          status\n          createdAt\n        }\n        builds {\n          id\n          status\n          startedAt\n          finishedAt\n        }\n        lastDeployedAt\n        createdAt\n      }\n      databases {\n        id\n        name\n        version\n        instances\n        status\n        size\n        createdAt\n      }\n    }\n  }\n": typeof types.EnvironmentDocument,
    "\n  query ProjectEnvironments($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n": typeof types.ProjectEnvironmentsDocument,
    "\n  mutation CompletePlanCheckout($sessionId: String!) {\n    completePlanCheckout(sessionId: $sessionId) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      hasPaymentMethod\n    }\n  }\n": typeof types.CompletePlanCheckoutDocument,
    "\n  query ProjectPage($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n": typeof types.ProjectPageDocument,
    "\n  query ProjectSettings($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n": typeof types.ProjectSettingsDocument,
    "\n  mutation DeleteProject($id: ProjectID!) {\n    deleteProject(id: $id)\n  }\n": typeof types.DeleteProjectDocument,
    "\n  mutation DeleteEnvironment($environment: EnvironmentID!) {\n    deleteEnvironment(environment: $environment)\n  }\n": typeof types.DeleteEnvironmentDocument,
    "\n  query EnvironmentResources($environment: EnvironmentID!) {\n    environmentResources(environment: $environment) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n": typeof types.EnvironmentResourcesDocument,
    "\n  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {\n    setEnvironmentResources(input: $input) {\n      id\n      resourceTier\n    }\n  }\n": typeof types.SetEnvironmentResourcesDocument,
    "\n  query Projects {\n    projects {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n        services {\n          id\n          name\n          sourceUrl\n        }\n      }\n    }\n  }\n": typeof types.ProjectsDocument,
    "\n  mutation UpdateWorkspace($input: UpdateWorkspaceInput!) {\n    updateWorkspace(input: $input) {\n      id\n      name\n    }\n  }\n": typeof types.UpdateWorkspaceDocument,
    "\n  mutation DeleteWorkspace {\n    deleteWorkspace\n  }\n": typeof types.DeleteWorkspaceDocument,
    "\n  mutation InviteMember($input: InviteMemberInput!) {\n    inviteMember(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n": typeof types.InviteMemberDocument,
    "\n  mutation RemoveMember($userId: ID!) {\n    removeMember(userId: $userId)\n  }\n": typeof types.RemoveMemberDocument,
    "\n  mutation UpdateMemberRole($input: UpdateMemberRoleInput!) {\n    updateMemberRole(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n": typeof types.UpdateMemberRoleDocument,
    "\n  mutation ChangePlan($plan: Plan!) {\n    changePlan(plan: $plan) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n    }\n  }\n": typeof types.ChangePlanDocument,
    "\n  mutation BillingPortalUrl {\n    billingPortalUrl {\n      url\n    }\n  }\n": typeof types.BillingPortalUrlDocument,
};
const documents: Documents = {
    "\n  query Workspaces {\n    workspaces {\n      id\n      name\n      personal\n    }\n  }\n": types.WorkspacesDocument,
    "\n  query ProjectsForNav {\n    projects {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n": types.ProjectsForNavDocument,
    "\n  query GitHubConnected {\n    githubConnected\n  }\n": types.GitHubConnectedDocument,
    "\n  query GitHubSources {\n    githubSources {\n      accountLogin\n      accountAvatarUrl\n      accountType\n    }\n  }\n": types.GitHubSourcesDocument,
    "\n  query GitHubRepositories($account: String!) {\n    githubRepositories(account: $account) {\n      id\n      name\n      fullName\n      htmlUrl\n      defaultBranch\n      private\n    }\n  }\n": types.GitHubRepositoriesDocument,
    "\n  mutation CreateProject($input: CreateProjectInput!) {\n    createProject(input: $input) {\n      id\n      name\n      environments {\n        id\n        name\n      }\n    }\n  }\n": types.CreateProjectDocument,
    "\n  mutation AddService($environmentId: EnvironmentID!, $input: AddServiceInput!) {\n    addService(environment: $environmentId, input: $input) {\n      id\n      name\n    }\n  }\n": types.AddServiceDocument,
    "\n  query DetectServices($repositoryUrl: String!) {\n    detectServices(repositoryUrl: $repositoryUrl) {\n      name\n      language\n      framework\n      startCommand\n      suggestedPort\n    }\n  }\n": types.DetectServicesDocument,
    "\n  query SearchImages($query: String!) {\n    searchImages(query: $query) {\n      name\n      description\n      starCount\n      pullCount\n      official\n    }\n  }\n": types.SearchImagesDocument,
    "\n  mutation CreateDatabase($input: CreateDatabaseInput!) {\n    createDatabase(input: $input) {\n      id\n      name\n      version\n      instances\n      size\n    }\n  }\n": types.CreateDatabaseDocument,
    "\n  mutation CreateEnvironment($input: CreateEnvironmentInput!) {\n    createEnvironment(input: $input) {\n      id\n      name\n      resourceTier\n    }\n  }\n": types.CreateEnvironmentDocument,
    "\n  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {\n    createWorkspaceCheckout(input: $input) {\n      url\n    }\n  }\n": types.CreateWorkspaceCheckoutDocument,
    "\n  query SharedVariables($environment: EnvironmentID!) {\n    sharedVariables(environment: $environment) {\n      key\n      value\n    }\n  }\n": types.SharedVariablesDocument,
    "\n  mutation SetSharedVariables($environment: EnvironmentID!, $variables: [VariableInput!]!) {\n    setSharedVariables(environment: $environment, variables: $variables)\n  }\n": types.SetSharedVariablesDocument,
    "\n  query Subscription {\n    subscription {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n      hasPaymentMethod\n    }\n  }\n": types.SubscriptionDocument,
    "\n  query UsageSummary {\n    usageSummary {\n      resourceCostCents\n      creditsCents\n      estimatedTotalCents\n    }\n  }\n": types.UsageSummaryDocument,
    "\n  mutation CreatePlanCheckout($plan: Plan!) {\n    createPlanCheckout(plan: $plan) {\n      url\n    }\n  }\n": types.CreatePlanCheckoutDocument,
    "\n  query DatabasePublic($database: DatabaseID!) {\n    database(id: $database) {\n      id\n      public\n    }\n  }\n": types.DatabasePublicDocument,
    "\n  query DatabaseCredentials($database: DatabaseID!) {\n    databaseCredentials(database: $database) {\n      type\n      host\n      port\n      dbname\n      user\n      password\n      uri\n    }\n  }\n": types.DatabaseCredentialsDocument,
    "\n  mutation ExposeDatabase($database: DatabaseID!) {\n    exposeDatabase(database: $database) {\n      id\n      public\n    }\n  }\n": types.ExposeDatabaseDocument,
    "\n  mutation UnexposeDatabase($database: DatabaseID!) {\n    unexposeDatabase(database: $database) {\n      id\n      public\n    }\n  }\n": types.UnexposeDatabaseDocument,
    "\n  mutation ExecuteQuery($database: DatabaseID!, $query: String!) {\n    executeQuery(database: $database, query: $query) {\n      columns\n      rows\n      affectedRows\n    }\n  }\n": types.ExecuteQueryDocument,
    "\n  mutation DeleteDatabase($database: DatabaseID!) {\n    deleteDatabase(database: $database)\n  }\n": types.DeleteDatabaseDocument,
    "\n  query DatabaseTables($database: DatabaseID!) {\n    databaseTables(database: $database) {\n      name\n      schema\n      estimatedRows\n      columns {\n        name\n        type\n        nullable\n        primaryKey\n      }\n    }\n  }\n": types.DatabaseTablesDocument,
    "\n  query DatabaseTableData(\n    $database: DatabaseID!\n    $table: String!\n    $schema: String\n    $limit: Int\n    $offset: Int\n  ) {\n    databaseTableData(\n      database: $database\n      table: $table\n      schema: $schema\n      limit: $limit\n      offset: $offset\n    ) {\n      columns\n      rows\n      totalEstimatedRows\n    }\n  }\n": types.DatabaseTableDataDocument,
    "\n  mutation RemoveService($service: ServiceID!) {\n    removeService(service: $service)\n  }\n": types.RemoveServiceDocument,
    "\n  mutation SetCustomStartCommand($service: ServiceID!, $command: String!) {\n    setCustomStartCommand(service: $service, command: $command) {\n      id\n    }\n  }\n": types.SetCustomStartCommandDocument,
    "\n  mutation GenerateDomain($service: ServiceID!) {\n    generateDomain(service: $service) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n": types.GenerateDomainDocument,
    "\n  mutation AddCustomDomain($service: ServiceID!, $hostname: String!) {\n    addCustomDomain(service: $service, hostname: $hostname) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n": types.AddCustomDomainDocument,
    "\n  mutation RemoveDomain($service: ServiceID!, $hostname: String!) {\n    removeDomain(service: $service, hostname: $hostname) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n": types.RemoveDomainDocument,
    "\n  mutation SetServiceScaling($input: SetServiceScalingInput!) {\n    setServiceScaling(input: $input) {\n      id\n      replicas {\n        desired\n        ready\n      }\n      autoscaling {\n        minReplicas\n        maxReplicas\n        targetCpu\n      }\n    }\n  }\n": types.SetServiceScalingDocument,
    "\n  mutation SetServicePort($service: ServiceID!, $port: Int) {\n    setServicePort(service: $service, port: $port) {\n      id\n      port\n    }\n  }\n": types.SetServicePortDocument,
    "\n  mutation SetServiceResources($service: ServiceID!, $resources: ResourcesInput!) {\n    setServiceResources(service: $service, resources: $resources) {\n      id\n      resources {\n        cpu\n        memory\n      }\n    }\n  }\n": types.SetServiceResourcesDocument,
    "\n  query ServiceVariables($service: ServiceID!) {\n    serviceVariables(service: $service) {\n      key\n      value\n      fromShared\n      databaseRef {\n        database\n        key\n      }\n    }\n  }\n": types.ServiceVariablesDocument,
    "\n  mutation SetServiceVariables($service: ServiceID!, $variables: [ServiceVariableInput!]!) {\n    setServiceVariables(service: $service, variables: $variables)\n  }\n": types.SetServiceVariablesDocument,
    "\n  subscription BuildLogs($id: BuildID!) {\n    buildLogs(id: $id)\n  }\n": types.BuildLogsDocument,
    "\n  query CanvasServiceBuilds($id: ServiceID!) {\n    service(id: $id) {\n      id\n      builds {\n        id\n        status\n        startedAt\n        finishedAt\n      }\n    }\n  }\n": types.CanvasServiceBuildsDocument,
    "\n  mutation Deploy($service: ServiceID!, $gitRef: String) {\n    deploy(service: $service, gitRef: $gitRef) {\n      id\n      status\n      startedAt\n      finishedAt\n    }\n  }\n": types.DeployDocument,
    "\n  query BuildStatus($id: BuildID!) {\n    build(id: $id) {\n      id\n      status\n      startedAt\n      finishedAt\n    }\n  }\n": types.BuildStatusDocument,
    "\n  subscription ServiceLogs($service: ServiceID!, $tailLines: Int) {\n    serviceLogs(service: $service, tailLines: $tailLines) {\n      line\n      pod\n    }\n  }\n": types.ServiceLogsDocument,
    "\n  query Workspace {\n    workspace {\n      id\n      name\n      personal\n      suspended\n      members {\n        id\n        email\n        name\n        role\n      }\n    }\n  }\n": types.WorkspaceDocument,
    "\n  mutation CompleteWorkspaceCheckout($sessionId: String!) {\n    completeWorkspaceCheckout(sessionId: $sessionId) {\n      id\n      name\n      personal\n    }\n  }\n": types.CompleteWorkspaceCheckoutDocument,
    "\n  query Environment($environment: EnvironmentID!) {\n    environment(environment: $environment) {\n      id\n      name\n      resourceTier\n      services {\n        id\n        name\n        status\n        replicas {\n          desired\n          ready\n        }\n        autoscaling {\n          minReplicas\n          maxReplicas\n          targetCpu\n        }\n        port\n        endpoints {\n          host\n          port\n          protocol\n          type\n          dns {\n            status\n            requiredRecords {\n              type\n              host\n              value\n            }\n          }\n          tls\n        }\n        sourceUrl\n        contextPath\n        resources {\n          cpu\n          memory\n        }\n        command\n        defaultCommand\n        activeDeployment {\n          id\n          image\n          imageDigest\n          commit\n          commitMessage\n          ref\n          status\n          createdAt\n        }\n        deployments {\n          id\n          image\n          imageDigest\n          commit\n          commitMessage\n          ref\n          status\n          createdAt\n        }\n        builds {\n          id\n          status\n          startedAt\n          finishedAt\n        }\n        lastDeployedAt\n        createdAt\n      }\n      databases {\n        id\n        name\n        version\n        instances\n        status\n        size\n        createdAt\n      }\n    }\n  }\n": types.EnvironmentDocument,
    "\n  query ProjectEnvironments($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n": types.ProjectEnvironmentsDocument,
    "\n  mutation CompletePlanCheckout($sessionId: String!) {\n    completePlanCheckout(sessionId: $sessionId) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      hasPaymentMethod\n    }\n  }\n": types.CompletePlanCheckoutDocument,
    "\n  query ProjectPage($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n": types.ProjectPageDocument,
    "\n  query ProjectSettings($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n": types.ProjectSettingsDocument,
    "\n  mutation DeleteProject($id: ProjectID!) {\n    deleteProject(id: $id)\n  }\n": types.DeleteProjectDocument,
    "\n  mutation DeleteEnvironment($environment: EnvironmentID!) {\n    deleteEnvironment(environment: $environment)\n  }\n": types.DeleteEnvironmentDocument,
    "\n  query EnvironmentResources($environment: EnvironmentID!) {\n    environmentResources(environment: $environment) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n": types.EnvironmentResourcesDocument,
    "\n  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {\n    setEnvironmentResources(input: $input) {\n      id\n      resourceTier\n    }\n  }\n": types.SetEnvironmentResourcesDocument,
    "\n  query Projects {\n    projects {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n        services {\n          id\n          name\n          sourceUrl\n        }\n      }\n    }\n  }\n": types.ProjectsDocument,
    "\n  mutation UpdateWorkspace($input: UpdateWorkspaceInput!) {\n    updateWorkspace(input: $input) {\n      id\n      name\n    }\n  }\n": types.UpdateWorkspaceDocument,
    "\n  mutation DeleteWorkspace {\n    deleteWorkspace\n  }\n": types.DeleteWorkspaceDocument,
    "\n  mutation InviteMember($input: InviteMemberInput!) {\n    inviteMember(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n": types.InviteMemberDocument,
    "\n  mutation RemoveMember($userId: ID!) {\n    removeMember(userId: $userId)\n  }\n": types.RemoveMemberDocument,
    "\n  mutation UpdateMemberRole($input: UpdateMemberRoleInput!) {\n    updateMemberRole(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n": types.UpdateMemberRoleDocument,
    "\n  mutation ChangePlan($plan: Plan!) {\n    changePlan(plan: $plan) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n    }\n  }\n": types.ChangePlanDocument,
    "\n  mutation BillingPortalUrl {\n    billingPortalUrl {\n      url\n    }\n  }\n": types.BillingPortalUrlDocument,
};

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = graphql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function graphql(source: string): unknown;

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Workspaces {\n    workspaces {\n      id\n      name\n      personal\n    }\n  }\n"): (typeof documents)["\n  query Workspaces {\n    workspaces {\n      id\n      name\n      personal\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ProjectsForNav {\n    projects {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n"): (typeof documents)["\n  query ProjectsForNav {\n    projects {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GitHubConnected {\n    githubConnected\n  }\n"): (typeof documents)["\n  query GitHubConnected {\n    githubConnected\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GitHubSources {\n    githubSources {\n      accountLogin\n      accountAvatarUrl\n      accountType\n    }\n  }\n"): (typeof documents)["\n  query GitHubSources {\n    githubSources {\n      accountLogin\n      accountAvatarUrl\n      accountType\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GitHubRepositories($account: String!) {\n    githubRepositories(account: $account) {\n      id\n      name\n      fullName\n      htmlUrl\n      defaultBranch\n      private\n    }\n  }\n"): (typeof documents)["\n  query GitHubRepositories($account: String!) {\n    githubRepositories(account: $account) {\n      id\n      name\n      fullName\n      htmlUrl\n      defaultBranch\n      private\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateProject($input: CreateProjectInput!) {\n    createProject(input: $input) {\n      id\n      name\n      environments {\n        id\n        name\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation CreateProject($input: CreateProjectInput!) {\n    createProject(input: $input) {\n      id\n      name\n      environments {\n        id\n        name\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation AddService($environmentId: EnvironmentID!, $input: AddServiceInput!) {\n    addService(environment: $environmentId, input: $input) {\n      id\n      name\n    }\n  }\n"): (typeof documents)["\n  mutation AddService($environmentId: EnvironmentID!, $input: AddServiceInput!) {\n    addService(environment: $environmentId, input: $input) {\n      id\n      name\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DetectServices($repositoryUrl: String!) {\n    detectServices(repositoryUrl: $repositoryUrl) {\n      name\n      language\n      framework\n      startCommand\n      suggestedPort\n    }\n  }\n"): (typeof documents)["\n  query DetectServices($repositoryUrl: String!) {\n    detectServices(repositoryUrl: $repositoryUrl) {\n      name\n      language\n      framework\n      startCommand\n      suggestedPort\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query SearchImages($query: String!) {\n    searchImages(query: $query) {\n      name\n      description\n      starCount\n      pullCount\n      official\n    }\n  }\n"): (typeof documents)["\n  query SearchImages($query: String!) {\n    searchImages(query: $query) {\n      name\n      description\n      starCount\n      pullCount\n      official\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateDatabase($input: CreateDatabaseInput!) {\n    createDatabase(input: $input) {\n      id\n      name\n      version\n      instances\n      size\n    }\n  }\n"): (typeof documents)["\n  mutation CreateDatabase($input: CreateDatabaseInput!) {\n    createDatabase(input: $input) {\n      id\n      name\n      version\n      instances\n      size\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateEnvironment($input: CreateEnvironmentInput!) {\n    createEnvironment(input: $input) {\n      id\n      name\n      resourceTier\n    }\n  }\n"): (typeof documents)["\n  mutation CreateEnvironment($input: CreateEnvironmentInput!) {\n    createEnvironment(input: $input) {\n      id\n      name\n      resourceTier\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {\n    createWorkspaceCheckout(input: $input) {\n      url\n    }\n  }\n"): (typeof documents)["\n  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {\n    createWorkspaceCheckout(input: $input) {\n      url\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query SharedVariables($environment: EnvironmentID!) {\n    sharedVariables(environment: $environment) {\n      key\n      value\n    }\n  }\n"): (typeof documents)["\n  query SharedVariables($environment: EnvironmentID!) {\n    sharedVariables(environment: $environment) {\n      key\n      value\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetSharedVariables($environment: EnvironmentID!, $variables: [VariableInput!]!) {\n    setSharedVariables(environment: $environment, variables: $variables)\n  }\n"): (typeof documents)["\n  mutation SetSharedVariables($environment: EnvironmentID!, $variables: [VariableInput!]!) {\n    setSharedVariables(environment: $environment, variables: $variables)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Subscription {\n    subscription {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n      hasPaymentMethod\n    }\n  }\n"): (typeof documents)["\n  query Subscription {\n    subscription {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n      hasPaymentMethod\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query UsageSummary {\n    usageSummary {\n      resourceCostCents\n      creditsCents\n      estimatedTotalCents\n    }\n  }\n"): (typeof documents)["\n  query UsageSummary {\n    usageSummary {\n      resourceCostCents\n      creditsCents\n      estimatedTotalCents\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreatePlanCheckout($plan: Plan!) {\n    createPlanCheckout(plan: $plan) {\n      url\n    }\n  }\n"): (typeof documents)["\n  mutation CreatePlanCheckout($plan: Plan!) {\n    createPlanCheckout(plan: $plan) {\n      url\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DatabasePublic($database: DatabaseID!) {\n    database(id: $database) {\n      id\n      public\n    }\n  }\n"): (typeof documents)["\n  query DatabasePublic($database: DatabaseID!) {\n    database(id: $database) {\n      id\n      public\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DatabaseCredentials($database: DatabaseID!) {\n    databaseCredentials(database: $database) {\n      type\n      host\n      port\n      dbname\n      user\n      password\n      uri\n    }\n  }\n"): (typeof documents)["\n  query DatabaseCredentials($database: DatabaseID!) {\n    databaseCredentials(database: $database) {\n      type\n      host\n      port\n      dbname\n      user\n      password\n      uri\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation ExposeDatabase($database: DatabaseID!) {\n    exposeDatabase(database: $database) {\n      id\n      public\n    }\n  }\n"): (typeof documents)["\n  mutation ExposeDatabase($database: DatabaseID!) {\n    exposeDatabase(database: $database) {\n      id\n      public\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation UnexposeDatabase($database: DatabaseID!) {\n    unexposeDatabase(database: $database) {\n      id\n      public\n    }\n  }\n"): (typeof documents)["\n  mutation UnexposeDatabase($database: DatabaseID!) {\n    unexposeDatabase(database: $database) {\n      id\n      public\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation ExecuteQuery($database: DatabaseID!, $query: String!) {\n    executeQuery(database: $database, query: $query) {\n      columns\n      rows\n      affectedRows\n    }\n  }\n"): (typeof documents)["\n  mutation ExecuteQuery($database: DatabaseID!, $query: String!) {\n    executeQuery(database: $database, query: $query) {\n      columns\n      rows\n      affectedRows\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation DeleteDatabase($database: DatabaseID!) {\n    deleteDatabase(database: $database)\n  }\n"): (typeof documents)["\n  mutation DeleteDatabase($database: DatabaseID!) {\n    deleteDatabase(database: $database)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DatabaseTables($database: DatabaseID!) {\n    databaseTables(database: $database) {\n      name\n      schema\n      estimatedRows\n      columns {\n        name\n        type\n        nullable\n        primaryKey\n      }\n    }\n  }\n"): (typeof documents)["\n  query DatabaseTables($database: DatabaseID!) {\n    databaseTables(database: $database) {\n      name\n      schema\n      estimatedRows\n      columns {\n        name\n        type\n        nullable\n        primaryKey\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DatabaseTableData(\n    $database: DatabaseID!\n    $table: String!\n    $schema: String\n    $limit: Int\n    $offset: Int\n  ) {\n    databaseTableData(\n      database: $database\n      table: $table\n      schema: $schema\n      limit: $limit\n      offset: $offset\n    ) {\n      columns\n      rows\n      totalEstimatedRows\n    }\n  }\n"): (typeof documents)["\n  query DatabaseTableData(\n    $database: DatabaseID!\n    $table: String!\n    $schema: String\n    $limit: Int\n    $offset: Int\n  ) {\n    databaseTableData(\n      database: $database\n      table: $table\n      schema: $schema\n      limit: $limit\n      offset: $offset\n    ) {\n      columns\n      rows\n      totalEstimatedRows\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation RemoveService($service: ServiceID!) {\n    removeService(service: $service)\n  }\n"): (typeof documents)["\n  mutation RemoveService($service: ServiceID!) {\n    removeService(service: $service)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetCustomStartCommand($service: ServiceID!, $command: String!) {\n    setCustomStartCommand(service: $service, command: $command) {\n      id\n    }\n  }\n"): (typeof documents)["\n  mutation SetCustomStartCommand($service: ServiceID!, $command: String!) {\n    setCustomStartCommand(service: $service, command: $command) {\n      id\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation GenerateDomain($service: ServiceID!) {\n    generateDomain(service: $service) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation GenerateDomain($service: ServiceID!) {\n    generateDomain(service: $service) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation AddCustomDomain($service: ServiceID!, $hostname: String!) {\n    addCustomDomain(service: $service, hostname: $hostname) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation AddCustomDomain($service: ServiceID!, $hostname: String!) {\n    addCustomDomain(service: $service, hostname: $hostname) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation RemoveDomain($service: ServiceID!, $hostname: String!) {\n    removeDomain(service: $service, hostname: $hostname) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation RemoveDomain($service: ServiceID!, $hostname: String!) {\n    removeDomain(service: $service, hostname: $hostname) {\n      id\n      endpoints {\n        host\n        port\n        type\n        protocol\n        dns {\n          status\n          requiredRecords {\n            type\n            host\n            value\n          }\n        }\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetServiceScaling($input: SetServiceScalingInput!) {\n    setServiceScaling(input: $input) {\n      id\n      replicas {\n        desired\n        ready\n      }\n      autoscaling {\n        minReplicas\n        maxReplicas\n        targetCpu\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation SetServiceScaling($input: SetServiceScalingInput!) {\n    setServiceScaling(input: $input) {\n      id\n      replicas {\n        desired\n        ready\n      }\n      autoscaling {\n        minReplicas\n        maxReplicas\n        targetCpu\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetServicePort($service: ServiceID!, $port: Int) {\n    setServicePort(service: $service, port: $port) {\n      id\n      port\n    }\n  }\n"): (typeof documents)["\n  mutation SetServicePort($service: ServiceID!, $port: Int) {\n    setServicePort(service: $service, port: $port) {\n      id\n      port\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetServiceResources($service: ServiceID!, $resources: ResourcesInput!) {\n    setServiceResources(service: $service, resources: $resources) {\n      id\n      resources {\n        cpu\n        memory\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation SetServiceResources($service: ServiceID!, $resources: ResourcesInput!) {\n    setServiceResources(service: $service, resources: $resources) {\n      id\n      resources {\n        cpu\n        memory\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ServiceVariables($service: ServiceID!) {\n    serviceVariables(service: $service) {\n      key\n      value\n      fromShared\n      databaseRef {\n        database\n        key\n      }\n    }\n  }\n"): (typeof documents)["\n  query ServiceVariables($service: ServiceID!) {\n    serviceVariables(service: $service) {\n      key\n      value\n      fromShared\n      databaseRef {\n        database\n        key\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetServiceVariables($service: ServiceID!, $variables: [ServiceVariableInput!]!) {\n    setServiceVariables(service: $service, variables: $variables)\n  }\n"): (typeof documents)["\n  mutation SetServiceVariables($service: ServiceID!, $variables: [ServiceVariableInput!]!) {\n    setServiceVariables(service: $service, variables: $variables)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  subscription BuildLogs($id: BuildID!) {\n    buildLogs(id: $id)\n  }\n"): (typeof documents)["\n  subscription BuildLogs($id: BuildID!) {\n    buildLogs(id: $id)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query CanvasServiceBuilds($id: ServiceID!) {\n    service(id: $id) {\n      id\n      builds {\n        id\n        status\n        startedAt\n        finishedAt\n      }\n    }\n  }\n"): (typeof documents)["\n  query CanvasServiceBuilds($id: ServiceID!) {\n    service(id: $id) {\n      id\n      builds {\n        id\n        status\n        startedAt\n        finishedAt\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation Deploy($service: ServiceID!, $gitRef: String) {\n    deploy(service: $service, gitRef: $gitRef) {\n      id\n      status\n      startedAt\n      finishedAt\n    }\n  }\n"): (typeof documents)["\n  mutation Deploy($service: ServiceID!, $gitRef: String) {\n    deploy(service: $service, gitRef: $gitRef) {\n      id\n      status\n      startedAt\n      finishedAt\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query BuildStatus($id: BuildID!) {\n    build(id: $id) {\n      id\n      status\n      startedAt\n      finishedAt\n    }\n  }\n"): (typeof documents)["\n  query BuildStatus($id: BuildID!) {\n    build(id: $id) {\n      id\n      status\n      startedAt\n      finishedAt\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  subscription ServiceLogs($service: ServiceID!, $tailLines: Int) {\n    serviceLogs(service: $service, tailLines: $tailLines) {\n      line\n      pod\n    }\n  }\n"): (typeof documents)["\n  subscription ServiceLogs($service: ServiceID!, $tailLines: Int) {\n    serviceLogs(service: $service, tailLines: $tailLines) {\n      line\n      pod\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Workspace {\n    workspace {\n      id\n      name\n      personal\n      suspended\n      members {\n        id\n        email\n        name\n        role\n      }\n    }\n  }\n"): (typeof documents)["\n  query Workspace {\n    workspace {\n      id\n      name\n      personal\n      suspended\n      members {\n        id\n        email\n        name\n        role\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CompleteWorkspaceCheckout($sessionId: String!) {\n    completeWorkspaceCheckout(sessionId: $sessionId) {\n      id\n      name\n      personal\n    }\n  }\n"): (typeof documents)["\n  mutation CompleteWorkspaceCheckout($sessionId: String!) {\n    completeWorkspaceCheckout(sessionId: $sessionId) {\n      id\n      name\n      personal\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Environment($environment: EnvironmentID!) {\n    environment(environment: $environment) {\n      id\n      name\n      resourceTier\n      services {\n        id\n        name\n        status\n        replicas {\n          desired\n          ready\n        }\n        autoscaling {\n          minReplicas\n          maxReplicas\n          targetCpu\n        }\n        port\n        endpoints {\n          host\n          port\n          protocol\n          type\n          dns {\n            status\n            requiredRecords {\n              type\n              host\n              value\n            }\n          }\n          tls\n        }\n        sourceUrl\n        contextPath\n        resources {\n          cpu\n          memory\n        }\n        command\n        defaultCommand\n        activeDeployment {\n          id\n          image\n          imageDigest\n          commit\n          commitMessage\n          ref\n          status\n          createdAt\n        }\n        deployments {\n          id\n          image\n          imageDigest\n          commit\n          commitMessage\n          ref\n          status\n          createdAt\n        }\n        builds {\n          id\n          status\n          startedAt\n          finishedAt\n        }\n        lastDeployedAt\n        createdAt\n      }\n      databases {\n        id\n        name\n        version\n        instances\n        status\n        size\n        createdAt\n      }\n    }\n  }\n"): (typeof documents)["\n  query Environment($environment: EnvironmentID!) {\n    environment(environment: $environment) {\n      id\n      name\n      resourceTier\n      services {\n        id\n        name\n        status\n        replicas {\n          desired\n          ready\n        }\n        autoscaling {\n          minReplicas\n          maxReplicas\n          targetCpu\n        }\n        port\n        endpoints {\n          host\n          port\n          protocol\n          type\n          dns {\n            status\n            requiredRecords {\n              type\n              host\n              value\n            }\n          }\n          tls\n        }\n        sourceUrl\n        contextPath\n        resources {\n          cpu\n          memory\n        }\n        command\n        defaultCommand\n        activeDeployment {\n          id\n          image\n          imageDigest\n          commit\n          commitMessage\n          ref\n          status\n          createdAt\n        }\n        deployments {\n          id\n          image\n          imageDigest\n          commit\n          commitMessage\n          ref\n          status\n          createdAt\n        }\n        builds {\n          id\n          status\n          startedAt\n          finishedAt\n        }\n        lastDeployedAt\n        createdAt\n      }\n      databases {\n        id\n        name\n        version\n        instances\n        status\n        size\n        createdAt\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ProjectEnvironments($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n"): (typeof documents)["\n  query ProjectEnvironments($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CompletePlanCheckout($sessionId: String!) {\n    completePlanCheckout(sessionId: $sessionId) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      hasPaymentMethod\n    }\n  }\n"): (typeof documents)["\n  mutation CompletePlanCheckout($sessionId: String!) {\n    completePlanCheckout(sessionId: $sessionId) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      hasPaymentMethod\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ProjectPage($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n"): (typeof documents)["\n  query ProjectPage($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ProjectSettings($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n"): (typeof documents)["\n  query ProjectSettings($id: ProjectID!) {\n    project(id: $id) {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation DeleteProject($id: ProjectID!) {\n    deleteProject(id: $id)\n  }\n"): (typeof documents)["\n  mutation DeleteProject($id: ProjectID!) {\n    deleteProject(id: $id)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation DeleteEnvironment($environment: EnvironmentID!) {\n    deleteEnvironment(environment: $environment)\n  }\n"): (typeof documents)["\n  mutation DeleteEnvironment($environment: EnvironmentID!) {\n    deleteEnvironment(environment: $environment)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query EnvironmentResources($environment: EnvironmentID!) {\n    environmentResources(environment: $environment) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n"): (typeof documents)["\n  query EnvironmentResources($environment: EnvironmentID!) {\n    environmentResources(environment: $environment) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {\n    setEnvironmentResources(input: $input) {\n      id\n      resourceTier\n    }\n  }\n"): (typeof documents)["\n  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {\n    setEnvironmentResources(input: $input) {\n      id\n      resourceTier\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Projects {\n    projects {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n        services {\n          id\n          name\n          sourceUrl\n        }\n      }\n    }\n  }\n"): (typeof documents)["\n  query Projects {\n    projects {\n      id\n      name\n      environments {\n        id\n        name\n        resourceTier\n        services {\n          id\n          name\n          sourceUrl\n        }\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation UpdateWorkspace($input: UpdateWorkspaceInput!) {\n    updateWorkspace(input: $input) {\n      id\n      name\n    }\n  }\n"): (typeof documents)["\n  mutation UpdateWorkspace($input: UpdateWorkspaceInput!) {\n    updateWorkspace(input: $input) {\n      id\n      name\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation DeleteWorkspace {\n    deleteWorkspace\n  }\n"): (typeof documents)["\n  mutation DeleteWorkspace {\n    deleteWorkspace\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation InviteMember($input: InviteMemberInput!) {\n    inviteMember(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n"): (typeof documents)["\n  mutation InviteMember($input: InviteMemberInput!) {\n    inviteMember(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation RemoveMember($userId: ID!) {\n    removeMember(userId: $userId)\n  }\n"): (typeof documents)["\n  mutation RemoveMember($userId: ID!) {\n    removeMember(userId: $userId)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation UpdateMemberRole($input: UpdateMemberRoleInput!) {\n    updateMemberRole(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n"): (typeof documents)["\n  mutation UpdateMemberRole($input: UpdateMemberRoleInput!) {\n    updateMemberRole(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation ChangePlan($plan: Plan!) {\n    changePlan(plan: $plan) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n    }\n  }\n"): (typeof documents)["\n  mutation ChangePlan($plan: Plan!) {\n    changePlan(plan: $plan) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation BillingPortalUrl {\n    billingPortalUrl {\n      url\n    }\n  }\n"): (typeof documents)["\n  mutation BillingPortalUrl {\n    billingPortalUrl {\n      url\n    }\n  }\n"];

export function graphql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;