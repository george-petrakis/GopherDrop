<template>
  <v-app>
    <v-main class="d-flex flex-column">
      <v-container fluid class="pa-0 flex-grow-1">
        <div class="gd-utility-wrap">
          <div class="gd-utility-line">
            <router-link to="/" class="gd-brand" @click="requestFormReset">
              <img src="./assets/Images/logo.png" alt="" height="24" width="24" />
              <span class="gd-brand__word font-display">GopherDrop</span>
            </router-link>

            <v-tooltip :text="isDarkMode ? 'Switch to Light Mode' : 'Switch to Dark Mode'">
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  variant="text"
                  density="comfortable"
                  icon
                  :aria-label="isDarkMode ? 'Switch to Light Mode' : 'Switch to Dark Mode'"
                  @click="toggleTheme"
                >
                  <v-icon>{{ isDarkMode ? 'mdi-weather-sunny' : 'mdi-weather-night' }}</v-icon>
                </v-btn>
              </template>
            </v-tooltip>
          </div>
        </div>

        <v-container class="mt-4">
          <router-view />
        </v-container>
      </v-container>

      <v-footer color="transparent" class="justify-center gd-footer pa-4">
        <span>
          © {{ new Date().getFullYear() }} GopherDrop | self-hosted & open source |
          <a
            href="https://github.com/kek-Sec/gopherdrop"
            target="_blank"
            rel="noopener noreferrer"
          >
            GitHub Repository
          </a>
        </span>
      </v-footer>
    </v-main>
  </v-app>
</template>

<script setup>
/**
 * The root component of the application.
 * Provides a header, navigation, and footer.
 */
import { ref, onMounted, computed } from 'vue';
import { formStore } from './stores/formStore.js';
import { useTheme } from 'vuetify';

const themeInstance = useTheme();
const isDarkMode = computed(() => themeInstance.global.current.value.dark);

// Ground color per theme, kept in sync with the <meta name="theme-color">
// so the mobile browser chrome matches the app's active theme (which is
// driven by localStorage, not the OS prefers-color-scheme).
const themeColors = {
  customLightTheme: '#FBFAFE',
  customDarkTheme: '#141218',
};

function updateThemeColorMeta(themeName) {
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) {
    meta.setAttribute('content', themeColors[themeName] || themeColors.customLightTheme);
  }
}

// On component mount, check localStorage for a saved theme
onMounted(() => {
  const savedTheme = localStorage.getItem('theme');
  if (savedTheme) {
    themeInstance.global.name.value = savedTheme;
  }
  updateThemeColorMeta(themeInstance.global.name.value);
});

function toggleTheme() {
  const newTheme = isDarkMode.value ? 'customLightTheme' : 'customDarkTheme';
  themeInstance.global.name.value = newTheme;
  // Save the new theme preference to localStorage
  localStorage.setItem('theme', newTheme);
  updateThemeColorMeta(newTheme);
}

/**
 * Triggers the form reset via the store.
 * This will be picked up by a watcher in Create.vue.
 */
function requestFormReset() {
  formStore.triggerReset();
}
</script>

<style scoped>
.v-footer a {
  font-weight: 500;
}

.v-main {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.gd-utility-wrap {
  max-width: 1120px;
  margin: 0 auto;
  padding: 0 24px;
}

.gd-utility-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
}

.gd-brand {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
}

.gd-brand__word {
  font-weight: 600;
  font-size: 1rem;
  color: rgba(var(--v-theme-on-surface), 0.85);
}

.gd-footer {
  margin-top: 64px;
  border-top: 1px solid rgba(var(--v-theme-on-surface), 0.1);
  font-size: 0.8125rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
}
</style>