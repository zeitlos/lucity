import { onBeforeUnmount, onMounted } from 'vue';

const OPEN_LAYER = [
  '[role="dialog"][data-state="open"]',
  '[role="alertdialog"][data-state="open"]',
  '[role="menu"][data-state="open"]',
  '[role="listbox"][data-state="open"]',
  '[aria-modal="true"]',
].join(', ');

const stack: Array<() => void> = [];
let listening = false;

function isEditable(target: EventTarget | null): target is HTMLElement {
  return target instanceof HTMLElement
    && (target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName));
}

function onKeydown(event: KeyboardEvent) {
  if (event.key !== 'Escape' || event.defaultPrevented || stack.length === 0) return;
  if (document.querySelector(OPEN_LAYER)) return;
  if (isEditable(event.target)) {
    event.target.blur();
    return;
  }
  event.preventDefault();
  event.stopPropagation();
  stack[stack.length - 1]!();
}

/**
 * Closes the topmost panel on Escape. Escape belongs to an open dialog, menu,
 * select or popover first, and leaves a focused text field before it closes anything.
 */
export function useEscapeToClose(close: () => void) {
  onMounted(() => {
    stack.push(close);
    if (!listening) {
      window.addEventListener('keydown', onKeydown, { capture: true });
      listening = true;
    }
  });

  onBeforeUnmount(() => {
    const index = stack.lastIndexOf(close);
    if (index !== -1) stack.splice(index, 1);
  });
}
