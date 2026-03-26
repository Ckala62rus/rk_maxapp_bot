// Vite configuration for Vue dev server and proxy.
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    // Allow all tuna subdomains in dev (tunnel hosts).
    allowedHosts: [".tuna.am", ".ru.tuna.am"],
    proxy: {
      // Proxy API calls to backend container in dev.
      "/api": {
        target: "http://backend-dev-live:3000",
        changeOrigin: true,
        timeout: 120000,
        proxyTimeout: 120000,
      },
    },
  },
});
