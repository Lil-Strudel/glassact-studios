import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import checker from "vite-plugin-checker";

export default defineConfig({
  plugins: [
    tanstackRouter({
      target: "solid",
      autoCodeSplitting: true,
    }),
    solidPlugin(),
    checker({
      typescript: true,
    }),
  ],
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
