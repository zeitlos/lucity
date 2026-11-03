import { ref } from 'vue';

const open = ref(false);

const modKey = /Mac|iPhone|iPad/.test(navigator.userAgent) ? '⌘' : 'Ctrl ';

export function useCommandPalette() {
  return {
    open,
    modKey,
    show() {
      open.value = true;
    },
  };
}
