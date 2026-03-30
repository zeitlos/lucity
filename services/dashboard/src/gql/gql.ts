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
    "\n  query Subscription {\n    subscription {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n      hasPaymentMethod\n    }\n  }\n": typeof types.SubscriptionDocument,
    "\n  query UsageSummary {\n    usageSummary {\n      resourceCostCents\n      creditsCents\n      estimatedTotalCents\n    }\n  }\n": typeof types.UsageSummaryDocument,
    "\n  mutation ChangePlan($plan: Plan!) {\n    changePlan(plan: $plan) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n    }\n  }\n": typeof types.ChangePlanDocument,
    "\n  mutation BillingPortalUrl {\n    billingPortalUrl {\n      url\n    }\n  }\n": typeof types.BillingPortalUrlDocument,
    "\n  mutation CreatePlanCheckout($plan: Plan!) {\n    createPlanCheckout(plan: $plan) {\n      url\n    }\n  }\n": typeof types.CreatePlanCheckoutDocument,
    "\n  mutation CompletePlanCheckout($sessionId: String!) {\n    completePlanCheckout(sessionId: $sessionId) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      hasPaymentMethod\n    }\n  }\n": typeof types.CompletePlanCheckoutDocument,
    "\n  query EnvironmentResources($projectId: ID!, $environment: String!) {\n    environmentResources(projectId: $projectId, environment: $environment) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n": typeof types.EnvironmentResourcesDocument,
    "\n  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {\n    setEnvironmentResources(input: $input) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n": typeof types.SetEnvironmentResourcesDocument,
    "\n  mutation CreateDatabase($input: CreateDatabaseInput!) {\n    createDatabase(input: $input) {\n      name\n      version\n      instances\n      size\n    }\n  }\n": typeof types.CreateDatabaseDocument,
    "\n  mutation DeleteDatabase($projectId: ID!, $name: String!) {\n    deleteDatabase(projectId: $projectId, name: $name)\n  }\n": typeof types.DeleteDatabaseDocument,
    "\n  query DatabaseTables($projectId: ID!, $environment: String!, $database: String!) {\n    databaseTables(projectId: $projectId, environment: $environment, database: $database) {\n      name\n      schema\n      estimatedRows\n      columns {\n        name\n        type\n        nullable\n        primaryKey\n      }\n    }\n  }\n": typeof types.DatabaseTablesDocument,
    "\n  query DatabaseTableData(\n    $projectId: ID!\n    $environment: String!\n    $database: String!\n    $table: String!\n    $schema: String\n    $limit: Int\n    $offset: Int\n  ) {\n    databaseTableData(\n      projectId: $projectId\n      environment: $environment\n      database: $database\n      table: $table\n      schema: $schema\n      limit: $limit\n      offset: $offset\n    ) {\n      columns\n      rows\n      totalEstimatedRows\n    }\n  }\n": typeof types.DatabaseTableDataDocument,
    "\n  query DatabaseCredentials($projectId: ID!, $environment: String!, $database: String!) {\n    databaseCredentials(projectId: $projectId, environment: $environment, database: $database) {\n      host\n      port\n      dbname\n      user\n      password\n      uri\n    }\n  }\n": typeof types.DatabaseCredentialsDocument,
    "\n  mutation ExecuteQuery($input: DatabaseQueryInput!) {\n    executeQuery(input: $input) {\n      columns\n      rows\n      affectedRows\n    }\n  }\n": typeof types.ExecuteQueryDocument,
    "\n  query GitHubConnected {\n    githubConnected\n  }\n": typeof types.GitHubConnectedDocument,
    "\n  query GitHubSources {\n    githubSources {\n      id\n      accountLogin\n      accountAvatarUrl\n      accountType\n    }\n  }\n": typeof types.GitHubSourcesDocument,
    "\n  query GitHubRepositories($installationId: ID!) {\n    githubRepositories(installationId: $installationId) {\n      id\n      name\n      fullName\n      htmlUrl\n      defaultBranch\n      private\n    }\n  }\n": typeof types.GitHubRepositoriesDocument,
    "\n  query Projects {\n    projects {\n      id\n      name\n      createdAt\n      environments {\n        id\n        name\n        syncStatus\n        resourceTier\n        services {\n          name\n          sourceUrl\n        }\n      }\n      databases {\n        name\n        version\n      }\n    }\n  }\n": typeof types.ProjectsDocument,
    "\n  mutation CreateProject($input: CreateProjectInput!) {\n    createProject(input: $input) {\n      id\n      name\n    }\n  }\n": typeof types.CreateProjectDocument,
    "\n  mutation DeleteProject($id: ID!) {\n    deleteProject(id: $id)\n  }\n": typeof types.DeleteProjectDocument,
    "\n  mutation CreateEnvironment($input: CreateEnvironmentInput!) {\n    createEnvironment(input: $input) {\n      id\n      name\n      namespace\n      ephemeral\n      syncStatus\n      resourceTier\n    }\n  }\n": typeof types.CreateEnvironmentDocument,
    "\n  mutation DeleteEnvironment($projectId: ID!, $environment: String!) {\n    deleteEnvironment(projectId: $projectId, environment: $environment)\n  }\n": typeof types.DeleteEnvironmentDocument,
    "\n  mutation SetServiceScaling($input: SetServiceScalingInput!) {\n    setServiceScaling(input: $input) {\n      replicas\n      autoscaling {\n        enabled\n        minReplicas\n        maxReplicas\n        targetCPU\n      }\n    }\n  }\n": typeof types.SetServiceScalingDocument,
    "\n  query Project($id: ID!) {\n    project(id: $id) {\n      id\n      name\n      createdAt\n      environments {\n        id\n        name\n        namespace\n        ephemeral\n        syncStatus\n        resourceTier\n        services {\n          id\n          name\n          environment\n          image\n          port\n          framework\n          startCommand\n          sourceUrl\n          contextPath\n          customStartCommand\n          imageTag\n          ready\n          replicas\n          scaling {\n            replicas\n            autoscaling {\n              enabled\n              minReplicas\n              maxReplicas\n              targetCPU\n            }\n          }\n          resources {\n            cpuMillicores\n            memoryMB\n            cpuLimitMillicores\n            memoryLimitMB\n          }\n          domains {\n            hostname\n            type\n            dnsStatus\n            tlsStatus\n          }\n          deployments {\n            id\n            imageTag\n            active\n            timestamp\n            revision\n            message\n            sourceCommitMessage\n            sourceUrl\n          }\n        }\n        databases {\n          name\n          environment\n          ready\n          instances\n          version\n          size\n          volume {\n            name\n            size\n            requestedSize\n            usedBytes\n            capacityBytes\n          }\n        }\n      }\n      databases {\n        name\n        version\n        instances\n        size\n      }\n    }\n  }\n": typeof types.ProjectDocument,
    "\n  query SearchImages($query: String!) {\n    searchImages(query: $query) {\n      name\n      description\n      starCount\n      pullCount\n      official\n    }\n  }\n": typeof types.SearchImagesDocument,
    "\n  query DetectServices($installationId: ID!, $repository: String!) {\n    detectServices(installationId: $installationId, repository: $repository) {\n      name\n      language\n      framework\n      startCommand\n      suggestedPort\n    }\n  }\n": typeof types.DetectServicesDocument,
    "\n  mutation AddService($input: AddServiceInput!) {\n    addService(input: $input) {\n      id\n      name\n      environment\n      image\n      port\n      framework\n      startCommand\n      sourceUrl\n      contextPath\n      customStartCommand\n      imageTag\n      initialDeploy {\n        id\n        phase\n      }\n    }\n  }\n": typeof types.AddServiceDocument,
    "\n  mutation SetCustomStartCommand($projectId: ID!, $environment: String!, $service: String!, $command: String!) {\n    setCustomStartCommand(projectId: $projectId, environment: $environment, service: $service, command: $command)\n  }\n": typeof types.SetCustomStartCommandDocument,
    "\n  mutation RemoveService($projectId: ID!, $environment: String!, $service: String!) {\n    removeService(projectId: $projectId, environment: $environment, service: $service)\n  }\n": typeof types.RemoveServiceDocument,
    "\n  mutation Deploy($input: DeployInput!) {\n    deploy(input: $input) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n    }\n  }\n": typeof types.DeployDocument,
    "\n  query DeployStatus($id: ID!) {\n    deployStatus(id: $id) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n      rolloutHealth\n      rolloutMessage\n    }\n  }\n": typeof types.DeployStatusDocument,
    "\n  query ActiveDeployment($projectId: ID!, $service: String!, $environment: String!) {\n    activeDeployment(projectId: $projectId, service: $service, environment: $environment) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n      rolloutHealth\n      rolloutMessage\n    }\n  }\n": typeof types.ActiveDeploymentDocument,
    "\n  mutation Rollback($input: RollbackInput!) {\n    rollback(input: $input)\n  }\n": typeof types.RollbackDocument,
    "\n  mutation GenerateDomain($input: GenerateDomainInput!) {\n    generateDomain(input: $input) {\n      hostname\n      type\n      dnsStatus\n      tlsStatus\n    }\n  }\n": typeof types.GenerateDomainDocument,
    "\n  mutation AddCustomDomain($input: AddCustomDomainInput!) {\n    addCustomDomain(input: $input) {\n      hostname\n      type\n      dnsStatus\n      tlsStatus\n    }\n  }\n": typeof types.AddCustomDomainDocument,
    "\n  mutation RemoveDomain($input: RemoveDomainInput!) {\n    removeDomain(input: $input)\n  }\n": typeof types.RemoveDomainDocument,
    "\n  query CheckDnsStatus($hostname: String!) {\n    checkDnsStatus(hostname: $hostname) {\n      hostname\n      status\n      cnameTarget\n      expectedTarget\n      message\n      tlsStatus\n    }\n  }\n": typeof types.CheckDnsStatusDocument,
    "\n  query PlatformConfig {\n    platformConfig {\n      workloadDomain\n      domainTarget\n      ipAddress\n    }\n  }\n": typeof types.PlatformConfigDocument,
    "\n  subscription DeployLogs($id: ID!) {\n    deployLogs(id: $id)\n  }\n": typeof types.DeployLogsDocument,
    "\n  subscription ServiceLogs($projectId: ID!, $service: String!, $environment: String!, $tailLines: Int) {\n    serviceLogs(projectId: $projectId, service: $service, environment: $environment, tailLines: $tailLines) {\n      line\n      pod\n    }\n  }\n": typeof types.ServiceLogsDocument,
    "\n  query SharedVariables($projectId: ID!, $environment: String!) {\n    sharedVariables(projectId: $projectId, environment: $environment) {\n      key\n      value\n    }\n  }\n": typeof types.SharedVariablesDocument,
    "\n  mutation SetSharedVariables($projectId: ID!, $environment: String!, $variables: [VariableInput!]!) {\n    setSharedVariables(projectId: $projectId, environment: $environment, variables: $variables)\n  }\n": typeof types.SetSharedVariablesDocument,
    "\n  query ServiceVariables($projectId: ID!, $environment: String!, $service: String!) {\n    serviceVariables(projectId: $projectId, environment: $environment, service: $service) {\n      key\n      value\n      fromShared\n      databaseRef {\n        database\n        key\n      }\n      serviceRef {\n        service\n      }\n    }\n  }\n": typeof types.ServiceVariablesDocument,
    "\n  mutation SetServiceVariables($projectId: ID!, $environment: String!, $service: String!, $variables: [ServiceVariableInput!]!) {\n    setServiceVariables(projectId: $projectId, environment: $environment, service: $service, variables: $variables)\n  }\n": typeof types.SetServiceVariablesDocument,
    "\n  query Workspaces {\n    workspaces {\n      id\n      name\n      personal\n    }\n  }\n": typeof types.WorkspacesDocument,
    "\n  query Workspace {\n    workspace {\n      id\n      name\n      personal\n      suspended\n      members {\n        id\n        email\n        name\n        role\n      }\n    }\n  }\n": typeof types.WorkspaceDocument,
    "\n  mutation CreateWorkspace($input: CreateWorkspaceInput!) {\n    createWorkspace(input: $input) {\n      id\n      name\n      personal\n    }\n  }\n": typeof types.CreateWorkspaceDocument,
    "\n  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {\n    createWorkspaceCheckout(input: $input) {\n      url\n    }\n  }\n": typeof types.CreateWorkspaceCheckoutDocument,
    "\n  mutation CompleteWorkspaceCheckout($sessionId: String!) {\n    completeWorkspaceCheckout(sessionId: $sessionId) {\n      id\n      name\n      personal\n    }\n  }\n": typeof types.CompleteWorkspaceCheckoutDocument,
    "\n  mutation UpdateWorkspace($input: UpdateWorkspaceInput!) {\n    updateWorkspace(input: $input) {\n      id\n      name\n    }\n  }\n": typeof types.UpdateWorkspaceDocument,
    "\n  mutation DeleteWorkspace {\n    deleteWorkspace\n  }\n": typeof types.DeleteWorkspaceDocument,
    "\n  mutation InviteMember($input: InviteMemberInput!) {\n    inviteMember(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n": typeof types.InviteMemberDocument,
    "\n  mutation RemoveMember($userId: ID!) {\n    removeMember(userId: $userId)\n  }\n": typeof types.RemoveMemberDocument,
    "\n  mutation UpdateMemberRole($input: UpdateMemberRoleInput!) {\n    updateMemberRole(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n": typeof types.UpdateMemberRoleDocument,
};
const documents: Documents = {
    "\n  query Subscription {\n    subscription {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n      hasPaymentMethod\n    }\n  }\n": types.SubscriptionDocument,
    "\n  query UsageSummary {\n    usageSummary {\n      resourceCostCents\n      creditsCents\n      estimatedTotalCents\n    }\n  }\n": types.UsageSummaryDocument,
    "\n  mutation ChangePlan($plan: Plan!) {\n    changePlan(plan: $plan) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n    }\n  }\n": types.ChangePlanDocument,
    "\n  mutation BillingPortalUrl {\n    billingPortalUrl {\n      url\n    }\n  }\n": types.BillingPortalUrlDocument,
    "\n  mutation CreatePlanCheckout($plan: Plan!) {\n    createPlanCheckout(plan: $plan) {\n      url\n    }\n  }\n": types.CreatePlanCheckoutDocument,
    "\n  mutation CompletePlanCheckout($sessionId: String!) {\n    completePlanCheckout(sessionId: $sessionId) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      hasPaymentMethod\n    }\n  }\n": types.CompletePlanCheckoutDocument,
    "\n  query EnvironmentResources($projectId: ID!, $environment: String!) {\n    environmentResources(projectId: $projectId, environment: $environment) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n": types.EnvironmentResourcesDocument,
    "\n  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {\n    setEnvironmentResources(input: $input) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n": types.SetEnvironmentResourcesDocument,
    "\n  mutation CreateDatabase($input: CreateDatabaseInput!) {\n    createDatabase(input: $input) {\n      name\n      version\n      instances\n      size\n    }\n  }\n": types.CreateDatabaseDocument,
    "\n  mutation DeleteDatabase($projectId: ID!, $name: String!) {\n    deleteDatabase(projectId: $projectId, name: $name)\n  }\n": types.DeleteDatabaseDocument,
    "\n  query DatabaseTables($projectId: ID!, $environment: String!, $database: String!) {\n    databaseTables(projectId: $projectId, environment: $environment, database: $database) {\n      name\n      schema\n      estimatedRows\n      columns {\n        name\n        type\n        nullable\n        primaryKey\n      }\n    }\n  }\n": types.DatabaseTablesDocument,
    "\n  query DatabaseTableData(\n    $projectId: ID!\n    $environment: String!\n    $database: String!\n    $table: String!\n    $schema: String\n    $limit: Int\n    $offset: Int\n  ) {\n    databaseTableData(\n      projectId: $projectId\n      environment: $environment\n      database: $database\n      table: $table\n      schema: $schema\n      limit: $limit\n      offset: $offset\n    ) {\n      columns\n      rows\n      totalEstimatedRows\n    }\n  }\n": types.DatabaseTableDataDocument,
    "\n  query DatabaseCredentials($projectId: ID!, $environment: String!, $database: String!) {\n    databaseCredentials(projectId: $projectId, environment: $environment, database: $database) {\n      host\n      port\n      dbname\n      user\n      password\n      uri\n    }\n  }\n": types.DatabaseCredentialsDocument,
    "\n  mutation ExecuteQuery($input: DatabaseQueryInput!) {\n    executeQuery(input: $input) {\n      columns\n      rows\n      affectedRows\n    }\n  }\n": types.ExecuteQueryDocument,
    "\n  query GitHubConnected {\n    githubConnected\n  }\n": types.GitHubConnectedDocument,
    "\n  query GitHubSources {\n    githubSources {\n      id\n      accountLogin\n      accountAvatarUrl\n      accountType\n    }\n  }\n": types.GitHubSourcesDocument,
    "\n  query GitHubRepositories($installationId: ID!) {\n    githubRepositories(installationId: $installationId) {\n      id\n      name\n      fullName\n      htmlUrl\n      defaultBranch\n      private\n    }\n  }\n": types.GitHubRepositoriesDocument,
    "\n  query Projects {\n    projects {\n      id\n      name\n      createdAt\n      environments {\n        id\n        name\n        syncStatus\n        resourceTier\n        services {\n          name\n          sourceUrl\n        }\n      }\n      databases {\n        name\n        version\n      }\n    }\n  }\n": types.ProjectsDocument,
    "\n  mutation CreateProject($input: CreateProjectInput!) {\n    createProject(input: $input) {\n      id\n      name\n    }\n  }\n": types.CreateProjectDocument,
    "\n  mutation DeleteProject($id: ID!) {\n    deleteProject(id: $id)\n  }\n": types.DeleteProjectDocument,
    "\n  mutation CreateEnvironment($input: CreateEnvironmentInput!) {\n    createEnvironment(input: $input) {\n      id\n      name\n      namespace\n      ephemeral\n      syncStatus\n      resourceTier\n    }\n  }\n": types.CreateEnvironmentDocument,
    "\n  mutation DeleteEnvironment($projectId: ID!, $environment: String!) {\n    deleteEnvironment(projectId: $projectId, environment: $environment)\n  }\n": types.DeleteEnvironmentDocument,
    "\n  mutation SetServiceScaling($input: SetServiceScalingInput!) {\n    setServiceScaling(input: $input) {\n      replicas\n      autoscaling {\n        enabled\n        minReplicas\n        maxReplicas\n        targetCPU\n      }\n    }\n  }\n": types.SetServiceScalingDocument,
    "\n  query Project($id: ID!) {\n    project(id: $id) {\n      id\n      name\n      createdAt\n      environments {\n        id\n        name\n        namespace\n        ephemeral\n        syncStatus\n        resourceTier\n        services {\n          id\n          name\n          environment\n          image\n          port\n          framework\n          startCommand\n          sourceUrl\n          contextPath\n          customStartCommand\n          imageTag\n          ready\n          replicas\n          scaling {\n            replicas\n            autoscaling {\n              enabled\n              minReplicas\n              maxReplicas\n              targetCPU\n            }\n          }\n          resources {\n            cpuMillicores\n            memoryMB\n            cpuLimitMillicores\n            memoryLimitMB\n          }\n          domains {\n            hostname\n            type\n            dnsStatus\n            tlsStatus\n          }\n          deployments {\n            id\n            imageTag\n            active\n            timestamp\n            revision\n            message\n            sourceCommitMessage\n            sourceUrl\n          }\n        }\n        databases {\n          name\n          environment\n          ready\n          instances\n          version\n          size\n          volume {\n            name\n            size\n            requestedSize\n            usedBytes\n            capacityBytes\n          }\n        }\n      }\n      databases {\n        name\n        version\n        instances\n        size\n      }\n    }\n  }\n": types.ProjectDocument,
    "\n  query SearchImages($query: String!) {\n    searchImages(query: $query) {\n      name\n      description\n      starCount\n      pullCount\n      official\n    }\n  }\n": types.SearchImagesDocument,
    "\n  query DetectServices($installationId: ID!, $repository: String!) {\n    detectServices(installationId: $installationId, repository: $repository) {\n      name\n      language\n      framework\n      startCommand\n      suggestedPort\n    }\n  }\n": types.DetectServicesDocument,
    "\n  mutation AddService($input: AddServiceInput!) {\n    addService(input: $input) {\n      id\n      name\n      environment\n      image\n      port\n      framework\n      startCommand\n      sourceUrl\n      contextPath\n      customStartCommand\n      imageTag\n      initialDeploy {\n        id\n        phase\n      }\n    }\n  }\n": types.AddServiceDocument,
    "\n  mutation SetCustomStartCommand($projectId: ID!, $environment: String!, $service: String!, $command: String!) {\n    setCustomStartCommand(projectId: $projectId, environment: $environment, service: $service, command: $command)\n  }\n": types.SetCustomStartCommandDocument,
    "\n  mutation RemoveService($projectId: ID!, $environment: String!, $service: String!) {\n    removeService(projectId: $projectId, environment: $environment, service: $service)\n  }\n": types.RemoveServiceDocument,
    "\n  mutation Deploy($input: DeployInput!) {\n    deploy(input: $input) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n    }\n  }\n": types.DeployDocument,
    "\n  query DeployStatus($id: ID!) {\n    deployStatus(id: $id) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n      rolloutHealth\n      rolloutMessage\n    }\n  }\n": types.DeployStatusDocument,
    "\n  query ActiveDeployment($projectId: ID!, $service: String!, $environment: String!) {\n    activeDeployment(projectId: $projectId, service: $service, environment: $environment) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n      rolloutHealth\n      rolloutMessage\n    }\n  }\n": types.ActiveDeploymentDocument,
    "\n  mutation Rollback($input: RollbackInput!) {\n    rollback(input: $input)\n  }\n": types.RollbackDocument,
    "\n  mutation GenerateDomain($input: GenerateDomainInput!) {\n    generateDomain(input: $input) {\n      hostname\n      type\n      dnsStatus\n      tlsStatus\n    }\n  }\n": types.GenerateDomainDocument,
    "\n  mutation AddCustomDomain($input: AddCustomDomainInput!) {\n    addCustomDomain(input: $input) {\n      hostname\n      type\n      dnsStatus\n      tlsStatus\n    }\n  }\n": types.AddCustomDomainDocument,
    "\n  mutation RemoveDomain($input: RemoveDomainInput!) {\n    removeDomain(input: $input)\n  }\n": types.RemoveDomainDocument,
    "\n  query CheckDnsStatus($hostname: String!) {\n    checkDnsStatus(hostname: $hostname) {\n      hostname\n      status\n      cnameTarget\n      expectedTarget\n      message\n      tlsStatus\n    }\n  }\n": types.CheckDnsStatusDocument,
    "\n  query PlatformConfig {\n    platformConfig {\n      workloadDomain\n      domainTarget\n      ipAddress\n    }\n  }\n": types.PlatformConfigDocument,
    "\n  subscription DeployLogs($id: ID!) {\n    deployLogs(id: $id)\n  }\n": types.DeployLogsDocument,
    "\n  subscription ServiceLogs($projectId: ID!, $service: String!, $environment: String!, $tailLines: Int) {\n    serviceLogs(projectId: $projectId, service: $service, environment: $environment, tailLines: $tailLines) {\n      line\n      pod\n    }\n  }\n": types.ServiceLogsDocument,
    "\n  query SharedVariables($projectId: ID!, $environment: String!) {\n    sharedVariables(projectId: $projectId, environment: $environment) {\n      key\n      value\n    }\n  }\n": types.SharedVariablesDocument,
    "\n  mutation SetSharedVariables($projectId: ID!, $environment: String!, $variables: [VariableInput!]!) {\n    setSharedVariables(projectId: $projectId, environment: $environment, variables: $variables)\n  }\n": types.SetSharedVariablesDocument,
    "\n  query ServiceVariables($projectId: ID!, $environment: String!, $service: String!) {\n    serviceVariables(projectId: $projectId, environment: $environment, service: $service) {\n      key\n      value\n      fromShared\n      databaseRef {\n        database\n        key\n      }\n      serviceRef {\n        service\n      }\n    }\n  }\n": types.ServiceVariablesDocument,
    "\n  mutation SetServiceVariables($projectId: ID!, $environment: String!, $service: String!, $variables: [ServiceVariableInput!]!) {\n    setServiceVariables(projectId: $projectId, environment: $environment, service: $service, variables: $variables)\n  }\n": types.SetServiceVariablesDocument,
    "\n  query Workspaces {\n    workspaces {\n      id\n      name\n      personal\n    }\n  }\n": types.WorkspacesDocument,
    "\n  query Workspace {\n    workspace {\n      id\n      name\n      personal\n      suspended\n      members {\n        id\n        email\n        name\n        role\n      }\n    }\n  }\n": types.WorkspaceDocument,
    "\n  mutation CreateWorkspace($input: CreateWorkspaceInput!) {\n    createWorkspace(input: $input) {\n      id\n      name\n      personal\n    }\n  }\n": types.CreateWorkspaceDocument,
    "\n  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {\n    createWorkspaceCheckout(input: $input) {\n      url\n    }\n  }\n": types.CreateWorkspaceCheckoutDocument,
    "\n  mutation CompleteWorkspaceCheckout($sessionId: String!) {\n    completeWorkspaceCheckout(sessionId: $sessionId) {\n      id\n      name\n      personal\n    }\n  }\n": types.CompleteWorkspaceCheckoutDocument,
    "\n  mutation UpdateWorkspace($input: UpdateWorkspaceInput!) {\n    updateWorkspace(input: $input) {\n      id\n      name\n    }\n  }\n": types.UpdateWorkspaceDocument,
    "\n  mutation DeleteWorkspace {\n    deleteWorkspace\n  }\n": types.DeleteWorkspaceDocument,
    "\n  mutation InviteMember($input: InviteMemberInput!) {\n    inviteMember(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n": types.InviteMemberDocument,
    "\n  mutation RemoveMember($userId: ID!) {\n    removeMember(userId: $userId)\n  }\n": types.RemoveMemberDocument,
    "\n  mutation UpdateMemberRole($input: UpdateMemberRoleInput!) {\n    updateMemberRole(input: $input) {\n      id\n      email\n      name\n      role\n    }\n  }\n": types.UpdateMemberRoleDocument,
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
export function graphql(source: "\n  query Subscription {\n    subscription {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n      hasPaymentMethod\n    }\n  }\n"): (typeof documents)["\n  query Subscription {\n    subscription {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n      hasPaymentMethod\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query UsageSummary {\n    usageSummary {\n      resourceCostCents\n      creditsCents\n      estimatedTotalCents\n    }\n  }\n"): (typeof documents)["\n  query UsageSummary {\n    usageSummary {\n      resourceCostCents\n      creditsCents\n      estimatedTotalCents\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation ChangePlan($plan: Plan!) {\n    changePlan(plan: $plan) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n    }\n  }\n"): (typeof documents)["\n  mutation ChangePlan($plan: Plan!) {\n    changePlan(plan: $plan) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      creditExpiry\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation BillingPortalUrl {\n    billingPortalUrl {\n      url\n    }\n  }\n"): (typeof documents)["\n  mutation BillingPortalUrl {\n    billingPortalUrl {\n      url\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreatePlanCheckout($plan: Plan!) {\n    createPlanCheckout(plan: $plan) {\n      url\n    }\n  }\n"): (typeof documents)["\n  mutation CreatePlanCheckout($plan: Plan!) {\n    createPlanCheckout(plan: $plan) {\n      url\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CompletePlanCheckout($sessionId: String!) {\n    completePlanCheckout(sessionId: $sessionId) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      hasPaymentMethod\n    }\n  }\n"): (typeof documents)["\n  mutation CompletePlanCheckout($sessionId: String!) {\n    completePlanCheckout(sessionId: $sessionId) {\n      plan\n      status\n      currentPeriodEnd\n      creditAmountCents\n      hasPaymentMethod\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query EnvironmentResources($projectId: ID!, $environment: String!) {\n    environmentResources(projectId: $projectId, environment: $environment) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n"): (typeof documents)["\n  query EnvironmentResources($projectId: ID!, $environment: String!) {\n    environmentResources(projectId: $projectId, environment: $environment) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {\n    setEnvironmentResources(input: $input) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {\n    setEnvironmentResources(input: $input) {\n      tier\n      allocation {\n        cpuMillicores\n        memoryMB\n        diskMB\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateDatabase($input: CreateDatabaseInput!) {\n    createDatabase(input: $input) {\n      name\n      version\n      instances\n      size\n    }\n  }\n"): (typeof documents)["\n  mutation CreateDatabase($input: CreateDatabaseInput!) {\n    createDatabase(input: $input) {\n      name\n      version\n      instances\n      size\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation DeleteDatabase($projectId: ID!, $name: String!) {\n    deleteDatabase(projectId: $projectId, name: $name)\n  }\n"): (typeof documents)["\n  mutation DeleteDatabase($projectId: ID!, $name: String!) {\n    deleteDatabase(projectId: $projectId, name: $name)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DatabaseTables($projectId: ID!, $environment: String!, $database: String!) {\n    databaseTables(projectId: $projectId, environment: $environment, database: $database) {\n      name\n      schema\n      estimatedRows\n      columns {\n        name\n        type\n        nullable\n        primaryKey\n      }\n    }\n  }\n"): (typeof documents)["\n  query DatabaseTables($projectId: ID!, $environment: String!, $database: String!) {\n    databaseTables(projectId: $projectId, environment: $environment, database: $database) {\n      name\n      schema\n      estimatedRows\n      columns {\n        name\n        type\n        nullable\n        primaryKey\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DatabaseTableData(\n    $projectId: ID!\n    $environment: String!\n    $database: String!\n    $table: String!\n    $schema: String\n    $limit: Int\n    $offset: Int\n  ) {\n    databaseTableData(\n      projectId: $projectId\n      environment: $environment\n      database: $database\n      table: $table\n      schema: $schema\n      limit: $limit\n      offset: $offset\n    ) {\n      columns\n      rows\n      totalEstimatedRows\n    }\n  }\n"): (typeof documents)["\n  query DatabaseTableData(\n    $projectId: ID!\n    $environment: String!\n    $database: String!\n    $table: String!\n    $schema: String\n    $limit: Int\n    $offset: Int\n  ) {\n    databaseTableData(\n      projectId: $projectId\n      environment: $environment\n      database: $database\n      table: $table\n      schema: $schema\n      limit: $limit\n      offset: $offset\n    ) {\n      columns\n      rows\n      totalEstimatedRows\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DatabaseCredentials($projectId: ID!, $environment: String!, $database: String!) {\n    databaseCredentials(projectId: $projectId, environment: $environment, database: $database) {\n      host\n      port\n      dbname\n      user\n      password\n      uri\n    }\n  }\n"): (typeof documents)["\n  query DatabaseCredentials($projectId: ID!, $environment: String!, $database: String!) {\n    databaseCredentials(projectId: $projectId, environment: $environment, database: $database) {\n      host\n      port\n      dbname\n      user\n      password\n      uri\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation ExecuteQuery($input: DatabaseQueryInput!) {\n    executeQuery(input: $input) {\n      columns\n      rows\n      affectedRows\n    }\n  }\n"): (typeof documents)["\n  mutation ExecuteQuery($input: DatabaseQueryInput!) {\n    executeQuery(input: $input) {\n      columns\n      rows\n      affectedRows\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GitHubConnected {\n    githubConnected\n  }\n"): (typeof documents)["\n  query GitHubConnected {\n    githubConnected\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GitHubSources {\n    githubSources {\n      id\n      accountLogin\n      accountAvatarUrl\n      accountType\n    }\n  }\n"): (typeof documents)["\n  query GitHubSources {\n    githubSources {\n      id\n      accountLogin\n      accountAvatarUrl\n      accountType\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GitHubRepositories($installationId: ID!) {\n    githubRepositories(installationId: $installationId) {\n      id\n      name\n      fullName\n      htmlUrl\n      defaultBranch\n      private\n    }\n  }\n"): (typeof documents)["\n  query GitHubRepositories($installationId: ID!) {\n    githubRepositories(installationId: $installationId) {\n      id\n      name\n      fullName\n      htmlUrl\n      defaultBranch\n      private\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Projects {\n    projects {\n      id\n      name\n      createdAt\n      environments {\n        id\n        name\n        syncStatus\n        resourceTier\n        services {\n          name\n          sourceUrl\n        }\n      }\n      databases {\n        name\n        version\n      }\n    }\n  }\n"): (typeof documents)["\n  query Projects {\n    projects {\n      id\n      name\n      createdAt\n      environments {\n        id\n        name\n        syncStatus\n        resourceTier\n        services {\n          name\n          sourceUrl\n        }\n      }\n      databases {\n        name\n        version\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateProject($input: CreateProjectInput!) {\n    createProject(input: $input) {\n      id\n      name\n    }\n  }\n"): (typeof documents)["\n  mutation CreateProject($input: CreateProjectInput!) {\n    createProject(input: $input) {\n      id\n      name\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation DeleteProject($id: ID!) {\n    deleteProject(id: $id)\n  }\n"): (typeof documents)["\n  mutation DeleteProject($id: ID!) {\n    deleteProject(id: $id)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateEnvironment($input: CreateEnvironmentInput!) {\n    createEnvironment(input: $input) {\n      id\n      name\n      namespace\n      ephemeral\n      syncStatus\n      resourceTier\n    }\n  }\n"): (typeof documents)["\n  mutation CreateEnvironment($input: CreateEnvironmentInput!) {\n    createEnvironment(input: $input) {\n      id\n      name\n      namespace\n      ephemeral\n      syncStatus\n      resourceTier\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation DeleteEnvironment($projectId: ID!, $environment: String!) {\n    deleteEnvironment(projectId: $projectId, environment: $environment)\n  }\n"): (typeof documents)["\n  mutation DeleteEnvironment($projectId: ID!, $environment: String!) {\n    deleteEnvironment(projectId: $projectId, environment: $environment)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetServiceScaling($input: SetServiceScalingInput!) {\n    setServiceScaling(input: $input) {\n      replicas\n      autoscaling {\n        enabled\n        minReplicas\n        maxReplicas\n        targetCPU\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation SetServiceScaling($input: SetServiceScalingInput!) {\n    setServiceScaling(input: $input) {\n      replicas\n      autoscaling {\n        enabled\n        minReplicas\n        maxReplicas\n        targetCPU\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Project($id: ID!) {\n    project(id: $id) {\n      id\n      name\n      createdAt\n      environments {\n        id\n        name\n        namespace\n        ephemeral\n        syncStatus\n        resourceTier\n        services {\n          id\n          name\n          environment\n          image\n          port\n          framework\n          startCommand\n          sourceUrl\n          contextPath\n          customStartCommand\n          imageTag\n          ready\n          replicas\n          scaling {\n            replicas\n            autoscaling {\n              enabled\n              minReplicas\n              maxReplicas\n              targetCPU\n            }\n          }\n          resources {\n            cpuMillicores\n            memoryMB\n            cpuLimitMillicores\n            memoryLimitMB\n          }\n          domains {\n            hostname\n            type\n            dnsStatus\n            tlsStatus\n          }\n          deployments {\n            id\n            imageTag\n            active\n            timestamp\n            revision\n            message\n            sourceCommitMessage\n            sourceUrl\n          }\n        }\n        databases {\n          name\n          environment\n          ready\n          instances\n          version\n          size\n          volume {\n            name\n            size\n            requestedSize\n            usedBytes\n            capacityBytes\n          }\n        }\n      }\n      databases {\n        name\n        version\n        instances\n        size\n      }\n    }\n  }\n"): (typeof documents)["\n  query Project($id: ID!) {\n    project(id: $id) {\n      id\n      name\n      createdAt\n      environments {\n        id\n        name\n        namespace\n        ephemeral\n        syncStatus\n        resourceTier\n        services {\n          id\n          name\n          environment\n          image\n          port\n          framework\n          startCommand\n          sourceUrl\n          contextPath\n          customStartCommand\n          imageTag\n          ready\n          replicas\n          scaling {\n            replicas\n            autoscaling {\n              enabled\n              minReplicas\n              maxReplicas\n              targetCPU\n            }\n          }\n          resources {\n            cpuMillicores\n            memoryMB\n            cpuLimitMillicores\n            memoryLimitMB\n          }\n          domains {\n            hostname\n            type\n            dnsStatus\n            tlsStatus\n          }\n          deployments {\n            id\n            imageTag\n            active\n            timestamp\n            revision\n            message\n            sourceCommitMessage\n            sourceUrl\n          }\n        }\n        databases {\n          name\n          environment\n          ready\n          instances\n          version\n          size\n          volume {\n            name\n            size\n            requestedSize\n            usedBytes\n            capacityBytes\n          }\n        }\n      }\n      databases {\n        name\n        version\n        instances\n        size\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query SearchImages($query: String!) {\n    searchImages(query: $query) {\n      name\n      description\n      starCount\n      pullCount\n      official\n    }\n  }\n"): (typeof documents)["\n  query SearchImages($query: String!) {\n    searchImages(query: $query) {\n      name\n      description\n      starCount\n      pullCount\n      official\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DetectServices($installationId: ID!, $repository: String!) {\n    detectServices(installationId: $installationId, repository: $repository) {\n      name\n      language\n      framework\n      startCommand\n      suggestedPort\n    }\n  }\n"): (typeof documents)["\n  query DetectServices($installationId: ID!, $repository: String!) {\n    detectServices(installationId: $installationId, repository: $repository) {\n      name\n      language\n      framework\n      startCommand\n      suggestedPort\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation AddService($input: AddServiceInput!) {\n    addService(input: $input) {\n      id\n      name\n      environment\n      image\n      port\n      framework\n      startCommand\n      sourceUrl\n      contextPath\n      customStartCommand\n      imageTag\n      initialDeploy {\n        id\n        phase\n      }\n    }\n  }\n"): (typeof documents)["\n  mutation AddService($input: AddServiceInput!) {\n    addService(input: $input) {\n      id\n      name\n      environment\n      image\n      port\n      framework\n      startCommand\n      sourceUrl\n      contextPath\n      customStartCommand\n      imageTag\n      initialDeploy {\n        id\n        phase\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetCustomStartCommand($projectId: ID!, $environment: String!, $service: String!, $command: String!) {\n    setCustomStartCommand(projectId: $projectId, environment: $environment, service: $service, command: $command)\n  }\n"): (typeof documents)["\n  mutation SetCustomStartCommand($projectId: ID!, $environment: String!, $service: String!, $command: String!) {\n    setCustomStartCommand(projectId: $projectId, environment: $environment, service: $service, command: $command)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation RemoveService($projectId: ID!, $environment: String!, $service: String!) {\n    removeService(projectId: $projectId, environment: $environment, service: $service)\n  }\n"): (typeof documents)["\n  mutation RemoveService($projectId: ID!, $environment: String!, $service: String!) {\n    removeService(projectId: $projectId, environment: $environment, service: $service)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation Deploy($input: DeployInput!) {\n    deploy(input: $input) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n    }\n  }\n"): (typeof documents)["\n  mutation Deploy($input: DeployInput!) {\n    deploy(input: $input) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query DeployStatus($id: ID!) {\n    deployStatus(id: $id) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n      rolloutHealth\n      rolloutMessage\n    }\n  }\n"): (typeof documents)["\n  query DeployStatus($id: ID!) {\n    deployStatus(id: $id) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n      rolloutHealth\n      rolloutMessage\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ActiveDeployment($projectId: ID!, $service: String!, $environment: String!) {\n    activeDeployment(projectId: $projectId, service: $service, environment: $environment) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n      rolloutHealth\n      rolloutMessage\n    }\n  }\n"): (typeof documents)["\n  query ActiveDeployment($projectId: ID!, $service: String!, $environment: String!) {\n    activeDeployment(projectId: $projectId, service: $service, environment: $environment) {\n      id\n      phase\n      buildId\n      imageRef\n      digest\n      error\n      startedAt\n      rolloutHealth\n      rolloutMessage\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation Rollback($input: RollbackInput!) {\n    rollback(input: $input)\n  }\n"): (typeof documents)["\n  mutation Rollback($input: RollbackInput!) {\n    rollback(input: $input)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation GenerateDomain($input: GenerateDomainInput!) {\n    generateDomain(input: $input) {\n      hostname\n      type\n      dnsStatus\n      tlsStatus\n    }\n  }\n"): (typeof documents)["\n  mutation GenerateDomain($input: GenerateDomainInput!) {\n    generateDomain(input: $input) {\n      hostname\n      type\n      dnsStatus\n      tlsStatus\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation AddCustomDomain($input: AddCustomDomainInput!) {\n    addCustomDomain(input: $input) {\n      hostname\n      type\n      dnsStatus\n      tlsStatus\n    }\n  }\n"): (typeof documents)["\n  mutation AddCustomDomain($input: AddCustomDomainInput!) {\n    addCustomDomain(input: $input) {\n      hostname\n      type\n      dnsStatus\n      tlsStatus\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation RemoveDomain($input: RemoveDomainInput!) {\n    removeDomain(input: $input)\n  }\n"): (typeof documents)["\n  mutation RemoveDomain($input: RemoveDomainInput!) {\n    removeDomain(input: $input)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query CheckDnsStatus($hostname: String!) {\n    checkDnsStatus(hostname: $hostname) {\n      hostname\n      status\n      cnameTarget\n      expectedTarget\n      message\n      tlsStatus\n    }\n  }\n"): (typeof documents)["\n  query CheckDnsStatus($hostname: String!) {\n    checkDnsStatus(hostname: $hostname) {\n      hostname\n      status\n      cnameTarget\n      expectedTarget\n      message\n      tlsStatus\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query PlatformConfig {\n    platformConfig {\n      workloadDomain\n      domainTarget\n      ipAddress\n    }\n  }\n"): (typeof documents)["\n  query PlatformConfig {\n    platformConfig {\n      workloadDomain\n      domainTarget\n      ipAddress\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  subscription DeployLogs($id: ID!) {\n    deployLogs(id: $id)\n  }\n"): (typeof documents)["\n  subscription DeployLogs($id: ID!) {\n    deployLogs(id: $id)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  subscription ServiceLogs($projectId: ID!, $service: String!, $environment: String!, $tailLines: Int) {\n    serviceLogs(projectId: $projectId, service: $service, environment: $environment, tailLines: $tailLines) {\n      line\n      pod\n    }\n  }\n"): (typeof documents)["\n  subscription ServiceLogs($projectId: ID!, $service: String!, $environment: String!, $tailLines: Int) {\n    serviceLogs(projectId: $projectId, service: $service, environment: $environment, tailLines: $tailLines) {\n      line\n      pod\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query SharedVariables($projectId: ID!, $environment: String!) {\n    sharedVariables(projectId: $projectId, environment: $environment) {\n      key\n      value\n    }\n  }\n"): (typeof documents)["\n  query SharedVariables($projectId: ID!, $environment: String!) {\n    sharedVariables(projectId: $projectId, environment: $environment) {\n      key\n      value\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetSharedVariables($projectId: ID!, $environment: String!, $variables: [VariableInput!]!) {\n    setSharedVariables(projectId: $projectId, environment: $environment, variables: $variables)\n  }\n"): (typeof documents)["\n  mutation SetSharedVariables($projectId: ID!, $environment: String!, $variables: [VariableInput!]!) {\n    setSharedVariables(projectId: $projectId, environment: $environment, variables: $variables)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ServiceVariables($projectId: ID!, $environment: String!, $service: String!) {\n    serviceVariables(projectId: $projectId, environment: $environment, service: $service) {\n      key\n      value\n      fromShared\n      databaseRef {\n        database\n        key\n      }\n      serviceRef {\n        service\n      }\n    }\n  }\n"): (typeof documents)["\n  query ServiceVariables($projectId: ID!, $environment: String!, $service: String!) {\n    serviceVariables(projectId: $projectId, environment: $environment, service: $service) {\n      key\n      value\n      fromShared\n      databaseRef {\n        database\n        key\n      }\n      serviceRef {\n        service\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation SetServiceVariables($projectId: ID!, $environment: String!, $service: String!, $variables: [ServiceVariableInput!]!) {\n    setServiceVariables(projectId: $projectId, environment: $environment, service: $service, variables: $variables)\n  }\n"): (typeof documents)["\n  mutation SetServiceVariables($projectId: ID!, $environment: String!, $service: String!, $variables: [ServiceVariableInput!]!) {\n    setServiceVariables(projectId: $projectId, environment: $environment, service: $service, variables: $variables)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Workspaces {\n    workspaces {\n      id\n      name\n      personal\n    }\n  }\n"): (typeof documents)["\n  query Workspaces {\n    workspaces {\n      id\n      name\n      personal\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Workspace {\n    workspace {\n      id\n      name\n      personal\n      suspended\n      members {\n        id\n        email\n        name\n        role\n      }\n    }\n  }\n"): (typeof documents)["\n  query Workspace {\n    workspace {\n      id\n      name\n      personal\n      suspended\n      members {\n        id\n        email\n        name\n        role\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateWorkspace($input: CreateWorkspaceInput!) {\n    createWorkspace(input: $input) {\n      id\n      name\n      personal\n    }\n  }\n"): (typeof documents)["\n  mutation CreateWorkspace($input: CreateWorkspaceInput!) {\n    createWorkspace(input: $input) {\n      id\n      name\n      personal\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {\n    createWorkspaceCheckout(input: $input) {\n      url\n    }\n  }\n"): (typeof documents)["\n  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {\n    createWorkspaceCheckout(input: $input) {\n      url\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CompleteWorkspaceCheckout($sessionId: String!) {\n    completeWorkspaceCheckout(sessionId: $sessionId) {\n      id\n      name\n      personal\n    }\n  }\n"): (typeof documents)["\n  mutation CompleteWorkspaceCheckout($sessionId: String!) {\n    completeWorkspaceCheckout(sessionId: $sessionId) {\n      id\n      name\n      personal\n    }\n  }\n"];
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

export function graphql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;