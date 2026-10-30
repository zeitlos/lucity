import { toast } from 'vue-sonner';
import { markRaw, type Component } from 'vue';
import { WifiOff } from '@lucide/vue';
import ToastContent from './ToastContent.vue';
import { describeError } from '@/lib/errors';
import { openBugReport } from '@/composables/useReportBug';
import { gateFromError } from '@/composables/useGate';

export { default as Sonner } from './Sonner.vue';

interface ToastOptions {
  description?: string;
  action?: { label: string; onClick: () => void };
  copy?: boolean;
  icon?: Component;
  duration?: number;
}

function show(tone: 'ok' | 'danger' | 'neutral', title: string, options: ToastOptions = {}) {
  const { duration, ...props } = options;
  return toast.custom(markRaw(ToastContent), {
    duration,
    componentProps: { tone, title, ...props },
  });
}

export function successToast(title: string, options?: ToastOptions) {
  return show('ok', title, options);
}

export function infoToast(title: string, options?: ToastOptions) {
  return show('neutral', title, options);
}

export function errorToast(title: string, options?: ToastOptions) {
  return show('danger', title, { copy: true, ...options });
}

/**
 * Toast for a failed action. Explicit and unexpected errors stay until dismissed;
 * network errors fade, since they usually clear up on their own.
 */
export function failureToast(title: string, error: unknown) {
  if (gateFromError(error)) return;

  const { kind, message, detail } = describeError(error);
  return errorToast(title, {
    description: message,
    icon: kind === 'network' ? WifiOff : undefined,
    duration: kind === 'network' ? 8000 : Infinity,
    action: kind === 'unexpected'
      ? { label: 'Report', onClick: () => openBugReport({ title, error: detail }) }
      : undefined,
  });
}
