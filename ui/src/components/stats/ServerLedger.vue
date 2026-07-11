<template>
  <div v-if="loading" class="server-ledger server-ledger--loading" aria-busy="true" aria-live="polite">
    <div class="server-ledger__skeleton-eyebrow" />
    <div class="server-ledger__skeleton-hero" />
    <div class="server-ledger__skeleton-row">
      <div class="server-ledger__skeleton-tile" v-for="n in 4" :key="n" />
    </div>
    <div class="server-ledger__skeleton-chart" />
  </div>

  <div v-else-if="stats" class="server-ledger gd-fade-in">
    <div class="server-ledger__eyebrow font-display">THIS SERVER</div>

    <StatTile
      class="server-ledger__hero"
      :class="{ 'server-ledger__hero--glow': glowing }"
      label="Secrets delivered"
      :value="local.total"
      :caption="heroCaption"
      hero
    />

    <v-divider class="server-ledger__divider" />

    <div class="server-ledger__grid">
      <StatTile label="Text secrets" :value="local.texts" swatch="text" />
      <StatTile label="Files shared" :value="local.files" swatch="file" />
      <StatTile label="Data encrypted" :value="local.bytes" format="bytes" />
      <StatTile label="Held right now" :value="local.active" caption="awaiting pickup" />
    </div>

    <SplitBar :texts="local.texts" :files="local.files" class="server-ledger__section" />

    <ActivityChart :daily="local.daily" class="server-ledger__section" />

    <MilestoneMeter :total="local.total" class="server-ledger__section" />
  </div>

  <!-- error/null: render nothing so the form column can center itself -->
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { getStats } from '../../services/api.js';
import StatTile from './StatTile.vue';
import ActivityChart from './ActivityChart.vue';
import SplitBar from './SplitBar.vue';
import MilestoneMeter from './MilestoneMeter.vue';

const loading = ref(true);
const stats = ref(null);
const glowing = ref(false);

// Local reactive copy of the lifetime figures + daily buckets, seeded from
// the fetched stats. tick() mutates these so StatTiles (which read only
// from `local`) animate a live +1 without waiting on the server.
const local = reactive({
  total: 0,
  texts: 0,
  files: 0,
  bytes: 0,
  active: 0,
  daily: [],
});

function prefersReducedMotion() {
  return (
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  );
}

onMounted(async () => {
  const result = await getStats();

  if (result) {
    stats.value = result;
    local.total = result.lifetime?.total ?? 0;
    local.texts = result.lifetime?.texts ?? 0;
    local.files = result.lifetime?.files ?? 0;
    local.bytes = result.lifetime?.bytes ?? 0;
    local.active = result.active?.total ?? 0;
    local.daily = (result.daily ?? []).map((day) => ({ ...day }));
  }

  loading.value = false;
});

const isEmpty = computed(() => local.total === 0);

const sinceLabel = computed(() => {
  if (!stats.value?.since) return '';
  const date = new Date(stats.value.since);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
});

const heroCaption = computed(() =>
  isEmpty.value ? 'Be the first to drop a secret' : `since ${sinceLabel.value}`
);

let glowTimeout = null;

/**
 * Optimistically bumps the local ledger after a successful create, so the
 * hero figure visibly counts up without waiting on a re-fetch.
 * @param {"text"|"file"} type
 */
function tick(type) {
  if (!stats.value) return;

  local.total += 1;

  const key = type === 'file' ? 'files' : 'texts';
  local[key] += 1;

  const today = local.daily[local.daily.length - 1];
  if (today) {
    today[key] = (today[key] || 0) + 1;
  }

  if (!prefersReducedMotion()) {
    glowing.value = true;
    clearTimeout(glowTimeout);
    glowTimeout = setTimeout(() => {
      glowing.value = false;
    }, 400);
  }
}

defineExpose({ tick });
</script>

<style scoped>
.server-ledger {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 8px 0;
}

.server-ledger__eyebrow {
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: rgba(var(--v-theme-on-surface), 0.55);
}

.server-ledger__divider {
  opacity: 0.6;
}

.server-ledger__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
}

.server-ledger__section {
  margin-top: 4px;
}

.server-ledger__hero :deep(.stat-tile__value) {
  transition: text-shadow 0.4s ease, color 0.4s ease;
}

.server-ledger__hero--glow :deep(.stat-tile__value) {
  color: rgb(var(--v-theme-primary));
  text-shadow: 0 0 18px rgba(var(--v-theme-primary), 0.55);
}

@media (prefers-reduced-motion: reduce) {
  .server-ledger__hero :deep(.stat-tile__value) {
    transition: none;
  }
}

/* Loading skeleton */
.server-ledger--loading {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 8px 0;
}

.server-ledger__skeleton-eyebrow,
.server-ledger__skeleton-hero,
.server-ledger__skeleton-tile,
.server-ledger__skeleton-chart {
  background: rgba(var(--v-theme-on-surface), 0.08);
  border-radius: 8px;
  animation: server-ledger-pulse 1.4s ease-in-out infinite;
}

.server-ledger__skeleton-eyebrow {
  width: 30%;
  height: 12px;
}

.server-ledger__skeleton-hero {
  width: 60%;
  height: 48px;
}

.server-ledger__skeleton-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
}

.server-ledger__skeleton-tile {
  height: 40px;
}

.server-ledger__skeleton-chart {
  height: 150px;
}

@keyframes server-ledger-pulse {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 1; }
}

@media (prefers-reduced-motion: reduce) {
  .server-ledger__skeleton-eyebrow,
  .server-ledger__skeleton-hero,
  .server-ledger__skeleton-tile,
  .server-ledger__skeleton-chart {
    animation: none;
  }
}
</style>
