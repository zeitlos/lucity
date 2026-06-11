/* eslint-disable */
import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
export type Maybe<T> = T | null;
export type InputMaybe<T> = T | null | undefined;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  DatabaseID: { input: string; output: string; }
  DeploymentID: { input: string; output: string; }
  Duration: { input: string; output: string; }
  EnvironmentID: { input: string; output: string; }
  ProjectID: { input: string; output: string; }
  ServiceID: { input: string; output: string; }
  Time: { input: string; output: string; }
  VolumeID: { input: string; output: string; }
};

export type AddServiceInput = {
  contextPath?: InputMaybe<Scalars['String']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  installationId?: InputMaybe<Scalars['ID']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  repository?: InputMaybe<Scalars['String']['input']>;
};

export type AutoscalingConfig = {
  __typename?: 'AutoscalingConfig';
  enabled: Scalars['Boolean']['output'];
  maxReplicas: Scalars['Int']['output'];
  minReplicas: Scalars['Int']['output'];
  targetCPU: Scalars['Int']['output'];
};

export type AutoscalingInput = {
  enabled: Scalars['Boolean']['input'];
  maxReplicas: Scalars['Int']['input'];
  minReplicas: Scalars['Int']['input'];
  targetCPU: Scalars['Int']['input'];
};

export type AutoscalingSettings = {
  __typename?: 'AutoscalingSettings';
  maxReplicas: Scalars['Int']['output'];
  minReplicas: Scalars['Int']['output'];
  targetCpu: Scalars['Int']['output'];
};

export type BillingPortalUrl = {
  __typename?: 'BillingPortalUrl';
  url: Scalars['String']['output'];
};

export type BillingSubscription = {
  __typename?: 'BillingSubscription';
  creditAmountCents: Scalars['Int']['output'];
  creditExpiry?: Maybe<Scalars['Time']['output']>;
  currentPeriodEnd: Scalars['Time']['output'];
  hasPaymentMethod: Scalars['Boolean']['output'];
  plan?: Maybe<Plan>;
  status: SubscriptionStatus;
};

export type Build = {
  __typename?: 'Build';
  finishedAt?: Maybe<Scalars['Time']['output']>;
  id: Scalars['String']['output'];
  startedAt: Scalars['Time']['output'];
  status: BuildStatus;
};

export enum BuildStatus {
  Cancelled = 'CANCELLED',
  Failed = 'FAILED',
  Queued = 'QUEUED',
  Running = 'RUNNING',
  Succeeded = 'SUCCEEDED'
}

export type CheckoutSession = {
  __typename?: 'CheckoutSession';
  url: Scalars['String']['output'];
};

export type CreateDatabaseInput = {
  environment: Scalars['EnvironmentID']['input'];
  instances?: InputMaybe<Scalars['Int']['input']>;
  name: Scalars['String']['input'];
  size?: InputMaybe<Scalars['String']['input']>;
  version?: InputMaybe<Scalars['String']['input']>;
};

export type CreateEnvironmentInput = {
  fromEnvironment?: InputMaybe<Scalars['EnvironmentID']['input']>;
  name: Scalars['String']['input'];
  project: Scalars['ProjectID']['input'];
  tier?: InputMaybe<ResourceTier>;
};

export type CreateProjectInput = {
  /** Optional URL-safe slug. Auto-derived from name if omitted. */
  id?: InputMaybe<Scalars['String']['input']>;
  /** Human-readable project name (e.g. "My API"). */
  name: Scalars['String']['input'];
};

export type CreateWorkspaceCheckoutInput = {
  id: Scalars['String']['input'];
  name: Scalars['String']['input'];
  plan: Plan;
};

export type Database = {
  __typename?: 'Database';
  createdAt: Scalars['Time']['output'];
  id: Scalars['DatabaseID']['output'];
  instances: Scalars['Int']['output'];
  name: Scalars['String']['output'];
  size: Scalars['String']['output'];
  status: DatabaseStatus;
  version: Scalars['String']['output'];
};

export type DatabaseColumn = {
  __typename?: 'DatabaseColumn';
  name: Scalars['String']['output'];
  nullable: Scalars['Boolean']['output'];
  primaryKey: Scalars['Boolean']['output'];
  type: Scalars['String']['output'];
};

export type DatabaseCredentials = {
  __typename?: 'DatabaseCredentials';
  dbname: Scalars['String']['output'];
  host: Scalars['String']['output'];
  password: Scalars['String']['output'];
  port: Scalars['String']['output'];
  uri: Scalars['String']['output'];
  user: Scalars['String']['output'];
};

/** A reference to a CNPG database secret key (resolved at pod startup via secretKeyRef). */
export type DatabaseRef = {
  __typename?: 'DatabaseRef';
  database: Scalars['DatabaseID']['output'];
  key: Scalars['String']['output'];
};

export type DatabaseRefInput = {
  database: Scalars['DatabaseID']['input'];
  key: Scalars['String']['input'];
};

export enum DatabaseStatus {
  Degraded = 'DEGRADED',
  Failed = 'FAILED',
  Healthy = 'HEALTHY',
  Pending = 'PENDING',
  Stopped = 'STOPPED'
}

export type DatabaseTable = {
  __typename?: 'DatabaseTable';
  columns: Array<DatabaseColumn>;
  estimatedRows: Scalars['Int']['output'];
  name: Scalars['String']['output'];
  schema: Scalars['String']['output'];
};

export type DatabaseTableData = {
  __typename?: 'DatabaseTableData';
  columns: Array<Scalars['String']['output']>;
  rows: Array<Maybe<Array<Maybe<Scalars['String']['output']>>>>;
  totalEstimatedRows: Scalars['Int']['output'];
};

export type Deployment = {
  __typename?: 'Deployment';
  buildId: Scalars['String']['output'];
  command: Scalars['String']['output'];
  commit: Scalars['String']['output'];
  commitMessage: Scalars['String']['output'];
  contextPath: Scalars['String']['output'];
  createdAt: Scalars['Time']['output'];
  deployedBy: Scalars['String']['output'];
  id: Scalars['DeploymentID']['output'];
  image: Scalars['String']['output'];
  imageDigest?: Maybe<Scalars['String']['output']>;
  ref: Scalars['String']['output'];
  replicas: ReplicaCount;
  resources: Resources;
  sourceUrl: Scalars['String']['output'];
  status: DeploymentStatus;
};

export enum DeploymentStatus {
  Active = 'ACTIVE',
  Deploying = 'DEPLOYING',
  Failed = 'FAILED',
  Superseded = 'SUPERSEDED'
}

export type DetectedService = {
  __typename?: 'DetectedService';
  framework: Scalars['String']['output'];
  language: Scalars['String']['output'];
  name: Scalars['String']['output'];
  startCommand: Scalars['String']['output'];
  suggestedPort: Scalars['Int']['output'];
};

export type DnsRecord = {
  __typename?: 'DnsRecord';
  host: Scalars['String']['output'];
  type: DnsRecordType;
  value: Scalars['String']['output'];
};

export enum DnsRecordType {
  A = 'A',
  Cname = 'CNAME',
  Txt = 'TXT'
}

export type DnsState = {
  __typename?: 'DnsState';
  requiredRecords: Array<DnsRecord>;
  status: DnsStatus;
};

export enum DnsStatus {
  Error = 'ERROR',
  Misconfigured = 'MISCONFIGURED',
  Pending = 'PENDING',
  Valid = 'VALID'
}

export enum DomainType {
  Custom = 'CUSTOM',
  Platform = 'PLATFORM'
}

export type Endpoint = {
  __typename?: 'Endpoint';
  dns: DnsState;
  host: Scalars['String']['output'];
  port: Scalars['Int']['output'];
  protocol: Protocol;
  tls: TlsStatus;
  type: EndpointType;
};

export enum EndpointType {
  Custom = 'CUSTOM',
  Internal = 'INTERNAL',
  Platform = 'PLATFORM'
}

export type Environment = {
  __typename?: 'Environment';
  databases: Array<Database>;
  id: Scalars['EnvironmentID']['output'];
  name: Scalars['String']['output'];
  resourceTier: ResourceTier;
  services: Array<Service>;
};

export type EnvironmentResources = {
  __typename?: 'EnvironmentResources';
  allocation: ResourceAllocation;
  tier: ResourceTier;
};

export enum GitHubAccountType {
  Organization = 'ORGANIZATION',
  User = 'USER'
}

export type GitHubInstallation = {
  __typename?: 'GitHubInstallation';
  accountAvatarUrl: Scalars['String']['output'];
  accountLogin: Scalars['String']['output'];
  accountType: GitHubAccountType;
  id: Scalars['ID']['output'];
};

export type GitHubRepository = {
  __typename?: 'GitHubRepository';
  defaultBranch: Scalars['String']['output'];
  fullName: Scalars['String']['output'];
  htmlUrl: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  private: Scalars['Boolean']['output'];
};

/** A container image from a public registry (Docker Hub). */
export type ImageSearchResult = {
  __typename?: 'ImageSearchResult';
  description: Scalars['String']['output'];
  name: Scalars['String']['output'];
  official: Scalars['Boolean']['output'];
  pullCount: Scalars['Int']['output'];
  starCount: Scalars['Int']['output'];
};

export type InviteMemberInput = {
  email: Scalars['String']['input'];
  role: WorkspaceRole;
};

export type Mutation = {
  __typename?: 'Mutation';
  addCustomDomain: Service;
  addService: Service;
  billingPortalUrl: BillingPortalUrl;
  changePlan: BillingSubscription;
  completePlanCheckout: BillingSubscription;
  completeWorkspaceCheckout: Workspace;
  createDatabase: Database;
  createEnvironment: Environment;
  createPlanCheckout: CheckoutSession;
  createProject: Project;
  createWorkspaceCheckout: CheckoutSession;
  deleteDatabase: Scalars['Boolean']['output'];
  deleteEnvironment: Scalars['Boolean']['output'];
  deleteProject: Scalars['Boolean']['output'];
  deleteWorkspace: Scalars['Boolean']['output'];
  deploy: Build;
  executeQuery: QueryResult;
  generateDomain: Service;
  inviteMember: WorkspaceMember;
  removeDomain: Service;
  removeMember: Scalars['Boolean']['output'];
  removeService: Scalars['Boolean']['output'];
  rollback: Scalars['Boolean']['output'];
  setCustomStartCommand: Service;
  setEnvironmentResources: Environment;
  setServicePort: Service;
  setServiceScaling: Service;
  setServiceVariables: Scalars['Boolean']['output'];
  setSharedVariables: Scalars['Boolean']['output'];
  updateMemberRole: WorkspaceMember;
  updateWorkspace: Workspace;
};


export type MutationAddCustomDomainArgs = {
  hostname: Scalars['String']['input'];
  service: Scalars['ServiceID']['input'];
};


export type MutationAddServiceArgs = {
  environment: Scalars['EnvironmentID']['input'];
  input: AddServiceInput;
};


export type MutationChangePlanArgs = {
  plan: Plan;
};


export type MutationCompletePlanCheckoutArgs = {
  sessionId: Scalars['String']['input'];
};


export type MutationCompleteWorkspaceCheckoutArgs = {
  sessionId: Scalars['String']['input'];
};


export type MutationCreateDatabaseArgs = {
  input: CreateDatabaseInput;
};


export type MutationCreateEnvironmentArgs = {
  input: CreateEnvironmentInput;
};


export type MutationCreatePlanCheckoutArgs = {
  plan: Plan;
};


export type MutationCreateProjectArgs = {
  input: CreateProjectInput;
};


export type MutationCreateWorkspaceCheckoutArgs = {
  input: CreateWorkspaceCheckoutInput;
};


export type MutationDeleteDatabaseArgs = {
  database: Scalars['DatabaseID']['input'];
};


export type MutationDeleteEnvironmentArgs = {
  environment: Scalars['EnvironmentID']['input'];
};


export type MutationDeleteProjectArgs = {
  id: Scalars['ProjectID']['input'];
};


export type MutationDeployArgs = {
  gitRef?: InputMaybe<Scalars['String']['input']>;
  service: Scalars['ServiceID']['input'];
};


export type MutationExecuteQueryArgs = {
  database: Scalars['DatabaseID']['input'];
  query: Scalars['String']['input'];
};


export type MutationGenerateDomainArgs = {
  service: Scalars['ServiceID']['input'];
};


export type MutationInviteMemberArgs = {
  input: InviteMemberInput;
};


export type MutationRemoveDomainArgs = {
  hostname: Scalars['String']['input'];
  service: Scalars['ServiceID']['input'];
};


export type MutationRemoveMemberArgs = {
  userId: Scalars['ID']['input'];
};


export type MutationRemoveServiceArgs = {
  service: Scalars['ServiceID']['input'];
};


export type MutationRollbackArgs = {
  deployment: Scalars['DeploymentID']['input'];
};


export type MutationSetCustomStartCommandArgs = {
  command: Scalars['String']['input'];
  service: Scalars['ServiceID']['input'];
};


export type MutationSetEnvironmentResourcesArgs = {
  input: SetEnvironmentResourcesInput;
};


export type MutationSetServicePortArgs = {
  port?: InputMaybe<Scalars['Int']['input']>;
  service: Scalars['ServiceID']['input'];
};


export type MutationSetServiceScalingArgs = {
  input: SetServiceScalingInput;
};


export type MutationSetServiceVariablesArgs = {
  service: Scalars['ServiceID']['input'];
  variables: Array<ServiceVariableInput>;
};


export type MutationSetSharedVariablesArgs = {
  environment: Scalars['EnvironmentID']['input'];
  variables: Array<VariableInput>;
};


export type MutationUpdateMemberRoleArgs = {
  input: UpdateMemberRoleInput;
};


export type MutationUpdateWorkspaceArgs = {
  input: UpdateWorkspaceInput;
};

export enum Plan {
  Hobby = 'HOBBY',
  Pro = 'PRO'
}

export type Project = {
  __typename?: 'Project';
  environments: Array<Environment>;
  id: Scalars['ProjectID']['output'];
  name: Scalars['String']['output'];
};

export enum Protocol {
  Http = 'HTTP',
  Https = 'HTTPS',
  Tcp = 'TCP'
}

export type Query = {
  __typename?: 'Query';
  build: Build;
  database: Database;
  databaseCredentials: DatabaseCredentials;
  databaseTableData: DatabaseTableData;
  databaseTables: Array<DatabaseTable>;
  deployment: Deployment;
  detectServices: Array<DetectedService>;
  environment: Environment;
  environmentResources?: Maybe<EnvironmentResources>;
  environments: Array<Environment>;
  /** Whether the current user has connected their GitHub account. */
  githubConnected: Scalars['Boolean']['output'];
  /** Repos from a specific installation. */
  githubRepositories: Array<GitHubRepository>;
  /** User's accessible GitHub App installations. Requires connected GitHub account. */
  githubSources: Array<GitHubInstallation>;
  me: User;
  project: Project;
  projects: Array<Project>;
  /** Search Docker Hub for public container images. */
  searchImages: Array<ImageSearchResult>;
  service: Service;
  serviceVariables: Array<ServiceVariable>;
  sharedVariables: Array<Variable>;
  subscription?: Maybe<BillingSubscription>;
  usageSummary?: Maybe<UsageSummary>;
  workspace: Workspace;
  workspaces: Array<Workspace>;
};


export type QueryBuildArgs = {
  id: Scalars['String']['input'];
};


export type QueryDatabaseArgs = {
  id: Scalars['DatabaseID']['input'];
};


export type QueryDatabaseCredentialsArgs = {
  database: Scalars['DatabaseID']['input'];
};


export type QueryDatabaseTableDataArgs = {
  database: Scalars['DatabaseID']['input'];
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
  schema?: InputMaybe<Scalars['String']['input']>;
  table: Scalars['String']['input'];
};


export type QueryDatabaseTablesArgs = {
  database: Scalars['DatabaseID']['input'];
};


export type QueryDeploymentArgs = {
  id: Scalars['DeploymentID']['input'];
};


export type QueryDetectServicesArgs = {
  installationId: Scalars['ID']['input'];
  repositoryUrl: Scalars['String']['input'];
};


export type QueryEnvironmentArgs = {
  environment: Scalars['EnvironmentID']['input'];
};


export type QueryEnvironmentResourcesArgs = {
  environment: Scalars['EnvironmentID']['input'];
};


export type QueryEnvironmentsArgs = {
  project: Scalars['ProjectID']['input'];
};


export type QueryGithubRepositoriesArgs = {
  installationId: Scalars['ID']['input'];
};


export type QueryProjectArgs = {
  id: Scalars['ProjectID']['input'];
};


export type QuerySearchImagesArgs = {
  query: Scalars['String']['input'];
};


export type QueryServiceArgs = {
  id: Scalars['ServiceID']['input'];
};


export type QueryServiceVariablesArgs = {
  service: Scalars['ServiceID']['input'];
};


export type QuerySharedVariablesArgs = {
  environment: Scalars['EnvironmentID']['input'];
};

export type QueryResult = {
  __typename?: 'QueryResult';
  affectedRows: Scalars['Int']['output'];
  columns: Array<Scalars['String']['output']>;
  rows: Array<Maybe<Array<Maybe<Scalars['String']['output']>>>>;
};

export type ReplicaCount = {
  __typename?: 'ReplicaCount';
  desired: Scalars['Int']['output'];
  ready: Scalars['Int']['output'];
};

export type ResourceAllocation = {
  __typename?: 'ResourceAllocation';
  cpuMillicores: Scalars['Int']['output'];
  diskMB: Scalars['Int']['output'];
  memoryMB: Scalars['Int']['output'];
};

export enum ResourceTier {
  Eco = 'ECO',
  Production = 'PRODUCTION'
}

export type Resources = {
  __typename?: 'Resources';
  cpu: Scalars['String']['output'];
  memory: Scalars['String']['output'];
};

export enum Role {
  Admin = 'ADMIN',
  Anonymous = 'ANONYMOUS',
  User = 'USER'
}

export type ScalingConfig = {
  __typename?: 'ScalingConfig';
  autoscaling?: Maybe<AutoscalingConfig>;
  replicas: Scalars['Int']['output'];
};

export type Service = {
  __typename?: 'Service';
  activeDeployment?: Maybe<Deployment>;
  autoscaling?: Maybe<AutoscalingSettings>;
  builds: Array<Build>;
  command: Scalars['String']['output'];
  contextPath: Scalars['String']['output'];
  createdAt: Scalars['Time']['output'];
  defaultCommand: Scalars['String']['output'];
  deployments: Array<Deployment>;
  endpoints: Array<Endpoint>;
  id: Scalars['ServiceID']['output'];
  lastDeployedAt?: Maybe<Scalars['Time']['output']>;
  name: Scalars['String']['output'];
  port: Scalars['Int']['output'];
  replicas: ReplicaCount;
  resources: Resources;
  sourceUrl: Scalars['String']['output'];
  status: ServiceStatus;
};

export type ServiceLogEntry = {
  __typename?: 'ServiceLogEntry';
  /** Log line text. Prefixed with [pod-suffix] when multiple replicas exist. */
  line: Scalars['String']['output'];
  /** Name of the pod that produced this line. */
  pod: Scalars['String']['output'];
};

export type ServiceResources = {
  __typename?: 'ServiceResources';
  cpuLimitMillicores: Scalars['Int']['output'];
  cpuMillicores: Scalars['Int']['output'];
  memoryLimitMB: Scalars['Int']['output'];
  memoryMB: Scalars['Int']['output'];
};

export enum ServiceStatus {
  Degraded = 'DEGRADED',
  Deploying = 'DEPLOYING',
  Failed = 'FAILED',
  Healthy = 'HEALTHY',
  Stopped = 'STOPPED'
}

export type ServiceVariable = {
  __typename?: 'ServiceVariable';
  databaseRef?: Maybe<DatabaseRef>;
  fromShared: Scalars['Boolean']['output'];
  key: Scalars['String']['output'];
  value: Scalars['String']['output'];
};

export type ServiceVariableInput = {
  /** Reference to a database secret key. */
  databaseRef?: InputMaybe<DatabaseRefInput>;
  /** If true, value is resolved from the shared variable with the same key. */
  fromShared?: InputMaybe<Scalars['Boolean']['input']>;
  key: Scalars['String']['input'];
  /** Direct value. Required when no ref is set. */
  value?: InputMaybe<Scalars['String']['input']>;
};

export type SetEnvironmentResourcesInput = {
  cpuMillicores: Scalars['Int']['input'];
  diskMB: Scalars['Int']['input'];
  environment: Scalars['EnvironmentID']['input'];
  memoryMB: Scalars['Int']['input'];
  tier: ResourceTier;
};

export type SetServiceScalingInput = {
  autoscaling?: InputMaybe<AutoscalingInput>;
  replicas: Scalars['Int']['input'];
  service: Scalars['ServiceID']['input'];
};

export type Subscription = {
  __typename?: 'Subscription';
  buildLogs: Scalars['String']['output'];
  serviceLogs: ServiceLogEntry;
};


export type SubscriptionBuildLogsArgs = {
  id: Scalars['String']['input'];
};


export type SubscriptionServiceLogsArgs = {
  service: Scalars['ServiceID']['input'];
  tailLines?: InputMaybe<Scalars['Int']['input']>;
};

export enum SubscriptionStatus {
  Active = 'ACTIVE',
  Canceled = 'CANCELED',
  Incomplete = 'INCOMPLETE',
  PastDue = 'PAST_DUE',
  Trialing = 'TRIALING'
}

export enum SyncStatus {
  Degraded = 'DEGRADED',
  OutOfSync = 'OUT_OF_SYNC',
  Progressing = 'PROGRESSING',
  Synced = 'SYNCED',
  Unknown = 'UNKNOWN'
}

export enum TlsStatus {
  Active = 'ACTIVE',
  Error = 'ERROR',
  None = 'NONE',
  Provisioning = 'PROVISIONING'
}

export type UpdateMemberRoleInput = {
  role: WorkspaceRole;
  userId: Scalars['ID']['input'];
};

export type UpdateWorkspaceInput = {
  name: Scalars['String']['input'];
};

export type UsageSummary = {
  __typename?: 'UsageSummary';
  creditsCents: Scalars['Int']['output'];
  estimatedTotalCents: Scalars['Int']['output'];
  resourceCostCents: Scalars['Int']['output'];
};

export type User = {
  __typename?: 'User';
  avatarUrl: Scalars['String']['output'];
  email?: Maybe<Scalars['String']['output']>;
  name?: Maybe<Scalars['String']['output']>;
  workspaces: Array<WorkspaceMembership>;
};

export type Variable = {
  __typename?: 'Variable';
  key: Scalars['String']['output'];
  value: Scalars['String']['output'];
};

export type VariableInput = {
  key: Scalars['String']['input'];
  value: Scalars['String']['input'];
};

export type Volume = {
  __typename?: 'Volume';
  capacityBytes: Scalars['Int']['output'];
  id: Scalars['VolumeID']['output'];
  name: Scalars['String']['output'];
  requestedSize: Scalars['String']['output'];
  size: Scalars['String']['output'];
  usedBytes: Scalars['Int']['output'];
};

export type Workspace = {
  __typename?: 'Workspace';
  id: Scalars['ID']['output'];
  members: Array<WorkspaceMember>;
  name: Scalars['String']['output'];
  personal: Scalars['Boolean']['output'];
  suspended: Scalars['Boolean']['output'];
};

export type WorkspaceMember = {
  __typename?: 'WorkspaceMember';
  email: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  name?: Maybe<Scalars['String']['output']>;
  role: WorkspaceRole;
};

export type WorkspaceMembership = {
  __typename?: 'WorkspaceMembership';
  role: WorkspaceRole;
  workspace: Scalars['String']['output'];
};

export enum WorkspaceRole {
  Admin = 'ADMIN',
  User = 'USER'
}

export type WorkspacesQueryVariables = Exact<{ [key: string]: never; }>;


export type WorkspacesQuery = { __typename?: 'Query', workspaces: Array<{ __typename?: 'Workspace', id: string, name: string, personal: boolean }> };

export type ProjectsForNavQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectsForNavQuery = { __typename?: 'Query', projects: Array<{ __typename?: 'Project', id: string, name: string, environments: Array<{ __typename?: 'Environment', id: string, name: string, resourceTier: ResourceTier }> }> };

export type GitHubConnectedQueryVariables = Exact<{ [key: string]: never; }>;


export type GitHubConnectedQuery = { __typename?: 'Query', githubConnected: boolean };

export type GitHubSourcesQueryVariables = Exact<{ [key: string]: never; }>;


export type GitHubSourcesQuery = { __typename?: 'Query', githubSources: Array<{ __typename?: 'GitHubInstallation', id: string, accountLogin: string, accountAvatarUrl: string, accountType: GitHubAccountType }> };

export type GitHubRepositoriesQueryVariables = Exact<{
  installationId: Scalars['ID']['input'];
}>;


export type GitHubRepositoriesQuery = { __typename?: 'Query', githubRepositories: Array<{ __typename?: 'GitHubRepository', id: string, name: string, fullName: string, htmlUrl: string, defaultBranch: string, private: boolean }> };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject: { __typename?: 'Project', id: string, name: string, environments: Array<{ __typename?: 'Environment', id: string, name: string }> } };

export type AddServiceMutationVariables = Exact<{
  environmentId: Scalars['EnvironmentID']['input'];
  input: AddServiceInput;
}>;


export type AddServiceMutation = { __typename?: 'Mutation', addService: { __typename?: 'Service', id: string, name: string } };

export type DetectServicesQueryVariables = Exact<{
  installationId: Scalars['ID']['input'];
  repositoryUrl: Scalars['String']['input'];
}>;


export type DetectServicesQuery = { __typename?: 'Query', detectServices: Array<{ __typename?: 'DetectedService', name: string, language: string, framework: string, startCommand: string, suggestedPort: number }> };

export type SearchImagesQueryVariables = Exact<{
  query: Scalars['String']['input'];
}>;


export type SearchImagesQuery = { __typename?: 'Query', searchImages: Array<{ __typename?: 'ImageSearchResult', name: string, description: string, starCount: number, pullCount: number, official: boolean }> };

export type CreateDatabaseMutationVariables = Exact<{
  input: CreateDatabaseInput;
}>;


export type CreateDatabaseMutation = { __typename?: 'Mutation', createDatabase: { __typename?: 'Database', id: string, name: string, version: string, instances: number, size: string } };

export type CreateEnvironmentMutationVariables = Exact<{
  input: CreateEnvironmentInput;
}>;


export type CreateEnvironmentMutation = { __typename?: 'Mutation', createEnvironment: { __typename?: 'Environment', id: string, name: string, resourceTier: ResourceTier } };

export type CreateWorkspaceCheckoutMutationVariables = Exact<{
  input: CreateWorkspaceCheckoutInput;
}>;


export type CreateWorkspaceCheckoutMutation = { __typename?: 'Mutation', createWorkspaceCheckout: { __typename?: 'CheckoutSession', url: string } };

export type SharedVariablesQueryVariables = Exact<{
  environment: Scalars['EnvironmentID']['input'];
}>;


export type SharedVariablesQuery = { __typename?: 'Query', sharedVariables: Array<{ __typename?: 'Variable', key: string, value: string }> };

export type SetSharedVariablesMutationVariables = Exact<{
  environment: Scalars['EnvironmentID']['input'];
  variables: Array<VariableInput> | VariableInput;
}>;


export type SetSharedVariablesMutation = { __typename?: 'Mutation', setSharedVariables: boolean };

export type SubscriptionQueryVariables = Exact<{ [key: string]: never; }>;


export type SubscriptionQuery = { __typename?: 'Query', subscription?: { __typename?: 'BillingSubscription', plan?: Plan | null, status: SubscriptionStatus, currentPeriodEnd: string, creditAmountCents: number, creditExpiry?: string | null, hasPaymentMethod: boolean } | null };

export type UsageSummaryQueryVariables = Exact<{ [key: string]: never; }>;


export type UsageSummaryQuery = { __typename?: 'Query', usageSummary?: { __typename?: 'UsageSummary', resourceCostCents: number, creditsCents: number, estimatedTotalCents: number } | null };

export type CreatePlanCheckoutMutationVariables = Exact<{
  plan: Plan;
}>;


export type CreatePlanCheckoutMutation = { __typename?: 'Mutation', createPlanCheckout: { __typename?: 'CheckoutSession', url: string } };

export type DatabaseCredentialsQueryVariables = Exact<{
  database: Scalars['DatabaseID']['input'];
}>;


export type DatabaseCredentialsQuery = { __typename?: 'Query', databaseCredentials: { __typename?: 'DatabaseCredentials', host: string, port: string, dbname: string, user: string, password: string, uri: string } };

export type ExecuteQueryMutationVariables = Exact<{
  database: Scalars['DatabaseID']['input'];
  query: Scalars['String']['input'];
}>;


export type ExecuteQueryMutation = { __typename?: 'Mutation', executeQuery: { __typename?: 'QueryResult', columns: Array<string>, rows: Array<Array<string | null> | null>, affectedRows: number } };

export type DeleteDatabaseMutationVariables = Exact<{
  database: Scalars['DatabaseID']['input'];
}>;


export type DeleteDatabaseMutation = { __typename?: 'Mutation', deleteDatabase: boolean };

export type DatabaseTablesQueryVariables = Exact<{
  database: Scalars['DatabaseID']['input'];
}>;


export type DatabaseTablesQuery = { __typename?: 'Query', databaseTables: Array<{ __typename?: 'DatabaseTable', name: string, schema: string, estimatedRows: number, columns: Array<{ __typename?: 'DatabaseColumn', name: string, type: string, nullable: boolean, primaryKey: boolean }> }> };

export type DatabaseTableDataQueryVariables = Exact<{
  database: Scalars['DatabaseID']['input'];
  table: Scalars['String']['input'];
  schema?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
}>;


export type DatabaseTableDataQuery = { __typename?: 'Query', databaseTableData: { __typename?: 'DatabaseTableData', columns: Array<string>, rows: Array<Array<string | null> | null>, totalEstimatedRows: number } };

export type RemoveServiceMutationVariables = Exact<{
  service: Scalars['ServiceID']['input'];
}>;


export type RemoveServiceMutation = { __typename?: 'Mutation', removeService: boolean };

export type SetCustomStartCommandMutationVariables = Exact<{
  service: Scalars['ServiceID']['input'];
  command: Scalars['String']['input'];
}>;


export type SetCustomStartCommandMutation = { __typename?: 'Mutation', setCustomStartCommand: { __typename?: 'Service', id: string } };

export type GenerateDomainMutationVariables = Exact<{
  service: Scalars['ServiceID']['input'];
}>;


export type GenerateDomainMutation = { __typename?: 'Mutation', generateDomain: { __typename?: 'Service', id: string, endpoints: Array<{ __typename?: 'Endpoint', host: string, port: number, type: EndpointType, protocol: Protocol, dns: { __typename?: 'DnsState', status: DnsStatus, requiredRecords: Array<{ __typename?: 'DnsRecord', type: DnsRecordType, host: string, value: string }> } }> } };

export type AddCustomDomainMutationVariables = Exact<{
  service: Scalars['ServiceID']['input'];
  hostname: Scalars['String']['input'];
}>;


export type AddCustomDomainMutation = { __typename?: 'Mutation', addCustomDomain: { __typename?: 'Service', id: string, endpoints: Array<{ __typename?: 'Endpoint', host: string, port: number, type: EndpointType, protocol: Protocol, dns: { __typename?: 'DnsState', status: DnsStatus, requiredRecords: Array<{ __typename?: 'DnsRecord', type: DnsRecordType, host: string, value: string }> } }> } };

export type RemoveDomainMutationVariables = Exact<{
  service: Scalars['ServiceID']['input'];
  hostname: Scalars['String']['input'];
}>;


export type RemoveDomainMutation = { __typename?: 'Mutation', removeDomain: { __typename?: 'Service', id: string, endpoints: Array<{ __typename?: 'Endpoint', host: string, port: number, type: EndpointType, protocol: Protocol, dns: { __typename?: 'DnsState', status: DnsStatus, requiredRecords: Array<{ __typename?: 'DnsRecord', type: DnsRecordType, host: string, value: string }> } }> } };

export type SetServiceScalingMutationVariables = Exact<{
  input: SetServiceScalingInput;
}>;


export type SetServiceScalingMutation = { __typename?: 'Mutation', setServiceScaling: { __typename?: 'Service', id: string, replicas: { __typename?: 'ReplicaCount', desired: number, ready: number }, autoscaling?: { __typename?: 'AutoscalingSettings', minReplicas: number, maxReplicas: number, targetCpu: number } | null } };

export type SetServicePortMutationVariables = Exact<{
  service: Scalars['ServiceID']['input'];
  port?: InputMaybe<Scalars['Int']['input']>;
}>;


export type SetServicePortMutation = { __typename?: 'Mutation', setServicePort: { __typename?: 'Service', id: string, port: number } };

export type ServiceVariablesQueryVariables = Exact<{
  service: Scalars['ServiceID']['input'];
}>;


export type ServiceVariablesQuery = { __typename?: 'Query', serviceVariables: Array<{ __typename?: 'ServiceVariable', key: string, value: string, fromShared: boolean, databaseRef?: { __typename?: 'DatabaseRef', database: string, key: string } | null }> };

export type SetServiceVariablesMutationVariables = Exact<{
  service: Scalars['ServiceID']['input'];
  variables: Array<ServiceVariableInput> | ServiceVariableInput;
}>;


export type SetServiceVariablesMutation = { __typename?: 'Mutation', setServiceVariables: boolean };

export type BuildLogsSubscriptionVariables = Exact<{
  id: Scalars['String']['input'];
}>;


export type BuildLogsSubscription = { __typename?: 'Subscription', buildLogs: string };

export type CanvasServiceBuildsQueryVariables = Exact<{
  id: Scalars['ServiceID']['input'];
}>;


export type CanvasServiceBuildsQuery = { __typename?: 'Query', service: { __typename?: 'Service', id: string, builds: Array<{ __typename?: 'Build', id: string, status: BuildStatus, startedAt: string, finishedAt?: string | null }> } };

export type DeployMutationVariables = Exact<{
  service: Scalars['ServiceID']['input'];
  gitRef?: InputMaybe<Scalars['String']['input']>;
}>;


export type DeployMutation = { __typename?: 'Mutation', deploy: { __typename?: 'Build', id: string, status: BuildStatus, startedAt: string, finishedAt?: string | null } };

export type BuildStatusQueryVariables = Exact<{
  id: Scalars['String']['input'];
}>;


export type BuildStatusQuery = { __typename?: 'Query', build: { __typename?: 'Build', id: string, status: BuildStatus, startedAt: string, finishedAt?: string | null } };

export type ServiceLogsSubscriptionVariables = Exact<{
  service: Scalars['ServiceID']['input'];
  tailLines?: InputMaybe<Scalars['Int']['input']>;
}>;


export type ServiceLogsSubscription = { __typename?: 'Subscription', serviceLogs: { __typename?: 'ServiceLogEntry', line: string, pod: string } };

export type WorkspaceQueryVariables = Exact<{ [key: string]: never; }>;


export type WorkspaceQuery = { __typename?: 'Query', workspace: { __typename?: 'Workspace', id: string, name: string, personal: boolean, suspended: boolean, members: Array<{ __typename?: 'WorkspaceMember', id: string, email: string, name?: string | null, role: WorkspaceRole }> } };

export type CompleteWorkspaceCheckoutMutationVariables = Exact<{
  sessionId: Scalars['String']['input'];
}>;


export type CompleteWorkspaceCheckoutMutation = { __typename?: 'Mutation', completeWorkspaceCheckout: { __typename?: 'Workspace', id: string, name: string, personal: boolean } };

export type EnvironmentQueryVariables = Exact<{
  environment: Scalars['EnvironmentID']['input'];
}>;


export type EnvironmentQuery = { __typename?: 'Query', environment: { __typename?: 'Environment', id: string, name: string, resourceTier: ResourceTier, services: Array<{ __typename?: 'Service', id: string, name: string, status: ServiceStatus, port: number, sourceUrl: string, contextPath: string, command: string, defaultCommand: string, lastDeployedAt?: string | null, createdAt: string, replicas: { __typename?: 'ReplicaCount', desired: number, ready: number }, autoscaling?: { __typename?: 'AutoscalingSettings', minReplicas: number, maxReplicas: number, targetCpu: number } | null, endpoints: Array<{ __typename?: 'Endpoint', host: string, port: number, protocol: Protocol, type: EndpointType, tls: TlsStatus, dns: { __typename?: 'DnsState', status: DnsStatus, requiredRecords: Array<{ __typename?: 'DnsRecord', type: DnsRecordType, host: string, value: string }> } }>, resources: { __typename?: 'Resources', cpu: string, memory: string }, activeDeployment?: { __typename?: 'Deployment', id: string, image: string, imageDigest?: string | null, commit: string, commitMessage: string, ref: string, status: DeploymentStatus, createdAt: string, deployedBy: string } | null, deployments: Array<{ __typename?: 'Deployment', id: string, image: string, imageDigest?: string | null, commit: string, commitMessage: string, ref: string, status: DeploymentStatus, createdAt: string, deployedBy: string }>, builds: Array<{ __typename?: 'Build', id: string, status: BuildStatus, startedAt: string, finishedAt?: string | null }> }>, databases: Array<{ __typename?: 'Database', id: string, name: string, version: string, instances: number, status: DatabaseStatus, size: string, createdAt: string }> } };

export type ProjectEnvironmentsQueryVariables = Exact<{
  id: Scalars['ProjectID']['input'];
}>;


export type ProjectEnvironmentsQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, environments: Array<{ __typename?: 'Environment', id: string, name: string, resourceTier: ResourceTier }> } };

export type CompletePlanCheckoutMutationVariables = Exact<{
  sessionId: Scalars['String']['input'];
}>;


export type CompletePlanCheckoutMutation = { __typename?: 'Mutation', completePlanCheckout: { __typename?: 'BillingSubscription', plan?: Plan | null, status: SubscriptionStatus, currentPeriodEnd: string, creditAmountCents: number, hasPaymentMethod: boolean } };

export type ProjectPageQueryVariables = Exact<{
  id: Scalars['ProjectID']['input'];
}>;


export type ProjectPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, environments: Array<{ __typename?: 'Environment', id: string, name: string, resourceTier: ResourceTier }> } };

export type ProjectSettingsQueryVariables = Exact<{
  id: Scalars['ProjectID']['input'];
}>;


export type ProjectSettingsQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, environments: Array<{ __typename?: 'Environment', id: string, name: string, resourceTier: ResourceTier }> } };

export type DeleteProjectMutationVariables = Exact<{
  id: Scalars['ProjectID']['input'];
}>;


export type DeleteProjectMutation = { __typename?: 'Mutation', deleteProject: boolean };

export type DeleteEnvironmentMutationVariables = Exact<{
  environment: Scalars['EnvironmentID']['input'];
}>;


export type DeleteEnvironmentMutation = { __typename?: 'Mutation', deleteEnvironment: boolean };

export type EnvironmentResourcesQueryVariables = Exact<{
  environment: Scalars['EnvironmentID']['input'];
}>;


export type EnvironmentResourcesQuery = { __typename?: 'Query', environmentResources?: { __typename?: 'EnvironmentResources', tier: ResourceTier, allocation: { __typename?: 'ResourceAllocation', cpuMillicores: number, memoryMB: number, diskMB: number } } | null };

export type SetEnvironmentResourcesMutationVariables = Exact<{
  input: SetEnvironmentResourcesInput;
}>;


export type SetEnvironmentResourcesMutation = { __typename?: 'Mutation', setEnvironmentResources: { __typename?: 'Environment', id: string, resourceTier: ResourceTier } };

export type ProjectsQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectsQuery = { __typename?: 'Query', projects: Array<{ __typename?: 'Project', id: string, name: string, environments: Array<{ __typename?: 'Environment', id: string, name: string, resourceTier: ResourceTier, services: Array<{ __typename?: 'Service', id: string, name: string, sourceUrl: string }> }> }> };

export type UpdateWorkspaceMutationVariables = Exact<{
  input: UpdateWorkspaceInput;
}>;


export type UpdateWorkspaceMutation = { __typename?: 'Mutation', updateWorkspace: { __typename?: 'Workspace', id: string, name: string } };

export type DeleteWorkspaceMutationVariables = Exact<{ [key: string]: never; }>;


export type DeleteWorkspaceMutation = { __typename?: 'Mutation', deleteWorkspace: boolean };

export type InviteMemberMutationVariables = Exact<{
  input: InviteMemberInput;
}>;


export type InviteMemberMutation = { __typename?: 'Mutation', inviteMember: { __typename?: 'WorkspaceMember', id: string, email: string, name?: string | null, role: WorkspaceRole } };

export type RemoveMemberMutationVariables = Exact<{
  userId: Scalars['ID']['input'];
}>;


export type RemoveMemberMutation = { __typename?: 'Mutation', removeMember: boolean };

export type UpdateMemberRoleMutationVariables = Exact<{
  input: UpdateMemberRoleInput;
}>;


export type UpdateMemberRoleMutation = { __typename?: 'Mutation', updateMemberRole: { __typename?: 'WorkspaceMember', id: string, email: string, name?: string | null, role: WorkspaceRole } };

export type ChangePlanMutationVariables = Exact<{
  plan: Plan;
}>;


export type ChangePlanMutation = { __typename?: 'Mutation', changePlan: { __typename?: 'BillingSubscription', plan?: Plan | null, status: SubscriptionStatus, currentPeriodEnd: string, creditAmountCents: number, creditExpiry?: string | null } };

export type BillingPortalUrlMutationVariables = Exact<{ [key: string]: never; }>;


export type BillingPortalUrlMutation = { __typename?: 'Mutation', billingPortalUrl: { __typename?: 'BillingPortalUrl', url: string } };


export const WorkspacesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Workspaces"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspaces"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}}]}}]}}]} as unknown as DocumentNode<WorkspacesQuery, WorkspacesQueryVariables>;
export const ProjectsForNavDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ProjectsForNav"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"projects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]}}]} as unknown as DocumentNode<ProjectsForNavQuery, ProjectsForNavQueryVariables>;
export const GitHubConnectedDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubConnected"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubConnected"}}]}}]} as unknown as DocumentNode<GitHubConnectedQuery, GitHubConnectedQueryVariables>;
export const GitHubSourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubSources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubSources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"accountLogin"}},{"kind":"Field","name":{"kind":"Name","value":"accountAvatarUrl"}},{"kind":"Field","name":{"kind":"Name","value":"accountType"}}]}}]}}]} as unknown as DocumentNode<GitHubSourcesQuery, GitHubSourcesQueryVariables>;
export const GitHubRepositoriesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubRepositories"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"installationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubRepositories"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"installationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"installationId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"fullName"}},{"kind":"Field","name":{"kind":"Name","value":"htmlUrl"}},{"kind":"Field","name":{"kind":"Name","value":"defaultBranch"}},{"kind":"Field","name":{"kind":"Name","value":"private"}}]}}]}}]} as unknown as DocumentNode<GitHubRepositoriesQuery, GitHubRepositoriesQueryVariables>;
export const CreateProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateProject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateProjectInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createProject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]}}]} as unknown as DocumentNode<CreateProjectMutation, CreateProjectMutationVariables>;
export const AddServiceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddService"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environmentId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AddServiceInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addService"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environmentId"}}},{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]} as unknown as DocumentNode<AddServiceMutation, AddServiceMutationVariables>;
export const DetectServicesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DetectServices"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"installationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"repositoryUrl"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"detectServices"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"installationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"installationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"repositoryUrl"},"value":{"kind":"Variable","name":{"kind":"Name","value":"repositoryUrl"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"language"}},{"kind":"Field","name":{"kind":"Name","value":"framework"}},{"kind":"Field","name":{"kind":"Name","value":"startCommand"}},{"kind":"Field","name":{"kind":"Name","value":"suggestedPort"}}]}}]}}]} as unknown as DocumentNode<DetectServicesQuery, DetectServicesQueryVariables>;
export const SearchImagesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SearchImages"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"query"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"searchImages"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"query"},"value":{"kind":"Variable","name":{"kind":"Name","value":"query"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"description"}},{"kind":"Field","name":{"kind":"Name","value":"starCount"}},{"kind":"Field","name":{"kind":"Name","value":"pullCount"}},{"kind":"Field","name":{"kind":"Name","value":"official"}}]}}]}}]} as unknown as DocumentNode<SearchImagesQuery, SearchImagesQueryVariables>;
export const CreateDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateDatabaseInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"instances"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<CreateDatabaseMutation, CreateDatabaseMutationVariables>;
export const CreateEnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateEnvironment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateEnvironmentInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createEnvironment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]} as unknown as DocumentNode<CreateEnvironmentMutation, CreateEnvironmentMutationVariables>;
export const CreateWorkspaceCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateWorkspaceCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateWorkspaceCheckoutInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createWorkspaceCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<CreateWorkspaceCheckoutMutation, CreateWorkspaceCheckoutMutationVariables>;
export const SharedVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SharedVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"sharedVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]} as unknown as DocumentNode<SharedVariablesQuery, SharedVariablesQueryVariables>;
export const SetSharedVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetSharedVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"variables"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VariableInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setSharedVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"variables"},"value":{"kind":"Variable","name":{"kind":"Name","value":"variables"}}}]}]}}]} as unknown as DocumentNode<SetSharedVariablesMutation, SetSharedVariablesMutationVariables>;
export const SubscriptionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditExpiry"}},{"kind":"Field","name":{"kind":"Name","value":"hasPaymentMethod"}}]}}]}}]} as unknown as DocumentNode<SubscriptionQuery, SubscriptionQueryVariables>;
export const UsageSummaryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"UsageSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"usageSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"resourceCostCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditsCents"}},{"kind":"Field","name":{"kind":"Name","value":"estimatedTotalCents"}}]}}]}}]} as unknown as DocumentNode<UsageSummaryQuery, UsageSummaryQueryVariables>;
export const CreatePlanCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreatePlanCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"plan"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Plan"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createPlanCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"plan"},"value":{"kind":"Variable","name":{"kind":"Name","value":"plan"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<CreatePlanCheckoutMutation, CreatePlanCheckoutMutationVariables>;
export const DatabaseCredentialsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseCredentials"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseCredentials"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"dbname"}},{"kind":"Field","name":{"kind":"Name","value":"user"}},{"kind":"Field","name":{"kind":"Name","value":"password"}},{"kind":"Field","name":{"kind":"Name","value":"uri"}}]}}]}}]} as unknown as DocumentNode<DatabaseCredentialsQuery, DatabaseCredentialsQueryVariables>;
export const ExecuteQueryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ExecuteQuery"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"query"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"executeQuery"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"query"},"value":{"kind":"Variable","name":{"kind":"Name","value":"query"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"columns"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"affectedRows"}}]}}]}}]} as unknown as DocumentNode<ExecuteQueryMutation, ExecuteQueryMutationVariables>;
export const DeleteDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}]}]}}]} as unknown as DocumentNode<DeleteDatabaseMutation, DeleteDatabaseMutationVariables>;
export const DatabaseTablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseTables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseTables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"schema"}},{"kind":"Field","name":{"kind":"Name","value":"estimatedRows"}},{"kind":"Field","name":{"kind":"Name","value":"columns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"nullable"}},{"kind":"Field","name":{"kind":"Name","value":"primaryKey"}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseTablesQuery, DatabaseTablesQueryVariables>;
export const DatabaseTableDataDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseTableData"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"table"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"schema"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseTableData"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"table"},"value":{"kind":"Variable","name":{"kind":"Name","value":"table"}}},{"kind":"Argument","name":{"kind":"Name","value":"schema"},"value":{"kind":"Variable","name":{"kind":"Name","value":"schema"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"columns"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"totalEstimatedRows"}}]}}]}}]} as unknown as DocumentNode<DatabaseTableDataQuery, DatabaseTableDataQueryVariables>;
export const RemoveServiceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveService"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeService"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}]}]}}]} as unknown as DocumentNode<RemoveServiceMutation, RemoveServiceMutationVariables>;
export const SetCustomStartCommandDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetCustomStartCommand"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"command"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setCustomStartCommand"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"command"},"value":{"kind":"Variable","name":{"kind":"Name","value":"command"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}}]}}]}}]} as unknown as DocumentNode<SetCustomStartCommandMutation, SetCustomStartCommandMutationVariables>;
export const GenerateDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"GenerateDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"generateDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<GenerateDomainMutation, GenerateDomainMutationVariables>;
export const AddCustomDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddCustomDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addCustomDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"hostname"},"value":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<AddCustomDomainMutation, AddCustomDomainMutationVariables>;
export const RemoveDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"hostname"},"value":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<RemoveDomainMutation, RemoveDomainMutationVariables>;
export const SetServiceScalingDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceScaling"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"SetServiceScalingInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceScaling"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"autoscaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"minReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"maxReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"targetCpu"}}]}}]}}]}}]} as unknown as DocumentNode<SetServiceScalingMutation, SetServiceScalingMutationVariables>;
export const SetServicePortDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServicePort"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"port"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServicePort"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"port"},"value":{"kind":"Variable","name":{"kind":"Name","value":"port"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"port"}}]}}]}}]} as unknown as DocumentNode<SetServicePortMutation, SetServicePortMutationVariables>;
export const ServiceVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ServiceVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serviceVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"value"}},{"kind":"Field","name":{"kind":"Name","value":"fromShared"}},{"kind":"Field","name":{"kind":"Name","value":"databaseRef"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"}},{"kind":"Field","name":{"kind":"Name","value":"key"}}]}}]}}]}}]} as unknown as DocumentNode<ServiceVariablesQuery, ServiceVariablesQueryVariables>;
export const SetServiceVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"variables"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceVariableInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"variables"},"value":{"kind":"Variable","name":{"kind":"Name","value":"variables"}}}]}]}}]} as unknown as DocumentNode<SetServiceVariablesMutation, SetServiceVariablesMutationVariables>;
export const BuildLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"BuildLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"buildLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<BuildLogsSubscription, BuildLogsSubscriptionVariables>;
export const CanvasServiceBuildsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"CanvasServiceBuilds"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"builds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}}]}}]}}]} as unknown as DocumentNode<CanvasServiceBuildsQuery, CanvasServiceBuildsQueryVariables>;
export const DeployDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Deploy"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"gitRef"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deploy"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"gitRef"},"value":{"kind":"Variable","name":{"kind":"Name","value":"gitRef"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}}]}}]} as unknown as DocumentNode<DeployMutation, DeployMutationVariables>;
export const BuildStatusDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"BuildStatus"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"build"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}}]}}]} as unknown as DocumentNode<BuildStatusQuery, BuildStatusQueryVariables>;
export const ServiceLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"ServiceLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"tailLines"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serviceLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"tailLines"},"value":{"kind":"Variable","name":{"kind":"Name","value":"tailLines"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"line"}},{"kind":"Field","name":{"kind":"Name","value":"pod"}}]}}]}}]} as unknown as DocumentNode<ServiceLogsSubscription, ServiceLogsSubscriptionVariables>;
export const WorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}},{"kind":"Field","name":{"kind":"Name","value":"suspended"}},{"kind":"Field","name":{"kind":"Name","value":"members"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]}}]} as unknown as DocumentNode<WorkspaceQuery, WorkspaceQueryVariables>;
export const CompleteWorkspaceCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CompleteWorkspaceCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"completeWorkspaceCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"sessionId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}}]}}]}}]} as unknown as DocumentNode<CompleteWorkspaceCheckoutMutation, CompleteWorkspaceCheckoutMutationVariables>;
export const EnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Environment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"EnvironmentID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"desired"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}}]}},{"kind":"Field","name":{"kind":"Name","value":"autoscaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"minReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"maxReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"targetCpu"}}]}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"endpoints"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"protocol"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"dns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"requiredRecords"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"tls"}}]}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}},{"kind":"Field","name":{"kind":"Name","value":"contextPath"}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpu"}},{"kind":"Field","name":{"kind":"Name","value":"memory"}}]}},{"kind":"Field","name":{"kind":"Name","value":"command"}},{"kind":"Field","name":{"kind":"Name","value":"defaultCommand"}},{"kind":"Field","name":{"kind":"Name","value":"activeDeployment"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"imageDigest"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"commitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"deployedBy"}}]}},{"kind":"Field","name":{"kind":"Name","value":"deployments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"imageDigest"}},{"kind":"Field","name":{"kind":"Name","value":"commit"}},{"kind":"Field","name":{"kind":"Name","value":"commitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"ref"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"deployedBy"}}]}},{"kind":"Field","name":{"kind":"Name","value":"builds"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"finishedAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"lastDeployedAt"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"instances"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentQuery, EnvironmentQueryVariables>;
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