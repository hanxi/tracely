import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import ui from '@nuxt/ui/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  base: '/static/',
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  plugins: [
    ui({
      ui: {
        modal: {
          slots: {
            overlay: 'z-[100] fixed inset-0 bg-elevated/75',
            content: 'z-[101]'
          }
        },
        select: {
          slots: {
            content: 'z-[60]'
          }
        },
        dropdownMenu: {
          slots: {
            content: 'z-[60]'
          }
        }
      }
    }),
    vue(),
  ],
  server: {
    proxy: {
      '/api': { target: 'http://localhost:3001', changeOrigin: true },
      '/auth': { target: 'http://localhost:3001', changeOrigin: true },
    },
  },
  optimizeDeps: {
    exclude: ['@tailwindcss/oxide'],
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    chunkSizeWarningLimit: 800,
    rollupOptions: {
      output: {
        manualChunks: {
          'vue-vendor': ['vue', 'vue-router', 'pinia'],
        },
      },
    },
  },
})
