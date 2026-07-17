<template>
  <div class="activity-chart" ref="wrapRef">
    <svg
      class="activity-chart__svg"
      :viewBox="`0 0 ${viewBoxWidth} ${viewBoxHeight}`"
      preserveAspectRatio="none"
      role="img"
      :aria-label="ariaLabel"
    >
      <defs>
        <linearGradient :id="gradTextId" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" class="activity-chart__grad-text-top" />
          <stop offset="1" class="activity-chart__grad-text-bot" />
        </linearGradient>
        <linearGradient :id="gradFileId" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" class="activity-chart__grad-file-top" />
          <stop offset="1" class="activity-chart__grad-file-bot" />
        </linearGradient>
      </defs>

      <!-- gridlines -->
      <g class="activity-chart__grid">
        <line
          v-for="(tick, i) in ticks"
          :key="`grid-${i}`"
          :x1="leftPad"
          :x2="viewBoxWidth - rightPad"
          :y1="yForValue(tick)"
          :y2="yForValue(tick)"
          :class="tick === 0 ? 'activity-chart__baseline' : 'activity-chart__gridline'"
        />
        <text
          v-for="(tick, i) in ticks"
          :key="`ticklabel-${i}`"
          :x="leftPad - 6"
          :y="yForValue(tick) + 3"
          class="activity-chart__ticklabel"
          text-anchor="end"
        >{{ compactNumber(tick) }}</text>
      </g>

      <!-- columns -->
      <g v-for="(day, i) in daily" :key="day.date || i">
        <g
          class="activity-chart__col"
          :class="{ 'activity-chart__col--hover': hoveredIndex === i }"
          :style="colStyle(i)"
        >
          <template v-if="!isEmpty">
            <path
              v-if="segmentHeight(day, 'files') > 0"
              :d="topRectPath(
                slotX(i) + colOffset(i),
                yForValue(day.texts || 0) - segmentHeight(day, 'files') - (segmentHeight(day, 'texts') > 0 ? gap : 0),
                colWidth(i),
                segmentHeight(day, 'files'),
                4
              )"
              class="activity-chart__seg activity-chart__seg--file"
              :class="{ 'activity-chart__seg--hover': hoveredIndex === i }"
            />
            <path
              v-if="segmentHeight(day, 'texts') > 0"
              :d="topRectPath(
                slotX(i) + colOffset(i),
                baselineY - segmentHeight(day, 'texts'),
                colWidth(i),
                segmentHeight(day, 'texts'),
                segmentHeight(day, 'files') > 0 ? 0 : 4
              )"
              class="activity-chart__seg activity-chart__seg--text"
              :class="{ 'activity-chart__seg--hover': hoveredIndex === i }"
            />
          </template>
          <path
            v-else
            :d="topRectPath(slotX(i) + colOffset(i), baselineY - ghostHeight(i), colWidth(i), ghostHeight(i), 4)"
            class="activity-chart__seg activity-chart__seg--ghost"
          />
        </g>

        <!-- hit target: full slot band, full plot height -->
        <rect
          :x="slotX(i)"
          :y="topPad"
          :width="slotWidth"
          :height="plotHeight"
          class="activity-chart__hit"
          tabindex="0"
          :aria-label="dayTooltip(day)"
          @mouseenter="hoveredIndex = i"
          @mouseleave="hoveredIndex = null"
          @focus="hoveredIndex = i"
          @blur="hoveredIndex = null"
        />
      </g>
    </svg>

    <div
      v-if="hoveredIndex !== null"
      class="activity-chart__tooltip"
      :style="tooltipStyle"
    >
      {{ dayTooltip(daily[hoveredIndex]) }}
    </div>

    <div v-if="isEmpty" class="activity-chart__caption">Activity will appear here</div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';
import { compactNumber } from '../../utils/format.js';

const props = defineProps({
  daily: { type: Array, default: () => [] },
});

const wrapRef = ref(null);
const hoveredIndex = ref(null);

// Namespaced gradient ids so the defs never clash with anything else on the page.
const gradTextId = 'activity-chart-grad-text';
const gradFileId = 'activity-chart-grad-file';

// Staggered rise: each column starts a beat after the previous so the row
// unfurls left-to-right. Capped so a long history never drags the reveal out.
function colStyle(i) {
  return { animationDelay: `${Math.min(i * 22, 420)}ms` };
}

const viewBoxWidth = 600;
const viewBoxHeight = 150;
const leftPad = 30;
const rightPad = 4;
const topPad = 10;
const bottomPad = 6;
const gap = 2;

const plotWidth = viewBoxWidth - leftPad - rightPad;
const plotHeight = viewBoxHeight - topPad - bottomPad;
const baselineY = topPad + plotHeight;

const dayCount = computed(() => Math.max(props.daily.length, 1));
const slotWidth = computed(() => plotWidth / dayCount.value);

const totals = computed(() => {
  let texts = 0;
  let files = 0;
  for (const day of props.daily) {
    texts += Number(day?.texts) || 0;
    files += Number(day?.files) || 0;
  }
  return { texts, files, all: texts + files };
});

const isEmpty = computed(() => totals.value.all === 0);

const maxDailyTotal = computed(() => {
  let max = 0;
  for (const day of props.daily) {
    const total = (Number(day?.texts) || 0) + (Number(day?.files) || 0);
    if (total > max) max = total;
  }
  return max;
});

/** "Nice" rounded step for an axis, biased toward whole-number counts. */
function niceStep(rawStep) {
  if (rawStep <= 0) return 1;
  const exponent = Math.floor(Math.log10(rawStep));
  const fraction = rawStep / Math.pow(10, exponent);
  let niceFraction;
  if (fraction < 1.5) niceFraction = 1;
  else if (fraction < 3) niceFraction = 2;
  else if (fraction < 7) niceFraction = 5;
  else niceFraction = 10;
  return Math.max(1, Math.round(niceFraction * Math.pow(10, exponent)));
}

const ticks = computed(() => {
  const tickCount = 4; // baseline (0) + 3 gridlines
  if (maxDailyTotal.value <= 0) {
    return [0, 1, 2, 3];
  }
  const step = niceStep(maxDailyTotal.value / (tickCount - 1));
  return Array.from({ length: tickCount }, (_, i) => step * i);
});

const scaleMax = computed(() => ticks.value[ticks.value.length - 1] || 1);

function yForValue(value) {
  const ratio = scaleMax.value > 0 ? value / scaleMax.value : 0;
  return baselineY - ratio * plotHeight;
}

function segmentHeight(day, key) {
  const value = Number(day?.[key]) || 0;
  return baselineY - yForValue(value);
}

function slotX(i) {
  return leftPad + i * slotWidth.value;
}

function colWidth(i) {
  return Math.min(24, slotWidth.value - gap);
}

function colOffset(i) {
  return (slotWidth.value - colWidth(i)) / 2;
}

function ghostHeight(i) {
  // Deterministic gentle variation so the empty state doesn't look like a flat wall.
  return plotHeight * (0.12 + 0.35 * Math.abs(Math.sin(i * 1.7 + 1)));
}

/** Rect path with rounded top corners and a square baseline. */
function topRectPath(x, y, width, height, radius) {
  const r = Math.max(0, Math.min(radius, height / 2, width / 2));
  if (r === 0) {
    return `M${x},${y} H${x + width} V${y + height} H${x} Z`;
  }
  return [
    `M${x},${y + r}`,
    `A${r},${r} 0 0 1 ${x + r},${y}`,
    `H${x + width - r}`,
    `A${r},${r} 0 0 1 ${x + width},${y + r}`,
    `V${y + height}`,
    `H${x}`,
    `Z`,
  ].join(' ');
}

function formatDayLabel(dateStr) {
  if (!dateStr) return '';
  const [y, m, d] = String(dateStr).split('-').map(Number);
  if (!y || !m || !d) return dateStr;
  const date = new Date(y, m - 1, d);
  return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
}

function dayTooltip(day) {
  if (!day) return '';
  const texts = Number(day.texts) || 0;
  const files = Number(day.files) || 0;
  return `${formatDayLabel(day.date)} – ${texts} text${texts === 1 ? '' : 's'}, ${files} file${files === 1 ? '' : 's'}`;
}

const ariaLabel = computed(() => {
  const { texts, files, all } = totals.value;
  return `${all} secrets over the last ${props.daily.length} days, ${texts} text and ${files} files`;
});

const tooltipStyle = computed(() => {
  if (hoveredIndex.value === null) return {};
  const centerX = slotX(hoveredIndex.value) + slotWidth.value / 2;
  const leftPercent = (centerX / viewBoxWidth) * 100;
  return {
    left: `${leftPercent}%`,
    top: `${(topPad / viewBoxHeight) * 100}%`,
  };
});
</script>

<style scoped>
.activity-chart {
  position: relative;
  width: 100%;
}

.activity-chart__svg {
  width: 100%;
  height: 150px;
  display: block;
  overflow: visible;
}

.activity-chart__gridline {
  stroke: rgba(var(--v-theme-on-surface), 0.08);
  stroke-width: 1;
  vector-effect: non-scaling-stroke;
}

.activity-chart__baseline {
  stroke: rgba(var(--v-theme-on-surface), 0.2);
  stroke-width: 1;
  vector-effect: non-scaling-stroke;
}

.activity-chart__ticklabel {
  font-size: 9px;
  fill: rgba(var(--v-theme-on-surface), 0.5);
}

/* Vertical gradients: saturated at the top edge, softening toward the
   baseline so each bar reads as rising out of the ground. Theme-driven, so
   they flip with light/dark automatically. */
.activity-chart__grad-text-top { stop-color: rgb(var(--v-theme-chart-text)); stop-opacity: 1; }
.activity-chart__grad-text-bot { stop-color: rgb(var(--v-theme-chart-text)); stop-opacity: 0.5; }
.activity-chart__grad-file-top { stop-color: rgb(var(--v-theme-chart-file)); stop-opacity: 1; }
.activity-chart__grad-file-bot { stop-color: rgb(var(--v-theme-chart-file)); stop-opacity: 0.5; }

.activity-chart__seg--text {
  fill: url(#activity-chart-grad-text);
}

.activity-chart__seg--file {
  fill: url(#activity-chart-grad-file);
}

.activity-chart__seg--ghost {
  fill: rgba(var(--v-theme-on-surface), 0.08);
}

.activity-chart__seg {
  transition: opacity 0.15s ease;
}

.activity-chart__seg--hover {
  opacity: 1;
}

/* Each column grows up from the baseline; the per-column delay is set inline
   so the row unfurls left-to-right. transform-box keeps the origin pinned to
   the group's own bottom edge regardless of the non-uniform viewBox scaling. */
.activity-chart__col {
  transform-box: fill-box;
  transform-origin: bottom;
  animation: gd-bar-grow 560ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.activity-chart__col--hover {
  filter: drop-shadow(0 0 5px rgba(var(--v-theme-chart-text), 0.45));
}

.activity-chart__hit {
  fill: transparent;
  cursor: pointer;
  outline: none;
}

.activity-chart__hit:focus-visible {
  fill: rgba(var(--v-theme-on-surface), 0.04);
}

.activity-chart__tooltip {
  position: absolute;
  transform: translate(-50%, -100%);
  background: rgb(var(--v-theme-surface));
  color: rgb(var(--v-theme-on-surface));
  border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
  border-radius: 6px;
  padding: 4px 8px;
  font-size: 12px;
  white-space: nowrap;
  pointer-events: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  z-index: 1;
}

.activity-chart__caption {
  text-align: center;
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.5);
  margin-top: 4px;
}

@media (prefers-reduced-motion: reduce) {
  .activity-chart__seg {
    transition: none;
  }
  .activity-chart__col {
    animation: none;
  }
}
</style>
