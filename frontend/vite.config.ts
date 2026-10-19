import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

// Frontend runs on 4069, backend on 3069.
// /api is proxied in dev so the browser never needs CORS.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 4069,
    host: true,
    strictPort: true,
    proxy: {
      '/api': {
        target: process.env.VITE_PROXY_TARGET ?? 'http://localhost:3069',
        changeOrigin: true,
      },
    },
  },
  preview: {
    port: 4069,
    host: true,
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
})
