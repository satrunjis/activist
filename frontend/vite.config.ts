import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { realpathSync } from "node:fs";

const rootDir = realpathSync(process.cwd());

export default defineConfig({
  root: rootDir,
  plugins: [react(), tailwindcss()],
  build: {
    chunkSizeWarningLimit: 2500
  }
});
