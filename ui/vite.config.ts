import { defineConfig } from "vite";
import { coverageConfigDefaults, configDefaults } from 'vitest/config'

import react from "@vitejs/plugin-react";
import oxlintPlugin from 'vite-plugin-oxlint'

export default defineConfig({
  base: '/immotep/',
  plugins: [react(), oxlintPlugin()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        secure: false,
      },
    },
  },
  test: {
    environment: 'happy-dom',
    globals: true,
    root: __dirname,
    setupFiles: ['./vitest.setup.ts'],
    include: ['./frontend/**/*.test.{js,jsx}'],
    exclude: [...configDefaults.exclude, './mock/**'],
    coverage: {
      thresholds: {
        statements: 50,
        functions: 90,
        branches: 80,
        lines: 50,
      },
      watermarks: {
        statements: [50, 60],
        functions: [50, 60],
        branches: [50, 60],
        lines: [50, 60]
      },
      exclude: [...coverageConfigDefaults.exclude, './mock/**'],
    },
  },
  
});
