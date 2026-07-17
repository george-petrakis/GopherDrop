<template>
  <div class="split-bar">
    <div class="split-bar__labels">
      <span class="split-bar__label">Text {{ textPercent }}%</span>
      <span class="split-bar__label">Files {{ filePercent }}%</span>
    </div>

    <svg
      class="split-bar__svg"
      viewBox="0 0 200 12"
      preserveAspectRatio="none"
      role="img"
      :aria-label="`Text ${textPercent}%, Files ${filePercent}%`"
    >
      <defs>
        <linearGradient id="split-bar-grad-text" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" class="split-bar__grad-text-top" />
          <stop offset="1" class="split-bar__grad-text-bot" />
        </linearGradient>
        <linearGradient id="split-bar-grad-file" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" class="split-bar__grad-file-top" />
          <stop offset="1" class="split-bar__grad-file-bot" />
        </linearGradient>
      </defs>

      <rect
        v-if="total === 0"
        x="0"
        y="0"
        width="200"
        height="12"
        rx="6"
        class="split-bar__track"
      />
      <g v-else class="split-bar__fill">
        <path
          v-if="textWidth > 0"
          :d="textPath"
          class="split-bar__seg split-bar__seg--text"
        />
        <path
          v-if="fileWidth > 0"
          :d="filePath"
          class="split-bar__seg split-bar__seg--file"
        />
      </g>
    </svg>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  texts: { type: Number, default: 0 },
  files: { type: Number, default: 0 },
});

const barWidth = 200;
const barHeight = 12;
const radius = barHeight / 2;
const gap = 2;

const total = computed(() => (Number(props.texts) || 0) + (Number(props.files) || 0));

const textPercent = computed(() =>
  total.value > 0 ? Math.round(((Number(props.texts) || 0) / total.value) * 100) : 0
);
const filePercent = computed(() =>
  total.value > 0 ? Math.round(((Number(props.files) || 0) / total.value) * 100) : 0
);

const usableWidth = barWidth - gap;

const textWidth = computed(() =>
  total.value > 0 ? usableWidth * ((Number(props.texts) || 0) / total.value) : 0
);
const fileWidth = computed(() =>
  total.value > 0 ? usableWidth * ((Number(props.files) || 0) / total.value) : 0
);

/** Rounded-left / square-right rect path. */
function leftRoundedPath(width, roundRight) {
  const r = Math.max(0, Math.min(radius, width / 2, barHeight / 2));
  if (roundRight) {
    // The only segment present - round both ends (a full pill).
    return `M${r},0 H${width - r} A${r},${r} 0 0 1 ${width},${r} V${barHeight - r} A${r},${r} 0 0 1 ${width - r},${barHeight} H${r} A${r},${r} 0 0 1 0,${barHeight - r} V${r} A${r},${r} 0 0 1 ${r},0 Z`;
  }
  return `M${r},0 H${width} V${barHeight} H${r} A${r},${r} 0 0 1 0,${barHeight - r} V${r} A${r},${r} 0 0 1 ${r},0 Z`;
}

/** Square-left / rounded-right rect path, offset to start at x. */
function rightRoundedPath(x, width, roundLeft) {
  const r = Math.max(0, Math.min(radius, width / 2, barHeight / 2));
  if (roundLeft) {
    // The only segment present - round both ends (a full pill).
    return `M${x + r},0 H${x + width - r} A${r},${r} 0 0 1 ${x + width},${r} V${barHeight - r} A${r},${r} 0 0 1 ${x + width - r},${barHeight} H${x + r} A${r},${r} 0 0 1 ${x},${barHeight - r} V${r} A${r},${r} 0 0 1 ${x + r},0 Z`;
  }
  return `M${x},0 H${x + width - r} A${r},${r} 0 0 1 ${x + width},${r} V${barHeight - r} A${r},${r} 0 0 1 ${x + width - r},${barHeight} H${x} Z`;
}

const textPath = computed(() => leftRoundedPath(textWidth.value, fileWidth.value === 0));
const filePath = computed(() =>
  rightRoundedPath(textWidth.value > 0 ? textWidth.value + gap : 0, fileWidth.value, textWidth.value === 0)
);
</script>

<style scoped>
.split-bar {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.split-bar__labels {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.87);
}

.split-bar__svg {
  width: 100%;
  height: 12px;
  display: block;
}

.split-bar__track {
  fill: rgba(var(--v-theme-on-surface), 0.08);
}

.split-bar__grad-text-top { stop-color: rgb(var(--v-theme-chart-text)); stop-opacity: 1; }
.split-bar__grad-text-bot { stop-color: rgb(var(--v-theme-chart-text)); stop-opacity: 0.72; }
.split-bar__grad-file-top { stop-color: rgb(var(--v-theme-chart-file)); stop-opacity: 1; }
.split-bar__grad-file-bot { stop-color: rgb(var(--v-theme-chart-file)); stop-opacity: 0.72; }

.split-bar__seg--text {
  fill: url(#split-bar-grad-text);
}

.split-bar__seg--file {
  fill: url(#split-bar-grad-file);
}

/* Sweep the ratio in from the left on reveal. */
.split-bar__fill {
  transform-box: fill-box;
  transform-origin: left;
  animation: gd-fill-x 620ms cubic-bezier(0.22, 1, 0.36, 1) both;
  animation-delay: 120ms;
}

@media (prefers-reduced-motion: reduce) {
  .split-bar__fill {
    animation: none;
  }
}
</style>
