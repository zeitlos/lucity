import { execSync } from 'node:child_process';
import { fileURLToPath, URL } from 'node:url';

import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import vueDevTools from 'vite-plugin-vue-devtools';
import tailwindcss from '@tailwindcss/vite';

// https://vite.dev/config/
const appVersion = process.env.APP_VERSION || (() => {
  try {
    return execSync('git describe --always --dirty', { encoding: 'utf-8' }).trim();
  } catch {
    return 'dev';
  }
})();

export default defineConfig({
  base: process.env.VITE_BASE_PATH || '/',
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  plugins: [
    vue(),
    vueDevTools(),
    tailwindcss(),
  ],
  build: {
    rollupOptions: {
      output: {
        advancedChunks: {
          groups: [
            { name: 'vue-vendor', test: /node_modules\/(?:vue|vue-router|@vue)\// },
            { name: 'apollo', test: /node_modules\/(?:@apollo\/client|graphql)\// },
            { name: 'ui', test: /node_modules\/reka-ui\// },
          ],
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/graphql': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
      },
      '/auth': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/version': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
