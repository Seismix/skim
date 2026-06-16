import { defineConfig, type PluginOption } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import { spawn, execFileSync, type ChildProcess } from "node:child_process";
import { existsSync, mkdirSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";

const BACKEND_PORT = 8799;

// Dev only: build and run the Go backend so the Svelte dev server (with HMR)
// has a working /api. Vite proxies /api/* to it. We run the compiled binary
// (not `go run`, which orphans its child) so killing it on exit triggers the
// binary's graceful shutdown and frees the port.
function goBackend(): PluginOption {
  let proc: ChildProcess | undefined;
  return {
    name: "skim-go-backend",
    apply: "serve",
    configureServer() {
      const web = import.meta.dirname;
      const repoRoot = resolve(web, "..");
      const dist = resolve(web, "dist");

      // go:embed needs web/dist to exist to compile; a stub suffices for dev
      // (the UI is served by Vite, not the Go backend, while developing).
      if (!existsSync(resolve(dist, "index.html"))) {
        mkdirSync(dist, { recursive: true });
        writeFileSync(resolve(dist, "index.html"), "<!-- skim dev stub -->");
      }

      const bin = resolve(repoRoot, "dist", process.platform === "win32" ? "skim-dev.exe" : "skim-dev");
      execFileSync("go", ["build", "-o", bin, "."], { cwd: repoRoot, stdio: "inherit" });
      // --quiet so the backend's own banner doesn't clutter the Vite output;
      // the line below is the single source of truth for the dev URLs.
      proc = spawn(bin, ["--no-open", "--quiet", "--port", String(BACKEND_PORT)], { stdio: "inherit" });

      console.log(`\n  skim backend ready on :${BACKEND_PORT} — proxied at /api. Open the UI at the Vite URL below.\n`);

      const stop = () => { if (proc && !proc.killed) proc.kill(); };
      process.once("exit", stop);
      process.once("SIGINT", () => { stop(); process.exit(0); });
      process.once("SIGTERM", () => { stop(); process.exit(0); });
    },
  };
}

export default defineConfig({
  plugins: [tailwindcss(), svelte(), goBackend()],
  base: "./",
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/api": `http://127.0.0.1:${BACKEND_PORT}`,
    },
  },
});
