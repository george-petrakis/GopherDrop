<template>
  <div class="stat-tile" :class="{ 'stat-tile--hero': hero }">
    <div class="stat-tile__eyebrow">
      <span
        v-if="swatch"
        class="stat-tile__swatch"
        :class="`stat-tile__swatch--${swatch}`"
        aria-hidden="true"
      />
      <span class="stat-tile__label">{{ label }}</span>
    </div>

    <div
      class="stat-tile__value"
      :class="{ 'font-display': hero }"
    >
      {{ displayValue }}
    </div>

    <div v-if="caption" class="stat-tile__caption">{{ caption }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useCountUp } from '../../composables/useCountUp.js';
import { compactNumber, formatBytes } from '../../utils/format.js';

const props = defineProps({
  label: { type: String, required: true },
  value: { type: Number, required: true },
  caption: { type: String, default: null },
  hero: { type: Boolean, default: false },
  swatch: { type: String, default: null }, // "text" | "file" | null
  format: { type: String, default: 'number' }, // "number" | "bytes"
});

const target = computed(() => props.value);
const animated = useCountUp(target, { durationMs: 700 });

const displayValue = computed(() => {
  const rounded = Math.round(animated.value);
  return props.format === 'bytes' ? formatBytes(rounded) : compactNumber(rounded);
});
</script>

<style scoped>
.stat-tile {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-tile__eyebrow {
  display: flex;
  align-items: center;
  gap: 6px;
}

.stat-tile__swatch {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.stat-tile__swatch--text {
  background-color: rgb(var(--v-theme-chart-text));
}

.stat-tile__swatch--file {
  background-color: rgb(var(--v-theme-chart-file));
}

.stat-tile__label {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: rgba(var(--v-theme-on-surface), 0.6);
}

.stat-tile__value {
  font-size: 1.5rem;
  font-weight: 600;
  line-height: 1.2;
  color: rgb(var(--v-theme-on-surface));
  font-variant-numeric: tabular-nums;
}

.stat-tile--hero .stat-tile__value {
  font-size: 3rem;
  line-height: 1.1;
}

.stat-tile__value.font-display {
  font-family: 'Space Grotesk Variable', 'Space Grotesk', sans-serif;
}

.stat-tile__caption {
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
}
</style>
