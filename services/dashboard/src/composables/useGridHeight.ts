import { onMounted, onUpdated, type ShallowRef } from 'vue';
import { gridSize } from '@/composables/useCanvasLayout';

export function useGridHeight(card: Readonly<ShallowRef<HTMLElement | null>>) {
  function snap() {
    const element = card.value;
    if (!element) return;
    element.style.height = '';
    element.style.height = `${Math.round(element.offsetHeight / gridSize) * gridSize}px`;
  }

  onMounted(snap);
  onUpdated(snap);
}
