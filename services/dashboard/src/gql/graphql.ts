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
  /** The user id an image-based service runs as, which also owns its mounted volumes. */
  user?: number | null | undefined;
  variables?: Array<VariableInput> | null | undefined;
};

export enum AgentApprovalDecision {
  /** Allows this one call. */
  AllowOnce = 'ALLOW_ONCE',
  Deny = 'DENY'
}

export enum AgentApprovalStatus {
  Allowed = 'ALLOWED',
  Denied = 'DENIED',
  Pending = 'PENDING'
}

export enum AgentAuditKind {
  /** A call to a model. */
  Model = 'MODEL',
  /** A call to a connector. */
  Tool = 'TOOL'
}

export enum AgentAuditResult {
  /** Refused because the agent isn't allowed to make the call. */
  Denied = 'DENIED',
  Failed = 'FAILED',
  Ok = 'OK'
}

export type AgentConnectorInput = {
  approval: Approval;
  connector: string;
  /** Tool names of the connector to hide from the agent. */
  hiddenTools?: Array<string> | null | undefined;
};

export type AgentModelInput = {
  name: string;
  provider: string;
};

export enum AgentNoticeSeverity {
  Error = 'ERROR',
  Info = 'INFO',
  Warning = 'WARNING'
}

export enum AgentSessionStatus {
  Idle = 'IDLE',
  WaitingForApproval = 'WAITING_FOR_APPROVAL',
  Working = 'WORKING'
}

export enum AgentStatus {
  /** The agent doesn't start. */
  Failed = 'FAILED',
  Idle = 'IDLE',
  /** The agent is starting, or can't be reached yet. */
  Starting = 'STARTING',
  /** At least one session waits for an approval. */
  WaitingForApproval = 'WAITING_FOR_APPROVAL',
  /** At least one session is working. */
  Working = 'WORKING'
}

export enum AgentToolCallStatus {
  Completed = 'COMPLETED',
  /** Someone in the workspace denied the call. */
  Denied = 'DENIED',
  Failed = 'FAILED',
  Running = 'RUNNING'
}

export enum AgentTurnEndReason {
  /** The agent finished its turn. */
  Completed = 'COMPLETED',
  Failed = 'FAILED',
  /** Someone stopped the session. */
  Stopped = 'STOPPED'
}

/** Whether an agent may use a connector's tools on its own. */
export enum Approval {
  /** The agent uses the tools without asking. */
  Allow = 'ALLOW',
  /** Every tool call waits for someone in the workspace to allow it. */
  Ask = 'ASK'
}

export type AutoscalingInput = {
  enabled: boolean;
  maxReplicas: number;
  minReplicas: number;
  targetCPU: number;
};

/** Whether a workspace can use something now, and what unlocks it otherwise. */
export enum Availability {
  /** It can be used. */
  Available = 'AVAILABLE',
  /** A card on file unlocks it. */
  PaymentMethodRequired = 'PAYMENT_METHOD_REQUIRED',
  /** A higher plan unlocks it. `plan` names the cheapest one. */
  PlanRequired = 'PLAN_REQUIRED',
  /** It isn't offered here. */
  Unavailable = 'UNAVAILABLE'
}

export enum BackupStatus {
  Completed = 'COMPLETED',
  Failed = 'FAILED',
  Pending = 'PENDING',
  Running = 'RUNNING'
}

export enum BackupTrigger {
  Manual = 'MANUAL',
  Scheduled = 'SCHEDULED'
}

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
  /** Not supported yet. Leave it unset, because new environments always start empty. */
  fromEnvironment?: string | null | undefined;
  name: string;
  project: string;
  /** Defaults to `ECO`. */
  tier?: ResourceTier | null | undefined;
};

export type CreateKeyValueStoreInput = {
  environment: string;
  name: string;
  size?: string | null | undefined;
};

export type CreateProjectInput = {
  /** The project id. Derived from `name` when left out. */
  id?: string | null | undefined;
  /** The display name of the project, such as `My API`. */
  name: string;
};

export type CreateWorkspaceCheckoutInput = {
  id: string;
  name: string;
  plan: Plan;
};

/** The outcome of the last check of a credential. */
export enum CredentialStatus {
  /** The service turned the credential down. */
  Rejected = 'REJECTED',
  /** The service couldn't be reached, or didn't answer as expected. */
  Unreachable = 'UNREACHABLE',
  Valid = 'VALID'
}

export type DatabaseAllowRuleInput = {
  /** A note on whose address this is, such as `Office`. */
  description: string;
  /** An IPv4 address, such as `203.0.113.7`, or a range in CIDR notation, such as `203.0.113.0/24`. `0.0.0.0/0` allows any address. */
  range: string;
};

export enum DatabaseInstanceRole {
  /** The instance that serves writes. */
  Primary = 'PRIMARY',
  /** An instance that replicates from the primary and takes over if it fails. */
  Replica = 'REPLICA'
}

/** A measurement of a database that can be charted over time. */
export enum DatabaseMetric {
  /** The share of reads served from memory instead of disk, from 0 to 1. */
  CacheHitRatio = 'CACHE_HIT_RATIO',
  /** Open connections, idle ones included. */
  Connections = 'CONNECTIONS',
  /** CPU in use, in cores. */
  CpuUsage = 'CPU_USAGE',
  /** Deadlocks detected in each interval of the range. */
  Deadlocks = 'DEADLOCKS',
  /** Memory in use, in bytes. */
  MemoryUsage = 'MEMORY_USAGE',
  /** How far a replica trails the primary, in seconds. */
  ReplicationLag = 'REPLICATION_LAG',
  /** Rows read by queries per second. */
  RowsRead = 'ROWS_READ',
  /** Rows inserted, updated or deleted per second. */
  RowsWritten = 'ROWS_WRITTEN',
  /** Space taken on the storage of the database, in bytes. */
  StorageUsed = 'STORAGE_USED',
  /** Transactions committed or rolled back per second. */
  Transactions = 'TRANSACTIONS'
}

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
  Alias = 'ALIAS',
  Cname = 'CNAME',
  Txt = 'TXT'
}

export enum DnsStatus {
  Error = 'ERROR',
  Misconfigured = 'MISCONFIGURED',
  Pending = 'PENDING',
  Valid = 'VALID'
}

/** How hard the model thinks before it answers. */
export enum Effort {
  High = 'HIGH',
  Low = 'LOW',
  Max = 'MAX',
  Medium = 'MEDIUM'
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

/** The coding agent that runs inside an agent. */
export enum Harness {
  /** Claude Code. Uses Anthropic models. */
  ClaudeCode = 'CLAUDE_CODE',
  /** Codex. Uses OpenAI models. */
  Codex = 'CODEX',
  /** OpenCode. Uses the models of every provider. */
  Opencode = 'OPENCODE'
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

/** The API a model provider speaks. */
export enum ModelApi {
  Anthropic = 'ANTHROPIC',
  Openai = 'OPENAI',
  /** Any API that speaks OpenAI's chat completions, such as Infomaniak, Scaleway or Mistral. Needs a `url`. */
  OpenaiCompatible = 'OPENAI_COMPATIBLE'
}

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

export type RestoreDatabaseInput = {
  database: string;
  name: string;
  /** Omit to restore the most recent state in the archive. */
  targetTime?: string | null | undefined;
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
  /** The name of the variable, such as `PORT` or `HOST`. */
  key: string;
  /** Links the variable to one from `availableVariables`. Set either `value` or `ref`. */
  ref?: string | null | undefined;
  /** A literal value. Set either `value` or `ref`. */
  value?: string | null | undefined;
};

/** `ECO` lets services share capacity and burst beyond what is reserved for them. `PRODUCTION` reserves the full CPU and memory of every service. */
export type SetEnvironmentResourcesInput = {
  environment: string;
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
  Trialing = 'TRIALING',
  TrialEnded = 'TRIAL_ENDED'
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
  /** The name of the variable, such as `PORT` or `HOST`. */
  key: string;
  value: string;
};

export enum VulnerabilitySeverity {
  Critical = 'CRITICAL',
  High = 'HIGH',
  Low = 'LOW',
  Medium = 'MEDIUM',
  Unknown = 'UNKNOWN'
}

export enum VulnerabilitySource {
  Application = 'APPLICATION',
  OperatingSystem = 'OPERATING_SYSTEM',
  Unknown = 'UNKNOWN'
}

export enum WorkspaceRole {
  Admin = 'ADMIN',
  User = 'USER'
}

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

export type AgentPaletteConnectorsQueryVariables = Exact<{ [key: string]: never; }>;


export type AgentPaletteConnectorsQuery = { connectors: Array<{ id: string, name: string, url: string }> };

export type CreateAgentMutationVariables = Exact<{
  environment: string;
  name: string;
  harness: Harness;
  instructions: string;
  connectors?: Array<AgentConnectorInput> | AgentConnectorInput | null | undefined;
}>;


export type CreateAgentMutation = { createAgent: { id: string, name: string } };

export type CreateEnvironmentMutationVariables = Exact<{
  input: CreateEnvironmentInput;
}>;


export type CreateEnvironmentMutation = { createEnvironment: { id: string, name: string, resourceTier: ResourceTier } };

export type CreateWorkspaceCheckoutMutationVariables = Exact<{
  input: CreateWorkspaceCheckoutInput;
}>;


export type CreateWorkspaceCheckoutMutation = { createWorkspaceCheckout: { url: string } };

export type GatePlansQueryVariables = Exact<{ [key: string]: never; }>;


export type GatePlansQuery = { plans: Array<{ plan: Plan | null, priceCents: number, currency: string, limits: { cpu: string, memory: string, replicas: number, volumes: boolean, volumeSize: string, databaseStorage: string, members: number | null, customDomains: number | null } }> };

export type GateSubscriptionQueryVariables = Exact<{ [key: string]: never; }>;


export type GateSubscriptionQuery = { subscription: { plan: Plan | null, status: SubscriptionStatus } | null };

export type GateMembersQueryVariables = Exact<{ [key: string]: never; }>;


export type GateMembersQuery = { workspace: { id: string, members: Array<{ id: string, name: string | null, email: string, role: WorkspaceRole }> } };

export type GateSelectPlanMutationVariables = Exact<{
  plan: Plan;
}>;


export type GateSelectPlanMutation = { selectPlan: { checkoutUrl: string | null, subscription: { plan: Plan | null, status: SubscriptionStatus } | null } };

export type AddPaymentMethodMutationVariables = Exact<{
  returnPath: string;
}>;


export type AddPaymentMethodMutation = { addPaymentMethod: { url: string } };

export type MountVolumeMutationVariables = Exact<{
  volume: string;
  service: string;
  path: string;
}>;


export type MountVolumeMutation = { mountVolume: { id: string, mount: { service: string, path: string } | null } };

export type CreateModelProviderMutationVariables = Exact<{
  name: string;
  api: ModelApi;
  url?: string | null | undefined;
  apiKey: string;
}>;


export type CreateModelProviderMutation = { createModelProvider: { id: string, name: string, api: ModelApi, url: string, keyHint: string | null, status: CredentialStatus, checkedAt: string | null, models: Array<string> } };

export type CreateConnectorMutationVariables = Exact<{
  name: string;
  url: string;
  credential?: string | null | undefined;
}>;


export type CreateConnectorMutation = { createConnector: { id: string, name: string, url: string, credentialHint: string | null, status: CredentialStatus, checkedAt: string | null, tools: Array<{ name: string, description: string | null }> } };

export type EjectProjectQueryVariables = Exact<{
  id: string;
}>;


export type EjectProjectQuery = { ejectProject: { filename: string, contentType: string, content: string } };

export type SubscriptionQueryVariables = Exact<{ [key: string]: never; }>;


export type SubscriptionQuery = { subscription: { plan: Plan | null, status: SubscriptionStatus, currentPeriodEnd: string, creditAmountCents: number, creditExpiry: string | null, hasPaymentMethod: boolean } | null };

export type UsageSummaryQueryVariables = Exact<{ [key: string]: never; }>;


export type UsageSummaryQuery = { usageSummary: { resourceCostCents: number, creditsCents: number, estimatedTotalCents: number } | null };

export type WelcomeSelectPlanMutationVariables = Exact<{
  plan: Plan;
}>;


export type WelcomeSelectPlanMutation = { selectPlan: { checkoutUrl: string | null } };

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

export type ConnectionsQueryVariables = Exact<{ [key: string]: never; }>;


export type ConnectionsQuery = { modelProviders: Array<{ id: string, name: string, api: ModelApi, status: CredentialStatus, models: Array<string> }>, connectors: Array<{ id: string, name: string, url: string, status: CredentialStatus, tools: Array<{ name: string }> }> };

export type ConnectorQueryVariables = Exact<{
  id: string;
}>;


export type ConnectorQuery = { connector: { id: string, name: string, url: string, credentialHint: string | null, status: CredentialStatus, checkedAt: string | null, tools: Array<{ name: string, description: string | null }>, agents: Array<{ id: string, name: string, harness: Harness }> } };

export type ReplaceConnectorCredentialMutationVariables = Exact<{
  connector: string;
  credential: string;
}>;


export type ReplaceConnectorCredentialMutation = { replaceConnectorCredential: { id: string, credentialHint: string | null, status: CredentialStatus, checkedAt: string | null, tools: Array<{ name: string, description: string | null }> } };

export type RefreshConnectorMutationVariables = Exact<{
  connector: string;
}>;


export type RefreshConnectorMutation = { refreshConnector: { id: string, status: CredentialStatus, checkedAt: string | null, tools: Array<{ name: string, description: string | null }> } };

export type DeleteConnectorMutationVariables = Exact<{
  connector: string;
}>;


export type DeleteConnectorMutation = { deleteConnector: boolean };

export type SetEnvironmentResourcesMutationVariables = Exact<{
  input: SetEnvironmentResourcesInput;
}>;


export type SetEnvironmentResourcesMutation = { setEnvironmentResources: { id: string, resourceTier: ResourceTier } };

export type DeleteEnvironmentMutationVariables = Exact<{
  environment: string;
}>;


export type DeleteEnvironmentMutation = { deleteEnvironment: boolean };

export type ModelProviderQueryVariables = Exact<{
  id: string;
}>;


export type ModelProviderQuery = { modelProvider: { id: string, name: string, api: ModelApi, url: string, keyHint: string | null, status: CredentialStatus, checkedAt: string | null, models: Array<string>, agents: Array<{ id: string, name: string, harness: Harness }> } };

export type ReplaceModelProviderKeyMutationVariables = Exact<{
  provider: string;
  apiKey: string;
}>;


export type ReplaceModelProviderKeyMutation = { replaceModelProviderKey: { id: string, keyHint: string | null, status: CredentialStatus, checkedAt: string | null, models: Array<string> } };

export type RefreshModelProviderMutationVariables = Exact<{
  provider: string;
}>;


export type RefreshModelProviderMutation = { refreshModelProvider: { id: string, status: CredentialStatus, checkedAt: string | null, models: Array<string> } };

export type DeleteModelProviderMutationVariables = Exact<{
  provider: string;
}>;


export type DeleteModelProviderMutation = { deleteModelProvider: boolean };

export type DeleteProjectMutationVariables = Exact<{
  id: string;
}>;


export type DeleteProjectMutation = { deleteProject: boolean };

export type SharedVariablesQueryVariables = Exact<{
  environment: string;
}>;


export type SharedVariablesQuery = { sharedVariables: Array<{ key: string, value: string }> };

export type SetSharedVariablesMutationVariables = Exact<{
  environment: string;
  variables: Array<VariableInput> | VariableInput;
}>;


export type SetSharedVariablesMutation = { setSharedVariables: boolean };

export type SelectPlanMutationVariables = Exact<{
  plan: Plan;
}>;


export type SelectPlanMutation = { selectPlan: { checkoutUrl: string | null, subscription: { plan: Plan | null, status: SubscriptionStatus, currentPeriodEnd: string, creditAmountCents: number, creditExpiry: string | null, hasPaymentMethod: boolean } | null } };

export type BillingPortalUrlMutationVariables = Exact<{ [key: string]: never; }>;


export type BillingPortalUrlMutation = { billingPortalUrl: { url: string } };

export type InviteMemberMutationVariables = Exact<{
  input: InviteMemberInput;
}>;


export type InviteMemberMutation = { inviteMember: { id: string, email: string, name: string | null, role: WorkspaceRole } };

export type RemoveMemberMutationVariables = Exact<{
  userId: string | number;
}>;


export type RemoveMemberMutation = { removeMember: boolean };

export type MembersPlanQueryVariables = Exact<{ [key: string]: never; }>;


export type MembersPlanQuery = { subscription: { plan: Plan | null } | null };

export type UpdateMemberRoleMutationVariables = Exact<{
  input: UpdateMemberRoleInput;
}>;


export type UpdateMemberRoleMutation = { updateMemberRole: { id: string, email: string, name: string | null, role: WorkspaceRole } };

export type UpdateWorkspaceMutationVariables = Exact<{
  input: UpdateWorkspaceInput;
}>;


export type UpdateWorkspaceMutation = { updateWorkspace: { id: string, name: string } };

export type DeleteWorkspaceMutationVariables = Exact<{ [key: string]: never; }>;


export type DeleteWorkspaceMutation = { deleteWorkspace: boolean };

export type WorkspaceConnectorsQueryVariables = Exact<{ [key: string]: never; }>;


export type WorkspaceConnectorsQuery = { connectors: Array<{ id: string, name: string, url: string, status: CredentialStatus }> };

export type SetAgentConnectorsMutationVariables = Exact<{
  agent: string;
  connectors: Array<AgentConnectorInput> | AgentConnectorInput;
}>;


export type SetAgentConnectorsMutation = { setAgentConnectors: { id: string, connectors: Array<{ approval: Approval, hiddenTools: Array<string>, connector: { id: string, name: string } }> } };

export type SetAgentNetworkMutationVariables = Exact<{
  agent: string;
  allowedHosts: Array<string> | string;
  privateNetwork: boolean;
}>;


export type SetAgentNetworkMutation = { setAgentNetwork: { id: string, allowedHosts: Array<string>, privateNetwork: boolean } };

export type AgentAuditQueryVariables = Exact<{
  agent: string;
  kinds?: Array<AgentAuditKind> | AgentAuditKind | null | undefined;
}>;


export type AgentAuditQuery = { agentAudit: Array<{ time: string, kind: AgentAuditKind, modelProvider: string | null, connector: string | null, target: string, inputTokens: number | null, outputTokens: number | null, result: AgentAuditResult }> };

export type AgentSessionsQueryVariables = Exact<{
  agent: string;
}>;


export type AgentSessionsQuery = { agent: { id: string, sessions: Array<{ id: string, title: string, startedAt: string, status: AgentSessionStatus, effort: Effort, startedBy: { id: string, name: string | null, email: string } | null, model: { name: string, provider: { id: string, name: string } } }> } };

export type AgentModelsQueryVariables = Exact<{
  agent: string;
}>;


export type AgentModelsQuery = { agentModels: Array<{ name: string, provider: { id: string, name: string } }> };

export type AgentSessionQueryVariables = Exact<{
  session: string;
}>;


export type AgentSessionQuery = { agentSession: { id: string, events: Array<
      | { __typename: 'AgentApproval', id: string, time: string, approval: string, title: string, toolCallId: string | null, approvalStatus: AgentApprovalStatus, decidedBy: { id: string, name: string | null, email: string } | null }
      | { __typename: 'AgentMessage', id: string, time: string, text: string }
      | { __typename: 'AgentNotice', id: string, time: string, severity: AgentNoticeSeverity, title: string, description: string | null }
      | { __typename: 'AgentToolCall', id: string, time: string, callId: string, connector: string | null, tool: string, title: string, status: AgentToolCallStatus, input: string | null, output: string | null }
      | { __typename: 'AgentTurnEnd', id: string, time: string, reason: AgentTurnEndReason, error: string | null }
      | { __typename: 'AgentUserMessage', id: string, time: string, text: string, author: { id: string, name: string | null, email: string } | null }
    > } };

export type AgentSessionEventsSubscriptionVariables = Exact<{
  session: string;
  after?: string | number | null | undefined;
}>;


export type AgentSessionEventsSubscription = { agentSessionEvents:
    | { __typename: 'AgentApproval', id: string, time: string, approval: string, title: string, toolCallId: string | null, approvalStatus: AgentApprovalStatus, decidedBy: { id: string, name: string | null, email: string } | null }
    | { __typename: 'AgentMessage', id: string, time: string, text: string }
    | { __typename: 'AgentNotice', id: string, time: string, severity: AgentNoticeSeverity, title: string, description: string | null }
    | { __typename: 'AgentToolCall', id: string, time: string, callId: string, connector: string | null, tool: string, title: string, status: AgentToolCallStatus, input: string | null, output: string | null }
    | { __typename: 'AgentTurnEnd', id: string, time: string, reason: AgentTurnEndReason, error: string | null }
    | { __typename: 'AgentUserMessage', id: string, time: string, text: string, author: { id: string, name: string | null, email: string } | null }
   };

export type StartAgentSessionMutationVariables = Exact<{
  agent: string;
  message: string;
  model?: AgentModelInput | null | undefined;
  effort?: Effort | null | undefined;
}>;


export type StartAgentSessionMutation = { startAgentSession: { id: string, title: string, startedAt: string, status: AgentSessionStatus, effort: Effort, startedBy: { id: string, name: string | null, email: string } | null, model: { name: string, provider: { id: string, name: string } } } };

export type SendAgentMessageMutationVariables = Exact<{
  session: string;
  message: string;
}>;


export type SendAgentMessageMutation = { sendAgentMessage: { id: string, status: AgentSessionStatus } };

export type SetAgentSessionModelMutationVariables = Exact<{
  session: string;
  model: AgentModelInput;
  effort: Effort;
}>;


export type SetAgentSessionModelMutation = { setAgentSessionModel: { id: string, effort: Effort, model: { name: string, provider: { id: string, name: string } } } };

export type AnswerAgentApprovalMutationVariables = Exact<{
  session: string;
  approval: string | number;
  decision: AgentApprovalDecision;
}>;


export type AnswerAgentApprovalMutation = { answerAgentApproval: { id: string, status: AgentSessionStatus } };

export type StopAgentSessionMutationVariables = Exact<{
  session: string;
}>;


export type StopAgentSessionMutation = { stopAgentSession: { id: string, status: AgentSessionStatus } };

export type UpdateAgentMutationVariables = Exact<{
  agent: string;
  harness: Harness;
  instructions: string;
}>;


export type UpdateAgentMutation = { updateAgent: { id: string, harness: Harness, instructions: string, status: AgentStatus } };

export type DeleteAgentMutationVariables = Exact<{
  agent: string;
}>;


export type DeleteAgentMutation = { deleteAgent: boolean };

export type BucketCredentialsQueryVariables = Exact<{
  bucket: string;
}>;


export type BucketCredentialsQuery = { bucketCredentials: { endpoint: string, region: string, bucket: string, accessKeyId: string, secretAccessKey: string } };

export type SetBucketPublicMutationVariables = Exact<{
  bucket: string;
  public: boolean;
}>;


export type SetBucketPublicMutation = { setBucketPublic: { id: string, public: boolean, publicEndpoint: string } };

export type BucketObjectsQueryVariables = Exact<{
  bucket: string;
  prefix?: string | null | undefined;
}>;


export type BucketObjectsQuery = { bucketObjects: { prefix: string, folders: Array<{ prefix: string }>, objects: Array<{ key: string, size: number, lastModified: string }> } };

export type BucketObjectUploadUrlMutationVariables = Exact<{
  bucket: string;
  key: string;
}>;


export type BucketObjectUploadUrlMutation = { bucketObjectUploadUrl: string };

export type BucketObjectDownloadUrlQueryVariables = Exact<{
  bucket: string;
  key: string;
}>;


export type BucketObjectDownloadUrlQuery = { bucketObjectDownloadUrl: string };

export type DeleteBucketObjectMutationVariables = Exact<{
  bucket: string;
  key: string;
}>;


export type DeleteBucketObjectMutation = { deleteBucketObject: boolean };

export type DeleteBucketMutationVariables = Exact<{
  bucket: string;
}>;


export type DeleteBucketMutation = { deleteBucket: boolean };

export type BuildLogsStatusQueryVariables = Exact<{
  id: string;
}>;


export type BuildLogsStatusQuery = { build: { id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null } };

export type AddDatabaseAllowRuleMutationVariables = Exact<{
  database: string;
  rule: DatabaseAllowRuleInput;
}>;


export type AddDatabaseAllowRuleMutation = { addDatabaseAllowRule: { id: string, allowRules: Array<{ range: string, description: string }> } };

export type RemoveDatabaseAllowRuleMutationVariables = Exact<{
  database: string;
  range: string;
}>;


export type RemoveDatabaseAllowRuleMutation = { removeDatabaseAllowRule: { id: string, allowRules: Array<{ range: string, description: string }> } };

export type DatabaseBackupsQueryVariables = Exact<{
  database: string;
}>;


export type DatabaseBackupsQuery = { database: { id: string, name: string, backups: { enabled: boolean, retentionDays: number, earliestRestorePoint: string | null, latestRestorePoint: string | null, lastBackupAt: string | null, backups: Array<{ id: string, status: BackupStatus, trigger: BackupTrigger, createdAt: string, startedAt: string | null, finishedAt: string | null, error: string | null }> } } };

export type CreateDatabaseBackupMutationVariables = Exact<{
  database: string;
}>;


export type CreateDatabaseBackupMutation = { createDatabaseBackup: { id: string, status: BackupStatus } };

export type RestoreDatabaseMutationVariables = Exact<{
  input: RestoreDatabaseInput;
}>;


export type RestoreDatabaseMutation = { restoreDatabase: { clampedToLatest: boolean, database: { id: string, name: string, status: DatabaseStatus } } };

export type DatabasePublicQueryVariables = Exact<{
  database: string;
}>;


export type DatabasePublicQuery = { clientAddress: string | null, database: { id: string, public: boolean, allowRules: Array<{ range: string, description: string }> } };

export type DatabaseCredentialsQueryVariables = Exact<{
  database: string;
}>;


export type DatabaseCredentialsQuery = { databaseCredentials: Array<{ type: EndpointType, host: string, port: string, dbname: string, user: string, password: string, uri: string }> };

export type ExposeDatabaseMutationVariables = Exact<{
  database: string;
}>;


export type ExposeDatabaseMutation = { exposeDatabase: { id: string, public: boolean } };

export type AllowOnExposeMutationVariables = Exact<{
  database: string;
  rule: DatabaseAllowRuleInput;
}>;


export type AllowOnExposeMutation = { addDatabaseAllowRule: { id: string, allowRules: Array<{ range: string, description: string }> } };

export type UnexposeDatabaseMutationVariables = Exact<{
  database: string;
}>;


export type UnexposeDatabaseMutation = { unexposeDatabase: { id: string, public: boolean } };

export type DatabaseMetricsQueryVariables = Exact<{
  id: string;
  range: MetricsRange;
  grouping: MetricGrouping;
}>;


export type DatabaseMetricsQuery = { database: { id: string, instances: number, size: string, connectionLimit: number | null, resources: { cpu: string, memory: string }, metrics: Array<{ metric: DatabaseMetric, instance: string | null, role: DatabaseInstanceRole | null, points: Array<{ timestamp: string, value: number | null }> }> } };

export type DatabaseQueryStatisticsQueryVariables = Exact<{
  id: string;
}>;


export type DatabaseQueryStatisticsQuery = { database: { id: string, queryStatistics: { enabled: boolean, since: string | null, queries: Array<{ query: string, calls: number, rows: number, totalMilliseconds: number, meanMilliseconds: number, share: number }> } } };

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

export type ServiceSecurityQueryVariables = Exact<{
  id: string;
}>;


export type ServiceSecurityQuery = { service: { id: string, sourceUrl: string, vulnerabilityReport: { summary: { critical: number, high: number, medium: number, low: number, unknown: number, total: number } } | null, secretScanReport: { commit: string, scannedAt: string, findings: Array<{ rule: string, file: string, line: number, commit: string, secret: string, author: string | null, url: string | null, verified: boolean }> } | null } };

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


export type GenerateDomainMutation = { generateDomain: { id: string, endpoints: Array<{ host: string, port: number, type: EndpointType, protocol: Protocol, redirectTo: string | null, dns: { status: DnsStatus, requiredRecords: Array<{ type: DnsRecordType, host: string, value: string }> } }> } };

export type AddCustomDomainMutationVariables = Exact<{
  service: string;
  hostname: string;
  redirectTo?: string | null | undefined;
}>;


export type AddCustomDomainMutation = { addCustomDomain: { id: string, endpoints: Array<{ host: string, port: number, type: EndpointType, protocol: Protocol, redirectTo: string | null, dns: { status: DnsStatus, requiredRecords: Array<{ type: DnsRecordType, host: string, value: string }> } }> } };

export type RemoveDomainMutationVariables = Exact<{
  service: string;
  hostname: string;
}>;


export type RemoveDomainMutation = { removeDomain: { id: string, endpoints: Array<{ host: string, port: number, type: EndpointType, protocol: Protocol, redirectTo: string | null, dns: { status: DnsStatus, requiredRecords: Array<{ type: DnsRecordType, host: string, value: string }> } }> } };

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

export type SetServiceUserMutationVariables = Exact<{
  service: string;
  user?: number | null | undefined;
}>;


export type SetServiceUserMutation = { setServiceUser: { id: string, user: number | null } };

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

export type ServiceVulnerabilitiesQueryVariables = Exact<{
  id: string;
}>;


export type ServiceVulnerabilitiesQuery = { service: { id: string, vulnerabilityReport: { image: string, summary: { critical: number, high: number, medium: number, low: number, unknown: number, total: number }, vulnerabilities: Array<{ id: string, severity: VulnerabilitySeverity, source: VulnerabilitySource, title: string | null, reference: string | null, packages: Array<{ name: string, installedVersion: string, fixedVersion: string | null, path: string | null }> }> } | null } };

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

export type CapabilitiesQueryVariables = Exact<{ [key: string]: never; }>;


export type CapabilitiesQuery = { workspace: { id: string, capabilities: { services: { availability: Availability, plan: Plan | null, replicas: number, cpu: { min: string, max: string, upgrades: Array<{ plan: Plan, max: string }> }, memory: { min: string, max: string, upgrades: Array<{ plan: Plan, max: string }> } }, databases: { availability: Availability, plan: Plan | null, cpu: { min: string, max: string, upgrades: Array<{ plan: Plan, max: string }> }, memory: { min: string, max: string, upgrades: Array<{ plan: Plan, max: string }> }, storage: { min: string, max: string, upgrades: Array<{ plan: Plan, max: string }> } }, keyValueStores: { availability: Availability, plan: Plan | null, size: { min: string, max: string, upgrades: Array<{ plan: Plan, max: string }> } }, volumes: { availability: Availability, plan: Plan | null, size: { min: string, max: string, upgrades: Array<{ plan: Plan, max: string }> } }, buckets: { availability: Availability, plan: Plan | null }, agents: { availability: Availability, plan: Plan | null }, members: { availability: Availability, plan: Plan | null, limit: number | null, upgrades: Array<{ plan: Plan, limit: number | null }> }, customDomains: { availability: Availability, plan: Plan | null, limit: number | null, upgrades: Array<{ plan: Plan, limit: number | null }> } } } };

export type DeployMutationVariables = Exact<{
  service: string;
  gitRef?: string | null | undefined;
}>;


export type DeployMutation = { deploy: { id: string, status: ReleaseStatus, build: { id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null } | null } };

export type BuildStatusQueryVariables = Exact<{
  id: string;
}>;


export type BuildStatusQuery = { build: { id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null } };

export type BuildLogsSubscriptionVariables = Exact<{
  id: string;
}>;


export type BuildLogsSubscription = { buildLogs: string };

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

export type CompletePaymentMethodCheckoutMutationVariables = Exact<{
  sessionId: string;
}>;


export type CompletePaymentMethodCheckoutMutation = { completePaymentMethodCheckout: { id: string } };

export type CompleteWorkspaceCheckoutMutationVariables = Exact<{
  sessionId: string;
}>;


export type CompleteWorkspaceCheckoutMutation = { completeWorkspaceCheckout: { id: string, name: string, personal: boolean } };

export type EnvironmentQueryVariables = Exact<{
  environment: string;
}>;


export type EnvironmentQuery = { environment: { id: string, name: string, resourceTier: ResourceTier, services: Array<{ id: string, name: string, status: ServiceStatus, port: number, sourceUrl: string, branch: string | null, autoDeploy: boolean, ciDeploy: boolean, contextPath: string, command: string, user: number | null, lastDeployedAt: string | null, createdAt: string, replicas: { desired: number, ready: number }, autoscaling: { minReplicas: number, maxReplicas: number, targetCpu: number } | null, endpoints: Array<{ host: string, port: number, protocol: Protocol, type: EndpointType, redirectTo: string | null, tls: TlsStatus, dns: { status: DnsStatus, requiredRecords: Array<{ type: DnsRecordType, host: string, value: string }> } }>, resources: { cpu: string, memory: string }, healthCheck: { path: string, port: number, initialDelaySeconds: number, periodSeconds: number, timeoutSeconds: number, failureThreshold: number, startupFailureThreshold: number } | null, activeDeployment: { id: string, image: string, imageDigest: string | null, commit: string, commitMessage: string, ref: string, status: DeploymentStatus, createdAt: string, replicas: { desired: number, ready: number }, rollout: { status: RolloutStatus, reason: RolloutReason | null, message: string | null, restarts: number, startedAt: string } | null } | null, deployments: Array<{ id: string, image: string, imageDigest: string | null, commit: string, commitMessage: string, ref: string, status: DeploymentStatus, createdAt: string }>, builds: Array<{ id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null }>, releases: Array<{ id: string, status: ReleaseStatus, createdAt: string, trigger: { kind: ReleaseTriggerKind, actor: string | null }, source: { provider: SourceProvider, repository: string, url: string, ref: string, contextPath: string, commit: { sha: string, message: string, url: string | null } } | null, build: { id: string, status: BuildStatus, startedAt: string | null, finishedAt: string | null } | null, deploy: { id: string, status: DeployStatus, startedAt: string | null, finishedAt: string | null } | null, scan: { id: string, status: ScanStatus, findingsCount: number | null, verifiedCount: number | null, startedAt: string | null, finishedAt: string | null } | null, deployment: { id: string, image: string, imageDigest: string | null, commit: string, commitMessage: string, ref: string, status: DeploymentStatus, createdAt: string, replicas: { desired: number, ready: number }, rollout: { status: RolloutStatus, reason: RolloutReason | null, message: string | null, restarts: number, startedAt: string } | null } | null }> }>, databases: Array<{ id: string, name: string, version: string, instances: number, status: DatabaseStatus, size: string, createdAt: string }>, keyValueStores: Array<{ id: string, name: string, version: string, status: DatabaseStatus, size: string, createdAt: string }>, buckets: Array<{ id: string, name: string, region: string, endpoint: string, publicEndpoint: string, status: BucketStatus, sizeBytes: number, objectCount: number, public: boolean, createdAt: string }>, volumes: Array<{ id: string, name: string, size: string, mount: { service: string, path: string } | null }> } };

export type EnvironmentVolumeUsageQueryVariables = Exact<{
  environment: string;
}>;


export type EnvironmentVolumeUsageQuery = { environment: { id: string, volumes: Array<{ id: string, size: string, metrics: Array<{ points: Array<{ value: number | null }> }> }> } };

export type EnvironmentDefaultCommandsQueryVariables = Exact<{
  environment: string;
}>;


export type EnvironmentDefaultCommandsQuery = { environment: { id: string, services: Array<{ id: string, defaultCommand: string }> } };

export type EnvironmentAgentsQueryVariables = Exact<{
  environment: string;
}>;


export type EnvironmentAgentsQuery = { environment: { id: string, agents: Array<{ id: string, name: string, harness: Harness, instructions: string, status: AgentStatus, allowedHosts: Array<string>, privateNetwork: boolean, connectors: Array<{ approval: Approval, hiddenTools: Array<string>, connector: { id: string, name: string } }> }> } };

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


export type ProjectPageQuery = { project: { id: string, name: string, environments: Array<{ id: string, name: string, resourceTier: ResourceTier, services: Array<{ id: string, name: string, status: ServiceStatus, sourceUrl: string }>, databases: Array<{ id: string, name: string, status: DatabaseStatus }>, keyValueStores: Array<{ id: string, name: string, status: DatabaseStatus }>, buckets: Array<{ id: string, name: string, status: BucketStatus }>, volumes: Array<{ id: string, name: string }> }> } };

export type ProjectsQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectsQuery = { projects: Array<{ id: string, name: string, environments: Array<{ id: string, name: string, resourceTier: ResourceTier, services: Array<{ id: string, name: string, sourceUrl: string }>, databases: Array<{ id: string, name: string }>, keyValueStores: Array<{ id: string, name: string }>, buckets: Array<{ id: string, name: string }>, volumes: Array<{ id: string, name: string }> }> }> };


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
export const AgentPaletteConnectorsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AgentPaletteConnectors"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"connectors"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<AgentPaletteConnectorsQuery, AgentPaletteConnectorsQueryVariables>;
export const CreateAgentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateAgent"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"name"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"harness"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Harness"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"instructions"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"connectors"}},"type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentConnectorInput"}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createAgent"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"name"},"value":{"kind":"Variable","name":{"kind":"Name","value":"name"}}},{"kind":"Argument","name":{"kind":"Name","value":"harness"},"value":{"kind":"Variable","name":{"kind":"Name","value":"harness"}}},{"kind":"Argument","name":{"kind":"Name","value":"instructions"},"value":{"kind":"Variable","name":{"kind":"Name","value":"instructions"}}},{"kind":"Argument","name":{"kind":"Name","value":"connectors"},"value":{"kind":"Variable","name":{"kind":"Name","value":"connectors"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]} as unknown as DocumentNode<CreateAgentMutation, CreateAgentMutationVariables>;
export const CreateEnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateEnvironment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateEnvironmentInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createEnvironment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]} as unknown as DocumentNode<CreateEnvironmentMutation, CreateEnvironmentMutationVariables>;
export const CreateWorkspaceCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateWorkspaceCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateWorkspaceCheckoutInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createWorkspaceCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<CreateWorkspaceCheckoutMutation, CreateWorkspaceCheckoutMutationVariables>;
export const GatePlansDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GatePlans"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plans"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"priceCents"}},{"kind":"Field","name":{"kind":"Name","value":"currency"}},{"kind":"Field","name":{"kind":"Name","value":"limits"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"}},{"kind":"Field","name":{"kind":"Name","value":"volumes"}},{"kind":"Field","name":{"kind":"Name","value":"volumeSize"}},{"kind":"Field","name":{"kind":"Name","value":"databaseStorage"}},{"kind":"Field","name":{"kind":"Name","value":"members"}},{"kind":"Field","name":{"kind":"Name","value":"customDomains"}}]}}]}}]}}]} as unknown as DocumentNode<GatePlansQuery, GatePlansQueryVariables>;
export const GateSubscriptionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GateSubscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<GateSubscriptionQuery, GateSubscriptionQueryVariables>;
export const GateMembersDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GateMembers"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"members"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]}}]} as unknown as DocumentNode<GateMembersQuery, GateMembersQueryVariables>;
export const GateSelectPlanDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"GateSelectPlan"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"plan"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Plan"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"selectPlan"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"plan"},"value":{"kind":"Variable","name":{"kind":"Name","value":"plan"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"checkoutUrl"}},{"kind":"Field","name":{"kind":"Name","value":"subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]}}]} as unknown as DocumentNode<GateSelectPlanMutation, GateSelectPlanMutationVariables>;
export const AddPaymentMethodDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddPaymentMethod"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"returnPath"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addPaymentMethod"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"returnPath"},"value":{"kind":"Variable","name":{"kind":"Name","value":"returnPath"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<AddPaymentMethodMutation, AddPaymentMethodMutationVariables>;
export const MountVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"MountVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"volume"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"path"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"mountVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"volume"},"value":{"kind":"Variable","name":{"kind":"Name","value":"volume"}}},{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"path"},"value":{"kind":"Variable","name":{"kind":"Name","value":"path"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"mount"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"}},{"kind":"Field","name":{"kind":"Name","value":"path"}}]}}]}}]}}]} as unknown as DocumentNode<MountVolumeMutation, MountVolumeMutationVariables>;
export const CreateModelProviderDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateModelProvider"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"name"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"api"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ModelApi"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"url"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"apiKey"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createModelProvider"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"name"},"value":{"kind":"Variable","name":{"kind":"Name","value":"name"}}},{"kind":"Argument","name":{"kind":"Name","value":"api"},"value":{"kind":"Variable","name":{"kind":"Name","value":"api"}}},{"kind":"Argument","name":{"kind":"Name","value":"url"},"value":{"kind":"Variable","name":{"kind":"Name","value":"url"}}},{"kind":"Argument","name":{"kind":"Name","value":"apiKey"},"value":{"kind":"Variable","name":{"kind":"Name","value":"apiKey"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"api"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"keyHint"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"checkedAt"}},{"kind":"Field","name":{"kind":"Name","value":"models"}}]}}]}}]} as unknown as DocumentNode<CreateModelProviderMutation, CreateModelProviderMutationVariables>;
export const CreateConnectorDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateConnector"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"name"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"url"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"credential"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createConnector"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"name"},"value":{"kind":"Variable","name":{"kind":"Name","value":"name"}}},{"kind":"Argument","name":{"kind":"Name","value":"url"},"value":{"kind":"Variable","name":{"kind":"Name","value":"url"}}},{"kind":"Argument","name":{"kind":"Name","value":"credential"},"value":{"kind":"Variable","name":{"kind":"Name","value":"credential"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"credentialHint"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"checkedAt"}},{"kind":"Field","name":{"kind":"Name","value":"tools"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}}]}}]}}]} as unknown as DocumentNode<CreateConnectorMutation, CreateConnectorMutationVariables>;
export const EjectProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EjectProject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"ejectProject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"filename"}},{"kind":"Field","name":{"kind":"Name","value":"contentType"}},{"kind":"Field","name":{"kind":"Name","value":"content"}}]}}]}}]} as unknown as DocumentNode<EjectProjectQuery, EjectProjectQueryVariables>;
export const SubscriptionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditExpiry"}},{"kind":"Field","name":{"kind":"Name","value":"hasPaymentMethod"}}]}}]}}]} as unknown as DocumentNode<SubscriptionQuery, SubscriptionQueryVariables>;
export const UsageSummaryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"UsageSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"usageSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"resourceCostCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditsCents"}},{"kind":"Field","name":{"kind":"Name","value":"estimatedTotalCents"}}]}}]}}]} as unknown as DocumentNode<UsageSummaryQuery, UsageSummaryQueryVariables>;
export const WelcomeSelectPlanDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"WelcomeSelectPlan"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"plan"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Plan"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"selectPlan"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"plan"},"value":{"kind":"Variable","name":{"kind":"Name","value":"plan"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"checkoutUrl"}}]}}]}}]} as unknown as DocumentNode<WelcomeSelectPlanMutation, WelcomeSelectPlanMutationVariables>;
export const ApiTokensDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ApiTokens"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"apiTokens"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}}]}}]} as unknown as DocumentNode<ApiTokensQuery, ApiTokensQueryVariables>;
export const CreateApiTokenDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateApiToken"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateApiTokenInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createApiToken"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"apiToken"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"token"}}]}}]}}]} as unknown as DocumentNode<CreateApiTokenMutation, CreateApiTokenMutationVariables>;
export const RevokeApiTokenDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RevokeApiToken"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"revokeApiToken"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<RevokeApiTokenMutation, RevokeApiTokenMutationVariables>;
export const ConnectionsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Connections"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"modelProviders"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"api"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"models"}}]}},{"kind":"Field","name":{"kind":"Name","value":"connectors"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"tools"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]} as unknown as DocumentNode<ConnectionsQuery, ConnectionsQueryVariables>;
export const ConnectorDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Connector"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ConnectorID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"connector"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"credentialHint"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"checkedAt"}},{"kind":"Field","name":{"kind":"Name","value":"tools"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}},{"kind":"Field","name":{"kind":"Name","value":"agents"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"harness"}}]}}]}}]}}]} as unknown as DocumentNode<ConnectorQuery, ConnectorQueryVariables>;
export const ReplaceConnectorCredentialDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ReplaceConnectorCredential"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"connector"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ConnectorID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"credential"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"replaceConnectorCredential"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"connector"},"value":{"kind":"Variable","name":{"kind":"Name","value":"connector"}}},{"kind":"Argument","name":{"kind":"Name","value":"credential"},"value":{"kind":"Variable","name":{"kind":"Name","value":"credential"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"credentialHint"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"checkedAt"}},{"kind":"Field","name":{"kind":"Name","value":"tools"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}}]}}]}}]} as unknown as DocumentNode<ReplaceConnectorCredentialMutation, ReplaceConnectorCredentialMutationVariables>;
export const RefreshConnectorDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RefreshConnector"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"connector"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ConnectorID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"refreshConnector"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"connector"},"value":{"kind":"Variable","name":{"kind":"Name","value":"connector"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"checkedAt"}},{"kind":"Field","name":{"kind":"Name","value":"tools"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}}]}}]}}]} as unknown as DocumentNode<RefreshConnectorMutation, RefreshConnectorMutationVariables>;
export const DeleteConnectorDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteConnector"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"connector"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ConnectorID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteConnector"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"connector"},"value":{"kind":"Variable","name":{"kind":"Name","value":"connector"}}}]}]}}]} as unknown as DocumentNode<DeleteConnectorMutation, DeleteConnectorMutationVariables>;
export const SetEnvironmentResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetEnvironmentResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"SetEnvironmentResourcesInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setEnvironmentResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]} as unknown as DocumentNode<SetEnvironmentResourcesMutation, SetEnvironmentResourcesMutationVariables>;
export const DeleteEnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteEnvironment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteEnvironment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}]}]}}]} as unknown as DocumentNode<DeleteEnvironmentMutation, DeleteEnvironmentMutationVariables>;
export const ModelProviderDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ModelProvider"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ModelProviderID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"modelProvider"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"api"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"keyHint"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"checkedAt"}},{"kind":"Field","name":{"kind":"Name","value":"models"}},{"kind":"Field","name":{"kind":"Name","value":"agents"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"harness"}}]}}]}}]}}]} as unknown as DocumentNode<ModelProviderQuery, ModelProviderQueryVariables>;
export const ReplaceModelProviderKeyDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ReplaceModelProviderKey"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"provider"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ModelProviderID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"apiKey"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"replaceModelProviderKey"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"provider"},"value":{"kind":"Variable","name":{"kind":"Name","value":"provider"}}},{"kind":"Argument","name":{"kind":"Name","value":"apiKey"},"value":{"kind":"Variable","name":{"kind":"Name","value":"apiKey"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"keyHint"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"checkedAt"}},{"kind":"Field","name":{"kind":"Name","value":"models"}}]}}]}}]} as unknown as DocumentNode<ReplaceModelProviderKeyMutation, ReplaceModelProviderKeyMutationVariables>;
export const RefreshModelProviderDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RefreshModelProvider"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"provider"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ModelProviderID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"refreshModelProvider"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"provider"},"value":{"kind":"Variable","name":{"kind":"Name","value":"provider"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"checkedAt"}},{"kind":"Field","name":{"kind":"Name","value":"models"}}]}}]}}]} as unknown as DocumentNode<RefreshModelProviderMutation, RefreshModelProviderMutationVariables>;
export const DeleteModelProviderDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteModelProvider"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"provider"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ModelProviderID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteModelProvider"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"provider"},"value":{"kind":"Variable","name":{"kind":"Name","value":"provider"}}}]}]}}]} as unknown as DocumentNode<DeleteModelProviderMutation, DeleteModelProviderMutationVariables>;
export const DeleteProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteProject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteProject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<DeleteProjectMutation, DeleteProjectMutationVariables>;
export const SharedVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SharedVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"sharedVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]} as unknown as DocumentNode<SharedVariablesQuery, SharedVariablesQueryVariables>;
export const SetSharedVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetSharedVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"variables"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VariableInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setSharedVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"variables"},"value":{"kind":"Variable","name":{"kind":"Name","value":"variables"}}}]}]}}]} as unknown as DocumentNode<SetSharedVariablesMutation, SetSharedVariablesMutationVariables>;
export const SelectPlanDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SelectPlan"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"plan"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Plan"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"selectPlan"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"plan"},"value":{"kind":"Variable","name":{"kind":"Name","value":"plan"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"checkoutUrl"}},{"kind":"Field","name":{"kind":"Name","value":"subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditExpiry"}},{"kind":"Field","name":{"kind":"Name","value":"hasPaymentMethod"}}]}}]}}]}}]} as unknown as DocumentNode<SelectPlanMutation, SelectPlanMutationVariables>;
export const BillingPortalUrlDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"BillingPortalUrl"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"billingPortalUrl"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<BillingPortalUrlMutation, BillingPortalUrlMutationVariables>;
export const InviteMemberDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"InviteMember"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"InviteMemberInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"inviteMember"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]} as unknown as DocumentNode<InviteMemberMutation, InviteMemberMutationVariables>;
export const RemoveMemberDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveMember"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"userId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeMember"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"userId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"userId"}}}]}]}}]} as unknown as DocumentNode<RemoveMemberMutation, RemoveMemberMutationVariables>;
export const MembersPlanDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"MembersPlan"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}}]}}]}}]} as unknown as DocumentNode<MembersPlanQuery, MembersPlanQueryVariables>;
export const UpdateMemberRoleDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateMemberRole"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateMemberRoleInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateMemberRole"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]} as unknown as DocumentNode<UpdateMemberRoleMutation, UpdateMemberRoleMutationVariables>;
export const UpdateWorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateWorkspace"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateWorkspaceInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateWorkspace"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]} as unknown as DocumentNode<UpdateWorkspaceMutation, UpdateWorkspaceMutationVariables>;
export const DeleteWorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteWorkspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteWorkspace"}}]}}]} as unknown as DocumentNode<DeleteWorkspaceMutation, DeleteWorkspaceMutationVariables>;
export const WorkspaceConnectorsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"WorkspaceConnectors"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"connectors"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<WorkspaceConnectorsQuery, WorkspaceConnectorsQueryVariables>;
export const SetAgentConnectorsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetAgentConnectors"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"agent"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"connectors"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentConnectorInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setAgentConnectors"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"agent"},"value":{"kind":"Variable","name":{"kind":"Name","value":"agent"}}},{"kind":"Argument","name":{"kind":"Name","value":"connectors"},"value":{"kind":"Variable","name":{"kind":"Name","value":"connectors"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"connectors"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"approval"}},{"kind":"Field","name":{"kind":"Name","value":"hiddenTools"}},{"kind":"Field","name":{"kind":"Name","value":"connector"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]} as unknown as DocumentNode<SetAgentConnectorsMutation, SetAgentConnectorsMutationVariables>;
export const SetAgentNetworkDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetAgentNetwork"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"agent"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"allowedHosts"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"privateNetwork"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setAgentNetwork"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"agent"},"value":{"kind":"Variable","name":{"kind":"Name","value":"agent"}}},{"kind":"Argument","name":{"kind":"Name","value":"allowedHosts"},"value":{"kind":"Variable","name":{"kind":"Name","value":"allowedHosts"}}},{"kind":"Argument","name":{"kind":"Name","value":"privateNetwork"},"value":{"kind":"Variable","name":{"kind":"Name","value":"privateNetwork"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"allowedHosts"}},{"kind":"Field","name":{"kind":"Name","value":"privateNetwork"}}]}}]}}]} as unknown as DocumentNode<SetAgentNetworkMutation, SetAgentNetworkMutationVariables>;
export const AgentAuditDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AgentAudit"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"agent"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"kinds"}},"type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentAuditKind"}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"agentAudit"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"agent"},"value":{"kind":"Variable","name":{"kind":"Name","value":"agent"}}},{"kind":"Argument","name":{"kind":"Name","value":"kinds"},"value":{"kind":"Variable","name":{"kind":"Name","value":"kinds"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"kind"}},{"kind":"Field","name":{"kind":"Name","value":"modelProvider"}},{"kind":"Field","name":{"kind":"Name","value":"connector"}},{"kind":"Field","name":{"kind":"Name","value":"target"}},{"kind":"Field","name":{"kind":"Name","value":"inputTokens"}},{"kind":"Field","name":{"kind":"Name","value":"outputTokens"}},{"kind":"Field","name":{"kind":"Name","value":"result"}}]}}]}}]} as unknown as DocumentNode<AgentAuditQuery, AgentAuditQueryVariables>;
export const AgentSessionsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AgentSessions"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"agent"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"agent"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"agent"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"sessions"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"effort"}},{"kind":"Field","name":{"kind":"Name","value":"startedBy"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}}]}},{"kind":"Field","name":{"kind":"Name","value":"model"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"provider"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<AgentSessionsQuery, AgentSessionsQueryVariables>;
export const AgentModelsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AgentModels"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"agent"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"agentModels"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"agent"},"value":{"kind":"Variable","name":{"kind":"Name","value":"agent"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"provider"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]} as unknown as DocumentNode<AgentModelsQuery, AgentModelsQueryVariables>;
export const AgentSessionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AgentSession"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"session"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentSessionID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"agentSession"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"session"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"events"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentUserMessage"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"text"}},{"kind":"Field","name":{"kind":"Name","value":"author"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}}]}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentMessage"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"text"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentToolCall"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"callId"}},{"kind":"Field","name":{"kind":"Name","value":"connector"}},{"kind":"Field","name":{"kind":"Name","value":"tool"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"input"}},{"kind":"Field","name":{"kind":"Name","value":"output"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentApproval"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"approval"}},{"kind":"Field","alias":{"kind":"Name","value":"toolCallId"},"name":{"kind":"Name","value":"callId"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","alias":{"kind":"Name","value":"approvalStatus"},"name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"decidedBy"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}}]}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentNotice"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"severity"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentTurnEnd"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"reason"}},{"kind":"Field","name":{"kind":"Name","value":"error"}}]}}]}}]}}]}}]} as unknown as DocumentNode<AgentSessionQuery, AgentSessionQueryVariables>;
export const AgentSessionEventsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"AgentSessionEvents"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"session"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentSessionID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"after"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"agentSessionEvents"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"session"},"value":{"kind":"Variable","name":{"kind":"Name","value":"session"}}},{"kind":"Argument","name":{"kind":"Name","value":"after"},"value":{"kind":"Variable","name":{"kind":"Name","value":"after"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentUserMessage"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"text"}},{"kind":"Field","name":{"kind":"Name","value":"author"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}}]}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentMessage"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"text"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentToolCall"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"callId"}},{"kind":"Field","name":{"kind":"Name","value":"connector"}},{"kind":"Field","name":{"kind":"Name","value":"tool"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"input"}},{"kind":"Field","name":{"kind":"Name","value":"output"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentApproval"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"approval"}},{"kind":"Field","alias":{"kind":"Name","value":"toolCallId"},"name":{"kind":"Name","value":"callId"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","alias":{"kind":"Name","value":"approvalStatus"},"name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"decidedBy"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}}]}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentNotice"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"severity"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"AgentTurnEnd"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"time"}},{"kind":"Field","name":{"kind":"Name","value":"reason"}},{"kind":"Field","name":{"kind":"Name","value":"error"}}]}}]}}]}}]} as unknown as DocumentNode<AgentSessionEventsSubscription, AgentSessionEventsSubscriptionVariables>;
export const StartAgentSessionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"StartAgentSession"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"agent"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"message"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"model"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentModelInput"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"effort"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Effort"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"startAgentSession"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"agent"},"value":{"kind":"Variable","name":{"kind":"Name","value":"agent"}}},{"kind":"Argument","name":{"kind":"Name","value":"message"},"value":{"kind":"Variable","name":{"kind":"Name","value":"message"}}},{"kind":"Argument","name":{"kind":"Name","value":"model"},"value":{"kind":"Variable","name":{"kind":"Name","value":"model"}}},{"kind":"Argument","name":{"kind":"Name","value":"effort"},"value":{"kind":"Variable","name":{"kind":"Name","value":"effort"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"effort"}},{"kind":"Field","name":{"kind":"Name","value":"startedBy"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}}]}},{"kind":"Field","name":{"kind":"Name","value":"model"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"provider"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]} as unknown as DocumentNode<StartAgentSessionMutation, StartAgentSessionMutationVariables>;
export const SendAgentMessageDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SendAgentMessage"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"session"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentSessionID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"message"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"sendAgentMessage"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"session"},"value":{"kind":"Variable","name":{"kind":"Name","value":"session"}}},{"kind":"Argument","name":{"kind":"Name","value":"message"},"value":{"kind":"Variable","name":{"kind":"Name","value":"message"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<SendAgentMessageMutation, SendAgentMessageMutationVariables>;
export const SetAgentSessionModelDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetAgentSessionModel"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"session"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentSessionID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"model"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentModelInput"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"effort"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Effort"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setAgentSessionModel"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"session"},"value":{"kind":"Variable","name":{"kind":"Name","value":"session"}}},{"kind":"Argument","name":{"kind":"Name","value":"model"},"value":{"kind":"Variable","name":{"kind":"Name","value":"model"}}},{"kind":"Argument","name":{"kind":"Name","value":"effort"},"value":{"kind":"Variable","name":{"kind":"Name","value":"effort"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"effort"}},{"kind":"Field","name":{"kind":"Name","value":"model"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"provider"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]} as unknown as DocumentNode<SetAgentSessionModelMutation, SetAgentSessionModelMutationVariables>;
export const AnswerAgentApprovalDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AnswerAgentApproval"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"session"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentSessionID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"approval"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"decision"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentApprovalDecision"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"answerAgentApproval"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"session"},"value":{"kind":"Variable","name":{"kind":"Name","value":"session"}}},{"kind":"Argument","name":{"kind":"Name","value":"approval"},"value":{"kind":"Variable","name":{"kind":"Name","value":"approval"}}},{"kind":"Argument","name":{"kind":"Name","value":"decision"},"value":{"kind":"Variable","name":{"kind":"Name","value":"decision"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<AnswerAgentApprovalMutation, AnswerAgentApprovalMutationVariables>;
export const StopAgentSessionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"StopAgentSession"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"session"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentSessionID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"stopAgentSession"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"session"},"value":{"kind":"Variable","name":{"kind":"Name","value":"session"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<StopAgentSessionMutation, StopAgentSessionMutationVariables>;
export const UpdateAgentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateAgent"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"agent"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"harness"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Harness"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"instructions"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateAgent"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"agent"},"value":{"kind":"Variable","name":{"kind":"Name","value":"agent"}}},{"kind":"Argument","name":{"kind":"Name","value":"harness"},"value":{"kind":"Variable","name":{"kind":"Name","value":"harness"}}},{"kind":"Argument","name":{"kind":"Name","value":"instructions"},"value":{"kind":"Variable","name":{"kind":"Name","value":"instructions"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"harness"}},{"kind":"Field","name":{"kind":"Name","value":"instructions"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<UpdateAgentMutation, UpdateAgentMutationVariables>;
export const DeleteAgentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteAgent"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"agent"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AgentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteAgent"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"agent"},"value":{"kind":"Variable","name":{"kind":"Name","value":"agent"}}}]}]}}]} as unknown as DocumentNode<DeleteAgentMutation, DeleteAgentMutationVariables>;
export const BucketCredentialsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BucketCredentials"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"bucketCredentials"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"endpoint"}},{"kind":"Field","name":{"kind":"Name","value":"region"}},{"kind":"Field","name":{"kind":"Name","value":"bucket"}},{"kind":"Field","name":{"kind":"Name","value":"accessKeyId"}},{"kind":"Field","name":{"kind":"Name","value":"secretAccessKey"}}]}}]}}]} as unknown as DocumentNode<BucketCredentialsQuery, BucketCredentialsQueryVariables>;
export const SetBucketPublicDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetBucketPublic"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"public"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setBucketPublic"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}},{"kind":"Argument","name":{"kind":"Name","value":"public"},"value":{"kind":"Variable","name":{"kind":"Name","value":"public"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"public"}},{"kind":"Field","name":{"kind":"Name","value":"publicEndpoint"}}]}}]}}]} as unknown as DocumentNode<SetBucketPublicMutation, SetBucketPublicMutationVariables>;
export const BucketObjectsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BucketObjects"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"prefix"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"bucketObjects"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}},{"kind":"Argument","name":{"kind":"Name","value":"prefix"},"value":{"kind":"Variable","name":{"kind":"Name","value":"prefix"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"prefix"}},{"kind":"Field","name":{"kind":"Name","value":"folders"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"prefix"}}]}},{"kind":"Field","name":{"kind":"Name","value":"objects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"lastModified"}}]}}]}}]}}]} as unknown as DocumentNode<BucketObjectsQuery, BucketObjectsQueryVariables>;
export const BucketObjectUploadUrlDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"BucketObjectUploadUrl"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"key"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"bucketObjectUploadUrl"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}},{"kind":"Argument","name":{"kind":"Name","value":"key"},"value":{"kind":"Variable","name":{"kind":"Name","value":"key"}}}]}]}}]} as unknown as DocumentNode<BucketObjectUploadUrlMutation, BucketObjectUploadUrlMutationVariables>;
export const BucketObjectDownloadUrlDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BucketObjectDownloadUrl"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"key"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"bucketObjectDownloadUrl"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}},{"kind":"Argument","name":{"kind":"Name","value":"key"},"value":{"kind":"Variable","name":{"kind":"Name","value":"key"}}}]}]}}]} as unknown as DocumentNode<BucketObjectDownloadUrlQuery, BucketObjectDownloadUrlQueryVariables>;
export const DeleteBucketObjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteBucketObject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"key"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteBucketObject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}},{"kind":"Argument","name":{"kind":"Name","value":"key"},"value":{"kind":"Variable","name":{"kind":"Name","value":"key"}}}]}]}}]} as unknown as DocumentNode<DeleteBucketObjectMutation, DeleteBucketObjectMutationVariables>;
export const DeleteBucketDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteBucket"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BucketID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteBucket"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"bucket"},"value":{"kind":"Variable","name":{"kind":"Name","value":"bucket"}}}]}]}}]} as unknown as DocumentNode<DeleteBucketMutation, DeleteBucketMutationVariables>;
export const BuildLogsStatusDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BuildLogsStatus"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BuildID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"build"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}}]}}]} as unknown as DocumentNode<BuildLogsStatusQuery, BuildLogsStatusQueryVariables>;
export const AddDatabaseAllowRuleDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddDatabaseAllowRule"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"rule"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseAllowRuleInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addDatabaseAllowRule"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"rule"},"value":{"kind":"Variable","name":{"kind":"Name","value":"rule"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"allowRules"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"range"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}}]}}]}}]} as unknown as DocumentNode<AddDatabaseAllowRuleMutation, AddDatabaseAllowRuleMutationVariables>;
export const RemoveDatabaseAllowRuleDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveDatabaseAllowRule"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"range"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeDatabaseAllowRule"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"range"},"value":{"kind":"Variable","name":{"kind":"Name","value":"range"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"allowRules"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"range"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}}]}}]}}]} as unknown as DocumentNode<RemoveDatabaseAllowRuleMutation, RemoveDatabaseAllowRuleMutationVariables>;
export const DatabaseBackupsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseBackups"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"backups"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"retentionDays"}},{"kind":"Field","name":{"kind":"Name","value":"earliestRestorePoint"}},{"kind":"Field","name":{"kind":"Name","value":"latestRestorePoint"}},{"kind":"Field","name":{"kind":"Name","value":"lastBackupAt"}},{"kind":"Field","name":{"kind":"Name","value":"backups"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"trigger"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}},{"kind":"Field","name":{"kind":"Name","value":"error"}}]}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseBackupsQuery, DatabaseBackupsQueryVariables>;
export const CreateDatabaseBackupDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateDatabaseBackup"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createDatabaseBackup"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<CreateDatabaseBackupMutation, CreateDatabaseBackupMutationVariables>;
export const RestoreDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RestoreDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"RestoreDatabaseInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"restoreDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"clampedToLatest"}},{"kind":"Field","name":{"kind":"Name","value":"database"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]}}]} as unknown as DocumentNode<RestoreDatabaseMutation, RestoreDatabaseMutationVariables>;
export const DatabasePublicDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabasePublic"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"public"}},{"kind":"Field","name":{"kind":"Name","value":"allowRules"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"range"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"clientAddress"}}]}}]} as unknown as DocumentNode<DatabasePublicQuery, DatabasePublicQueryVariables>;
export const DatabaseCredentialsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseCredentials"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseCredentials"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"dbname"}},{"kind":"Field","name":{"kind":"Name","value":"user"}},{"kind":"Field","name":{"kind":"Name","value":"password"}},{"kind":"Field","name":{"kind":"Name","value":"uri"}}]}}]}}]} as unknown as DocumentNode<DatabaseCredentialsQuery, DatabaseCredentialsQueryVariables>;
export const ExposeDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ExposeDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"exposeDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"public"}}]}}]}}]} as unknown as DocumentNode<ExposeDatabaseMutation, ExposeDatabaseMutationVariables>;
export const AllowOnExposeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AllowOnExpose"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"rule"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseAllowRuleInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addDatabaseAllowRule"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"rule"},"value":{"kind":"Variable","name":{"kind":"Name","value":"rule"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"allowRules"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"range"}},{"kind":"Field","name":{"kind":"Name","value":"description"}}]}}]}}]}}]} as unknown as DocumentNode<AllowOnExposeMutation, AllowOnExposeMutationVariables>;
export const UnexposeDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UnexposeDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"unexposeDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"public"}}]}}]}}]} as unknown as DocumentNode<UnexposeDatabaseMutation, UnexposeDatabaseMutationVariables>;
export const DatabaseMetricsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseMetrics"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"range"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"MetricsRange"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"grouping"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"MetricGrouping"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"instances"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}},{"kind":"Field","name":{"kind":"Name","value":"connectionLimit"}},{"kind":"Field","name":{"kind":"Name","value":"metrics"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"metrics"},"value":{"kind":"ListValue","values":[{"kind":"EnumValue","value":"CPU_USAGE"},{"kind":"EnumValue","value":"MEMORY_USAGE"},{"kind":"EnumValue","value":"STORAGE_USED"},{"kind":"EnumValue","value":"CONNECTIONS"},{"kind":"EnumValue","value":"TRANSACTIONS"},{"kind":"EnumValue","value":"CACHE_HIT_RATIO"},{"kind":"EnumValue","value":"ROWS_READ"},{"kind":"EnumValue","value":"ROWS_WRITTEN"},{"kind":"EnumValue","value":"DEADLOCKS"},{"kind":"EnumValue","value":"REPLICATION_LAG"}]}},{"kind":"Argument","name":{"kind":"Name","value":"range"},"value":{"kind":"Variable","name":{"kind":"Name","value":"range"}}},{"kind":"Argument","name":{"kind":"Name","value":"grouping"},"value":{"kind":"Variable","name":{"kind":"Name","value":"grouping"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"metric"}},{"kind":"Field","name":{"kind":"Name","value":"instance"}},{"kind":"Field","name":{"kind":"Name","value":"role"}},{"kind":"Field","name":{"kind":"Name","value":"points"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseMetricsQuery, DatabaseMetricsQueryVariables>;
export const DatabaseQueryStatisticsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseQueryStatistics"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"queryStatistics"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"since"}},{"kind":"Field","name":{"kind":"Name","value":"queries"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"query"}},{"kind":"Field","name":{"kind":"Name","value":"calls"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"totalMilliseconds"}},{"kind":"Field","name":{"kind":"Name","value":"meanMilliseconds"}},{"kind":"Field","name":{"kind":"Name","value":"share"}}]}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseQueryStatisticsQuery, DatabaseQueryStatisticsQueryVariables>;
export const ExecuteQueryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ExecuteQuery"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"query"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"executeQuery"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"query"},"value":{"kind":"Variable","name":{"kind":"Name","value":"query"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"columns"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"affectedRows"}}]}}]}}]} as unknown as DocumentNode<ExecuteQueryMutation, ExecuteQueryMutationVariables>;
export const DatabaseResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseResourcesQuery, DatabaseResourcesQueryVariables>;
export const SetDatabaseResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetDatabaseResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"resources"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ResourcesInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setDatabaseResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"resources"},"value":{"kind":"Variable","name":{"kind":"Name","value":"resources"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}}]}}]}}]} as unknown as DocumentNode<SetDatabaseResourcesMutation, SetDatabaseResourcesMutationVariables>;
export const SetDatabaseStorageDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetDatabaseStorage"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"size"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setDatabaseStorage"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"size"},"value":{"kind":"Variable","name":{"kind":"Name","value":"size"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<SetDatabaseStorageMutation, SetDatabaseStorageMutationVariables>;
export const DeleteDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}]}]}}]} as unknown as DocumentNode<DeleteDatabaseMutation, DeleteDatabaseMutationVariables>;
export const DatabaseTablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseTables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseTables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"schema"}},{"kind":"Field","name":{"kind":"Name","value":"estimatedRows"}},{"kind":"Field","name":{"kind":"Name","value":"columns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"nullable"}},{"kind":"Field","name":{"kind":"Name","value":"primaryKey"}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseTablesQuery, DatabaseTablesQueryVariables>;
export const DatabaseTableDataDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseTableData"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"table"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"schema"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseTableData"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"table"},"value":{"kind":"Variable","name":{"kind":"Name","value":"table"}}},{"kind":"Argument","name":{"kind":"Name","value":"schema"},"value":{"kind":"Variable","name":{"kind":"Name","value":"schema"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"columns"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"totalEstimatedRows"}}]}}]}}]} as unknown as DocumentNode<DatabaseTableDataQuery, DatabaseTableDataQueryVariables>;
export const KeyValueStoreCredentialsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"KeyValueStoreCredentials"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"keyValueStore"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"KeyValueStoreID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"keyValueStoreCredentials"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"keyValueStore"},"value":{"kind":"Variable","name":{"kind":"Name","value":"keyValueStore"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"password"}},{"kind":"Field","name":{"kind":"Name","value":"uri"}}]}}]}}]} as unknown as DocumentNode<KeyValueStoreCredentialsQuery, KeyValueStoreCredentialsQueryVariables>;
export const DeleteKeyValueStoreDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteKeyValueStore"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"keyValueStore"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"KeyValueStoreID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteKeyValueStore"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"keyValueStore"},"value":{"kind":"Variable","name":{"kind":"Name","value":"keyValueStore"}}}]}]}}]} as unknown as DocumentNode<DeleteKeyValueStoreMutation, DeleteKeyValueStoreMutationVariables>;
export const ServiceSecurityDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ServiceSecurity"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}},{"kind":"Field","name":{"kind":"Name","value":"vulnerabilityReport"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"summary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"critical"}},{"kind":"Field","name":{"kind":"Name","value":"high"}},{"kind":"Field","name":{"kind":"Name","value":"medium"}},{"kind":"Field","name":{"kind":"Name","value":"low"}},{"kind":"Field","name":{"kind":"Name","value":"unknown"}},{"kind":"Field","name":{"kind":"Name","value":"total"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"secretScanReport"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"scannedAt"}},{"kind":"Field","name":{"kind":"Name","value":"findings"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"rule"}},{"kind":"Field","name":{"kind":"Name","value":"file"}},{"kind":"Field","name":{"kind":"Name","value":"line"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"secret"}},{"kind":"Field","name":{"kind":"Name","value":"author"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"verified"}}]}}]}}]}}]}}]} as unknown as DocumentNode<ServiceSecurityQuery, ServiceSecurityQueryVariables>;
export const ServiceMetricsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ServiceMetrics"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"range"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"MetricsRange"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"grouping"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"MetricGrouping"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}}]}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}},{"kind":"Field","name":{"kind":"Name","value":"deployments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"metrics"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"metrics"},"value":{"kind":"ListValue","values":[{"kind":"EnumValue","value":"CPU_USAGE"},{"kind":"EnumValue","value":"MEMORY_USAGE"}]}},{"kind":"Argument","name":{"kind":"Name","value":"range"},"value":{"kind":"Variable","name":{"kind":"Name","value":"range"}}},{"kind":"Argument","name":{"kind":"Name","value":"grouping"},"value":{"kind":"Variable","name":{"kind":"Name","value":"grouping"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"metric"}},{"kind":"Field","name":{"kind":"Name","value":"replica"}},{"kind":"Field","name":{"kind":"Name","value":"points"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]} as unknown as DocumentNode<ServiceMetricsQuery, ServiceMetricsQueryVariables>;
export const RemoveServiceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveService"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeService"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}]}]}}]} as unknown as DocumentNode<RemoveServiceMutation, RemoveServiceMutationVariables>;
export const SetCustomStartCommandDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetCustomStartCommand"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"command"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setCustomStartCommand"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"command"},"value":{"kind":"Variable","name":{"kind":"Name","value":"command"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}}]}}]}}]} as unknown as DocumentNode<SetCustomStartCommandMutation, SetCustomStartCommandMutationVariables>;
export const GenerateDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"GenerateDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"generateDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"redirectTo"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<GenerateDomainMutation, GenerateDomainMutationVariables>;
export const AddCustomDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddCustomDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"redirectTo"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addCustomDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"hostname"},"value":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}}},{"kind":"Argument","name":{"kind":"Name","value":"redirectTo"},"value":{"kind":"Variable","name":{"kind":"Name","value":"redirectTo"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"redirectTo"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<AddCustomDomainMutation, AddCustomDomainMutationVariables>;
export const RemoveDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"hostname"},"value":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"redirectTo"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<RemoveDomainMutation, RemoveDomainMutationVariables>;
export const SetServiceScalingDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceScaling"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"SetServiceScalingInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceScaling"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"autoscaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"minReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"maxReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"targetCpu"}}]}}]}}]}}]} as unknown as DocumentNode<SetServiceScalingMutation, SetServiceScalingMutationVariables>;
export const SetServicePortDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServicePort"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"port"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServicePort"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"port"},"value":{"kind":"Variable","name":{"kind":"Name","value":"port"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"port"}}]}}]}}]} as unknown as DocumentNode<SetServicePortMutation, SetServicePortMutationVariables>;
export const SetServiceHealthCheckDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceHealthCheck"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"healthCheck"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"HealthCheckInput"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceHealthCheck"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"healthCheck"},"value":{"kind":"Variable","name":{"kind":"Name","value":"healthCheck"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"healthCheck"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"path"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"initialDelaySeconds"}},{"kind":"Field","name":{"kind":"Name","value":"periodSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"timeoutSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"failureThreshold"}},{"kind":"Field","name":{"kind":"Name","value":"startupFailureThreshold"}}]}}]}}]}}]} as unknown as DocumentNode<SetServiceHealthCheckMutation, SetServiceHealthCheckMutationVariables>;
export const SetServiceBranchDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceBranch"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"branch"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceBranch"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"branch"},"value":{"kind":"Variable","name":{"kind":"Name","value":"branch"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"branch"}}]}}]}}]} as unknown as DocumentNode<SetServiceBranchMutation, SetServiceBranchMutationVariables>;
export const SetAutoDeployDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetAutoDeploy"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"enabled"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setAutoDeploy"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"enabled"},"value":{"kind":"Variable","name":{"kind":"Name","value":"enabled"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"autoDeploy"}}]}}]}}]} as unknown as DocumentNode<SetAutoDeployMutation, SetAutoDeployMutationVariables>;
export const SetCiDeployDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetCIDeploy"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"enabled"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setCIDeploy"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"enabled"},"value":{"kind":"Variable","name":{"kind":"Name","value":"enabled"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"ciDeploy"}}]}}]}}]} as unknown as DocumentNode<SetCiDeployMutation, SetCiDeployMutationVariables>;
export const RepositoryBranchesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"RepositoryBranches"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"repositoryUrl"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"repositoryBranches"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"repositoryUrl"},"value":{"kind":"Variable","name":{"kind":"Name","value":"repositoryUrl"}}}]}]}}]} as unknown as DocumentNode<RepositoryBranchesQuery, RepositoryBranchesQueryVariables>;
export const SetServiceResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"resources"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ResourcesInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"resources"},"value":{"kind":"Variable","name":{"kind":"Name","value":"resources"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}}]}}]}}]} as unknown as DocumentNode<SetServiceResourcesMutation, SetServiceResourcesMutationVariables>;
export const SetServiceUserDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceUser"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"user"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceUser"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"user"},"value":{"kind":"Variable","name":{"kind":"Name","value":"user"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"user"}}]}}]}}]} as unknown as DocumentNode<SetServiceUserMutation, SetServiceUserMutationVariables>;
export const ServiceVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ServiceVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serviceVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"value"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}}]}}]}}]} as unknown as DocumentNode<ServiceVariablesQuery, ServiceVariablesQueryVariables>;
export const AvailableVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"AvailableVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availableVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"source"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseSource"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","alias":{"kind":"Name","value":"databaseId"},"name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"KeyValueStoreSource"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","alias":{"kind":"Name","value":"keyValueStoreId"},"name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"BucketSource"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","alias":{"kind":"Name","value":"bucketId"},"name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"InlineFragment","typeCondition":{"kind":"NamedType","name":{"kind":"Name","value":"SharedSource"}},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]} as unknown as DocumentNode<AvailableVariablesQuery, AvailableVariablesQueryVariables>;
export const SetServiceVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"variables"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceVariableInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"variables"},"value":{"kind":"Variable","name":{"kind":"Name","value":"variables"}}}]}]}}]} as unknown as DocumentNode<SetServiceVariablesMutation, SetServiceVariablesMutationVariables>;
export const ServiceVulnerabilitiesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ServiceVulnerabilities"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"vulnerabilityReport"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"summary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"critical"}},{"kind":"Field","name":{"kind":"Name","value":"high"}},{"kind":"Field","name":{"kind":"Name","value":"medium"}},{"kind":"Field","name":{"kind":"Name","value":"low"}},{"kind":"Field","name":{"kind":"Name","value":"unknown"}},{"kind":"Field","name":{"kind":"Name","value":"total"}}]}},{"kind":"Field","name":{"kind":"Name","value":"vulnerabilities"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"severity"}},{"kind":"Field","name":{"kind":"Name","value":"source"}},{"kind":"Field","name":{"kind":"Name","value":"title"}},{"kind":"Field","name":{"kind":"Name","value":"reference"}},{"kind":"Field","name":{"kind":"Name","value":"packages"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"installedVersion"}},{"kind":"Field","name":{"kind":"Name","value":"fixedVersion"}},{"kind":"Field","name":{"kind":"Name","value":"path"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<ServiceVulnerabilitiesQuery, ServiceVulnerabilitiesQueryVariables>;
export const DeleteVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"volume"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"volume"},"value":{"kind":"Variable","name":{"kind":"Name","value":"volume"}}}]}]}}]} as unknown as DocumentNode<DeleteVolumeMutation, DeleteVolumeMutationVariables>;
export const ExpandVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ExpandVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"volume"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"size"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"expandVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"volume"},"value":{"kind":"Variable","name":{"kind":"Name","value":"volume"}}},{"kind":"Argument","name":{"kind":"Name","value":"size"},"value":{"kind":"Variable","name":{"kind":"Name","value":"size"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<ExpandVolumeMutation, ExpandVolumeMutationVariables>;
export const UnmountVolumeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UnmountVolume"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"volume"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"unmountVolume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"volume"},"value":{"kind":"Variable","name":{"kind":"Name","value":"volume"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"mount"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"}},{"kind":"Field","name":{"kind":"Name","value":"path"}}]}}]}}]}}]} as unknown as DocumentNode<UnmountVolumeMutation, UnmountVolumeMutationVariables>;
export const VolumeMetricsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"VolumeMetrics"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VolumeID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"range"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"MetricsRange"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"volume"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"metrics"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"metrics"},"value":{"kind":"ListValue","values":[{"kind":"EnumValue","value":"STORAGE_USED"}]}},{"kind":"Argument","name":{"kind":"Name","value":"range"},"value":{"kind":"Variable","name":{"kind":"Name","value":"range"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"metric"}},{"kind":"Field","name":{"kind":"Name","value":"points"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]} as unknown as DocumentNode<VolumeMetricsQuery, VolumeMetricsQueryVariables>;
export const CapabilitiesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Capabilities"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"capabilities"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availability"}},{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"cpu"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"min"}},{"kind":"Field","name":{"kind":"Name","value":"max"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"max"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"memory"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"min"}},{"kind":"Field","name":{"kind":"Name","value":"max"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"max"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"replicas"}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availability"}},{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"cpu"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"min"}},{"kind":"Field","name":{"kind":"Name","value":"max"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"max"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"memory"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"min"}},{"kind":"Field","name":{"kind":"Name","value":"max"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"max"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"storage"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"min"}},{"kind":"Field","name":{"kind":"Name","value":"max"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"max"}}]}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"keyValueStores"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availability"}},{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"size"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"min"}},{"kind":"Field","name":{"kind":"Name","value":"max"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"max"}}]}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"volumes"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availability"}},{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"size"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"min"}},{"kind":"Field","name":{"kind":"Name","value":"max"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"max"}}]}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"buckets"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availability"}},{"kind":"Field","name":{"kind":"Name","value":"plan"}}]}},{"kind":"Field","name":{"kind":"Name","value":"agents"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availability"}},{"kind":"Field","name":{"kind":"Name","value":"plan"}}]}},{"kind":"Field","name":{"kind":"Name","value":"members"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availability"}},{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"limit"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"limit"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"customDomains"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"availability"}},{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"limit"}},{"kind":"Field","name":{"kind":"Name","value":"upgrades"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"limit"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<CapabilitiesQuery, CapabilitiesQueryVariables>;
export const DeployDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Deploy"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"gitRef"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deploy"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"gitRef"},"value":{"kind":"Variable","name":{"kind":"Name","value":"gitRef"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"build"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}}]}}]}}]} as unknown as DocumentNode<DeployMutation, DeployMutationVariables>;
export const BuildStatusDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BuildStatus"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BuildID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"build"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}}]}}]} as unknown as DocumentNode<BuildStatusQuery, BuildStatusQueryVariables>;
export const BuildLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"BuildLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BuildID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"buildLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<BuildLogsSubscription, BuildLogsSubscriptionVariables>;
export const DeployLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"DeployLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DeployID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deployLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<DeployLogsSubscription, DeployLogsSubscriptionVariables>;
export const ScanLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"ScanLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ScanID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"scanLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<ScanLogsSubscription, ScanLogsSubscriptionVariables>;
export const ServiceLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"ServiceLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"tailLines"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serviceLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"tailLines"},"value":{"kind":"Variable","name":{"kind":"Name","value":"tailLines"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"line"}},{"kind":"Field","name":{"kind":"Name","value":"pod"}}]}}]}}]} as unknown as DocumentNode<ServiceLogsSubscription, ServiceLogsSubscriptionVariables>;
export const WorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}},{"kind":"Field","name":{"kind":"Name","value":"suspended"}},{"kind":"Field","name":{"kind":"Name","value":"members"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]}}]} as unknown as DocumentNode<WorkspaceQuery, WorkspaceQueryVariables>;
export const CompletePaymentMethodCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CompletePaymentMethodCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"completePaymentMethodCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"sessionId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}}]}}]}}]} as unknown as DocumentNode<CompletePaymentMethodCheckoutMutation, CompletePaymentMethodCheckoutMutationVariables>;
export const CompleteWorkspaceCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CompleteWorkspaceCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"completeWorkspaceCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"sessionId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}}]}}]}}]} as unknown as DocumentNode<CompleteWorkspaceCheckoutMutation, CompleteWorkspaceCheckoutMutationVariables>;
export const EnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Environment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"autoscaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"minReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"maxReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"targetCpu"}}]}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"redirectTo"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"tls"}}]}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}},{"kind":"Field","name":{"kind":"Name","value":"branch"}},{"kind":"Field","name":{"kind":"Name","value":"autoDeploy"}},{"kind":"Field","name":{"kind":"Name","value":"ciDeploy"}},{"kind":"Field","name":{"kind":"Name","value":"contextPath"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}},{"kind":"Field","name":{"kind":"Name","value":"command"}},{"kind":"Field","name":{"kind":"Name","value":"healthCheck"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"path"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"initialDelaySeconds"}},{"kind":"Field","name":{"kind":"Name","value":"periodSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"timeoutSeconds"}},{"kind":"Field","name":{"kind":"Name","value":"failureThreshold"}},{"kind":"Field","name":{"kind":"Name","value":"startupFailureThreshold"}}]}},{"kind":"Field","name":{"kind":"Name","value":"user"}},{"kind":"Field","name":{"kind":"Name","value":"activeDeployment"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"imageDigest"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"commitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"rollout"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"reason"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"restarts"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"deployments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"imageDigest"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"commitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"builds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"releases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"trigger"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"kind"}},{"kind":"Field","name":{"kind":"Name","value":"actor"}}]}},{"kind":"Field","name":{"kind":"Name","value":"source"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"provider"}},{"kind":"Field","name":{"kind":"Name","value":"repository"}},{"kind":"Field","name":{"kind":"Name","value":"url"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"contextPath"}},{"kind":"Field","name":{"kind":"Name","value":"commit"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"sha"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"build"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"deploy"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"scan"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"findingsCount"}},{"kind":"Field","name":{"kind":"Name","value":"verifiedCount"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"deployment"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"imageDigest"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"commitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"rollout"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"reason"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"restarts"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}}]}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"lastDeployedAt"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"instances"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"keyValueStores"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"buckets"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"region"}},{"kind":"Field","name":{"kind":"Name","value":"endpoint"}},{"kind":"Field","name":{"kind":"Name","value":"publicEndpoint"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"sizeBytes"}},{"kind":"Field","name":{"kind":"Name","value":"objectCount"}},{"kind":"Field","name":{"kind":"Name","value":"public"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"volumes"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"mount"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"}},{"kind":"Field","name":{"kind":"Name","value":"path"}}]}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentQuery, EnvironmentQueryVariables>;
export const EnvironmentVolumeUsageDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EnvironmentVolumeUsage"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"volumes"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"metrics"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"metrics"},"value":{"kind":"ListValue","values":[{"kind":"EnumValue","value":"STORAGE_USED"}]}},{"kind":"Argument","name":{"kind":"Name","value":"range"},"value":{"kind":"ObjectValue","fields":[{"kind":"ObjectField","name":{"kind":"Name","value":"window"},"value":{"kind":"EnumValue","value":"LAST_1H"}}]}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"points"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentVolumeUsageQuery, EnvironmentVolumeUsageQueryVariables>;
export const EnvironmentDefaultCommandsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EnvironmentDefaultCommands"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"defaultCommand"}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentDefaultCommandsQuery, EnvironmentDefaultCommandsQueryVariables>;
export const EnvironmentAgentsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EnvironmentAgents"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"agents"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"harness"}},{"kind":"Field","name":{"kind":"Name","value":"instructions"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"allowedHosts"}},{"kind":"Field","name":{"kind":"Name","value":"privateNetwork"}},{"kind":"Field","name":{"kind":"Name","value":"connectors"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"approval"}},{"kind":"Field","name":{"kind":"Name","value":"hiddenTools"}},{"kind":"Field","name":{"kind":"Name","value":"connector"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentAgentsQuery, EnvironmentAgentsQueryVariables>;
export const ProjectEnvironmentsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ProjectEnvironments"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"project"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]}}]} as unknown as DocumentNode<ProjectEnvironmentsQuery, ProjectEnvironmentsQueryVariables>;
export const CompletePlanCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CompletePlanCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"completePlanCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"sessionId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"hasPaymentMethod"}}]}}]}}]} as unknown as DocumentNode<CompletePlanCheckoutMutation, CompletePlanCheckoutMutationVariables>;
export const ProjectPageDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ProjectPage"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ProjectID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"project"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}},{"kind":"Field","name":{"kind":"Name","value":"keyValueStores"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}},{"kind":"Field","name":{"kind":"Name","value":"buckets"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"status"}}]}},{"kind":"Field","name":{"kind":"Name","value":"volumes"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]} as unknown as DocumentNode<ProjectPageQuery, ProjectPageQueryVariables>;
export const ProjectsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Projects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"projects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"Field","name":{"kind":"Name","value":"keyValueStores"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"Field","name":{"kind":"Name","value":"buckets"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}},{"kind":"Field","name":{"kind":"Name","value":"volumes"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]}}]} as unknown as DocumentNode<ProjectsQuery, ProjectsQueryVariables>;