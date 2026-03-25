import { onMounted, onUnmounted } from 'vue';

/**
 * Opens the GitHub App installation page in a centered popup window
 * and listens for the `github-app-installed` postMessage to trigger a callback.
 */
export function useGitHubInstall(onInstalled?: () => void) {
  function openInstallPopup() {
    const w = 600;
    const h = 700;
    const left = window.screenX + (window.outerWidth - w) / 2;
    const top = window.screenY + (window.outerHeight - h) / 2;
    window.open('/auth/github/install', 'github-install', `width=${w},height=${h},left=${left},top=${top}`);
  }

  function handleMessage(event: MessageEvent) {
    if (event.data === 'github-app-installed') {
      onInstalled?.();
    }
  }

  onMounted(() => window.addEventListener('message', handleMessage));
  onUnmounted(() => window.removeEventListener('message', handleMessage));

  return { openInstallPopup };
}
