<template>
  <v-container class="create-page">
    <div class="create-page__headline">
      <h1 class="font-display text-h4 text-md-h3 font-weight-bold">Share secrets that vanish.</h1>
      <p class="create-page__subhead">Encrypted on this server. Delivered once. Then destroyed.</p>
    </div>

    <v-row class="create-page__row">
      <v-col cols="12" md="7" order="1">
        <v-card class="pa-4 pa-md-8 animate__animated animate__fadeIn" elevation="6" rounded="lg">
          <v-card-title class="text-h5 text-md-h4 font-weight-bold text-center mb-4">Create a New Secret 🔑</v-card-title>
          <v-card-text>
            <v-form @submit.prevent="handleSubmit">
              <v-btn-toggle
                v-model="type"
                mandatory
                class="mb-4 d-flex justify-center"
                color="primary"
                rounded
                group
              >
                <v-btn value="text" class="px-6" rounded>
                  <v-icon left>mdi-text</v-icon> Text
                </v-btn>
                <v-btn value="file" class="px-6" rounded>
                  <v-icon left>mdi-file</v-icon> File
                </v-btn>
              </v-btn-toggle>

              <v-textarea
                v-if="type === 'text'"
                label="Text Secret"
                v-model="textSecret"
                required
                variant="outlined"
                rows="4"
              ></v-textarea>

              <v-file-input
                v-if="type === 'file'"
                label="Select File"
                prepend-icon="mdi-upload"
                v-model="files"
                show-size
                required
                variant="outlined"
              ></v-file-input>

              <PasswordInput v-model="password" class="mt-2" />

              <v-select
                label="Expiration"
                v-model="expires"
                :items="expirationOptions"
                required
                variant="outlined"
                class="mt-2"
              ></v-select>

              <v-checkbox
                v-model="oneTime"
                label="One-Time Retrieval"
                color="primary"
                class="mt-2"
              ></v-checkbox>

              <v-btn type="submit" color="primary" class="mt-4" block large rounded x-large height="50">Create Secret</v-btn>

              <v-alert v-if="errorMessage" type="error" class="mt-4 animate__animated animate__bounceIn" variant="tonal">
                {{ errorMessage }}
              </v-alert>
            </v-form>

            <SecretResult
              v-if="resultHash"
              :link="`${baseUrl}/view/${resultHash}`"
              @create-another="resetForm"
            />

            <v-overlay v-model="loading" class="align-center justify-center">
              <v-progress-circular indeterminate color="primary" size="64"></v-progress-circular>
            </v-overlay>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="5" order="2">
        <ServerLedger ref="ledger" />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, watch } from 'vue';
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

const expirationOptions = [
  { title: '1 Hour', value: '1h' },
  { title: '6 Hours', value: '6h' },
  { title: '12 Hours', value: '12h' },
  { title: '24 Hours', value: '24h' },
  { title: '3 Days', value: '72h' },
  { title: '1 Week', value: '168h' }
];

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
    const createdType = type.value;
    resultHash.value = result.hash;
    ledger.value?.tick(createdType);
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

.create-page__headline {
  text-align: center;
  margin-bottom: 24px;
}

.create-page__subhead {
  color: rgba(var(--v-theme-on-surface), 0.7);
  margin-top: 4px;
}

.create-page__row {
  align-items: flex-start;
}

.v-card {
  width: 100%;
}
</style>
