import {
  AgentAuditResult,
  AgentSessionStatus,
  AgentStatus,
  BackupStatus,
  CredentialStatus,
  BucketStatus,
  BuildStatus,
  DatabaseStatus,
  DnsStatus,
  ServiceStatus,
  SubscriptionStatus,
  TlsStatus,
} from '@/gql/graphql';

export type StatusTone = 'ok' | 'warn' | 'danger' | 'progress' | 'neutral';

export interface StatusView {
  tone: StatusTone;
  label: string;
}

const unknown: StatusView = { tone: 'neutral', label: 'Unknown' };

const service: Record<ServiceStatus, StatusView> = {
  [ServiceStatus.Healthy]: { tone: 'ok', label: 'Online' },
  [ServiceStatus.Building]: { tone: 'progress', label: 'Building' },
  [ServiceStatus.Deploying]: { tone: 'progress', label: 'Deploying' },
  [ServiceStatus.Degraded]: { tone: 'warn', label: 'Degraded' },
  [ServiceStatus.Failed]: { tone: 'danger', label: 'Failed' },
  [ServiceStatus.Stopped]: { tone: 'neutral', label: 'Stopped' },
};

const database: Record<DatabaseStatus, StatusView> = {
  [DatabaseStatus.Healthy]: { tone: 'ok', label: 'Online' },
  [DatabaseStatus.Pending]: { tone: 'progress', label: 'Provisioning' },
  [DatabaseStatus.Updating]: { tone: 'progress', label: 'Updating' },
  [DatabaseStatus.Degraded]: { tone: 'warn', label: 'Degraded' },
  [DatabaseStatus.Failed]: { tone: 'danger', label: 'Failed' },
  [DatabaseStatus.Stopped]: { tone: 'neutral', label: 'Stopped' },
};

const bucket: Record<BucketStatus, StatusView> = {
  [BucketStatus.Ready]: { tone: 'ok', label: 'Online' },
  [BucketStatus.Pending]: { tone: 'progress', label: 'Provisioning' },
  [BucketStatus.Failed]: { tone: 'danger', label: 'Failed' },
};

export type JobState = 'queued' | 'running' | 'succeeded' | 'failed' | 'cancelled' | 'skipped';

const job: Record<JobState, StatusView> = {
  queued: { tone: 'neutral', label: 'Queued' },
  running: { tone: 'progress', label: 'Running' },
  succeeded: { tone: 'ok', label: 'Succeeded' },
  failed: { tone: 'danger', label: 'Failed' },
  cancelled: { tone: 'neutral', label: 'Cancelled' },
  skipped: { tone: 'neutral', label: 'Skipped' },
};

const backupState: Record<BackupStatus, JobState> = {
  [BackupStatus.Pending]: 'queued',
  [BackupStatus.Running]: 'running',
  [BackupStatus.Completed]: 'succeeded',
  [BackupStatus.Failed]: 'failed',
};

const subscription: Record<SubscriptionStatus, StatusView> = {
  [SubscriptionStatus.Active]: { tone: 'ok', label: 'Active' },
  [SubscriptionStatus.Trialing]: { tone: 'progress', label: 'Trial' },
  [SubscriptionStatus.PastDue]: { tone: 'warn', label: 'Payment due' },
  [SubscriptionStatus.Incomplete]: { tone: 'warn', label: 'Payment due' },
  [SubscriptionStatus.Canceled]: { tone: 'neutral', label: 'Cancelled' },
};

const agent: Record<AgentStatus, StatusView> = {
  [AgentStatus.Starting]: { tone: 'progress', label: 'Starting' },
  [AgentStatus.Idle]: { tone: 'neutral', label: 'Idle' },
  [AgentStatus.Working]: { tone: 'progress', label: 'Working' },
  [AgentStatus.WaitingForApproval]: { tone: 'progress', label: 'Working' },
  [AgentStatus.Failed]: { tone: 'danger', label: 'Failed' },
};

const agentSession: Record<AgentSessionStatus, StatusView> = {
  [AgentSessionStatus.Idle]: { tone: 'neutral', label: 'Idle' },
  [AgentSessionStatus.Working]: { tone: 'progress', label: 'Working' },
  [AgentSessionStatus.WaitingForApproval]: { tone: 'warn', label: 'Waiting' },
};

const auditResult: Record<AgentAuditResult, StatusView> = {
  [AgentAuditResult.Ok]: { tone: 'neutral', label: 'OK' },
  [AgentAuditResult.Denied]: { tone: 'warn', label: 'Denied' },
  [AgentAuditResult.Failed]: { tone: 'danger', label: 'Failed' },
};

export function agentStatus(status: AgentStatus): StatusView {
  return agent[status] ?? unknown;
}

export function agentSessionStatus(status: AgentSessionStatus): StatusView {
  return agentSession[status] ?? unknown;
}

export function auditResultStatus(result: AgentAuditResult): StatusView {
  return auditResult[result] ?? unknown;
}

export function credentialStatus(status: CredentialStatus, credential: string): StatusView {
  switch (status) {
    case CredentialStatus.Valid:
      return { tone: 'ok', label: 'Connected' };
    case CredentialStatus.Rejected:
      return { tone: 'danger', label: `${credential} rejected` };
    case CredentialStatus.Unreachable:
      return { tone: 'warn', label: 'Unreachable' };
    default:
      return unknown;
  }
}

export function serviceStatus(status: ServiceStatus): StatusView {
  return service[status] ?? unknown;
}

export function databaseStatus(status: DatabaseStatus): StatusView {
  return database[status] ?? unknown;
}

export function bucketStatus(status: BucketStatus): StatusView {
  return bucket[status] ?? unknown;
}

export function jobStatus(state: JobState | BuildStatus): StatusView {
  return job[state.toLowerCase() as JobState] ?? unknown;
}

export function backupStatus(status: BackupStatus): StatusView {
  return job[backupState[status]] ?? unknown;
}

export function subscriptionStatus(status: SubscriptionStatus): StatusView {
  return subscription[status] ?? unknown;
}

export function dnsStatus(status: DnsStatus): StatusView {
  switch (status) {
    case DnsStatus.Valid:
      return { tone: 'ok', label: 'Verified' };
    case DnsStatus.Pending:
      return { tone: 'progress', label: 'Waiting for DNS' };
    case DnsStatus.Misconfigured:
      return { tone: 'warn', label: 'Misconfigured' };
    case DnsStatus.Error:
      return { tone: 'danger', label: 'Check failed' };
    default:
      return unknown;
  }
}

export function certificateStatus(status: TlsStatus): StatusView {
  switch (status) {
    case TlsStatus.Active:
      return { tone: 'ok', label: 'Issued' };
    case TlsStatus.Provisioning:
      return { tone: 'progress', label: 'Issuing' };
    case TlsStatus.Error:
      return { tone: 'danger', label: 'Certificate failed' };
    default:
      return { tone: 'neutral', label: 'Not issued yet' };
  }
}

export function domainStatus(dns: DnsStatus, tls: TlsStatus): StatusView {
  if (dns !== DnsStatus.Valid) return dnsStatus(dns);
  if (tls === TlsStatus.Active) return { tone: 'ok', label: 'Verified' };
  if (tls === TlsStatus.Error) return certificateStatus(tls);
  return { tone: 'progress', label: 'Issuing certificate' };
}
