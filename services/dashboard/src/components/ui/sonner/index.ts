import type { ExternalToast } from 'vue-sonner';
import { toast } from 'vue-sonner';
import { markRaw } from 'vue';
import ErrorToastComponent from './ErrorToast.vue';

export { default as Sonner } from './Sonner.vue';
export { toast };

export function errorToast(message: string, opts?: ExternalToast) {
  // Strip action/cancel/description from sonner options — our custom component renders them
  const { action: componentAction, cancel: _c, description, ...rest } = opts ?? {};
  void _c;
  const text = description ? `${message}: ${description}` : message;
  return toast.error(markRaw(ErrorToastComponent), {
    ...rest,
    componentProps: {
      title: message,
      description,
      copyText: text,
      action: componentAction,
    },
  });
}
