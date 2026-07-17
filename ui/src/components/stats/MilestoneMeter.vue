<template>
  <div class="milestone-meter">
    <svg
      class="milestone-meter__svg"
      viewBox="0 0 200 8"
      preserveAspectRatio="none"
      role="img"
      :aria-label="ariaLabel"
    >
      <defs>
        <linearGradient id="milestone-meter-grad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" class="milestone-meter__grad-start" />
          <stop offset="1" class="milestone-meter__grad-end" />
        </linearGradient>
      </defs>
      <path :d="trackPath" class="milestone-meter__track" />
      <path v-if="progress > 0" :d="fillPath" class="milestone-meter__fill" />
    </svg>

    <div class="milestone-meter__caption">{{ caption }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { compactNumber } from '../../utils/format.js';

const props = defineProps({
  total: { type: Number, default: 0 },
});

const LADDER = [10, 100, 500, 1000, 5000, 10000, 50000, 100000];

const barWidth = 200;
const barHeight = 8;
const radius = barHeight / 2;

const rungs = computed(() => {
  const total = Number(props.total) || 0;
  let prev = 0;
  for (const rung of LADDER) {
    if (total < rung) return { prev, next: rung, maxed: false };
    prev = rung;
  }
  return { prev, next: null, maxed: true };
});

const isFresh = computed(() => (Number(props.total) || 0) === 0);

const progress = computed(() => {
  const { prev, next, maxed } = rungs.value;
  if (maxed) return 1;
  const span = next - prev;
  if (span <= 0) return 0;
  const total = Number(props.total) || 0;
  return Math.min(1, Math.max(0, (total - prev) / span));
});

const caption = computed(() => {
  const total = Number(props.total) || 0;
  const { next, maxed } = rungs.value;
  if (isFresh.value) return 'First milestone: 10 secrets';
  if (maxed) return `Max milestone reached — ${compactNumber(total)} secrets`;
  const remaining = Math.max(0, next - total);
  return `${compactNumber(remaining)} to ${compactNumber(next)}`;
});

const ariaLabel = computed(() => `Milestone progress: ${caption.value}`);

/** Full rounded pill covering [x, x+width]. */
function pillPath(x, width) {
  const r = Math.min(radius, width / 2, barHeight / 2);
  if (r <= 0) return '';
  return `M${x + r},0 H${x + width - r} A${r},${r} 0 0 1 ${x + width},${r} V${barHeight - r} A${r},${r} 0 0 1 ${x + width - r},${barHeight} H${x + r} A${r},${r} 0 0 1 ${x},${barHeight - r} V${r} A${r},${r} 0 0 1 ${x + r},0 Z`;
}

/** Rounded-left / square-right rect, for a partial fill. */
function leftRoundedPath(width) {
  const r = Math.min(radius, width / 2, barHeight / 2);
  if (r <= 0) return '';
  return `M${r},0 H${width} V${barHeight} H${r} A${r},${r} 0 0 1 0,${barHeight - r} V${r} A${r},${r} 0 0 1 ${r},0 Z`;
}

const trackPath = computed(() => pillPath(0, barWidth));
const fillPath = computed(() => {
  const width = barWidth * progress.value;
  return progress.value >= 1 ? pillPath(0, barWidth) : leftRoundedPath(width);
});
</script>

<style scoped>
.milestone-meter {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.milestone-meter__svg {
  width: 100%;
  height: 8px;
  display: block;
}

.milestone-meter__track {
  fill: rgba(var(--v-theme-primary), 0.16);
}

.milestone-meter__grad-start { stop-color: rgb(var(--v-theme-primary)); stop-opacity: 0.65; }
.milestone-meter__grad-end { stop-color: rgb(var(--v-theme-primary)); stop-opacity: 1; }

.milestone-meter__fill {
  fill: url(#milestone-meter-grad);
  transform-box: fill-box;
  transform-origin: left;
  animation: gd-fill-x 700ms cubic-bezier(0.22, 1, 0.36, 1) both;
  animation-delay: 240ms;
}

@media (prefers-reduced-motion: reduce) {
  .milestone-meter__fill {
    animation: none;
  }
}

.milestone-meter__caption {
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
}
</style>
