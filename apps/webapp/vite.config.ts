import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
  plugins: [solidPlugin()],
  server: {
    port: 4000,
    proxy: {
      "/api": "http://localhost:4100",
    },
  },
  build: {
    target: "esnext",
  },
});
