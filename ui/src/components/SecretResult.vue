<template>
  <div class="gd-receipt gd-rise">
    <div class="gd-receipt__rule" aria-hidden="true"></div>

    <div class="gd-receipt__body">
      <div class="gd-receipt__heading font-display">Secret created</div>
      <p class="gd-receipt__hint">Share this link to view the secret:</p>

      <div class="gd-receipt__link-row">
        <span class="gd-receipt__link">{{ link }}</span>
        <v-tooltip text="Copy Link to Clipboard">
          <template v-slot:activator="{ props }">
            <v-btn
              variant="text"
              icon
              v-bind="props"
              @click="copyLink"
              size="small"
              aria-label="Copy link to clipboard"
            >
              <v-icon>mdi-content-copy</v-icon>
            </v-btn>
          </template>
        </v-tooltip>
      </div>

      <p v-if="recapText" class="gd-receipt__recap">{{ recapText }}</p>

      <v-btn
        color="primary"
        variant="flat"
        class="mt-4 gd-receipt__cta"
        height="44"
        @click="$emit('create-another')"
      >
        Create another
      </v-btn>
    </div>

    <v-snackbar v-model="snackbar" timeout="2000" color="success">
      Link copied to clipboard!
    </v-snackbar>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';

const props = defineProps({
  link: { type: String, required: true },
  // Optional recap props. When absent (or expires is falsy) the recap
  // line is simply omitted and the receipt still renders cleanly.
  expires: { type: String, default: null },
  oneTime: { type: Boolean, default: false },
});

defineEmits(['create-another']);

const snackbar = ref(false);

const recapText = computed(() => {
  if (!props.expires) return null;
  return props.oneTime
    ? `Expires in ${props.expires} · deletes after first view`
    : `Expires in ${props.expires}`;
});

function copyLink() {
  navigator.clipboard.writeText(props.link);
  snackbar.value = true;
}
</script>

<style scoped>
.gd-receipt {
  position: relative;
  display: flex;
  margin-top: 24px;
  border: 1px solid rgba(var(--v-theme-on-surface), 0.14);
  border-radius: 10px;
  overflow: hidden;
}

.gd-receipt__rule {
  width: 4px;
  flex-shrink: 0;
  background: rgb(var(--v-theme-success));
}

.gd-receipt__body {
  flex: 1;
  min-width: 0;
  padding: 20px 24px;
}

.gd-receipt__heading {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 8px;
}

.gd-receipt__hint {
  margin: 0 0 8px;
  font-size: 0.875rem;
  color: rgba(var(--v-theme-on-surface), 0.7);
}

.gd-receipt__link-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 8px 8px 12px;
  background: rgba(var(--v-theme-on-surface), 0.05);
  border-radius: 8px;
}

.gd-receipt__link {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: ui-monospace, 'SF Mono', Menlo, Consolas, monospace;
  font-size: 0.8125rem;
}

.gd-receipt__recap {
  margin: 12px 0 0;
  font-size: 0.8125rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
}

.gd-receipt__cta {
  border-radius: 10px;
}

@media (max-width: 599px) {
  .gd-receipt__body {
    padding: 16px;
  }
}
</style>
