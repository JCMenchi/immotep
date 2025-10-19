import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import oxlintPlugin from 'vite-plugin-oxlint'

export default defineConfig({
  base: '/immotep/',
  plugins: [react(), oxlintPlugin()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
});
