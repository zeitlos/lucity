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
  Duration: { input: any; output: any; }
  Time: { input: any; output: any; }
};

export type AddCustomDomainInput = {
  environment: Scalars['String']['input'];
  hostname: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
};

export type AddServiceInput = {
  contextPath?: InputMaybe<Scalars['String']['input']>;
  /** Custom command to start the service (e.g. "npm run start", "stress --cpu 2 --timeout 60s"). Overrides the image entrypoint. */
  customStartCommand?: InputMaybe<Scalars['String']['input']>;
  environment: Scalars['String']['input'];
  framework?: InputMaybe<Scalars['String']['input']>;
  /** External container image reference (e.g. nginx:1.25, ghcr.io/foo/bar:v1). Skips the build step. */
  image?: InputMaybe<Scalars['String']['input']>;
  /** GitHub App installation ID. Required when repository is set. */
  installationId?: InputMaybe<Scalars['ID']['input']>;
  /** Service name. If omitted when image is set, derived from the image (e.g. nginx:1.25 → nginx). */
  name?: InputMaybe<Scalars['String']['input']>;
  /** Container port. If omitted when image is set, uses well-known defaults (e.g. redis → 6379). */
  port?: InputMaybe<Scalars['Int']['input']>;
  projectId: Scalars['ID']['input'];
  /** GitHub repository in owner/repo format (e.g. "acme/myapp"). Requires installationId. The clone URL is constructed server-side. */
  repository?: InputMaybe<Scalars['String']['input']>;
  /** Auto-detected start command from the build system (e.g. railpack). Stored for UI display as the default. */
  startCommand?: InputMaybe<Scalars['String']['input']>;
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

export type CheckoutSession = {
  __typename?: 'CheckoutSession';
  url: Scalars['String']['output'];
};

export type CreateDatabaseInput = {
  instances?: InputMaybe<Scalars['Int']['input']>;
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  size?: InputMaybe<Scalars['String']['input']>;
  version?: InputMaybe<Scalars['String']['input']>;
};

export type CreateEnvironmentInput = {
  fromEnvironment?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
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

export type CreateWorkspaceInput = {
  id: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export type Database = {
  __typename?: 'Database';
  instances: Scalars['Int']['output'];
  name: Scalars['String']['output'];
  size: Scalars['String']['output'];
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

export type DatabaseInstance = {
  __typename?: 'DatabaseInstance';
  environment: Scalars['String']['output'];
  instances: Scalars['Int']['output'];
  name: Scalars['String']['output'];
  ready: Scalars['Boolean']['output'];
  size: Scalars['String']['output'];
  version: Scalars['String']['output'];
  volume?: Maybe<Volume>;
};

export type DatabaseQueryInput = {
  database: Scalars['String']['input'];
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  query: Scalars['String']['input'];
};

/** A reference to a CNPG database secret key (resolved at pod startup via secretKeyRef). */
export type DatabaseRef = {
  __typename?: 'DatabaseRef';
  database: Scalars['String']['output'];
  key: Scalars['String']['output'];
};

export type DatabaseRefInput = {
  database: Scalars['String']['input'];
  key: Scalars['String']['input'];
};

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

export type DeployInput = {
  environment: Scalars['String']['input'];
  gitRef?: InputMaybe<Scalars['String']['input']>;
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
};

export enum DeployPhase {
  Building = 'BUILDING',
  Cloning = 'CLONING',
  Deploying = 'DEPLOYING',
  Failed = 'FAILED',
  Pushing = 'PUSHING',
  Queued = 'QUEUED',
  Succeeded = 'SUCCEEDED'
}

export type DeployRun = {
  __typename?: 'DeployRun';
  buildId?: Maybe<Scalars['String']['output']>;
  digest?: Maybe<Scalars['String']['output']>;
  error?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  imageRef?: Maybe<Scalars['String']['output']>;
  phase: DeployPhase;
  /** Rollout health status from the deployment system. */
  rolloutHealth?: Maybe<SyncStatus>;
  /** Rollout status detail, e.g. ImagePullBackOff, CrashLoopBackOff. */
  rolloutMessage?: Maybe<Scalars['String']['output']>;
  /** When the deploy operation started. */
  startedAt?: Maybe<Scalars['Time']['output']>;
};

export type Deployment = {
  __typename?: 'Deployment';
  active: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  imageTag: Scalars['String']['output'];
  message?: Maybe<Scalars['String']['output']>;
  revision?: Maybe<Scalars['String']['output']>;
  /** First line of the source commit message (fetched from GitHub). */
  sourceCommitMessage?: Maybe<Scalars['String']['output']>;
  /** URL to the source commit on GitHub. */
  sourceUrl?: Maybe<Scalars['String']['output']>;
  timestamp?: Maybe<Scalars['Time']['output']>;
};

export type DetectedService = {
  __typename?: 'DetectedService';
  framework: Scalars['String']['output'];
  language: Scalars['String']['output'];
  name: Scalars['String']['output'];
  startCommand: Scalars['String']['output'];
  suggestedPort: Scalars['Int']['output'];
};

/** Result of a live DNS verification for a custom domain. */
export type DnsCheck = {
  __typename?: 'DnsCheck';
  /** The CNAME target found in DNS, if any. */
  cnameTarget?: Maybe<Scalars['String']['output']>;
  /** The expected CNAME target for this platform. */
  expectedTarget: Scalars['String']['output'];
  hostname: Scalars['String']['output'];
  /** Human-readable message explaining the current status. */
  message?: Maybe<Scalars['String']['output']>;
  /** Current DNS status after live lookup. */
  status: DnsStatus;
  /** TLS certificate provisioning status for custom domains. */
  tlsStatus?: Maybe<TlsStatus>;
};

export enum DnsStatus {
  Error = 'ERROR',
  Misconfigured = 'MISCONFIGURED',
  Pending = 'PENDING',
  Valid = 'VALID'
}

export type Domain = {
  __typename?: 'Domain';
  /** DNS resolution status. Always VALID for platform domains. Checked via DNS lookup for custom domains. */
  dnsStatus: DnsStatus;
  hostname: Scalars['String']['output'];
  /** TLS certificate status. NONE for platform domains (covered by wildcard). Checked via cert-manager for custom domains. */
  tlsStatus: TlsStatus;
  /** PLATFORM domains use wildcard DNS on the workload domain. CUSTOM domains require user DNS config. */
  type: DomainType;
};

export enum DomainType {
  Custom = 'CUSTOM',
  Platform = 'PLATFORM'
}

export type Environment = {
  __typename?: 'Environment';
  databases: Array<DatabaseInstance>;
  ephemeral: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  namespace: Scalars['String']['output'];
  resourceTier?: Maybe<ResourceTier>;
  services: Array<ServiceInstance>;
  syncStatus: SyncStatus;
};

export type EnvironmentResources = {
  __typename?: 'EnvironmentResources';
  allocation: ResourceAllocation;
  tier: ResourceTier;
};

export type GenerateDomainInput = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
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
  /** Add a custom domain hostname. */
  addCustomDomain: Domain;
  addService: ServiceInstance;
  billingPortalUrl: BillingPortalUrl;
  changePlan: BillingSubscription;
  completePlanCheckout: BillingSubscription;
  completeWorkspaceCheckout: Workspace;
  createDatabase: Database;
  createEnvironment: Environment;
  createPlanCheckout: CheckoutSession;
  createProject: Project;
  createWorkspace: Workspace;
  createWorkspaceCheckout: CheckoutSession;
  deleteDatabase: Scalars['Boolean']['output'];
  deleteEnvironment: Scalars['Boolean']['output'];
  deleteProject: Scalars['Boolean']['output'];
  deleteWorkspace: Scalars['Boolean']['output'];
  deploy: DeployRun;
  executeQuery: QueryResult;
  /** Generate a platform domain ({service}-{env}.{workloadDomain}). Handles collisions with a suffix. */
  generateDomain: Domain;
  inviteMember: WorkspaceMember;
  promote: ServiceInstance;
  /** Remove any domain (platform or custom). */
  removeDomain: Scalars['Boolean']['output'];
  removeMember: Scalars['Boolean']['output'];
  removeService: Scalars['Boolean']['output'];
  /** Roll back to a previous image tag without rebuilding. */
  rollback: Scalars['Boolean']['output'];
  /** Set or clear the custom start command for a service in an environment. Empty string clears it. */
  setCustomStartCommand: Scalars['Boolean']['output'];
  setEnvironmentResources: EnvironmentResources;
  setServiceScaling: ScalingConfig;
  /** Replace all variables for a service in an environment. */
  setServiceVariables: Scalars['Boolean']['output'];
  /** Replace all shared variables for an environment. Propagates changes to services referencing them. */
  setSharedVariables: Scalars['Boolean']['output'];
  updateMemberRole: WorkspaceMember;
  updateWorkspace: Workspace;
};


export type MutationAddCustomDomainArgs = {
  input: AddCustomDomainInput;
};


export type MutationAddServiceArgs = {
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


export type MutationCreateWorkspaceArgs = {
  input: CreateWorkspaceInput;
};


export type MutationCreateWorkspaceCheckoutArgs = {
  input: CreateWorkspaceCheckoutInput;
};


export type MutationDeleteDatabaseArgs = {
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
};


export type MutationDeleteEnvironmentArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
};


export type MutationDeleteProjectArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeployArgs = {
  input: DeployInput;
};


export type MutationExecuteQueryArgs = {
  input: DatabaseQueryInput;
};


export type MutationGenerateDomainArgs = {
  input: GenerateDomainInput;
};


export type MutationInviteMemberArgs = {
  input: InviteMemberInput;
};


export type MutationPromoteArgs = {
  input: PromoteInput;
};


export type MutationRemoveDomainArgs = {
  input: RemoveDomainInput;
};


export type MutationRemoveMemberArgs = {
  userId: Scalars['ID']['input'];
};


export type MutationRemoveServiceArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
};


export type MutationRollbackArgs = {
  input: RollbackInput;
};


export type MutationSetCustomStartCommandArgs = {
  command: Scalars['String']['input'];
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
};


export type MutationSetEnvironmentResourcesArgs = {
  input: SetEnvironmentResourcesInput;
};


export type MutationSetServiceScalingArgs = {
  input: SetServiceScalingInput;
};


export type MutationSetServiceVariablesArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
  variables: Array<ServiceVariableInput>;
};


export type MutationSetSharedVariablesArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
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

export type PlatformConfig = {
  __typename?: 'PlatformConfig';
  /** CNAME target for custom domains. Empty if not configured. */
  domainTarget: Scalars['String']['output'];
  /** Load balancer IP address for A record configuration (apex domains). */
  ipAddress: Scalars['String']['output'];
  workloadDomain: Scalars['String']['output'];
};

export type Project = {
  __typename?: 'Project';
  createdAt?: Maybe<Scalars['Time']['output']>;
  databases: Array<Database>;
  environments: Array<Environment>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
};

export type PromoteInput = {
  fromEnvironment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
  tier?: InputMaybe<ResourceTier>;
  toEnvironment: Scalars['String']['input'];
};

export type Query = {
  __typename?: 'Query';
  activeDeployment?: Maybe<DeployRun>;
  /** Perform a live DNS check for a custom domain. Returns current CNAME status and whether it points to the correct target. */
  checkDnsStatus: DnsCheck;
  databaseCredentials: DatabaseCredentials;
  databaseTableData: DatabaseTableData;
  databaseTables: Array<DatabaseTable>;
  databases: Array<Database>;
  deployStatus?: Maybe<DeployRun>;
  /** Detect services in a GitHub repository. Repository must be in owner/repo format. */
  detectServices: Array<DetectedService>;
  environmentResources?: Maybe<EnvironmentResources>;
  /** Whether the current user has connected their GitHub account. */
  githubConnected: Scalars['Boolean']['output'];
  /** Repos from a specific installation. */
  githubRepositories: Array<GitHubRepository>;
  /** User's accessible GitHub App installations. Requires connected GitHub account. */
  githubSources: Array<GitHubInstallation>;
  me: User;
  platformConfig: PlatformConfig;
  project?: Maybe<Project>;
  projects: Array<Project>;
  /** Search Docker Hub for public container images. */
  searchImages: Array<ImageSearchResult>;
  serviceVariables: Array<ServiceVariable>;
  sharedVariables: Array<Variable>;
  subscription?: Maybe<BillingSubscription>;
  usageSummary?: Maybe<UsageSummary>;
  workspace: Workspace;
  workspaces: Array<Workspace>;
};


export type QueryActiveDeploymentArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
};


export type QueryCheckDnsStatusArgs = {
  hostname: Scalars['String']['input'];
};


export type QueryDatabaseCredentialsArgs = {
  database: Scalars['String']['input'];
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
};


export type QueryDatabaseTableDataArgs = {
  database: Scalars['String']['input'];
  environment: Scalars['String']['input'];
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
  projectId: Scalars['ID']['input'];
  schema?: InputMaybe<Scalars['String']['input']>;
  table: Scalars['String']['input'];
};


export type QueryDatabaseTablesArgs = {
  database: Scalars['String']['input'];
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
};


export type QueryDatabasesArgs = {
  projectId: Scalars['ID']['input'];
};


export type QueryDeployStatusArgs = {
  id: Scalars['ID']['input'];
};


export type QueryDetectServicesArgs = {
  installationId: Scalars['ID']['input'];
  repository: Scalars['String']['input'];
};


export type QueryEnvironmentResourcesArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
};


export type QueryGithubRepositoriesArgs = {
  installationId: Scalars['ID']['input'];
};


export type QueryProjectArgs = {
  id: Scalars['ID']['input'];
};


export type QuerySearchImagesArgs = {
  query: Scalars['String']['input'];
};


export type QueryServiceVariablesArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
};


export type QuerySharedVariablesArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
};

export type QueryResult = {
  __typename?: 'QueryResult';
  affectedRows: Scalars['Int']['output'];
  columns: Array<Scalars['String']['output']>;
  rows: Array<Maybe<Array<Maybe<Scalars['String']['output']>>>>;
};

export type RemoveDomainInput = {
  environment: Scalars['String']['input'];
  hostname: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
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

export enum Role {
  Admin = 'ADMIN',
  Anonymous = 'ANONYMOUS',
  User = 'USER'
}

export type RollbackInput = {
  environment: Scalars['String']['input'];
  /** Image tag to roll back to (typically a short git SHA). */
  imageTag: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
};

export type ScalingConfig = {
  __typename?: 'ScalingConfig';
  autoscaling?: Maybe<AutoscalingConfig>;
  replicas: Scalars['Int']['output'];
};

export type ServiceInstance = {
  __typename?: 'ServiceInstance';
  contextPath?: Maybe<Scalars['String']['output']>;
  customStartCommand?: Maybe<Scalars['String']['output']>;
  deployments: Array<Deployment>;
  domains: Array<Domain>;
  environment: Scalars['String']['output'];
  framework?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  imageTag: Scalars['String']['output'];
  /** Deploy automatically triggered when the service was added. Null for image-based services. */
  initialDeploy?: Maybe<DeployRun>;
  name: Scalars['String']['output'];
  port?: Maybe<Scalars['Int']['output']>;
  ready: Scalars['Boolean']['output'];
  replicas: Scalars['Int']['output'];
  /** Compute resources allocated to this service. Null if the service has no running deployment. */
  resources?: Maybe<ServiceResources>;
  scaling: ScalingConfig;
  sourceUrl?: Maybe<Scalars['String']['output']>;
  startCommand?: Maybe<Scalars['String']['output']>;
};

export type ServiceLogEntry = {
  __typename?: 'ServiceLogEntry';
  /** Log line text. Prefixed with [pod-suffix] when multiple replicas exist. */
  line: Scalars['String']['output'];
  /** Name of the pod that produced this line. */
  pod: Scalars['String']['output'];
};

/** A reference to another service's internal URL (computed by Helm template). */
export type ServiceRef = {
  __typename?: 'ServiceRef';
  service: Scalars['String']['output'];
};

export type ServiceRefInput = {
  service: Scalars['String']['input'];
};

/** Compute resource allocation for a service instance. */
export type ServiceResources = {
  __typename?: 'ServiceResources';
  /** CPU limit in millicores. */
  cpuLimitMillicores: Scalars['Int']['output'];
  /** CPU allocation in millicores (e.g. 250 = 0.25 vCPU). */
  cpuMillicores: Scalars['Int']['output'];
  /** Memory limit in megabytes. */
  memoryLimitMB: Scalars['Int']['output'];
  /** Memory allocation in megabytes. */
  memoryMB: Scalars['Int']['output'];
};

export type ServiceVariable = {
  __typename?: 'ServiceVariable';
  databaseRef?: Maybe<DatabaseRef>;
  fromShared: Scalars['Boolean']['output'];
  key: Scalars['String']['output'];
  serviceRef?: Maybe<ServiceRef>;
  value: Scalars['String']['output'];
};

export type ServiceVariableInput = {
  /** Reference to a database secret key. */
  databaseRef?: InputMaybe<DatabaseRefInput>;
  /** If true, value is resolved from the shared variable with the same key. */
  fromShared?: InputMaybe<Scalars['Boolean']['input']>;
  key: Scalars['String']['input'];
  /** Reference to another service's internal URL. */
  serviceRef?: InputMaybe<ServiceRefInput>;
  /** Direct value. Required when no ref is set. */
  value?: InputMaybe<Scalars['String']['input']>;
};

export type SetEnvironmentResourcesInput = {
  cpuMillicores: Scalars['Int']['input'];
  diskMB: Scalars['Int']['input'];
  environment: Scalars['String']['input'];
  memoryMB: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  tier: ResourceTier;
};

export type SetServiceScalingInput = {
  autoscaling?: InputMaybe<AutoscalingInput>;
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  replicas: Scalars['Int']['input'];
  service: Scalars['String']['input'];
};

export type Subscription = {
  __typename?: 'Subscription';
  /** Stream deploy log lines in real time. Emits existing lines then new lines as they arrive. */
  deployLogs: Scalars['String']['output'];
  /** Stream real-time stdout/stderr from running pods for a service. */
  serviceLogs: ServiceLogEntry;
};


export type SubscriptionDeployLogsArgs = {
  id: Scalars['ID']['input'];
};


export type SubscriptionServiceLogsArgs = {
  environment: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
  tailLines?: InputMaybe<Scalars['Int']['input']>;
};

export enum SubscriptionStatus {
  Active = 'ACTIVE',
  Canceled = 'CANCELED',
  Incomplete = 'INCOMPLETE',
  PastDue = 'PAST_DUE'
}

export enum SyncStatus {
  Degraded = 'DEGRADED',
  OutOfSync = 'OUT_OF_SYNC',
  Progressing = 'PROGRESSING',
  Synced = 'SYNCED',
  Unknown = 'UNKNOWN'
}

export enum TlsStatus {
  /** Certificate is active and TLS termination is working. */
  Active = 'ACTIVE',
  /** Certificate provisioning failed. */
  Error = 'ERROR',
  /** No certificate needed (platform domains use wildcard cert). */
  None = 'NONE',
  /** Certificate is being provisioned by cert-manager. */
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

export type SubscriptionQueryVariables = Exact<{ [key: string]: never; }>;


export type SubscriptionQuery = { __typename?: 'Query', subscription?: { __typename?: 'BillingSubscription', plan?: Plan | null, status: SubscriptionStatus, currentPeriodEnd: any, creditAmountCents: number, creditExpiry?: any | null, hasPaymentMethod: boolean } | null };

export type UsageSummaryQueryVariables = Exact<{ [key: string]: never; }>;


export type UsageSummaryQuery = { __typename?: 'Query', usageSummary?: { __typename?: 'UsageSummary', resourceCostCents: number, creditsCents: number, estimatedTotalCents: number } | null };

export type ChangePlanMutationVariables = Exact<{
  plan: Plan;
}>;


export type ChangePlanMutation = { __typename?: 'Mutation', changePlan: { __typename?: 'BillingSubscription', plan?: Plan | null, status: SubscriptionStatus, currentPeriodEnd: any, creditAmountCents: number, creditExpiry?: any | null } };

export type BillingPortalUrlMutationVariables = Exact<{ [key: string]: never; }>;


export type BillingPortalUrlMutation = { __typename?: 'Mutation', billingPortalUrl: { __typename?: 'BillingPortalUrl', url: string } };

export type CreatePlanCheckoutMutationVariables = Exact<{
  plan: Plan;
}>;


export type CreatePlanCheckoutMutation = { __typename?: 'Mutation', createPlanCheckout: { __typename?: 'CheckoutSession', url: string } };

export type CompletePlanCheckoutMutationVariables = Exact<{
  sessionId: Scalars['String']['input'];
}>;


export type CompletePlanCheckoutMutation = { __typename?: 'Mutation', completePlanCheckout: { __typename?: 'BillingSubscription', plan?: Plan | null, status: SubscriptionStatus, currentPeriodEnd: any, creditAmountCents: number, hasPaymentMethod: boolean } };

export type EnvironmentResourcesQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
}>;


export type EnvironmentResourcesQuery = { __typename?: 'Query', environmentResources?: { __typename?: 'EnvironmentResources', tier: ResourceTier, allocation: { __typename?: 'ResourceAllocation', cpuMillicores: number, memoryMB: number, diskMB: number } } | null };

export type SetEnvironmentResourcesMutationVariables = Exact<{
  input: SetEnvironmentResourcesInput;
}>;


export type SetEnvironmentResourcesMutation = { __typename?: 'Mutation', setEnvironmentResources: { __typename?: 'EnvironmentResources', tier: ResourceTier, allocation: { __typename?: 'ResourceAllocation', cpuMillicores: number, memoryMB: number, diskMB: number } } };

export type CreateDatabaseMutationVariables = Exact<{
  input: CreateDatabaseInput;
}>;


export type CreateDatabaseMutation = { __typename?: 'Mutation', createDatabase: { __typename?: 'Database', name: string, version: string, instances: number, size: string } };

export type DeleteDatabaseMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
}>;


export type DeleteDatabaseMutation = { __typename?: 'Mutation', deleteDatabase: boolean };

export type DatabaseTablesQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
  database: Scalars['String']['input'];
}>;


export type DatabaseTablesQuery = { __typename?: 'Query', databaseTables: Array<{ __typename?: 'DatabaseTable', name: string, schema: string, estimatedRows: number, columns: Array<{ __typename?: 'DatabaseColumn', name: string, type: string, nullable: boolean, primaryKey: boolean }> }> };

export type DatabaseTableDataQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
  database: Scalars['String']['input'];
  table: Scalars['String']['input'];
  schema?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
}>;


export type DatabaseTableDataQuery = { __typename?: 'Query', databaseTableData: { __typename?: 'DatabaseTableData', columns: Array<string>, rows: Array<Array<string | null> | null>, totalEstimatedRows: number } };

export type DatabaseCredentialsQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
  database: Scalars['String']['input'];
}>;


export type DatabaseCredentialsQuery = { __typename?: 'Query', databaseCredentials: { __typename?: 'DatabaseCredentials', host: string, port: string, dbname: string, user: string, password: string, uri: string } };

export type ExecuteQueryMutationVariables = Exact<{
  input: DatabaseQueryInput;
}>;


export type ExecuteQueryMutation = { __typename?: 'Mutation', executeQuery: { __typename?: 'QueryResult', columns: Array<string>, rows: Array<Array<string | null> | null>, affectedRows: number } };

export type GitHubConnectedQueryVariables = Exact<{ [key: string]: never; }>;


export type GitHubConnectedQuery = { __typename?: 'Query', githubConnected: boolean };

export type GitHubSourcesQueryVariables = Exact<{ [key: string]: never; }>;


export type GitHubSourcesQuery = { __typename?: 'Query', githubSources: Array<{ __typename?: 'GitHubInstallation', id: string, accountLogin: string, accountAvatarUrl: string, accountType: GitHubAccountType }> };

export type GitHubRepositoriesQueryVariables = Exact<{
  installationId: Scalars['ID']['input'];
}>;


export type GitHubRepositoriesQuery = { __typename?: 'Query', githubRepositories: Array<{ __typename?: 'GitHubRepository', id: string, name: string, fullName: string, htmlUrl: string, defaultBranch: string, private: boolean }> };

export type ProjectsQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectsQuery = { __typename?: 'Query', projects: Array<{ __typename?: 'Project', id: string, name: string, createdAt?: any | null, environments: Array<{ __typename?: 'Environment', id: string, name: string, syncStatus: SyncStatus, resourceTier?: ResourceTier | null, services: Array<{ __typename?: 'ServiceInstance', name: string, sourceUrl?: string | null }> }>, databases: Array<{ __typename?: 'Database', name: string, version: string }> }> };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject: { __typename?: 'Project', id: string, name: string } };

export type DeleteProjectMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteProjectMutation = { __typename?: 'Mutation', deleteProject: boolean };

export type CreateEnvironmentMutationVariables = Exact<{
  input: CreateEnvironmentInput;
}>;


export type CreateEnvironmentMutation = { __typename?: 'Mutation', createEnvironment: { __typename?: 'Environment', id: string, name: string, namespace: string, ephemeral: boolean, syncStatus: SyncStatus, resourceTier?: ResourceTier | null } };

export type DeleteEnvironmentMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
}>;


export type DeleteEnvironmentMutation = { __typename?: 'Mutation', deleteEnvironment: boolean };

export type SetServiceScalingMutationVariables = Exact<{
  input: SetServiceScalingInput;
}>;


export type SetServiceScalingMutation = { __typename?: 'Mutation', setServiceScaling: { __typename?: 'ScalingConfig', replicas: number, autoscaling?: { __typename?: 'AutoscalingConfig', enabled: boolean, minReplicas: number, maxReplicas: number, targetCPU: number } | null } };

export type ProjectQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type ProjectQuery = { __typename?: 'Query', project?: { __typename?: 'Project', id: string, name: string, createdAt?: any | null, environments: Array<{ __typename?: 'Environment', id: string, name: string, namespace: string, ephemeral: boolean, syncStatus: SyncStatus, resourceTier?: ResourceTier | null, services: Array<{ __typename?: 'ServiceInstance', id: string, name: string, environment: string, image: string, port?: number | null, framework?: string | null, startCommand?: string | null, sourceUrl?: string | null, contextPath?: string | null, customStartCommand?: string | null, imageTag: string, ready: boolean, replicas: number, scaling: { __typename?: 'ScalingConfig', replicas: number, autoscaling?: { __typename?: 'AutoscalingConfig', enabled: boolean, minReplicas: number, maxReplicas: number, targetCPU: number } | null }, resources?: { __typename?: 'ServiceResources', cpuMillicores: number, memoryMB: number, cpuLimitMillicores: number, memoryLimitMB: number } | null, domains: Array<{ __typename?: 'Domain', hostname: string, type: DomainType, dnsStatus: DnsStatus, tlsStatus: TlsStatus }>, deployments: Array<{ __typename?: 'Deployment', id: string, imageTag: string, active: boolean, timestamp?: any | null, revision?: string | null, message?: string | null, sourceCommitMessage?: string | null, sourceUrl?: string | null }> }>, databases: Array<{ __typename?: 'DatabaseInstance', name: string, environment: string, ready: boolean, instances: number, version: string, size: string, volume?: { __typename?: 'Volume', name: string, size: string, requestedSize: string, usedBytes: number, capacityBytes: number } | null }> }>, databases: Array<{ __typename?: 'Database', name: string, version: string, instances: number, size: string }> } | null };

export type SearchImagesQueryVariables = Exact<{
  query: Scalars['String']['input'];
}>;


export type SearchImagesQuery = { __typename?: 'Query', searchImages: Array<{ __typename?: 'ImageSearchResult', name: string, description: string, starCount: number, pullCount: number, official: boolean }> };

export type DetectServicesQueryVariables = Exact<{
  installationId: Scalars['ID']['input'];
  repository: Scalars['String']['input'];
}>;


export type DetectServicesQuery = { __typename?: 'Query', detectServices: Array<{ __typename?: 'DetectedService', name: string, language: string, framework: string, startCommand: string, suggestedPort: number }> };

export type AddServiceMutationVariables = Exact<{
  input: AddServiceInput;
}>;


export type AddServiceMutation = { __typename?: 'Mutation', addService: { __typename?: 'ServiceInstance', id: string, name: string, environment: string, image: string, port?: number | null, framework?: string | null, startCommand?: string | null, sourceUrl?: string | null, contextPath?: string | null, customStartCommand?: string | null, imageTag: string, initialDeploy?: { __typename?: 'DeployRun', id: string, phase: DeployPhase } | null } };

export type SetCustomStartCommandMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
  service: Scalars['String']['input'];
  command: Scalars['String']['input'];
}>;


export type SetCustomStartCommandMutation = { __typename?: 'Mutation', setCustomStartCommand: boolean };

export type RemoveServiceMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
  service: Scalars['String']['input'];
}>;


export type RemoveServiceMutation = { __typename?: 'Mutation', removeService: boolean };

export type DeployMutationVariables = Exact<{
  input: DeployInput;
}>;


export type DeployMutation = { __typename?: 'Mutation', deploy: { __typename?: 'DeployRun', id: string, phase: DeployPhase, buildId?: string | null, imageRef?: string | null, digest?: string | null, error?: string | null, startedAt?: any | null } };

export type DeployStatusQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeployStatusQuery = { __typename?: 'Query', deployStatus?: { __typename?: 'DeployRun', id: string, phase: DeployPhase, buildId?: string | null, imageRef?: string | null, digest?: string | null, error?: string | null, startedAt?: any | null, rolloutHealth?: SyncStatus | null, rolloutMessage?: string | null } | null };

export type ActiveDeploymentQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
  environment: Scalars['String']['input'];
}>;


export type ActiveDeploymentQuery = { __typename?: 'Query', activeDeployment?: { __typename?: 'DeployRun', id: string, phase: DeployPhase, buildId?: string | null, imageRef?: string | null, digest?: string | null, error?: string | null, startedAt?: any | null, rolloutHealth?: SyncStatus | null, rolloutMessage?: string | null } | null };

export type RollbackMutationVariables = Exact<{
  input: RollbackInput;
}>;


export type RollbackMutation = { __typename?: 'Mutation', rollback: boolean };

export type GenerateDomainMutationVariables = Exact<{
  input: GenerateDomainInput;
}>;


export type GenerateDomainMutation = { __typename?: 'Mutation', generateDomain: { __typename?: 'Domain', hostname: string, type: DomainType, dnsStatus: DnsStatus, tlsStatus: TlsStatus } };

export type AddCustomDomainMutationVariables = Exact<{
  input: AddCustomDomainInput;
}>;


export type AddCustomDomainMutation = { __typename?: 'Mutation', addCustomDomain: { __typename?: 'Domain', hostname: string, type: DomainType, dnsStatus: DnsStatus, tlsStatus: TlsStatus } };

export type RemoveDomainMutationVariables = Exact<{
  input: RemoveDomainInput;
}>;


export type RemoveDomainMutation = { __typename?: 'Mutation', removeDomain: boolean };

export type CheckDnsStatusQueryVariables = Exact<{
  hostname: Scalars['String']['input'];
}>;


export type CheckDnsStatusQuery = { __typename?: 'Query', checkDnsStatus: { __typename?: 'DnsCheck', hostname: string, status: DnsStatus, cnameTarget?: string | null, expectedTarget: string, message?: string | null, tlsStatus?: TlsStatus | null } };

export type PlatformConfigQueryVariables = Exact<{ [key: string]: never; }>;


export type PlatformConfigQuery = { __typename?: 'Query', platformConfig: { __typename?: 'PlatformConfig', workloadDomain: string, domainTarget: string, ipAddress: string } };

export type DeployLogsSubscriptionVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeployLogsSubscription = { __typename?: 'Subscription', deployLogs: string };

export type ServiceLogsSubscriptionVariables = Exact<{
  projectId: Scalars['ID']['input'];
  service: Scalars['String']['input'];
  environment: Scalars['String']['input'];
  tailLines?: InputMaybe<Scalars['Int']['input']>;
}>;


export type ServiceLogsSubscription = { __typename?: 'Subscription', serviceLogs: { __typename?: 'ServiceLogEntry', line: string, pod: string } };

export type SharedVariablesQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
}>;


export type SharedVariablesQuery = { __typename?: 'Query', sharedVariables: Array<{ __typename?: 'Variable', key: string, value: string }> };

export type SetSharedVariablesMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
  variables: Array<VariableInput> | VariableInput;
}>;


export type SetSharedVariablesMutation = { __typename?: 'Mutation', setSharedVariables: boolean };

export type ServiceVariablesQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
  service: Scalars['String']['input'];
}>;


export type ServiceVariablesQuery = { __typename?: 'Query', serviceVariables: Array<{ __typename?: 'ServiceVariable', key: string, value: string, fromShared: boolean, databaseRef?: { __typename?: 'DatabaseRef', database: string, key: string } | null, serviceRef?: { __typename?: 'ServiceRef', service: string } | null }> };

export type SetServiceVariablesMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  environment: Scalars['String']['input'];
  service: Scalars['String']['input'];
  variables: Array<ServiceVariableInput> | ServiceVariableInput;
}>;


export type SetServiceVariablesMutation = { __typename?: 'Mutation', setServiceVariables: boolean };

export type WorkspacesQueryVariables = Exact<{ [key: string]: never; }>;


export type WorkspacesQuery = { __typename?: 'Query', workspaces: Array<{ __typename?: 'Workspace', id: string, name: string, personal: boolean }> };

export type WorkspaceQueryVariables = Exact<{ [key: string]: never; }>;


export type WorkspaceQuery = { __typename?: 'Query', workspace: { __typename?: 'Workspace', id: string, name: string, personal: boolean, suspended: boolean, members: Array<{ __typename?: 'WorkspaceMember', id: string, email: string, name?: string | null, role: WorkspaceRole }> } };

export type CreateWorkspaceMutationVariables = Exact<{
  input: CreateWorkspaceInput;
}>;


export type CreateWorkspaceMutation = { __typename?: 'Mutation', createWorkspace: { __typename?: 'Workspace', id: string, name: string, personal: boolean } };

export type CreateWorkspaceCheckoutMutationVariables = Exact<{
  input: CreateWorkspaceCheckoutInput;
}>;


export type CreateWorkspaceCheckoutMutation = { __typename?: 'Mutation', createWorkspaceCheckout: { __typename?: 'CheckoutSession', url: string } };

export type CompleteWorkspaceCheckoutMutationVariables = Exact<{
  sessionId: Scalars['String']['input'];
}>;


export type CompleteWorkspaceCheckoutMutation = { __typename?: 'Mutation', completeWorkspaceCheckout: { __typename?: 'Workspace', id: string, name: string, personal: boolean } };

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


export const SubscriptionDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"subscription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditExpiry"}},{"kind":"Field","name":{"kind":"Name","value":"hasPaymentMethod"}}]}}]}}]} as unknown as DocumentNode<SubscriptionQuery, SubscriptionQueryVariables>;
export const UsageSummaryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"UsageSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"usageSummary"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"resourceCostCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditsCents"}},{"kind":"Field","name":{"kind":"Name","value":"estimatedTotalCents"}}]}}]}}]} as unknown as DocumentNode<UsageSummaryQuery, UsageSummaryQueryVariables>;
export const ChangePlanDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ChangePlan"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"plan"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Plan"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"changePlan"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"plan"},"value":{"kind":"Variable","name":{"kind":"Name","value":"plan"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"creditExpiry"}}]}}]}}]} as unknown as DocumentNode<ChangePlanMutation, ChangePlanMutationVariables>;
export const BillingPortalUrlDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"BillingPortalUrl"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"billingPortalUrl"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<BillingPortalUrlMutation, BillingPortalUrlMutationVariables>;
export const CreatePlanCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreatePlanCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"plan"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Plan"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createPlanCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"plan"},"value":{"kind":"Variable","name":{"kind":"Name","value":"plan"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<CreatePlanCheckoutMutation, CreatePlanCheckoutMutationVariables>;
export const CompletePlanCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CompletePlanCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"completePlanCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"sessionId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"plan"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"currentPeriodEnd"}},{"kind":"Field","name":{"kind":"Name","value":"creditAmountCents"}},{"kind":"Field","name":{"kind":"Name","value":"hasPaymentMethod"}}]}}]}}]} as unknown as DocumentNode<CompletePlanCheckoutMutation, CompletePlanCheckoutMutationVariables>;
export const EnvironmentResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"EnvironmentResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"environmentResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"tier"}},{"kind":"Field","name":{"kind":"Name","value":"allocation"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpuMillicores"}},{"kind":"Field","name":{"kind":"Name","value":"memoryMB"}},{"kind":"Field","name":{"kind":"Name","value":"diskMB"}}]}}]}}]}}]} as unknown as DocumentNode<EnvironmentResourcesQuery, EnvironmentResourcesQueryVariables>;
export const SetEnvironmentResourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetEnvironmentResources"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"SetEnvironmentResourcesInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setEnvironmentResources"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"tier"}},{"kind":"Field","name":{"kind":"Name","value":"allocation"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpuMillicores"}},{"kind":"Field","name":{"kind":"Name","value":"memoryMB"}},{"kind":"Field","name":{"kind":"Name","value":"diskMB"}}]}}]}}]}}]} as unknown as DocumentNode<SetEnvironmentResourcesMutation, SetEnvironmentResourcesMutationVariables>;
export const CreateDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateDatabaseInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"instances"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]} as unknown as DocumentNode<CreateDatabaseMutation, CreateDatabaseMutationVariables>;
export const DeleteDatabaseDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteDatabase"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"name"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteDatabase"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"name"},"value":{"kind":"Variable","name":{"kind":"Name","value":"name"}}}]}]}}]} as unknown as DocumentNode<DeleteDatabaseMutation, DeleteDatabaseMutationVariables>;
export const DatabaseTablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseTables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseTables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"schema"}},{"kind":"Field","name":{"kind":"Name","value":"estimatedRows"}},{"kind":"Field","name":{"kind":"Name","value":"columns"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"nullable"}},{"kind":"Field","name":{"kind":"Name","value":"primaryKey"}}]}}]}}]}}]} as unknown as DocumentNode<DatabaseTablesQuery, DatabaseTablesQueryVariables>;
export const DatabaseTableDataDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseTableData"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"table"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"schema"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"limit"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"offset"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseTableData"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}},{"kind":"Argument","name":{"kind":"Name","value":"table"},"value":{"kind":"Variable","name":{"kind":"Name","value":"table"}}},{"kind":"Argument","name":{"kind":"Name","value":"schema"},"value":{"kind":"Variable","name":{"kind":"Name","value":"schema"}}},{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"limit"}}},{"kind":"Argument","name":{"kind":"Name","value":"offset"},"value":{"kind":"Variable","name":{"kind":"Name","value":"offset"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"columns"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"totalEstimatedRows"}}]}}]}}]} as unknown as DocumentNode<DatabaseTableDataQuery, DatabaseTableDataQueryVariables>;
export const DatabaseCredentialsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DatabaseCredentials"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"database"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"databaseCredentials"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"database"},"value":{"kind":"Variable","name":{"kind":"Name","value":"database"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"host"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"dbname"}},{"kind":"Field","name":{"kind":"Name","value":"user"}},{"kind":"Field","name":{"kind":"Name","value":"password"}},{"kind":"Field","name":{"kind":"Name","value":"uri"}}]}}]}}]} as unknown as DocumentNode<DatabaseCredentialsQuery, DatabaseCredentialsQueryVariables>;
export const ExecuteQueryDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"ExecuteQuery"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DatabaseQueryInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"executeQuery"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"columns"}},{"kind":"Field","name":{"kind":"Name","value":"rows"}},{"kind":"Field","name":{"kind":"Name","value":"affectedRows"}}]}}]}}]} as unknown as DocumentNode<ExecuteQueryMutation, ExecuteQueryMutationVariables>;
export const GitHubConnectedDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubConnected"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubConnected"}}]}}]} as unknown as DocumentNode<GitHubConnectedQuery, GitHubConnectedQueryVariables>;
export const GitHubSourcesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubSources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubSources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"accountLogin"}},{"kind":"Field","name":{"kind":"Name","value":"accountAvatarUrl"}},{"kind":"Field","name":{"kind":"Name","value":"accountType"}}]}}]}}]} as unknown as DocumentNode<GitHubSourcesQuery, GitHubSourcesQueryVariables>;
export const GitHubRepositoriesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GitHubRepositories"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"installationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"githubRepositories"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"installationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"installationId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"fullName"}},{"kind":"Field","name":{"kind":"Name","value":"htmlUrl"}},{"kind":"Field","name":{"kind":"Name","value":"defaultBranch"}},{"kind":"Field","name":{"kind":"Name","value":"private"}}]}}]}}]} as unknown as DocumentNode<GitHubRepositoriesQuery, GitHubRepositoriesQueryVariables>;
export const ProjectsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Projects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"projects"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"syncStatus"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}}]}}]}}]}}]} as unknown as DocumentNode<ProjectsQuery, ProjectsQueryVariables>;
export const CreateProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateProject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateProjectInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createProject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]} as unknown as DocumentNode<CreateProjectMutation, CreateProjectMutationVariables>;
export const DeleteProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteProject"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteProject"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<DeleteProjectMutation, DeleteProjectMutationVariables>;
export const CreateEnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateEnvironment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateEnvironmentInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createEnvironment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"namespace"}},{"kind":"Field","name":{"kind":"Name","value":"ephemeral"}},{"kind":"Field","name":{"kind":"Name","value":"syncStatus"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}}]}}]}}]} as unknown as DocumentNode<CreateEnvironmentMutation, CreateEnvironmentMutationVariables>;
export const DeleteEnvironmentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteEnvironment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteEnvironment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}]}]}}]} as unknown as DocumentNode<DeleteEnvironmentMutation, DeleteEnvironmentMutationVariables>;
export const SetServiceScalingDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceScaling"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"SetServiceScalingInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceScaling"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"replicas"}},{"kind":"Field","name":{"kind":"Name","value":"autoscaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"minReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"maxReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"targetCPU"}}]}}]}}]}}]} as unknown as DocumentNode<SetServiceScalingMutation, SetServiceScalingMutationVariables>;
export const ProjectDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Project"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"project"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"environments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"namespace"}},{"kind":"Field","name":{"kind":"Name","value":"ephemeral"}},{"kind":"Field","name":{"kind":"Name","value":"syncStatus"}},{"kind":"Field","name":{"kind":"Name","value":"resourceTier"}},{"kind":"Field","name":{"kind":"Name","value":"services"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environment"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"framework"}},{"kind":"Field","name":{"kind":"Name","value":"startCommand"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}},{"kind":"Field","name":{"kind":"Name","value":"contextPath"}},{"kind":"Field","name":{"kind":"Name","value":"customStartCommand"}},{"kind":"Field","name":{"kind":"Name","value":"imageTag"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}},{"kind":"Field","name":{"kind":"Name","value":"replicas"}},{"kind":"Field","name":{"kind":"Name","value":"scaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"replicas"}},{"kind":"Field","name":{"kind":"Name","value":"autoscaling"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"enabled"}},{"kind":"Field","name":{"kind":"Name","value":"minReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"maxReplicas"}},{"kind":"Field","name":{"kind":"Name","value":"targetCPU"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"resources"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cpuMillicores"}},{"kind":"Field","name":{"kind":"Name","value":"memoryMB"}},{"kind":"Field","name":{"kind":"Name","value":"cpuLimitMillicores"}},{"kind":"Field","name":{"kind":"Name","value":"memoryLimitMB"}}]}},{"kind":"Field","name":{"kind":"Name","value":"domains"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"hostname"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"dnsStatus"}},{"kind":"Field","name":{"kind":"Name","value":"tlsStatus"}}]}},{"kind":"Field","name":{"kind":"Name","value":"deployments"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"imageTag"}},{"kind":"Field","name":{"kind":"Name","value":"active"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"revision"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"sourceCommitMessage"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environment"}},{"kind":"Field","name":{"kind":"Name","value":"ready"}},{"kind":"Field","name":{"kind":"Name","value":"instances"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"volume"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"size"}},{"kind":"Field","name":{"kind":"Name","value":"requestedSize"}},{"kind":"Field","name":{"kind":"Name","value":"usedBytes"}},{"kind":"Field","name":{"kind":"Name","value":"capacityBytes"}}]}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"databases"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"version"}},{"kind":"Field","name":{"kind":"Name","value":"instances"}},{"kind":"Field","name":{"kind":"Name","value":"size"}}]}}]}}]}}]} as unknown as DocumentNode<ProjectQuery, ProjectQueryVariables>;
export const SearchImagesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SearchImages"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"query"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"searchImages"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"query"},"value":{"kind":"Variable","name":{"kind":"Name","value":"query"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"description"}},{"kind":"Field","name":{"kind":"Name","value":"starCount"}},{"kind":"Field","name":{"kind":"Name","value":"pullCount"}},{"kind":"Field","name":{"kind":"Name","value":"official"}}]}}]}}]} as unknown as DocumentNode<SearchImagesQuery, SearchImagesQueryVariables>;
export const DetectServicesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DetectServices"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"installationId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"repository"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"detectServices"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"installationId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"installationId"}}},{"kind":"Argument","name":{"kind":"Name","value":"repository"},"value":{"kind":"Variable","name":{"kind":"Name","value":"repository"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"language"}},{"kind":"Field","name":{"kind":"Name","value":"framework"}},{"kind":"Field","name":{"kind":"Name","value":"startCommand"}},{"kind":"Field","name":{"kind":"Name","value":"suggestedPort"}}]}}]}}]} as unknown as DocumentNode<DetectServicesQuery, DetectServicesQueryVariables>;
export const AddServiceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddService"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AddServiceInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addService"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"environment"}},{"kind":"Field","name":{"kind":"Name","value":"image"}},{"kind":"Field","name":{"kind":"Name","value":"port"}},{"kind":"Field","name":{"kind":"Name","value":"framework"}},{"kind":"Field","name":{"kind":"Name","value":"startCommand"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}},{"kind":"Field","name":{"kind":"Name","value":"contextPath"}},{"kind":"Field","name":{"kind":"Name","value":"customStartCommand"}},{"kind":"Field","name":{"kind":"Name","value":"imageTag"}},{"kind":"Field","name":{"kind":"Name","value":"initialDeploy"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"phase"}}]}}]}}]}}]} as unknown as DocumentNode<AddServiceMutation, AddServiceMutationVariables>;
export const SetCustomStartCommandDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetCustomStartCommand"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"command"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setCustomStartCommand"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"command"},"value":{"kind":"Variable","name":{"kind":"Name","value":"command"}}}]}]}}]} as unknown as DocumentNode<SetCustomStartCommandMutation, SetCustomStartCommandMutationVariables>;
export const RemoveServiceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveService"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeService"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}]}]}}]} as unknown as DocumentNode<RemoveServiceMutation, RemoveServiceMutationVariables>;
export const DeployDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Deploy"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"DeployInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deploy"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"phase"}},{"kind":"Field","name":{"kind":"Name","value":"buildId"}},{"kind":"Field","name":{"kind":"Name","value":"imageRef"}},{"kind":"Field","name":{"kind":"Name","value":"digest"}},{"kind":"Field","name":{"kind":"Name","value":"error"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}}]}}]}}]} as unknown as DocumentNode<DeployMutation, DeployMutationVariables>;
export const DeployStatusDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"DeployStatus"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deployStatus"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"phase"}},{"kind":"Field","name":{"kind":"Name","value":"buildId"}},{"kind":"Field","name":{"kind":"Name","value":"imageRef"}},{"kind":"Field","name":{"kind":"Name","value":"digest"}},{"kind":"Field","name":{"kind":"Name","value":"error"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"rolloutHealth"}},{"kind":"Field","name":{"kind":"Name","value":"rolloutMessage"}}]}}]}}]} as unknown as DocumentNode<DeployStatusQuery, DeployStatusQueryVariables>;
export const ActiveDeploymentDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ActiveDeployment"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"activeDeployment"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"phase"}},{"kind":"Field","name":{"kind":"Name","value":"buildId"}},{"kind":"Field","name":{"kind":"Name","value":"imageRef"}},{"kind":"Field","name":{"kind":"Name","value":"digest"}},{"kind":"Field","name":{"kind":"Name","value":"error"}},{"kind":"Field","name":{"kind":"Name","value":"startedAt"}},{"kind":"Field","name":{"kind":"Name","value":"rolloutHealth"}},{"kind":"Field","name":{"kind":"Name","value":"rolloutMessage"}}]}}]}}]} as unknown as DocumentNode<ActiveDeploymentQuery, ActiveDeploymentQueryVariables>;
export const RollbackDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Rollback"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"RollbackInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"rollback"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}]}]}}]} as unknown as DocumentNode<RollbackMutation, RollbackMutationVariables>;
export const GenerateDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"GenerateDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"GenerateDomainInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"generateDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"hostname"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"dnsStatus"}},{"kind":"Field","name":{"kind":"Name","value":"tlsStatus"}}]}}]}}]} as unknown as DocumentNode<GenerateDomainMutation, GenerateDomainMutationVariables>;
export const AddCustomDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"AddCustomDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AddCustomDomainInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"addCustomDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"hostname"}},{"kind":"Field","name":{"kind":"Name","value":"type"}},{"kind":"Field","name":{"kind":"Name","value":"dnsStatus"}},{"kind":"Field","name":{"kind":"Name","value":"tlsStatus"}}]}}]}}]} as unknown as DocumentNode<AddCustomDomainMutation, AddCustomDomainMutationVariables>;
export const RemoveDomainDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveDomain"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"RemoveDomainInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeDomain"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}]}]}}]} as unknown as DocumentNode<RemoveDomainMutation, RemoveDomainMutationVariables>;
export const CheckDnsStatusDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"CheckDnsStatus"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"checkDnsStatus"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"hostname"},"value":{"kind":"Variable","name":{"kind":"Name","value":"hostname"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"hostname"}},{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"cnameTarget"}},{"kind":"Field","name":{"kind":"Name","value":"expectedTarget"}},{"kind":"Field","name":{"kind":"Name","value":"message"}},{"kind":"Field","name":{"kind":"Name","value":"tlsStatus"}}]}}]}}]} as unknown as DocumentNode<CheckDnsStatusQuery, CheckDnsStatusQueryVariables>;
export const PlatformConfigDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"PlatformConfig"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"platformConfig"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workloadDomain"}},{"kind":"Field","name":{"kind":"Name","value":"domainTarget"}},{"kind":"Field","name":{"kind":"Name","value":"ipAddress"}}]}}]}}]} as unknown as DocumentNode<PlatformConfigQuery, PlatformConfigQueryVariables>;
export const DeployLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"DeployLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deployLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}]}]}}]} as unknown as DocumentNode<DeployLogsSubscription, DeployLogsSubscriptionVariables>;
export const ServiceLogsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"ServiceLogs"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"tailLines"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serviceLogs"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"tailLines"},"value":{"kind":"Variable","name":{"kind":"Name","value":"tailLines"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"line"}},{"kind":"Field","name":{"kind":"Name","value":"pod"}}]}}]}}]} as unknown as DocumentNode<ServiceLogsSubscription, ServiceLogsSubscriptionVariables>;
export const SharedVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"SharedVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"sharedVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"value"}}]}}]}}]} as unknown as DocumentNode<SharedVariablesQuery, SharedVariablesQueryVariables>;
export const SetSharedVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetSharedVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"variables"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"VariableInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setSharedVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"variables"},"value":{"kind":"Variable","name":{"kind":"Name","value":"variables"}}}]}]}}]} as unknown as DocumentNode<SetSharedVariablesMutation, SetSharedVariablesMutationVariables>;
export const ServiceVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"ServiceVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"serviceVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"key"}},{"kind":"Field","name":{"kind":"Name","value":"value"}},{"kind":"Field","name":{"kind":"Name","value":"fromShared"}},{"kind":"Field","name":{"kind":"Name","value":"databaseRef"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"database"}},{"kind":"Field","name":{"kind":"Name","value":"key"}}]}},{"kind":"Field","name":{"kind":"Name","value":"serviceRef"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"service"}}]}}]}}]}}]} as unknown as DocumentNode<ServiceVariablesQuery, ServiceVariablesQueryVariables>;
export const SetServiceVariablesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"SetServiceVariables"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"environment"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"service"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"variables"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ServiceVariableInput"}}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"setServiceVariables"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"projectId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"projectId"}}},{"kind":"Argument","name":{"kind":"Name","value":"environment"},"value":{"kind":"Variable","name":{"kind":"Name","value":"environment"}}},{"kind":"Argument","name":{"kind":"Name","value":"service"},"value":{"kind":"Variable","name":{"kind":"Name","value":"service"}}},{"kind":"Argument","name":{"kind":"Name","value":"variables"},"value":{"kind":"Variable","name":{"kind":"Name","value":"variables"}}}]}]}}]} as unknown as DocumentNode<SetServiceVariablesMutation, SetServiceVariablesMutationVariables>;
export const WorkspacesDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Workspaces"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspaces"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}}]}}]}}]} as unknown as DocumentNode<WorkspacesQuery, WorkspacesQueryVariables>;
export const WorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"workspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}},{"kind":"Field","name":{"kind":"Name","value":"suspended"}},{"kind":"Field","name":{"kind":"Name","value":"members"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]}}]} as unknown as DocumentNode<WorkspaceQuery, WorkspaceQueryVariables>;
export const CreateWorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateWorkspace"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateWorkspaceInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createWorkspace"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}}]}}]}}]} as unknown as DocumentNode<CreateWorkspaceMutation, CreateWorkspaceMutationVariables>;
export const CreateWorkspaceCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CreateWorkspaceCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"CreateWorkspaceCheckoutInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"createWorkspaceCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"url"}}]}}]}}]} as unknown as DocumentNode<CreateWorkspaceCheckoutMutation, CreateWorkspaceCheckoutMutationVariables>;
export const CompleteWorkspaceCheckoutDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"CompleteWorkspaceCheckout"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"completeWorkspaceCheckout"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"sessionId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"sessionId"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"personal"}}]}}]}}]} as unknown as DocumentNode<CompleteWorkspaceCheckoutMutation, CompleteWorkspaceCheckoutMutationVariables>;
export const UpdateWorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateWorkspace"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateWorkspaceInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateWorkspace"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}}]}}]}}]} as unknown as DocumentNode<UpdateWorkspaceMutation, UpdateWorkspaceMutationVariables>;
export const DeleteWorkspaceDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"DeleteWorkspace"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"deleteWorkspace"}}]}}]} as unknown as DocumentNode<DeleteWorkspaceMutation, DeleteWorkspaceMutationVariables>;
export const InviteMemberDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"InviteMember"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"InviteMemberInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"inviteMember"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]} as unknown as DocumentNode<InviteMemberMutation, InviteMemberMutationVariables>;
export const RemoveMemberDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"RemoveMember"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"userId"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"removeMember"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"userId"},"value":{"kind":"Variable","name":{"kind":"Name","value":"userId"}}}]}]}}]} as unknown as DocumentNode<RemoveMemberMutation, RemoveMemberMutationVariables>;
export const UpdateMemberRoleDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"UpdateMemberRole"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UpdateMemberRoleInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"updateMemberRole"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"role"}}]}}]}}]} as unknown as DocumentNode<UpdateMemberRoleMutation, UpdateMemberRoleMutationVariables>;