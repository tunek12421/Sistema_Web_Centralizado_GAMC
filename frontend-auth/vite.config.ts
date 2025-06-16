import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0', // Importante para Docker
    port: 5173,
    watch: {
      usePolling: true, // Para hot reload en Docker
    },
    hmr: {
      port: 5173, // Hot Module Replacement
    }
  }
})