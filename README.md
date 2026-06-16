# skim

Preview how a link looks when shared on social media — **including `localhost`** — without deploying.

skim extracts OpenGraph / Twitter / standard meta tags from any URL and renders realistic preview cards for **Facebook, Twitter/X, LinkedIn, Discord, and Slack**, plus validation diagnostics and a raw meta-tag grid. It is a rebuild of an Electron app as a **single, dependency-free ~7 MB binary** that opens in your browser — no Chromium, no install.

## Why not Electron?

The original app wrapped a ~600-line web UI in a ~200 MB Electron shell. skim keeps the UI verbatim and replaces the shell with a tiny local HTTP server compiled to one native file. Drop the file in Slack; a colleague runs it and it works. No runtime to install.

## Quick start

Grab the binary for your OS from `dist/` (or build it — see below) and run it:

```bash
./skim
```

It starts a local server on a free port and opens your default browser. To pre-fill and auto-preview a URL:

```bash
./skim http://localhost:3000
```

You can type a bare host — `reddit.com` becomes `https://reddit.com`, while `localhost:3000` (and `127.0.0.1`, `*.localhost`) becomes `http://…`.

Flags: `--port N` (fixed port), `--no-open` (don't launch the browser), `--user-agent "…"`. skim binds loopback (`127.0.0.1`) only — it's a personal local tool. Press Ctrl+C to stop — the listener closes and the port frees.

By default skim sends the OpenGraph crawler User-Agent (`facebookexternalhit`), because many sites serve their share metadata **only** to recognized crawlers (Reddit, for example, returns an anti-bot landing page to anything else). This makes skim see what Facebook / Discord / Slack actually see. Override it with `--user-agent`.

## Why a local server at all?

The localhost feature is the whole point, and it dictates the architecture. A browser page can't `fetch()` an arbitrary origin's raw HTML (CORS + mixed content), so skim does the fetch **on your machine** and hands the parsed result to the UI. That fetch resolving to *your* `localhost` is exactly why skim must run locally rather than as a hosted website.

```text
skim binary
├── GET  /             → serves the embedded Svelte UI (web/dist)
├── POST /api/fetch-og → fetches the URL on THIS machine, parses meta, returns JSON
└── GET  /api/img      → proxies a preview image same-origin, so browser tracking
                          protection / hotlink checks don't block it
```

## Building

Building needs **Go** and **Node/npm** (Node only builds the UI; the shipped binary has no runtime deps).

```bash
make           # build the Svelte UI, then the native binary → dist/skim
make dist      # build the UI, then cross-compile static binaries for win/mac/linux → dist/
make web       # build just the Svelte UI → web/dist
make run ARGS="http://localhost:3000"   # build, then run the binary serving the built UI
```

The Svelte UI is compiled by Vite to static assets and embedded into the binary via `go:embed`, so distribution is still a single file. Cross-compilation needs no extra setup — Go's `GOOS`/`GOARCH` plus `CGO_ENABLED=0` produce fully static binaries.

## Developing the UI

For live frontend work with hot-reload:

```bash
make dev        # or: cd web && npm run dev
```

This runs the Vite dev server (HMR for `web/src/**`) and **automatically builds and runs the Go backend**, proxying `/api` to it — so the UI is fully functional while you edit. Vite prints the local URL (e.g. `http://localhost:5173`). Ctrl+C stops both; the backend frees its port on shutdown. Only the UI hot-reloads — change `main.go` and the dev server rebuilds the backend on restart.

## Tech stack

| Layer | Technology |
| ------ | ------ |
| Server + CLI | Go (`net/http`, `embed`) |
| HTML parsing | [goquery](https://github.com/PuerkitoBio/goquery) |
| UI | [Svelte 5](https://svelte.dev/) + [Vite](https://vite.dev/), built to static assets and embedded; Milk Interactive branding |

## Project layout

```text
skim/
├── main.go           # server + embed + browser launch + graceful shutdown
├── go.mod / go.sum
├── web/              # Svelte UI (Vite) — built to web/dist, embedded into the binary
│   ├── src/          #   App.svelte, components/, lib/, app.css
│   └── package.json
├── Makefile
└── dist/             # build output (gitignored)
```
