import { h } from 'vue';
import type { ExternalToast } from 'vue-sonner';
import { toast } from 'vue-sonner';
import { Clipboard } from 'lucide-vue-next';

export { default as Sonner } from './Sonner.vue';
export { toast };

export function errorToast(message: string, opts?: ExternalToast) {
  const text = opts?.description ? `${message}: ${opts.description}` : message;
  const copyAction = {
    label: h(Clipboard, { size: 14 }),
    onClick: () => navigator.clipboard.writeText(text),
  };
  return toast.error(message, {
    ...opts,
    ...(opts?.action ? { cancel: copyAction } : { action: copyAction }),
  });
}
