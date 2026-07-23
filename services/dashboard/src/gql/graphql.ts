/* eslint-disable */
/** Internal type. DO NOT USE DIRECTLY. */
type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
/** Internal type. DO NOT USE DIRECTLY. */
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
export type AddServiceInput = {
  contextPath?: string | null | undefined;
  image?: string | null | undefined;
  name?: string | null | undefined;
  repository?: string | null | undefined;
  resources?: ResourcesInput | null | undefined;
  variables?: Array<VariableInput> | null | undefined;
};

export type AutoscalingInput = {
  enabled: boolean;
  maxReplicas: number;
  minReplicas: number;
  targetCPU: number;
};

export enum BucketStatus {
  Failed = 'FAILED',
  Pending = 'PENDING',
  Ready = 'READY'
}

export enum BuildStatus {
  Cancelled = 'CANCELLED',
  Failed = 'FAILED',
  Queued = 'QUEUED',
  Running = 'RUNNING',
  Succeeded = 'SUCCEEDED'
}

export type CreateApiTokenInput = {
  name: string;
  role: WorkspaceRole;
};

export type CreateBucketInput = {
  environment: string;
  name: string;
};

export type CreateDatabaseInput = {
  environment: string;
  name: string;
  resources?: ResourcesInput | null | undefined;
  size?: string | null | undefined;
};

export type CreateEnvironmentInput = {
  fromEnvironment?: string | null | undefined;
  name: string;
  project: string;
  tier?: ResourceTier | null | undefined;
};

export type CreateKeyValueStoreInput = {
  environment: string;
  name: string;
  size?: string | null | undefined;
  version?: string | null | undefined;
};

export type CreateProjectInput = {
  /** Optional URL-safe slug. Auto-derived from name if omitted. */
  id?: string | null | undefined;
  /** Human-readable project name (e.g. "My API"). */
  name: string;
};

export type CreateWorkspaceCheckoutInput = {
  id: string;
  name: string;
  plan: Plan;
};

export enum DatabaseStatus {
  Degraded = 'DEGRADED',
  Failed = 'FAILED',
  Healthy = 'HEALTHY',
  Pending = 'PENDING',
  Stopped = 'STOPPED',
  Updating = 'UPDATING'
}

export enum DeployStatus {
  Failed = 'FAILED',
  Queued = 'QUEUED',
  Running = 'RUNNING',
  Skipped = 'SKIPPED',
  Succeeded = 'SUCCEEDED'
}

export enum DeploymentStatus {
  Active = 'ACTIVE',
  Deploying = 'DEPLOYING',
  Failed = 'FAILED',
  Superseded = 'SUPERSEDED'
}

export enum DnsRecordType {
  A = 'A',
  Cname = 'CNAME',
  Txt = 'TXT'
}

export enum DnsStatus {
  Error = 'ERROR',
  Misconfigured = 'MISCONFIGURED',
  Pending = 'PENDING',
  Valid = 'VALID'
}

export enum EndpointType {
  Custom = 'CUSTOM',
  Internal = 'INTERNAL',
  Platform = 'PLATFORM'
}

export enum GitHubAccountType {
  Organization = 'ORGANIZATION',
  User = 'USER'
}

export type HealthCheckInput = {
  failureThreshold?: number | null | undefined;
  initialDelaySeconds?: number | null | undefined;
  path: string;
  periodSeconds?: number | null | undefined;
  port?: number | null | undefined;
  startupFailureThreshold?: number | null | undefined;
  timeoutSeconds?: number | null | undefined;
};

export type InviteMemberInput = {
  email: string;
  role: WorkspaceRole;
};

export enum MetricGrouping {
  PerReplica = 'PER_REPLICA',
  Total = 'TOTAL'
}

export enum MetricWindow {
  Last_1H = 'LAST_1H',
  Last_6H = 'LAST_6H',
  Last_7D = 'LAST_7D',
  Last_24H = 'LAST_24H',
  Last_30D = 'LAST_30D'
}

export type MetricsRange = {
  window: MetricWindow;
};

export enum Plan {
  Hobby = 'HOBBY',
  Pro = 'PRO'
}

export enum Protocol {
  Http = 'HTTP',
  Https = 'HTTPS',
  Tcp = 'TCP'
}

export enum ReleaseStatus {
  Building = 'BUILDING',
  Cancelled = 'CANCELLED',
  Deploying = 'DEPLOYING',
  Failed = 'FAILED',
  Live = 'LIVE',
  Queued = 'QUEUED',
  Superseded = 'SUPERSEDED'
}

export enum ReleaseTriggerKind {
  Manual = 'MANUAL',
  Promotion = 'PROMOTION',
  Push = 'PUSH',
  Rollback = 'ROLLBACK'
}

export enum ResourceMetric {
  CpuUsage = 'CPU_USAGE',
  MemoryUsage = 'MEMORY_USAGE',
  StorageUsed = 'STORAGE_USED'
}

export enum ResourceTier {
  Eco = 'ECO',
  Production = 'PRODUCTION'
}

export type ResourcesInput = {
  cpu: string;
  memory: string;
};

export enum RolloutReason {
  ConfigError = 'CONFIG_ERROR',
  CrashLoop = 'CRASH_LOOP',
  DeadlineExceeded = 'DEADLINE_EXCEEDED',
  ImagePullFailed = 'IMAGE_PULL_FAILED',
  NotReady = 'NOT_READY',
  OomKilled = 'OOM_KILLED',
  QuotaExceeded = 'QUOTA_EXCEEDED',
  Unschedulable = 'UNSCHEDULABLE'
}

export enum RolloutStatus {
  Degraded = 'DEGRADED',
  Failed = 'FAILED',
  Progressing = 'PROGRESSING',
  Ready = 'READY',
  Superseded = 'SUPERSEDED'
}

export enum ScanStatus {
  Clean = 'CLEAN',
  Failed = 'FAILED',
  Findings = 'FINDINGS',
  Queued = 'QUEUED',
  Running = 'RUNNING'
}

export enum ServiceStatus {
  Building = 'BUILDING',
  Degraded = 'DEGRADED',
  Deploying = 'DEPLOYING',
  Failed = 'FAILED',
  Healthy = 'HEALTHY',
  Stopped = 'STOPPED'
}

export type ServiceVariableInput = {
  /** Key of the variable. e.g. PORT or HOST */
  key: string;
  /** Reference to an available variable. Required when no value is set. Mutually exclusive with value. */
  ref?: string | null | undefined;
  /** Literal value. Required when no ref is set. Mutually exclusive with ref. */
  value?: string | null | undefined;
};

export type SetEnvironmentResourcesInput = {
  cpuMillicores: number;
  diskMB: number;
  environment: string;
  memoryMB: number;
  tier: ResourceTier;
};

export type SetServiceScalingInput = {
  autoscaling?: AutoscalingInput | null | undefined;
  replicas: number;
  service: string;
};

export enum SourceProvider {
  Bitbucket = 'BITBUCKET',
  Github = 'GITHUB',
  Gitlab = 'GITLAB'
}

export enum SubscriptionStatus {
  Active = 'ACTIVE',
  Canceled = 'CANCELED',
  Incomplete = 'INCOMPLETE',
  PastDue = 'PAST_DUE',
  Trialing = 'TRIALING'
}

export enum TlsStatus {
  Active = 'ACTIVE',
  Error = 'ERROR',
  None = 'NONE',
  Provisioning = 'PROVISIONING'
}

export type UpdateMemberRoleInput = {
  role: WorkspaceRole;
  userId: string | number;
};

export type UpdateWorkspaceInput = {
  name: string;
};

export type VariableInput = {
  /** Key of the variable. e.g. PORT or HOST */
  key: string;
  value: string;
};

export enum WorkspaceRole {
  Admin = 'ADMIN',
  User = 'USER'
}

export type ApiTokensQueryVariables = Exact<{ [key: string]: never; }>;


export type ApiTokensQuery = { apiTokens: Array<{ id: string, name: string, role: WorkspaceRole, createdAt: string }> };

export type CreateApiTokenMutationVariables = Exact<{
  input: CreateApiTokenInput;
}>;


export type CreateApiTokenMutation = { createApiToken: { token: string, apiToken: { id: string, name: string, role: WorkspaceRole, createdAt: string } } };

export type RevokeApiTokenMutationVariables = Exact<{
  id: string | number;
}>;


export type RevokeApiTokenMutation = { revokeApiToken: boolean };

export type WorkspacesQueryVariables = Exact<{ [key: string]: never; }>;


export type WorkspacesQuery = { workspaces: Array<{ id: string, name: string, personal: boolean }> };

export type ProjectsForNavQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectsForNavQuery = { projects: Array<{ id: string, name: string, environments: Array<{ id: string, name: string, resourceTier: ResourceTier }> }> };

export type GitHubConnectedQueryVariables = Exact<{ [key: string]: never; }>;


export type GitHubConnectedQuery = { githubConnected: boolean };

export type GitHubSourcesQueryVariables = Exact<{ [key: string]: never; }>;


export type GitHubSourcesQuery = { githubSources: Array<{ accountLogin: string, accountAvatarUrl: string, accountType: GitHubAccountType }> };

export type GitHubRepositoriesQueryVariables = Exact<{
  account: string;
}>;


export type GitHubRepositoriesQuery = { githubRepositories: Array<{ id: string, name: string, fullName: string, htmlUrl: string, defaultBranch: string, private: boolean }> };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { createProject: { id: string, name: string, environments: Array<{ id: string, name: string }> } };

export type AddServiceMutationVariables = Exact<{
  environmentId: string;
  input: AddServiceInput;
}>;


export type AddServiceMutation = { addService: { id: string, name: string } };

export type DetectServicesQueryVariables = Exact<{
  repositoryUrl: string;
}>;


export type DetectServicesQuery = { detectServices: Array<{ name: string, language: string, framework: string, contextPath: string, startCommand: string, suggestedPort: number }> };

export type SearchImagesQueryVariables = Exact<{
  query: string;
}>;


export type SearchImagesQuery = { searchImages: Array<{ name: string, description: string, starCount: number, pullCount: number, official: boolean }> };

export type CreateDatabaseMutationVariables = Exact<{
  input: CreateDatabaseInput;
}>;


export type CreateDatabaseMutation = { createDatabase: { id: string, name: string, version: string, size: string } };

export type CreateKeyValueStoreMutationVariables = Exact<{
  input: CreateKeyValueStoreInput;
}>;


export type CreateKeyValueStoreMutation = { createKeyValueStore: { id: string, name: string, version: string, size: string } };

export type CreateBucketMutationVariables = Exact<{
  input: CreateBucketInput;
}>;


export type CreateBucketMutation = { createBucket: { id: string, name: string, region: string, endpoint: string } };

export type CreateVolumeMutationVariables = Exact<{
  environment: string;
  name: string;
  size: string;
}>;


export type CreateVolumeMutation = { createVolume: { id: string, name: string, size: string } };

export type CreateEnvironmentMutationVariables = Exact<{
  input: CreateEnvironmentInput;
}>;


export type CreateEnvironmentMutation = { createEnvironment: { id: string, name: string, resourceTier: ResourceTier } };

export type CreateWorkspaceCheckoutMutationVariables = Exact<{
  input: CreateWorkspaceCheckoutInput;
}>;


export type CreateWorkspaceCheckoutMutation = { createWorkspaceCheckout: { url: string } };

export type MountVolumeMutationVariables = Exact<{
  volume: string;
  service: string;
  path: string;
}>;


export type MountVolumeMutation = { mountVolume: { id: string, mount: { service: string, path: string } | null } };

export type EjectProjectQueryVariables = Exact<{
  id: string;
}>;


export type EjectProjectQuery = { ejectProject: { filename: string, contentType: string, content: string } };

export type SharedVariablesQueryVariables = Exact<{
  environment: string;
}>;


export type SharedVariablesQuery = { sharedVariables: Array<{ key: string, value: string }> };

export type SetSharedVariablesMutationVariables = Exact<{
  environment: string;
  variables: Array<VariableInput> | VariableInput;
}>;


export type SetSharedVariablesMutation = { setSharedVariables: boolean };

export type SubscriptionQueryVariables = Exact<{ [key: string]: never; }>;


export type SubscriptionQuery = { subscription: { plan: Plan | null, status: SubscriptionStatus, currentPeriodEnd: string, creditAmountCents: number, creditExpiry: string | null, hasPaymentMethod: boolean } | null };

export type UsageSummaryQueryVariables = Exact<{ [key: string]: never; }>;


export type UsageSummaryQuery = { usageSummary: { resourceCostCents: number, creditsCents: number, estimatedTotalCents: number } | null };

export type CreatePlanCheckoutMutationVariables = Exact<{
  plan: Plan;
}>;


export type CreatePlanCheckoutMutation = { createPlanCheckout: { url: string } };

export type BucketCredentialsQueryVariables = Exact<{
  bucket: string;
}>;


export type BucketCredentialsQuery = { bucketCredentials: { endpoint: string, region: string, bucket: string, accessKeyId: string, secretAccessKey: string } };

export type SetBucketPublicMutationVariables = Exact<{
  bucket: string;
  public: boolean;
}>;


export type SetBucketPublicMutation = { setBucketPublic: { id: string, public: boolean, publicEndpoint: string } };

export type DeleteBucketMutationVariables = Exact<{
  bucket: string;
}>;


export type DeleteBucketMutation = { deleteBucket: boolean };

export type DatabasePublicQueryVariables = Exact<{
  database: string;
}>;


export type DatabasePublicQuery = { database: { id: string, public: boolean } };

export type DatabaseCredentialsQueryVariables = Exact<{
  database: string;
}>;


export type DatabaseCredentialsQuery = { databaseCredentials: Array<{ type: EndpointType, host: string, port: string, dbname: string, user: string, password: string, uri: string }> };

export type ExposeDatabaseMutationVariables = Exact<{
  database: string;
}>;


export type ExposeDatabaseMutation = { exposeDatabase: { id: string, public: boolean } };

export type UnexposeDatabaseMutationVariables = Exact<{
  database: string;
}>;


export type UnexposeDatabaseMutation = { unexposeDatabase: { id: string, public: boolean } };

export type ExecuteQueryMutationVariables = Exact<{
  database: string;
  query: string;
}>;


export type ExecuteQueryMutation = { executeQuery: { columns: Array<string>, rows: Array<Array<string | null> | null>, affectedRows: number } };

export type DatabaseResourcesQueryVariables = Exact<{
  database: string;
}>;


export type DatabaseResourcesQuery = { database: { id: string, size: string, resources: { cpu: string, memory: string } } };

export type SetDatabaseResourcesMutationVariables = Exact<{
  database: string;
  resources: ResourcesInput;
}>;


export type SetDatabaseResourcesMutation = { setDatabaseResources: { id: string, status: DatabaseStatus, resources: { cpu: string, memory: string } } };

export type SetDatabaseStorageMutationVariables = Exact<{
  database: string;
  size: string;
}>;


export type SetDatabaseStorageMutation = { setDatabaseStorage: { id: string, status: DatabaseStatus, size: string } };

export type DeleteDatabaseMutationVariables = Exact<{
  database: string;
}>;


export type DeleteDatabaseMutation = { deleteDatabase: boolean };

export type DatabaseTablesQueryVariables = Exact<{
  database: string;
}>;


export type DatabaseTablesQuery = { databaseTables: Array<{ name: string, schema: string, estimatedRows: number, columns: Array<{ name: string, type: string, nullable: boolean, primaryKey: boolean }> }> };

export type DatabaseTableDataQueryVariables = Exact<{
  database: string;
  table: string;
  schema?: string | null | undefined;
  limit?: number | null | undefined;
  offset?: number | null | undefined;
}>;


export type DatabaseTableDataQuery = { databaseTableData: { columns: Array<string>, rows: Array<Array<string | null> | null>, totalEstimatedRows: number } };

export type KeyValueStoreCredentialsQueryVariables = Exact<{
  keyValueStore: string;
}>;


export type KeyValueStoreCredentialsQuery = { keyValueStoreCredentials: Array<{ type: EndpointType, host: string, port: string, password: string, uri: string }> };

export type DeleteKeyValueStoreMutationVariables = Exact<{
  keyValueStore: string;
}>;


export type DeleteKeyValueStoreMutation = { deleteKeyValueStore: boolean };

export type SecretScanReportQueryVariables = Exact<{
  id: string;
}>;


export type SecretScanReportQuery = { service: { id: string, secretScanReport: { commit: string, scannedAt: string, findings: Array<{ rule: string, file: string, line: number, commit: string, secret: string, author: string | null, url: string | null, verified: boolean }> } | null } };

export type ServiceMetricsQueryVariables = Exact<{
  id: string;
  range: MetricsRange;
  grouping: MetricGrouping;
}>;


export type ServiceMetricsQuery = { service: { id: string, replicas: { desired: number }, resources: { cpu: string, memory: string }, deployments: Array<{ createdAt: string }>, metrics: Array<{ metric: ResourceMetric, replica: string | null, points: Array<{ timestamp: string, value: number | null }> }> } };

export type RemoveServiceMutationVariables = Exact<{
  service: string;
}>;


export type RemoveServiceMutation = { removeService: boolean };

export type SetCustomStartCommandMutationVariables = Exact<{
  service: string;
  command: string;
}>;


export type SetCustomStartCommandMutation = { setCustomStartCommand: { id: string } };

export type GenerateDomainMutationVariables = Exact<{
  service: string;
}>;


export type GenerateDomainMutation = { generateDomain: { id: string, endpoints: Array<{ host: string, port: number, type: EndpointType, protocol: Protocol, dns: { status: DnsStatus, requiredRecords: Array<{ type: DnsRecordType, host: string, value: string }> } }> } };

export type AddCustomDomainMutationVariables = Exact<{
  service: string;
  hostname: string;
}>;


export type AddCustomDomainMutation = { addCustomDomain: { id: string, endpoints: Array<{ host: string, port: number, type: EndpointType, protocol: Protocol, dns: { status: DnsStatus, requiredRecords: Array<{ type: DnsRecordType, host: string, value: string }> } }> } };

export type RemoveDomainMutationVariables = Exact<{
  service: string;
  hostname: string;
}>;


export type RemoveDomainMutation = { removeDomain: { id: string, endpoints: Array<{ host: string, port: number, type: EndpointType, protocol: Protocol, dns: { status: DnsStatus, requiredRecords: Array<{ type: DnsRecordType, host: string, value: string }> } }> } };

export type SetServiceScalingMutationVariables = Exact<{
  input: SetServiceScalingInput;
}>;


export type SetServiceScalingMutation = { setServiceScaling: { id: string, replicas: { desired: number, ready: number }, autoscaling: { minReplicas: number, maxReplicas: number, targetCpu: number } | null } };

export type SetServicePortMutationVariables = Exact<{
  service: string;
  port?: number | null | undefined;
}>;


export type SetServicePortMutation = { setServicePort: { id: string, port: number } };

export type SetServiceHealthCheckMutationVariables = Exact<{
  service: string;
  healthCheck?: HealthCheckInput | null | undefined;
}>;


export type SetServiceHealthCheckMutation = { setServiceHealthCheck: { id: string, healthCheck: { path: string, port: number, initialDelaySeconds: number, periodSeconds: number, timeoutSeconds: number, failureThreshold: number, startupFailureThreshold: number } | null } };

export type SetServiceBranchMutationVariables = Exact<{
  service: string;
  branch?: string | null | undefined;
}>;


export type SetServiceBranchMutation = { setServiceBranch: { id: string, branch: string | null } };

export type SetAutoDeployMutationVariables = Exact<{
  service: string;
  enabled: boolean;
}>;


export type SetAutoDeployMutation = { setAutoDeploy: { id: string, autoDeploy: boolean } };

export type SetCiDeployMutationVariables = Exact<{
  service: string;
  enabled: boolean;
}>;


export type SetCiDeployMutation = { setCIDeploy: { id: string, ciDeploy: boolean } };

export type RepositoryBranchesQueryVariables = Exact<{
  repositoryUrl: string;
}>;


export type RepositoryBranchesQuery = { repositoryBranches: Array<string> };

export type SetServiceResourcesMutationVariables = Exact<{
  service: string;
  resources: ResourcesInput;
}>;


export type SetServiceResourcesMutation = { setServiceResources: { id: string, resources: { cpu: string, memory: string } } };

export type ServiceVariablesQueryVariables = Exact<{
  service: string;
}>;


export type ServiceVariablesQuery = { serviceVariables: Array<{ key: string, value: string | null, ref: string | null }> };

export type AvailableVariablesQueryVariables = Exact<{
  environment: string;
}>;


export type AvailableVariablesQuery = { availableVariables: Array<{ id: string, key: string, source:
      | { __typename: 'BucketSource', name: string, bucketId: string }
      | { __typename: 'DatabaseSource', name: string, databaseId: string }
      | { __typename: 'KeyValueStoreSource', name: string, keyValueStoreId: string }
      | { __typename: 'SharedSource', name: string }
     }> };

export type SetServiceVariablesMutationVariables = Exact<{
  service: string;
  variables: Array<ServiceVariableInput> | ServiceVariableInput;
}>;


export type SetServiceVariablesMutation = { setServiceVariables: boolean };

export type DeleteVolumeMutationVariables = Exact<{
  volume: string;
}>;


export type DeleteVolumeMutation = { deleteVolume: boolean };

export type ExpandVolumeMutationVariables = Exact<{
  volume: string;
  size: string;
}>;


export type ExpandVolumeMutation = { expandVolume: { id: string, size: string } };

export type UnmountVolumeMutationVariables = Exact<{
  volume: string;
}>;


export type UnmountVolumeMutation = { unmountVolume: { id: string, mount: { service: string, path: string } | null } };

export type VolumeMetricsQueryVariables = Exact<{
  id: string;
  range: MetricsRange;
}>;


export type VolumeMetricsQuery = { volume: { id: string, metrics: Array<{ metric: ResourceMetric, points: Array<{ timestamp: string, value: number | null }> }> } };

export type BuildLogsSubscriptionVariables = Exact<{
  id: string;
}>;


export type BuildLogsSubscription = { buildLogs: string };

export type DeployMutationVariables = Exact<{
  service: string;
  gitRef?: string | null | undefined;
}>;


export type DeployMutation = { deploy: { id: string, status: ReleaseStatus, build: { id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null } | null } };

export type BuildStatusQueryVariables = Exact<{
  id: string;
}>;


export type BuildStatusQuery = { build: { id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null } };

export type DeployLogsSubscriptionVariables = Exact<{
  id: string;
}>;


export type DeployLogsSubscription = { deployLogs: string };

export type ScanLogsSubscriptionVariables = Exact<{
  id: string;
}>;


export type ScanLogsSubscription = { scanLogs: string };

export type ServiceLogsSubscriptionVariables = Exact<{
  service: string;
  tailLines?: number | null | undefined;
}>;


export type ServiceLogsSubscription = { serviceLogs: { line: string, pod: string } };

export type WorkspaceQueryVariables = Exact<{ [key: string]: never; }>;


export type WorkspaceQuery = { workspace: { id: string, name: string, personal: boolean, suspended: boolean, members: Array<{ id: string, email: string, name: string | null, role: WorkspaceRole }> } };

export type CompleteWorkspaceCheckoutMutationVariables = Exact<{
  sessionId: string;
}>;


export type CompleteWorkspaceCheckoutMutation = { completeWorkspaceCheckout: { id: string, name: string, personal: boolean } };

export type EnvironmentQueryVariables = Exact<{
  environment: string;
}>;


export type EnvironmentQuery = { environment: { id: string, name: string, resourceTier: ResourceTier, services: Array<{ id: string, name: string, status: ServiceStatus, port: number, sourceUrl: string, branch: string | null, autoDeploy: boolean, ciDeploy: boolean, contextPath: string, command: string, lastDeployedAt: string | null, createdAt: string, replicas: { desired: number, ready: number }, autoscaling: { minReplicas: number, maxReplicas: number, targetCpu: number } | null, endpoints: Array<{ host: string, port: number, protocol: Protocol, type: EndpointType, tls: TlsStatus, dns: { status: DnsStatus, requiredRecords: Array<{ type: DnsRecordType, host: string, value: string }> } }>, resources: { cpu: string, memory: string }, healthCheck: { path: string, port: number, initialDelaySeconds: number, periodSeconds: number, timeoutSeconds: number, failureThreshold: number, startupFailureThreshold: number } | null, activeDeployment: { id: string, image: string, imageDigest: string | null, commit: string, commitMessage: string, ref: string, status: DeploymentStatus, createdAt: string, replicas: { desired: number, ready: number }, rollout: { status: RolloutStatus, reason: RolloutReason | null, message: string | null, restarts: number, startedAt: string } | null } | null, deployments: Array<{ id: string, image: string, imageDigest: string | null, commit: string, commitMessage: string, ref: string, status: DeploymentStatus, createdAt: string }>, builds: Array<{ id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null }>, releases: Array<{ id: string, status: ReleaseStatus, createdAt: string, trigger: { kind: ReleaseTriggerKind, actor: string | null }, source: { provider: SourceProvider, repository: string, url: string, ref: string, contextPath: string, commit: { sha: string, message: string, url: string | null } } | null, build: { id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null } | null, deploy: { id: string, status: DeployStatus, startedAt: string | null, finishedAt: string | null } | null, scan: { id: string, status: ScanStatus, findingsCount: number | null, verifiedCount: number | null, startedAt: string | null, finishedAt: string | null } | null, deployment: { id: string, image: string, imageDigest: string | null, commit: string, commitMessage: string, ref: string, status: DeploymentStatus, createdAt: string, replicas: { desired: number, ready: number }, rollout: { status: RolloutStatus, reason: RolloutReason | null, message: string | null, restarts: number, startedAt: string } | null } | null }> }>, databases: Array<{ id: string, name: string, version: string, instances: number, status: DatabaseStatus, size: string, createdAt: string }>, keyValueStores: Array<{ id: string, name: string, version: string, status: DatabaseStatus, size: string, createdAt: string }>, buckets: Array<{ id: string, name: string, region: string, endpoint: string, publicEndpoint: string, status: BucketStatus, sizeBytes: number, objectCount: number, public: boolean, createdAt: string }>, volumes: Array<{ id: string, name: string, size: string, mount: { service: string, path: string } | null }> } };

export type EnvironmentVolumeUsageQueryVariables = Exact<{
  environment: string;
}>;


export type EnvironmentVolumeUsageQuery = { environment: { id: string, volumes: Array<{ id: string, size: string, metrics: Array<{ points: Array<{ value: number | null }> }> }> } };

export type EnvironmentDefaultCommandsQueryVariables = Exact<{
  environment: string;
}>;


export type EnvironmentDefaultCommandsQuery = { environment: { id: string, services: Array<{ id: string, defaultCommand: string }> } };

export type ProjectEnvironmentsQueryVariables = Exact<{
  id: string;
}>;


export type ProjectEnvironmentsQuery = { project: { id: string, name: string, environments: Array<{ id: string, name: string, resourceTier: ResourceTier }> } };

export type CompletePlanCheckoutMutationVariables = Exact<{
  sessionId: string;
}>;


export type CompletePlanCheckoutMutation = { completePlanCheckout: { plan: Plan | null, status: SubscriptionStatus, currentPeriodEnd: string, creditAmountCents: number, hasPaymentMethod: boolean } };

export type ProjectPageQueryVariables = Exact<{
  id: string;
}>;


export type ProjectPageQuery = { project: { id: string, name: string, environments: Array<{ id: string, name: string, resourceTier: ResourceTier }> } };

export type ProjectSettingsQueryVariables = Exact<{
  id: string;
}>;


export type ProjectSettingsQuery = { project: { id: string, name: string, environments: Array<{ id: string, name: string, resourceTier: ResourceTier }> } };

export type DeleteProjectMutationVariables = Exact<{
  id: string;
}>;


export type DeleteProjectMutation = { deleteProject: boolean };

export type DeleteEnvironmentMutationVariables = Exact<{
  environment: string;
}>;


export type DeleteEnvironmentMutation = { deleteEnvironment: boolean };

export type EnvironmentResourcesQueryVariables = Exact<{
  environment: string;
}>;


export type EnvironmentResourcesQuery = { environmentResources: { tier: ResourceTier, allocation: { cpuMillicores: number, memoryMB: number, diskMB: number } } | null };

export type SetEnvironmentResourcesMutationVariables = Exact<{
  input: SetEnvironmentResourcesInput;
}>;


export type SetEnvironmentResourcesMutation = { setEnvironmentResources: { id: string, resourceTier: ResourceTier } };

export type ProjectsQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectsQuery = { projects: Array<{ id: string, name: string, environments: Array<{ id: string, name: string, resourceTier: ResourceTier, services: Array<{ id: string, name: string, sourceUrl: string }> }> }> };

export type UpdateWorkspaceMutationVariables = Exact<{
  input: UpdateWorkspaceInput;
}>;


export type UpdateWorkspaceMutation = { updateWorkspace: { id: string, name: string } };

export type DeleteWorkspaceMutationVariables = Exact<{ [key: string]: never; }>;


export type DeleteWorkspaceMutation = { deleteWorkspace: boolean };

export type InviteMemberMutationVariables = Exact<{
  input: InviteMemberInput;
}>;


export type InviteMemberMutation = { inviteMember: { id: string, email: string, name: string | null, role: WorkspaceRole } };

export type RemoveMemberMutationVariables = Exact<{
  userId: string | number;
}>;


export type RemoveMemberMutation = { removeMember: boolean };

export type UpdateMemberRoleMutationVariables = Exact<{
  input: UpdateMemberRoleInput;
}>;


export type UpdateMemberRoleMutation = { updateMemberRole: { id: string, email: string, name: string | null, role: WorkspaceRole } };

export type ChangePlanMutationVariables = Exact<{
  plan: Plan;
}>;


export type ChangePlanMutation = { changePlan: { plan: Plan | null, status: SubscriptionStatus, currentPeriodEnd: string, creditAmountCents: number, creditExpiry: string | null } };

export type BillingPortalUrlMutationVariables = Exact<{ [key: string]: never; }>;


export type BillingPortalUrlMutation = { billingPortalUrl: { url: string } };

export type BootstrapWorkspacesQueryVariables = Exact<{ [key: string]: never; }>;


export type BootstrapWorkspacesQuery = { workspaces: Array<{ id: string }> };


export const ApiTokensDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ApiTokens"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"apiTokens"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}}]}}]} as unknown as DocumentNode<ApiTokensQuery, ApiTokensQueryVariables>;
export const CreateApiTokenDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateApiToken"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateApiTokenInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createApiToken"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"apiToken"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"token"}}]}}]}}]} as unknown as DocumentNode<CreateApiTokenMutation, CreateApiTokenMutationVariables>;
export const RevokeApiTokenDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RevokeApiToken"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"revokeApiToken"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<RevokeApiTokenMutation, RevokeApiTokenMutationVariables>;
export const WorkspacesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Workspaces"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspaces"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}}]}}]}}]} as unknown as DocumentNode<WorkspacesQuery, WorkspacesQueryVariables>;
export const ProjectsForNavDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ProjectsForNav"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"projects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]}}]} as unknown as DocumentNode<ProjectsForNavQuery, ProjectsForNavQueryVariables>;
export const GitHubConnectedDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubConnected"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubConnected"}}]}}]} as unknown as DocumentNode<GitHubConnectedQuery, GitHubConnectedQueryVariables>;
export const GitHubSourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubSources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubSources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"accountLogin"}},{"kind":"Field","name":{"kind":"Name","value":"accountAvatarUrl"}},{"kind":"Field","name":{"kind":"Name","value":"accountType"}}]}}]}}]} as unknown as DocumentNode<GitHubSourcesQuery, GitHubSourcesQueryVariables>;
export const GitHubRepositoriesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubRepositories"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"account"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubRepositories"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"account"},"value":{"kind":"Variable","name":{"kind":"Name","value":"account"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"fullName"}},{"kind":"Field","name":{"kind":"Name","value":"htmlUrl"}},{"kind":"Field","name":{"kind":"Name","value":"defaultBranch"}},{"kind":"Field","name":{"kind":"Name","value":"private"}}]}}]}}]} as unknown as DocumentNode<GitHubRepositoriesQuery, GitHubRepositoriesQueryVariables>;
export const CreateProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateProject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateProjectInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createProject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]} as unknown as DocumentNode<CreateProjectMutation, CreateProjectMutationVariables>;
export const AddServiceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddService"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environmentId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AddServiceInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addService"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environmentId"}}},{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]} as unknown as DocumentNode<AddServiceMutation, AddServiceMutationVariables>;
export const DetectServicesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DetectServices"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"repositoryUrl"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"detectServices"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"repositoryUrl"},"value":{"kind":"Variable","name":{"kind":"Name","value":"repositoryUrl"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"language"}},{"kind":"Field","name":{"kind":"Name","value":"framework"}},{"kind":"Field","name":{"kind":"Name","value":"contextPath"}},{"kind":"Field","name":{"kind":"Name","value":"startCommand"}},{"kind":"Field","name":{"kind":"Name","value":"suggestedPort"}}]}}]}}]} as unknown as DocumentNode<DetectServicesQuery, DetectServicesQueryVariables>;
export const SearchImagesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SearchImages"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"query"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"searchImages"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"query"},"value":{"kind":"Variable","name":{"kind":"Name","value":"query"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"description"}},{"kind":"Field","name":{"kind":"Name","value":"starCount"}},{"kind":"Field","name":{"kind":"Name","value":"pullCount"}},{"kind":"Field","name":{"kind":"Name","value":"official"}}]}}]}}]} as unknown as DocumentNode<SearchImagesQuery, SearchImagesQueryVariables>;
export const CreateDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateDatabaseInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<CreateDatabaseMutation, CreateDatabaseMutationVariables>;
export const CreateKeyValueStoreDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateKeyValueStore"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateKeyValueStoreInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createKeyValueStore"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<CreateKeyValueStoreMutation, CreateKeyValueStoreMutationVariables>;
export const CreateBucketDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateBucket"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateBucketInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createBucket"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"region"}},{"kind":"Field","name":{"kind":"Name","value":"endpoint"}}]}}]}}]} as unknown as DocumentNode<CreateBucketMutation, CreateBucketMutationVariables>;
export const CreateVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"name"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"size"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"name"},"value":{"kind":"Variable","name":{"kind":"Name","value":"name"}}},{"kind":"Argument","name":{"kind":"Name","value":"size"},"value":{"kind":"Variable","name":{"kind":"Name","value":"size"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<CreateVolumeMutation, CreateVolumeMutationVariables>;
export const CreateEnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateEnvironment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateEnvironmentInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createEnvironment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]} as unknown as DocumentNode<CreateEnvironmentMutation, CreateEnvironmentMutationVariables>;
export const CreateWorkspaceCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateWorkspaceCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateWorkspaceCheckoutInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createWorkspaceCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<CreateWorkspaceCheckoutMutation, CreateWorkspaceCheckoutMutationVariables>;
export const MountVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"MountVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"volume"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"path"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"mountVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"volume"},"value":{"kind":"Variable","name":{"kind":"Name","value":"volume"}}},{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"path"},"value":{"kind":"Variable","name":{"kind":"Name","value":"path"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"mount"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"}},{"kind":"Field","name":{"kind":"Name","value":"path"}}]}}]}}]}}]} as unknown as DocumentNode<MountVolumeMutation, MountVolumeMutationVariables>;
export const EjectProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EjectProject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"ejectProject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"filename"}},{"kind":"Field","name":{"kind":"Name","value":"contentType"}},{"kind":"Field","name":{"kind":"Name","value":"content"}}]}}]}}]} as unknown as DocumentNode<EjectProjectQuery, EjectProjectQueryVariables>;
export const SharedVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SharedVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"sharedVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]} as unknown as DocumentNode<SharedVariablesQuery, SharedVariablesQueryVariables>;
export const SetSharedVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetSharedVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"variables"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VariableInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setSharedVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"variables"},"value":{"kind":"Variable","name":{"kind":"Name","value":"variables"}}}]}]}}]} as unknown as DocumentNode<SetSharedVariablesMutation, SetSharedVariablesMutationVariables>;
export const SubscriptionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditExpiry"}},{"kind":"Field","name":{"kind":"Name","value":"hasPaymentMethod"}}]}}]}}]} as unknown as DocumentNode<SubscriptionQuery, SubscriptionQueryVariables>;
export const UsageSummaryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"UsageSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"usageSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"resourceCostCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditsCents"}},{"kind":"Field","name":{"kind":"Name","value":"estimatedTotalCents"}}]}}]}}]} as unknown as DocumentNode<UsageSummaryQuery, UsageSummaryQueryVariables>;
export const CreatePlanCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreatePlanCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"plan"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Plan"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createPlanCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"plan"},"value":{"kind":"Variable","name":{"kind":"Name","value":"plan"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<CreatePlanCheckoutMutation, CreatePlanCheckoutMutationVariables>;
export const BucketCredentialsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BucketCredentials"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"bucketCredentials"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"endpoint"}},{"kind":"Field","name":{"kind":"Name","value":"region"}},{"kind":"Field","name":{"kind":"Name","value":"bucket"}},{"kind":"Field","name":{"kind":"Name","value":"accessKeyId"}},{"kind":"Field","name":{"kind":"Name","value":"secretAccessKey"}}]}}]}}]} as unknown as DocumentNode<BucketCredentialsQuery, BucketCredentialsQueryVariables>;
export const SetBucketPublicDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetBucketPublic"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"public"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setBucketPublic"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}},{"kind":"Argument","name":{"kind":"Name","value":"public"},"value":{"kind":"Variable","name":{"kind":"Name","value":"public"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"public"}},{"kind":"Field","name":{"kind":"Name","value":"publicEndpoint"}}]}}]}}]} as unknown as DocumentNode<SetBucketPublicMutation, SetBucketPublicMutationVariables>;
export const DeleteBucketDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteBucket"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteBucket"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}}]}]}}]} as unknown as DocumentNode<DeleteBucketMutation, DeleteBucketMutationVariables>;
export const DatabasePublicDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabasePublic"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"public"}}]}}]}}]} as unknown as DocumentNode<DatabasePublicQuery, DatabasePublicQueryVariables>;
export const DatabaseCredentialsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseCredentials"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseCredentials"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"dbname"}},{"kind":"Field","name":{"kind":"Name","value":"user"}},{"kind":"Field","name":{"kind":"Name","value":"password"}},{"kind":"Field","name":{"kind":"Name","value":"uri"}}]}}]}}]} as unknown as DocumentNode<DatabaseCredentialsQuery, DatabaseCredentialsQueryVariables>;
export const ExposeDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ExposeDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"exposeDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"public"}}]}}]}}]} as unknown as DocumentNode<ExposeDatabaseMutation, ExposeDatabaseMutationVariables>;
export const UnexposeDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UnexposeDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"unexposeDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"public"}}]}}]}}]} as unknown as DocumentNode<UnexposeDatabaseMutation, UnexposeDatabaseMutationVariables>;
export const ExecuteQueryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ExecuteQuery"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"query"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"executeQuery"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"query"},"value":{"kind":"Variable","name":{"kind":"Name","value":"query"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"columns"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"affectedRows"}}]}}]}}]} as unknown as DocumentNode<ExecuteQueryMutation, ExecuteQueryMutationVariables>;
export const DatabaseResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseResourcesQuery, DatabaseResourcesQueryVariables>;
export const SetDatabaseResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetDatabaseResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"resources"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ResourcesInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setDatabaseResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"resources"},"value":{"kind":"Variable","name":{"kind":"Name","value":"resources"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}}]}}]}}]} as unknown as DocumentNode<SetDatabaseResourcesMutation, SetDatabaseResourcesMutationVariables>;
export const SetDatabaseStorageDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetDatabaseStorage"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"size"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setDatabaseStorage"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"size"},"value":{"kind":"Variable","name":{"kind":"Name","value":"size"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<SetDatabaseStorageMutation, SetDatabaseStorageMutationVariables>;
export const DeleteDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}]}]}}]} as unknown as DocumentNode<DeleteDatabaseMutation, DeleteDatabaseMutationVariables>;
export const DatabaseTablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseTables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseTables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"schema"}},{"kind":"Field","name":{"kind":"Name","value":"estimatedRows"}},{"kind":"Field","name":{"kind":"Name","value":"columns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"nullable"}},{"kind":"Field","name":{"kind":"Name","value":"primaryKey"}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseTablesQuery, DatabaseTablesQueryVariables>;
export const DatabaseTableDataDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseTableData"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"table"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"schema"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseTableData"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"table"},"value":{"kind":"Variable","name":{"kind":"Name","value":"table"}}},{"kind":"Argument","name":{"kind":"Name","value":"schema"},"value":{"kind":"Variable","name":{"kind":"Name","value":"schema"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"columns"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"totalEstimatedRows"}}]}}]}}]} as unknown as DocumentNode<DatabaseTableDataQuery, DatabaseTableDataQueryVariables>;
export const KeyValueStoreCredentialsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"KeyValueStoreCredentials"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"keyValueStore"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"KeyValueStoreID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"keyValueStoreCredentials"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"keyValueStore"},"value":{"kind":"Variable","name":{"kind":"Name","value":"keyValueStore"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"password"}},{"kind":"Field","name":{"kind":"Name","value":"uri"}}]}}]}}]} as unknown as DocumentNode<KeyValueStoreCredentialsQuery, KeyValueStoreCredentialsQueryVariables>;
export const DeleteKeyValueStoreDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteKeyValueStore"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"keyValueStore"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"KeyValueStoreID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteKeyValueStore"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"keyValueStore"},"value":{"kind":"Variable","name":{"kind":"Name","value":"keyValueStore"}}}]}]}}]} as unknown as DocumentNode<DeleteKeyValueStoreMutation, DeleteKeyValueStoreMutationVariables>;
export const SecretScanReportDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SecretScanReport"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"secretScanReport"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"scannedAt"}},{"kind":"Field","name":{"kind":"Name","value":"findings"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"rule"}},{"kind":"Field","name":{"kind":"Name","value":"file"}},{"kind":"Field","name":{"kind":"Name","value":"line"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"secret"}},{"kind":"Field","name":{"kind":"Name","value":"author"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"verified"}}]}}]}}]}}]}}]} as unknown as DocumentNode<SecretScanReportQuery, SecretScanReportQueryVariables>;
export const ServiceMetricsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ServiceMetrics"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"range"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"MetricsRange"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"grouping"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"MetricGrouping"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}}]}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}},{"kind":"Field","name":{"kind":"Name","value":"deployments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"metrics"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"metrics"},"value":{"kind":"ListValue","values":[{"kind":"EnumValue","value":"CPU_USAGE"},{"kind":"EnumValue","value":"MEMORY_USAGE"}]}},{"kind":"Argument","name":{"kind":"Name","value":"range"},"value":{"kind":"Variable","name":{"kind":"Name","value":"range"}}},{"kind":"Argument","name":{"kind":"Name","value":"grouping"},"value":{"kind":"Variable","name":{"kind":"Name","value":"grouping"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"metric"}},{"kind":"Field","name":{"kind":"Name","value":"replica"}},{"kind":"Field","name":{"kind":"Name","value":"points"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]} as unknown as DocumentNode<ServiceMetricsQuery, ServiceMetricsQueryVariables>;
export const RemoveServiceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveService"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeService"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}]}]}}]} as unknown as DocumentNode<RemoveServiceMutation, RemoveServiceMutationVariables>;
export const SetCustomStartCommandDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetCustomStartCommand"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"command"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setCustomStartCommand"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"command"},"value":{"kind":"Variable","name":{"kind":"Name","value":"command"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}}]}}]}}]} as unknown as DocumentNode<SetCustomStartCommandMutation, SetCustomStartCommandMutationVariables>;
export const GenerateDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"GenerateDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"generateDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<GenerateDomainMutation, GenerateDomainMutationVariables>;
export const AddCustomDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddCustomDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addCustomDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"hostname"},"value":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<AddCustomDomainMutation, AddCustomDomainMutationVariables>;
export const RemoveDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"hostname"},"value":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<RemoveDomainMutation, RemoveDomainMutationVariables>;
export const SetServiceScalingDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceScaling"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"SetServiceScalingInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceScaling"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"autoscaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"minReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"maxReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"targetCpu"}}]}}]}}]}}]} as unknown as DocumentNode<SetServiceScalingMutation, SetServiceScalingMutationVariables>;
export const SetServicePortDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServicePort"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"port"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServicePort"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"port"},"value":{"kind":"Variable","name":{"kind":"Name","value":"port"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"port"}}]}}]}}]} as unknown as DocumentNode<SetServicePortMutation, SetServicePortMutationVariables>;
export const SetServiceHealthCheckDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceHealthCheck"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"healthCheck"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"HealthCheckInput"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceHealthCheck"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"healthCheck"},"value":{"kind":"Variable","name":{"kind":"Name","value":"healthCheck"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"healthCheck"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"path"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"initialDelaySeconds"}},{"kind":"Field","name":{"kind":"Name","value":"periodSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"timeoutSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"failureThreshold"}},{"kind":"Field","name":{"kind":"Name","value":"startupFailureThreshold"}}]}}]}}]}}]} as unknown as DocumentNode<SetServiceHealthCheckMutation, SetServiceHealthCheckMutationVariables>;
export const SetServiceBranchDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceBranch"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"branch"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceBranch"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"branch"},"value":{"kind":"Variable","name":{"kind":"Name","value":"branch"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"branch"}}]}}]}}]} as unknown as DocumentNode<SetServiceBranchMutation, SetServiceBranchMutationVariables>;
export const SetAutoDeployDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetAutoDeploy"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"enabled"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setAutoDeploy"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"enabled"},"value":{"kind":"Variable","name":{"kind":"Name","value":"enabled"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"autoDeploy"}}]}}]}}]} as unknown as DocumentNode<SetAutoDeployMutation, SetAutoDeployMutationVariables>;
export const SetCiDeployDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetCIDeploy"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"enabled"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setCIDeploy"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"enabled"},"value":{"kind":"Variable","name":{"kind":"Name","value":"enabled"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"ciDeploy"}}]}}]}}]} as unknown as DocumentNode<SetCiDeployMutation, SetCiDeployMutationVariables>;
export const RepositoryBranchesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"RepositoryBranches"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"repositoryUrl"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"repositoryBranches"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"repositoryUrl"},"value":{"kind":"Variable","name":{"kind":"Name","value":"repositoryUrl"}}}]}]}}]} as unknown as DocumentNode<RepositoryBranchesQuery, RepositoryBranchesQueryVariables>;
export const SetServiceResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"resources"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ResourcesInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"resources"},"value":{"kind":"Variable","name":{"kind":"Name","value":"resources"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}}]}}]}}]} as unknown as DocumentNode<SetServiceResourcesMutation, SetServiceResourcesMutationVariables>;
export const ServiceVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ServiceVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serviceVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"value"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}}]}}]}}]} as unknown as DocumentNode<ServiceVariablesQuery, ServiceVariablesQueryVariables>;
export const AvailableVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AvailableVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availableVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"source"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseSource"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","alias":{"kind":"Name","value":"databaseId"},"name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"KeyValueStoreSource"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","alias":{"kind":"Name","value":"keyValueStoreId"},"name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"BucketSource"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","alias":{"kind":"Name","value":"bucketId"},"name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"SharedSource"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]} as unknown as DocumentNode<AvailableVariablesQuery, AvailableVariablesQueryVariables>;
export const SetServiceVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"variables"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceVariableInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"variables"},"value":{"kind":"Variable","name":{"kind":"Name","value":"variables"}}}]}]}}]} as unknown as DocumentNode<SetServiceVariablesMutation, SetServiceVariablesMutationVariables>;
export const DeleteVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"volume"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"volume"},"value":{"kind":"Variable","name":{"kind":"Name","value":"volume"}}}]}]}}]} as unknown as DocumentNode<DeleteVolumeMutation, DeleteVolumeMutationVariables>;
export const ExpandVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ExpandVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"volume"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"size"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"expandVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"volume"},"value":{"kind":"Variable","name":{"kind":"Name","value":"volume"}}},{"kind":"Argument","name":{"kind":"Name","value":"size"},"value":{"kind":"Variable","name":{"kind":"Name","value":"size"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<ExpandVolumeMutation, ExpandVolumeMutationVariables>;
export const UnmountVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UnmountVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"volume"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"unmountVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"volume"},"value":{"kind":"Variable","name":{"kind":"Name","value":"volume"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"mount"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"}},{"kind":"Field","name":{"kind":"Name","value":"path"}}]}}]}}]}}]} as unknown as DocumentNode<UnmountVolumeMutation, UnmountVolumeMutationVariables>;
export const VolumeMetricsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"VolumeMetrics"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"range"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"MetricsRange"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"volume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"metrics"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"metrics"},"value":{"kind":"ListValue","values":[{"kind":"EnumValue","value":"STORAGE_USED"}]}},{"kind":"Argument","name":{"kind":"Name","value":"range"},"value":{"kind":"Variable","name":{"kind":"Name","value":"range"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"metric"}},{"kind":"Field","name":{"kind":"Name","value":"points"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]} as unknown as DocumentNode<VolumeMetricsQuery, VolumeMetricsQueryVariables>;
export const BuildLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"BuildLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BuildID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"buildLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<BuildLogsSubscription, BuildLogsSubscriptionVariables>;
export const DeployDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Deploy"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"gitRef"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deploy"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"gitRef"},"value":{"kind":"Variable","name":{"kind":"Name","value":"gitRef"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"build"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}}]}}]}}]} as unknown as DocumentNode<DeployMutation, DeployMutationVariables>;
export const BuildStatusDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BuildStatus"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BuildID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"build"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}}]}}]} as unknown as DocumentNode<BuildStatusQuery, BuildStatusQueryVariables>;
export const DeployLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"DeployLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DeployID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deployLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<DeployLogsSubscription, DeployLogsSubscriptionVariables>;
export const ScanLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"ScanLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ScanID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"scanLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<ScanLogsSubscription, ScanLogsSubscriptionVariables>;
export const ServiceLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"ServiceLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"tailLines"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serviceLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"tailLines"},"value":{"kind":"Variable","name":{"kind":"Name","value":"tailLines"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"line"}},{"kind":"Field","name":{"kind":"Name","value":"pod"}}]}}]}}]} as unknown as DocumentNode<ServiceLogsSubscription, ServiceLogsSubscriptionVariables>;
export const WorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}},{"kind":"Field","name":{"kind":"Name","value":"suspended"}},{"kind":"Field","name":{"kind":"Name","value":"members"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]}}]} as unknown as DocumentNode<WorkspaceQuery, WorkspaceQueryVariables>;
export const CompleteWorkspaceCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CompleteWorkspaceCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"completeWorkspaceCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"sessionId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}}]}}]}}]} as unknown as DocumentNode<CompleteWorkspaceCheckoutMutation, CompleteWorkspaceCheckoutMutationVariables>;
export const EnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Environment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"autoscaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"minReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"maxReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"targetCpu"}}]}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"tls"}}]}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}},{"kind":"Field","name":{"kind":"Name","value":"branch"}},{"kind":"Field","name":{"kind":"Name","value":"autoDeploy"}},{"kind":"Field","name":{"kind":"Name","value":"ciDeploy"}},{"kind":"Field","name":{"kind":"Name","value":"contextPath"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}},{"kind":"Field","name":{"kind":"Name","value":"command"}},{"kind":"Field","name":{"kind":"Name","value":"healthCheck"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"path"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"initialDelaySeconds"}},{"kind":"Field","name":{"kind":"Name","value":"periodSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"timeoutSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"failureThreshold"}},{"kind":"Field","name":{"kind":"Name","value":"startupFailureThreshold"}}]}},{"kind":"Field","name":{"kind":"Name","value":"activeDeployment"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"imageDigest"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"commitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"rollout"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"reason"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"restarts"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"deployments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"imageDigest"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"commitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"builds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"releases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"trigger"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"kind"}},{"kind":"Field","name":{"kind":"Name","value":"actor"}}]}},{"kind":"Field","name":{"kind":"Name","value":"source"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"provider"}},{"kind":"Field","name":{"kind":"Name","value":"repository"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"contextPath"}},{"kind":"Field","name":{"kind":"Name","value":"commit"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"sha"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"build"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"deploy"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"scan"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"findingsCount"}},{"kind":"Field","name":{"kind":"Name","value":"verifiedCount"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"deployment"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"imageDigest"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"commitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"rollout"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"reason"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"restarts"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}}]}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"lastDeployedAt"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"instances"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"keyValueStores"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"buckets"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"region"}},{"kind":"Field","name":{"kind":"Name","value":"endpoint"}},{"kind":"Field","name":{"kind":"Name","value":"publicEndpoint"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"sizeBytes"}},{"kind":"Field","name":{"kind":"Name","value":"objectCount"}},{"kind":"Field","name":{"kind":"Name","value":"public"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"volumes"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"mount"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"}},{"kind":"Field","name":{"kind":"Name","value":"path"}}]}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentQuery, EnvironmentQueryVariables>;
export const EnvironmentVolumeUsageDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EnvironmentVolumeUsage"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"volumes"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"metrics"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"metrics"},"value":{"kind":"ListValue","values":[{"kind":"EnumValue","value":"STORAGE_USED"}]}},{"kind":"Argument","name":{"kind":"Name","value":"range"},"value":{"kind":"ObjectValue","fields":[{"kind":"ObjectField","name":{"kind":"Name","value":"window"},"value":{"kind":"EnumValue","value":"LAST_1H"}}]}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"points"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentVolumeUsageQuery, EnvironmentVolumeUsageQueryVariables>;
export const EnvironmentDefaultCommandsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EnvironmentDefaultCommands"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"defaultCommand"}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentDefaultCommandsQuery, EnvironmentDefaultCommandsQueryVariables>;
export const ProjectEnvironmentsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ProjectEnvironments"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"project"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]}}]} as unknown as DocumentNode<ProjectEnvironmentsQuery, ProjectEnvironmentsQueryVariables>;
export const CompletePlanCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CompletePlanCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"completePlanCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"sessionId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"hasPaymentMethod"}}]}}]}}]} as unknown as DocumentNode<CompletePlanCheckoutMutation, CompletePlanCheckoutMutationVariables>;
export const ProjectPageDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ProjectPage"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"project"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]}}]} as unknown as DocumentNode<ProjectPageQuery, ProjectPageQueryVariables>;
export const ProjectSettingsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ProjectSettings"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"project"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]}}]} as unknown as DocumentNode<ProjectSettingsQuery, ProjectSettingsQueryVariables>;
export const DeleteProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteProject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteProject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<DeleteProjectMutation, DeleteProjectMutationVariables>;
export const DeleteEnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteEnvironment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteEnvironment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}]}]}}]} as unknown as DocumentNode<DeleteEnvironmentMutation, DeleteEnvironmentMutationVariables>;
export const EnvironmentResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EnvironmentResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environmentResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"tier"}},{"kind":"Field","name":{"kind":"Name","value":"allocation"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpuMillicores"}},{"kind":"Field","name":{"kind":"Name","value":"memoryMB"}},{"kind":"Field","name":{"kind":"Name","value":"diskMB"}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentResourcesQuery, EnvironmentResourcesQueryVariables>;
export const SetEnvironmentResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetEnvironmentResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"SetEnvironmentResourcesInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setEnvironmentResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]} as unknown as DocumentNode<SetEnvironmentResourcesMutation, SetEnvironmentResourcesMutationVariables>;
export const ProjectsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Projects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"projects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}}]}}]}}]}}]}}]} as unknown as DocumentNode<ProjectsQuery, ProjectsQueryVariables>;
export const UpdateWorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateWorkspace"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateWorkspaceInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateWorkspace"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]} as unknown as DocumentNode<UpdateWorkspaceMutation, UpdateWorkspaceMutationVariables>;
export const DeleteWorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteWorkspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteWorkspace"}}]}}]} as unknown as DocumentNode<DeleteWorkspaceMutation, DeleteWorkspaceMutationVariables>;
export const InviteMemberDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"InviteMember"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"InviteMemberInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"inviteMember"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]} as unknown as DocumentNode<InviteMemberMutation, InviteMemberMutationVariables>;
export const RemoveMemberDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveMember"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"userId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeMember"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"userId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"userId"}}}]}]}}]} as unknown as DocumentNode<RemoveMemberMutation, RemoveMemberMutationVariables>;
export const UpdateMemberRoleDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateMemberRole"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateMemberRoleInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateMemberRole"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]} as unknown as DocumentNode<UpdateMemberRoleMutation, UpdateMemberRoleMutationVariables>;
export const ChangePlanDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ChangePlan"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"plan"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Plan"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"changePlan"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"plan"},"value":{"kind":"Variable","name":{"kind":"Name","value":"plan"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditExpiry"}}]}}]}}]} as unknown as DocumentNode<ChangePlanMutation, ChangePlanMutationVariables>;
export const BillingPortalUrlDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"BillingPortalUrl"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"billingPortalUrl"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<BillingPortalUrlMutation, BillingPortalUrlMutationVariables>;
export const BootstrapWorkspacesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BootstrapWorkspaces"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspaces"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}}]}}]}}]} as unknown as DocumentNode<BootstrapWorkspacesQuery, BootstrapWorkspacesQueryVariables>;