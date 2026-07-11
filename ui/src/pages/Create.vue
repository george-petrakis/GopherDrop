<template>
  <v-container class="create-page">
    <v-row class="create-page__row">
      <v-col cols="12" md="7" order="1">
        <div class="create-page__title gd-rise">
          <h1 class="font-display create-page__h1">Drop a secret.</h1>
          <p class="create-page__tagline">Encrypted on this server &middot; delivered once &middot; then destroyed.</p>
        </div>

        <v-form v-if="!resultHash" @submit.prevent="handleSubmit" class="create-page__form gd-rise">
          <div class="gd-eyebrow-row">
            <span class="gd-eyebrow">Your secret</span>

            <v-btn-toggle
              v-model="type"
              mandatory
              density="compact"
              divided
              class="gd-mode-toggle"
            >
              <v-btn
                value="text"
                size="small"
                :variant="type === 'text' ? 'tonal' : 'outlined'"
                :color="type === 'text' ? 'primary' : undefined"
              >
                <v-icon start size="16">mdi-text</v-icon> Text
              </v-btn>
              <v-btn
                value="file"
                size="small"
                :variant="type === 'file' ? 'tonal' : 'outlined'"
                :color="type === 'file' ? 'primary' : undefined"
              >
                <v-icon start size="16">mdi-file</v-icon> File
              </v-btn>
            </v-btn-toggle>
          </div>

          <v-textarea
            v-if="type === 'text'"
            v-model="textSecret"
            required
            variant="plain"
            auto-grow
            no-resize
            rows="5"
            hide-details
            placeholder="Paste the secret. It never leaves this server unencrypted."
            class="gd-secret-well"
          ></v-textarea>

          <div
            v-if="type === 'file'"
            @dragover.prevent
            @dragenter.prevent
            @drop.prevent="onFileDrop"
          >
            <v-file-input
              v-model="files"
              label="Choose a file or drop it here"
              variant="plain"
              hide-details
              show-size
              required
              class="gd-secret-well gd-file-well"
            >
              <template v-slot:prepend-inner>
                <v-icon icon="mdi-tray-arrow-up" class="gd-file-well__icon" />
              </template>
            </v-file-input>
          </div>

          <div class="gd-eyebrow create-page__group-label">Delivery</div>

          <PasswordInput v-model="password" class="mb-4" />

          <div class="gd-delivery-row">
            <v-select
              label="Expires in"
              v-model="expires"
              :items="expirationOptions"
              required
              variant="outlined"
              density="comfortable"
              hide-details
              class="gd-delivery-row__select"
            ></v-select>

            <v-switch
              v-model="oneTime"
              label="Delete after first view"
              color="primary"
              inset
              hide-details
              class="gd-delivery-row__switch"
            ></v-switch>
          </div>

          <v-btn
            type="submit"
            color="primary"
            variant="flat"
            block
            height="48"
            class="create-page__submit"
          >
            Create secret
          </v-btn>

          <v-alert
            v-if="errorMessage"
            type="error"
            variant="tonal"
            class="mt-4 gd-rise"
          >
            {{ errorMessage }}
          </v-alert>
        </v-form>

        <SecretResult
          v-if="resultHash"
          :link="`${baseUrl}/view/${resultHash}`"
          :expires="createdExpiresLabel"
          :one-time="createdOneTime"
          @create-another="resetForm"
        />

        <v-overlay v-model="loading" class="align-center justify-center">
          <v-progress-circular indeterminate color="primary" size="64"></v-progress-circular>
        </v-overlay>
      </v-col>

      <v-col cols="12" md="5" order="2" class="create-page__ledger-col gd-rise">
        <ServerLedger ref="ledger" />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { createSend } from '../services/api.js';
import PasswordInput from '../components/PasswordInput.vue';
import SecretResult from '../components/SecretResult.vue';
import ServerLedger from '../components/stats/ServerLedger.vue';
import { formStore } from '../stores/formStore.js';

const type = ref('text');
const textSecret = ref('');
// v-file-input uses an array for its model, so initialize it as such.
const files = ref([]);
const password = ref('');
const oneTime = ref(false);
const expires = ref('24h');
const errorMessage = ref('');
const resultHash = ref('');
const baseUrl = window.location.origin;
const loading = ref(false);
const ledger = ref(null);

// Captured at the moment of a successful submit (before the fields are
// cleared) so SecretResult can render an accurate recap line.
const createdExpires = ref('');
const createdOneTime = ref(false);

const expirationOptions = [
  { title: '1 Hour', value: '1h' },
  { title: '6 Hours', value: '6h' },
  { title: '12 Hours', value: '12h' },
  { title: '24 Hours', value: '24h' },
  { title: '3 Days', value: '72h' },
  { title: '1 Week', value: '168h' }
];

const createdExpiresLabel = computed(() => {
  const match = expirationOptions.find((option) => option.value === createdExpires.value);
  return match ? match.title : createdExpires.value;
});

function resetForm() {
  type.value = 'text';
  textSecret.value = '';
  files.value = [];
  password.value = '';
  oneTime.value = false;
  expires.value = '24h';
  errorMessage.value = '';
  resultHash.value = '';
  loading.value = false;
}

watch(() => formStore.resetCounter, () => {
  resetForm();
});

// Accept a file dropped anywhere on the file well. Keeps files.value an
// array-with-a-File, matching what handleSubmit expects.
function onFileDrop(event) {
  const dropped = event.dataTransfer?.files;
  if (dropped && dropped.length) {
    files.value = [dropped[0]];
  }
}

async function handleSubmit() {
  errorMessage.value = '';
  resultHash.value = '';
  loading.value = true;

  const formData = new FormData();
  formData.append('type', type.value);

  if (type.value === 'text') {
    if (!textSecret.value.trim()) {
      errorMessage.value = 'Please enter some text';
      loading.value = false;
      return;
    }
    formData.append('data', textSecret.value);
  } else if (type.value === 'file') {
    // Ensure files.value is an array and has a File object
    const fileArr = Array.isArray(files.value) ? files.value : (files.value ? [files.value] : []);
    if (!fileArr.length || !(fileArr[0] instanceof File)) {
      errorMessage.value = 'Please select a file 😟';
      loading.value = false;
      return;
    }
    const fileToUpload = fileArr[0];
    formData.append('file', fileToUpload, fileToUpload.name);
  }

  if (password.value.trim()) {
    formData.append('password', password.value);
  }
  if (oneTime.value) {
    formData.append('onetime', 'true');
  }
  formData.append('expires', expires.value);

  try {
    const result = await createSend(formData);
    const createdTypeValue = type.value;
    createdExpires.value = expires.value;
    createdOneTime.value = oneTime.value;
    resultHash.value = result.hash;
    ledger.value?.tick(createdTypeValue);
    // Clear form inputs but keep the result hash visible
    type.value = 'text';
    textSecret.value = '';
    files.value = [];
    password.value = '';
    oneTime.value = false;
    expires.value = '24h';
  } catch (err) {
    errorMessage.value = err.message || 'Failed to create secret';
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.create-page {
  min-height: 85vh;
}

.create-page__row {
  align-items: flex-start;
}

.create-page__title {
  margin-bottom: 40px;
}

.create-page__h1 {
  font-weight: 700;
  font-size: 1.75rem;
  line-height: 1.2;
  text-align: left;
  margin: 0;
}

@media (min-width: 960px) {
  .create-page__h1 {
    font-size: 2rem;
  }
}

.create-page__tagline {
  margin: 8px 0 0;
  font-size: 0.9375rem;
  color: rgba(var(--v-theme-on-surface), 0.65);
}

.create-page__form {
  animation-delay: 60ms;
}

.create-page__ledger-col {
  animation-delay: 120ms;
  padding-top: 32px;
}

@media (min-width: 960px) {
  .create-page__ledger-col {
    padding-top: 0;
    padding-left: 48px;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.10);
  }
}

@media (max-width: 959.98px) {
  .create-page__ledger-col {
    margin-top: 8px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.10);
  }
}

/* "YOUR SECRET" eyebrow + Text|File segmented control */
.gd-eyebrow-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.create-page__group-label {
  margin: 24px 0 8px;
}

.gd-mode-toggle :deep(.v-btn) {
  text-transform: none;
  letter-spacing: normal;
}

/* Shared footprint for text + file wells so switching modes never jumps
   the layout. */
.gd-secret-well :deep(.v-field) {
  min-height: 160px;
  background: rgba(var(--v-theme-primary), 0.04);
  border: 1px solid rgba(var(--v-theme-on-surface), 0.14);
  border-radius: 10px;
  box-shadow: none;
}

.gd-secret-well :deep(.v-field.v-field--focused) {
  border: 1.5px solid rgb(var(--v-theme-primary));
  box-shadow: none;
}

.gd-secret-well :deep(textarea) {
  padding: 12px;
}

.gd-file-well :deep(.v-field__input) {
  align-items: center;
}

.gd-file-well__icon {
  color: rgba(var(--v-theme-on-surface), 0.5);
}

.gd-file-well :deep(.v-chip) {
  background: rgba(var(--v-theme-chart-file), 0.14);
  color: rgb(var(--v-theme-chart-file));
}

.gd-delivery-row {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-top: 8px;
}

.gd-delivery-row__select {
  flex: 1;
}

.gd-delivery-row__switch {
  flex-shrink: 0;
}

@media (max-width: 599px) {
  .gd-delivery-row {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
}

.create-page__submit {
  margin-top: 24px;
  border-radius: 10px;
}
</style>
