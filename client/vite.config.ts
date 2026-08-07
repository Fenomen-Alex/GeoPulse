import { defineConfig } from 'vite'
import solid from 'vite-plugin-solid'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig({
  plugins: [tailwindcss(), solid()],
  // Point Vite to the root directory where .env lives
  envDir: '../',
  define: {
    // Expose NODE_ENV to the client so NODE_ENV=test can opt into test-mode bypass
    'import.meta.env.VITE_NODE_ENV': JSON.stringify(process.env.NODE_ENV ?? ''),
  },
  server: {
    port: 5173,
    proxy: {
      // Proxy local API calls to Go backend during dev
      // Override VITE_PROXY_TARGET for containerized dev (e.g. http://api:8080)
      '/api': {
        target: process.env.VITE_PROXY_TARGET || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
