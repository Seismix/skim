<script lang="ts">
  import { onMount } from "svelte";
  import UrlBar from "./components/UrlBar.svelte";
  import Results from "./components/Results.svelte";
  import { fetchOg } from "./lib/api";
  import { getDomain } from "./lib/util";
  import type { OgData } from "./lib/types";

  let url = $state("");
  let loading = $state(false);
  let error = $state<string | null>(null);
  let metaCount = $state<number | null>(null);
  let domain = $state("");
  let target = $state(""); // the URL being skimmed, for the status line
  let data = $state<OgData | null>(null);
  let lastUrl = $state("");
  let metaEl = $state<HTMLElement | undefined>();
  let flashMeta = $state(false);
  let flashTimer: ReturnType<typeof setTimeout> | undefined;

  // Prefill + auto-run from ?url= (CLI / shared link).
  const preset = new URLSearchParams(location.search).get("url");
  if (preset) url = preset;

  async function run() {
    const u = url.trim();
    if (!u) return;
    loading = true;
    error = null;
    metaCount = null;
    data = null;
    target = u;
    try {
      const d = await fetchOg(u);
      lastUrl = u;
      data = d;
      metaCount = Object.keys(d.allMeta || {}).length;
      domain = getDomain(d.url || u);
    } catch (e) {
      error = (e as Error).message;
    } finally {
      loading = false;
    }
  }

  // Skim from the input: record the URL as a new history entry (so Back/Forward
  // step between skims and the address bar stays shareable), then fetch. Skip the
  // push when re-submitting the URL already showing, to avoid duplicate entries.
  function submit() {
    const u = url.trim();
    if (!u) return;
    const params = new URLSearchParams(location.search);
    if (params.get("url") !== u) {
      params.set("url", u);
      history.pushState({ url: u }, "", `?${params}`);
    }
    run();
  }

  // Back/Forward: restore the input from the URL and re-skim. No history write
  // here — we're moving through existing entries, not creating one.
  function onPopState() {
    const u = new URLSearchParams(location.search).get("url") || "";
    url = u;
    if (u) {
      run();
    } else {
      data = null;
      error = null;
      metaCount = null;
    }
  }

  function jumpToMeta() {
    // Default behavior ("auto") defers to CSS scroll-behavior: smooth, which the
    // prefers-reduced-motion media query downgrades to instant.
    metaEl?.scrollIntoView({ block: "start" });
    // Brief lime flash on the section title so the target is obvious even when
    // the page is already scrolled there and nothing visibly moves.
    flashMeta = true;
    clearTimeout(flashTimer);
    flashTimer = setTimeout(() => (flashMeta = false), 900);
  }

  onMount(() => {
    // The initial entry already carries ?url= (CLI / shared link), so just run —
    // pushing here would duplicate it. popstate then handles Back/Forward.
    window.addEventListener("popstate", onPopState);
    if (preset) run();
    return () => window.removeEventListener("popstate", onPopState);
  });
</script>

<div class="wrap">
  <header class="topbar">
    <span class="wordmark">skim<span class="caret" aria-hidden="true"></span></span>
    <span class="byline">Social preview inspector</span>
  </header>

  <section class="hero">
    <div class="eyebrow">Open Graph · Twitter · LinkedIn · Discord · Slack</div>
    <h1>See how a link looks when it&rsquo;s <em>shared</em> &mdash; before it ships.</h1>
  </section>

  <section class="console">
    <UrlBar bind:url {loading} onsubmit={submit} />
    <div class="status" class:err={!!error} role="status" aria-live="polite">
      {#if loading}
        skimming {target}<span class="tick"></span>
      {:else if error}
        error &mdash; {error}
      {:else if metaCount !== null}
        done ·
        {#if metaCount > 0}
          <button type="button" class="meta-jump" onclick={jumpToMeta}>{metaCount} meta tags</button>
        {:else}
          {metaCount} meta tags
        {/if}
        read from {domain}
      {/if}
    </div>
  </section>
</div>

<!-- Results sit outside .wrap so the wide-screen grid can break past the 1080px
     reading column; the header/hero/console/footer stay centered and narrow. -->
{#if data}
  <Results og={data} inputUrl={lastUrl} bind:metaEl flash={flashMeta} />
{/if}

<div class="wrap">
  <footer class="foot">
    <span>skim — a single-binary local tool. Fetches on your machine, so localhost works.</span>
  </footer>
</div>
