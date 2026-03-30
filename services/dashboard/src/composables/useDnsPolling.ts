import { reactive, onUnmounted } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { CheckDnsStatusDocument, DnsStatus, TlsStatus } from '@/gql/graphql';
import { toast } from '@/components/ui/sonner';

export interface DnsCheckResult {
  hostname: string;
  status: string;
  cnameTarget?: string | null;
  expectedTarget: string;
  message?: string | null;
  tlsStatus?: string | null;
}

const POLL_INTERVAL = 5000;

export function useDnsPolling() {
  let pollTimer: ReturnType<typeof setInterval> | null = null;
  const trackedHostnames = new Set<string>();

  const checks = reactive<Record<string, DnsCheckResult>>({});

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  async function checkAll() {
    const pending = [...trackedHostnames].filter(
      h => !checks[h] || checks[h].status !== DnsStatus.Valid || (checks[h].tlsStatus && checks[h].tlsStatus !== TlsStatus.Active),
    );

    if (pending.length === 0) {
      stopPolling();
      return;
    }

    await Promise.allSettled(
      pending.map(async (hostname) => {
        try {
          const { data } = await apolloClient.query({
            query: CheckDnsStatusDocument,
            variables: { hostname },
            fetchPolicy: 'network-only',
          });
          if (data?.checkDnsStatus) {
            const prev = checks[hostname]?.status;
            checks[hostname] = data.checkDnsStatus;
            if (prev && prev !== DnsStatus.Valid && data.checkDnsStatus.status === DnsStatus.Valid) {
              toast.success(`Domain verified: ${hostname}`);
            }
          }
        } catch {
          // Keep previous state on error
        }
      }),
    );

    // Stop if all tracked hostnames have VALID DNS and ACTIVE TLS
    const allValid = [...trackedHostnames].every(
      h => checks[h]?.status === DnsStatus.Valid && (!checks[h]?.tlsStatus || checks[h]?.tlsStatus === TlsStatus.Active),
    );
    if (allValid) {
      stopPolling();
    }
  }

  function startPolling() {
    if (pollTimer) return;
    checkAll();
    pollTimer = setInterval(checkAll, POLL_INTERVAL);
  }

  function trackHostnames(hostnames: string[]) {
    trackedHostnames.clear();
    for (const key of Object.keys(checks)) {
      delete checks[key];
    }
    for (const h of hostnames) {
      trackedHostnames.add(h);
    }
    if (hostnames.length > 0) {
      startPolling();
    } else {
      stopPolling();
    }
  }

  function addHostname(hostname: string) {
    trackedHostnames.add(hostname);
    startPolling();
  }

  function removeHostname(hostname: string) {
    trackedHostnames.delete(hostname);
    delete checks[hostname];
    if (trackedHostnames.size === 0) {
      stopPolling();
    }
  }

  function setTlsStatus(hostname: string, tlsStatus: TlsStatus) {
    if (checks[hostname]) {
      checks[hostname].tlsStatus = tlsStatus;
    }
  }

  onUnmounted(() => stopPolling());

  return {
    checks,
    trackHostnames,
    addHostname,
    removeHostname,
    setTlsStatus,
    stopPolling,
  };
}
