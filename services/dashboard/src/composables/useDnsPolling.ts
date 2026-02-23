import { reactive, onUnmounted } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { CheckDnsStatusQuery } from '@/graphql/services';
import { toast } from '@/components/ui/sonner';

export interface DnsCheckResult {
  hostname: string;
  status: 'VALID' | 'PENDING' | 'MISCONFIGURED' | 'ERROR';
  cnameTarget: string | null;
  expectedTarget: string;
  message: string | null;
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
      h => !checks[h] || checks[h].status !== 'VALID',
    );

    if (pending.length === 0) {
      stopPolling();
      return;
    }

    await Promise.allSettled(
      pending.map(async (hostname) => {
        try {
          const { data } = await apolloClient.query({
            query: CheckDnsStatusQuery,
            variables: { hostname },
            fetchPolicy: 'network-only',
          });
          if (data?.checkDnsStatus) {
            const prev = checks[hostname]?.status;
            checks[hostname] = data.checkDnsStatus;
            if (prev && prev !== 'VALID' && data.checkDnsStatus.status === 'VALID') {
              toast.success(`Domain verified: ${hostname}`);
            }
          }
        } catch {
          // Keep previous state on error
        }
      }),
    );

    // Stop if all tracked hostnames are now VALID
    const allValid = [...trackedHostnames].every(
      h => checks[h]?.status === 'VALID',
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

  onUnmounted(() => stopPolling());

  return {
    checks,
    trackHostnames,
    addHostname,
    removeHostname,
    stopPolling,
  };
}
