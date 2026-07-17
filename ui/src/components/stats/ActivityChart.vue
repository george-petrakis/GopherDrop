<template>
  <div class="activity-chart" ref="wrapRef">
    <div class="activity-chart__header">
      <span class="activity-chart__title gd-eyebrow">Daily activity</span>
      <span class="activity-chart__summary">{{ summaryText }}</span>
    </div>

    <div class="activity-chart__plot">
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
          :class="{
            'activity-chart__col--hover': hoveredIndex === i,
            'activity-chart__col--today': i === daily.length - 1 && !isEmpty,
          }"
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

      <!-- 7-day momentum: a trailing average riding over the bars, so a run
           of busy days reads as a rising trend and not just tall columns.
           pathLength normalizes the length so the draw-in works regardless of
           how many days are plotted. -->
      <path
        v-if="showTrend"
        :d="trendPath"
        class="activity-chart__trend"
        pathLength="1"
        fill="none"
      />
    </svg>

    <!-- "Now" marker for the trend, as HTML so the non-uniform viewBox scaling
         can't squash it into an ellipse. -->
    <div
      v-if="showTrend"
      class="activity-chart__trend-head"
      :style="trendHeadStyle"
      aria-hidden="true"
    />

    <div
      v-if="hoveredIndex !== null"
      class="activity-chart__tooltip"
      :style="tooltipStyle"
    >
      {{ dayTooltip(daily[hoveredIndex]) }}
    </div>
    </div>

    <div v-if="!isEmpty" class="activity-chart__axis">
      <span>{{ firstDateLabel }}</span>
      <span class="activity-chart__axis-today">Today</span>
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

/* ---- Period summary + momentum trend ---- */

const dayTotals = computed(() =>
  props.daily.map((d) => (Number(d?.texts) || 0) + (Number(d?.files) || 0))
);

const peak = computed(() => {
  let value = 0;
  let index = -1;
  dayTotals.value.forEach((total, i) => {
    if (total > value) {
      value = total;
      index = i;
    }
  });
  return { value, index };
});

// One-line context so the chart states what it's showing rather than leaving
// the reader to eyeball it: the period's volume and its busiest day.
const summaryText = computed(() => {
  if (isEmpty.value) return 'No drops yet';
  const total = totals.value.all;
  const noun = total === 1 ? 'secret' : 'secrets';
  const peakDay = formatShortDate(props.daily[peak.value.index]?.date);
  return `${compactNumber(total)} ${noun} · peak ${peak.value.value}${peakDay ? ` on ${peakDay}` : ''}`;
});

// Trailing average (up to a 7-day window) sampled at each day's centre. Riding
// it over the raw bars turns a cluster of busy days into a visible upswing.
const trendPoints = computed(() => {
  const totalsArr = dayTotals.value;
  const n = totalsArr.length;
  if (n === 0) return [];
  const window = Math.min(7, n);
  const points = [];
  for (let i = 0; i < n; i += 1) {
    let sum = 0;
    let count = 0;
    for (let j = Math.max(0, i - window + 1); j <= i; j += 1) {
      sum += totalsArr[j];
      count += 1;
    }
    const avg = count > 0 ? sum / count : 0;
    points.push({ x: slotX(i) + slotWidth.value / 2, y: yForValue(avg) });
  }
  return points;
});

// Needs at least a few days of real activity to mean anything.
const showTrend = computed(() => !isEmpty.value && trendPoints.value.length >= 3);

// Smooth cubic through the points with horizontal tangents at each node — a
// calm curve that reads as a trend line, not a jagged connect-the-dots.
const trendPath = computed(() => {
  const pts = trendPoints.value;
  if (pts.length < 2) return '';
  let d = `M${pts[0].x},${pts[0].y}`;
  for (let i = 1; i < pts.length; i += 1) {
    const midX = (pts[i - 1].x + pts[i].x) / 2;
    d += ` C${midX},${pts[i - 1].y} ${midX},${pts[i].y} ${pts[i].x},${pts[i].y}`;
  }
  return d;
});

function formatShortDate(dateStr) {
  if (!dateStr) return '';
  const [y, m, d] = String(dateStr).split('-').map(Number);
  if (!y || !m || !d) return '';
  const date = new Date(y, m - 1, d);
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

const firstDateLabel = computed(() => formatShortDate(props.daily[0]?.date));

const trendHeadStyle = computed(() => {
  const pts = trendPoints.value;
  if (!pts.length) return {};
  const last = pts[pts.length - 1];
  return {
    left: `${(last.x / viewBoxWidth) * 100}%`,
    top: `${(last.y / viewBoxHeight) * 100}%`,
  };
});
</script>

<style scoped>
.activity-chart {
  position: relative;
  width: 100%;
}

.activity-chart__header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.activity-chart__summary {
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
  font-variant-numeric: tabular-nums;
  text-align: right;
}

.activity-chart__plot {
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

/* Today reads as "live": a gentle standing glow, so the most recent day is
   findable at a glance without a hard outline. */
.activity-chart__col--today {
  filter: drop-shadow(0 0 4px rgba(var(--v-theme-on-surface), 0.28));
}

/* Momentum line: a calm near-neutral stroke that sits above the coloured bars
   as an analytical overlay, drawing itself in left-to-right after the bars. */
.activity-chart__trend {
  stroke: rgba(var(--v-theme-on-surface), 0.6);
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;
  vector-effect: non-scaling-stroke;
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: gd-draw 900ms ease-out 260ms forwards;
}

@keyframes gd-draw {
  to { stroke-dashoffset: 0; }
}

.activity-chart__trend-head {
  position: absolute;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  background: rgb(var(--v-theme-on-surface));
  box-shadow: 0 0 0 3px rgba(var(--v-theme-surface), 1), 0 0 8px rgba(var(--v-theme-on-surface), 0.5);
  pointer-events: none;
  animation: gd-fade-in 300ms ease-out 1100ms both;
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

.activity-chart__axis {
  display: flex;
  justify-content: space-between;
  margin-top: 6px;
  padding: 0 2px;
  font-size: 0.6875rem;
  color: rgba(var(--v-theme-on-surface), 0.5);
  font-variant-numeric: tabular-nums;
}

.activity-chart__axis-today {
  color: rgba(var(--v-theme-on-surface), 0.7);
  font-weight: 600;
}

@media (prefers-reduced-motion: reduce) {
  .activity-chart__seg {
    transition: none;
  }
  .activity-chart__col {
    animation: none;
  }
  .activity-chart__trend {
    animation: none;
    stroke-dashoffset: 0;
  }
  .activity-chart__trend-head {
    animation: none;
  }
}
</style>
