import { ref, onMounted, onUnmounted, type Ref } from 'vue';

/**
 * Returns a `visible` ref that becomes true (once) when the element
 * enters the viewport. Used to trigger bento card animations on scroll.
 */
export function useBentoVisible(el: Ref<HTMLElement | null>, threshold = 0.3) {
  const visible = ref(false);
  let observer: IntersectionObserver | null = null;

  onMounted(() => {
    if (!el.value) return;
    observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          visible.value = true;
          observer?.disconnect();
        }
      },
      { threshold },
    );
    observer.observe(el.value);
  });

  onUnmounted(() => {
    observer?.disconnect();
  });

  return visible;
}
