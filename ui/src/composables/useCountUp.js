import { ref, watch, onUnmounted, unref } from 'vue';

/**
 * Ease-out cubic - fast start, gentle settle.
 * @param {number} t progress in [0, 1]
 */
function easeOutCubic(t) {
  return 1 - Math.pow(1 - t, 3);
}

function prefersReducedMotion() {
  return (
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  );
}

/**
 * Animate a displayed number toward a source value whenever it changes.
 * Powers the "+1 tick" feel after creating a secret, and re-animates on
 * every subsequent change to the source.
 *
 * @param {import('vue').Ref<number>|(() => number)} source - ref or getter holding the target number
 * @param {{ durationMs?: number }} [options]
 * @returns {import('vue').Ref<number>} display - the animated, currently-rendered value
 */
export function useCountUp(source, options = {}) {
  const durationMs = options.durationMs ?? 700;

  const getTarget = () => {
    const value = typeof source === 'function' ? source() : unref(source);
    return Number(value) || 0;
  };

  const display = ref(getTarget());

  let rafId = null;

  function cancel() {
    if (rafId !== null) {
      cancelAnimationFrame(rafId);
      rafId = null;
    }
  }

  function animateTo(target) {
    cancel();

    if (prefersReducedMotion()) {
      display.value = target;
      return;
    }

    const start = display.value;
    const delta = target - start;
    if (delta === 0) return;

    const startTime = performance.now();

    const tick = (now) => {
      const elapsed = now - startTime;
      const progress = Math.min(elapsed / durationMs, 1);
      const eased = easeOutCubic(progress);
      display.value = start + delta * eased;

      if (progress < 1) {
        rafId = requestAnimationFrame(tick);
      } else {
        display.value = target;
        rafId = null;
      }
    };

    rafId = requestAnimationFrame(tick);
  }

  watch(
    typeof source === 'function' ? source : () => unref(source),
    (newValue) => {
      animateTo(Number(newValue) || 0);
    }
  );

  onUnmounted(cancel);

  return display;
}
