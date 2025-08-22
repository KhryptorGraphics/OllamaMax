import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'
import checker from 'vite-plugin-checker'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    checker({
      typescript: true,
      // Disable ESLint in dev server for Playwright webServer stability
      // Run lint separately in CI
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
      '@/components': resolve(__dirname, './src/components'),
      '@/services': resolve(__dirname, './src/services'),
      '@/hooks': resolve(__dirname, './src/hooks'),
      '@/utils': resolve(__dirname, './src/utils'),
      '@/types': resolve(__dirname, './src/types'),
      '@/stores': resolve(__dirname, './src/stores'),
      '@/assets': resolve(__dirname, './src/assets'),
      '@/theme': resolve(__dirname, './src/theme'),
    },
  },
  server: {
    port: 5174,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        secure: false,
      },
      '/ws': {
        target: 'ws://localhost:8083',
        ws: true,
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'esnext',
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: true,
    rollupOptions: {
      external: (id) => {
        // Exclude test files from build
        return id.includes('.test.') || id.includes('.spec.') || id.includes('/tests/');
      },
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          router: ['react-router-dom'],
          ui: ['styled-components'],
          state: ['zustand'],
        },
      },
    },
    // Security headers for production
    cssCodeSplit: true,
    reportCompressedSize: false,
    chunkSizeWarningLimit: 1000,
  },
  optimizeDeps: {
    include: ['react', 'react-dom', 'react-router-dom', 'styled-components', 'zustand'],
  },
  define: {
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version),
    __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
  },
  css: {
    modules: {
      localsConvention: 'camelCase',
    },
    devSourcemap: true,
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/tests/setup.ts'],
    css: true,
  },
})