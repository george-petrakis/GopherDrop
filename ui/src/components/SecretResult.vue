<template>
  <v-alert type="success" class="mt-6 animate__animated animate__fadeIn" variant="tonal">
    <div class="text-h6 mb-2">Secret Created!</div>
    <p>Share this link to view the secret:</p>
    <div class="d-flex align-center mt-2 pa-2" style="background-color: rgba(var(--v-theme-on-surface), 0.05); border-radius: 8px;">
      <span class="mr-2 text-truncate">{{ link }}</span>
      <v-spacer></v-spacer>
      <v-tooltip text="Copy Link to Clipboard">
        <template v-slot:activator="{ props }">
          <v-btn icon v-bind="props" @click="copyLink">
            <v-icon>mdi-content-copy</v-icon>
          </v-btn>
        </template>
      </v-tooltip>
    </div>

    <v-btn color="success" variant="outlined" class="mt-4" @click="$emit('create-another')">
      Create another
    </v-btn>

    <v-snackbar v-model="snackbar" timeout="2000" color="success">
      Link copied to clipboard!
    </v-snackbar>
  </v-alert>
</template>

<script setup>
import { ref } from 'vue';

const props = defineProps({
  link: { type: String, required: true },
});

defineEmits(['create-another']);

const snackbar = ref(false);

function copyLink() {
  navigator.clipboard.writeText(props.link);
  snackbar.value = true;
}
</script>
