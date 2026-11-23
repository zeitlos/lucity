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
  ServiceInstanceState,
  ServiceStatus,
  SubscriptionStatus,
  TlsStatus,
} from '@/gql/graphql';

export type StatusTone = 'ok' | 'warn' | 'danger' | 'progress' | 'neutral';

export type StatusGlyph =
  | 'online'
  | 'degraded'
  | 'failed'
  | 'stopped'
  | 'idle'
  | 'dot'
  | 'unknown'
  | 'snake'
  | 'ripple'
  | 'build';

export interface StatusView {
  tone: StatusTone;
  label: string;
  glyph?: StatusGlyph;
}

const unknown: StatusView = { tone: 'neutral', label: 'Unknown', glyph: 'unknown' };

const service: Record<ServiceStatus, StatusView> = {
  [ServiceStatus.Healthy]: { tone: 'ok', label: 'Online', glyph: 'online' },
  [ServiceStatus.Building]: { tone: 'progress', label: 'Building', glyph: 'build' },
  [ServiceStatus.Deploying]: { tone: 'progress', label: 'Deploying', glyph: 'snake' },
  [ServiceStatus.Degraded]: { tone: 'warn', label: 'Degraded', glyph: 'degraded' },
  [ServiceStatus.Failed]: { tone: 'danger', label: 'Failed', glyph: 'failed' },
  [ServiceStatus.Stopped]: { tone: 'neutral', label: 'Stopped', glyph: 'stopped' },
};

const instance: Record<ServiceInstanceState, StatusView> = {
  [ServiceInstanceState.Starting]: { tone: 'progress', label: 'Starting' },
  [ServiceInstanceState.Running]: { tone: 'neutral', label: 'Running' },
  [ServiceInstanceState.Stopping]: { tone: 'progress', label: 'Stopping' },
  [ServiceInstanceState.Stopped]: { tone: 'neutral', label: 'Stopped' },
};

const database: Record<DatabaseStatus, StatusView> = {
  [DatabaseStatus.Healthy]: { tone: 'ok', label: 'Online', glyph: 'online' },
  [DatabaseStatus.Pending]: { tone: 'progress', label: 'Starting', glyph: 'ripple' },
  [DatabaseStatus.Updating]: { tone: 'progress', label: 'Updating', glyph: 'snake' },
  [DatabaseStatus.Degraded]: { tone: 'warn', label: 'Degraded', glyph: 'degraded' },
  [DatabaseStatus.Failed]: { tone: 'danger', label: 'Failed', glyph: 'failed' },
  [DatabaseStatus.Stopped]: { tone: 'neutral', label: 'Stopped', glyph: 'stopped' },
};

const bucket: Record<BucketStatus, StatusView> = {
  [BucketStatus.Ready]: { tone: 'ok', label: 'Online', glyph: 'online' },
  [BucketStatus.Pending]: { tone: 'progress', label: 'Provisioning', glyph: 'build' },
  [BucketStatus.Failed]: { tone: 'danger', label: 'Failed', glyph: 'failed' },
};

export type JobState = 'queued' | 'running' | 'succeeded' | 'failed' | 'cancelled' | 'skipped' | 'superseded';

const job: Record<JobState, StatusView> = {
  queued: { tone: 'neutral', label: 'Queued' },
  running: { tone: 'progress', label: 'Running' },
  succeeded: { tone: 'ok', label: 'Succeeded' },
  failed: { tone: 'danger', label: 'Failed' },
  cancelled: { tone: 'neutral', label: 'Cancelled' },
  skipped: { tone: 'neutral', label: 'Skipped' },
  superseded: { tone: 'neutral', label: 'Superseded' },
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
  [SubscriptionStatus.TrialEnded]: { tone: 'warn', label: 'Trial ended' },
  [SubscriptionStatus.PastDue]: { tone: 'warn', label: 'Payment due' },
  [SubscriptionStatus.Incomplete]: { tone: 'warn', label: 'Payment due' },
  [SubscriptionStatus.Canceled]: { tone: 'neutral', label: 'Cancelled' },
};

const agent: Record<AgentStatus, StatusView> = {
  [AgentStatus.Starting]: { tone: 'progress', label: 'Starting', glyph: 'ripple' },
  [AgentStatus.Idle]: { tone: 'neutral', label: 'Idle', glyph: 'idle' },
  [AgentStatus.Working]: { tone: 'progress', label: 'Working', glyph: 'snake' },
  [AgentStatus.WaitingForApproval]: { tone: 'progress', label: 'Working', glyph: 'snake' },
  [AgentStatus.Failed]: { tone: 'danger', label: 'Failed', glyph: 'failed' },
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

export function instanceStatus(state: ServiceInstanceState): StatusView {
  return instance[state] ?? unknown;
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
