<script lang="ts">
    import { onMount } from "svelte";
    import UrlBar from "./components/UrlBar.svelte";
    import Results from "./components/Results.svelte";
    import { fetchOgBatch } from "./lib/api";
    import { getDomain } from "./lib/util";
    import { PLATFORMS, parsePlatformParam, serializePlatformParam, type PlatformId } from "./lib/platforms";
    import type { FetchResult } from "./lib/types";

    let loading = $state(false);
    let error = $state<string | null>(null); // whole-request failure (per-URL errors live on each result)
    let target = $state(""); // what's being skimmed, for the status line
    let results = $state<FetchResult[]>([]);
    let metaEl = $state<HTMLElement | undefined>();
    let flashMeta = $state(false);
    let flashTimer: ReturnType<typeof setTimeout> | undefined;

    // Each input box holds one URL; the "+" in UrlBar adds another box. A single box
    // may still carry a comma-separated list — handy for pasting one — which
    // expandList splits back out. Spaces are NOT separators, so a URL with a stray
    // space stays a single (likely invalid) entry rather than being torn in two.
    function expandList(value: string): string[] {
        return value
            .split(",")
            .map((u) => u.trim())
            .filter(Boolean);
    }

    // The flat list of URLs to skim, gathered from every box (each possibly a list).
    function cleanUrls(): string[] {
        return urls.flatMap(expandList);
    }

    // URLs carried in the address bar: one per repeated ?url= param (lists also split).
    function paramUrls(): string[] {
        return new URLSearchParams(location.search).getAll("url").flatMap(expandList);
    }

    const presetUrls = paramUrls();
    let urls = $state<string[]>(presetUrls.length ? presetUrls : [""]);

    // Which platforms render, driven by the toggle row in the hero. Purely a view
    // filter: the fetched metadata is platform-agnostic, so toggling only shows and
    // hides cards that are already built — never a refetch.
    let selected = $state<PlatformId[]>(parsePlatformParam(location.search));

    function toggle(id: PlatformId) {
        selected = selected.includes(id) ? selected.filter((x) => x !== id) : [...selected, id];
    }

    // Mirror the selection into ?platforms= so a view is shareable. replaceState,
    // not push: toggling a card off is a view preference, and pushing would bury
    // the Back button under an undo trail of clicks instead of stepping between
    // skims. Pushed entries (see submit) still carry whatever is set here.
    $effect(() => {
        const params = new URLSearchParams(location.search);
        const value = serializePlatformParam(selected);
        if (value === null) params.delete("platforms");
        else params.set("platforms", value);
        const qs = params.toString();
        const next = qs ? `?${qs}` : location.pathname;
        const current = location.search || location.pathname;
        if (next !== current) history.replaceState(history.state, "", next);
    });

    // The single-result view keeps the richer status line (meta count + jump). For
    // a batch we summarise counts instead, since there's no one section to jump to.
    let single = $derived(results.length === 1 ? results[0] : null);
    let metaCount = $derived(single?.data ? Object.keys(single.data.allMeta || {}).length : null);
    let domain = $derived(single?.data ? getDomain(single.data.url || single.url) : "");
    let okCount = $derived(results.filter((r) => r.data).length);
    let failCount = $derived(results.length - okCount);

    // Preview theme: render the platform cards in their light or dark variants.
    // skim's own chrome stays light regardless — only the cards opt into dark.
    // Persisted so the choice survives reloads.
    const THEME_KEY = "skim:previewTheme";
    let dark = $state(false);
    try {
        dark = localStorage.getItem(THEME_KEY) === "dark";
    } catch {
        /* storage may be blocked (private mode); default to light */
    }
    $effect(() => {
        try {
            localStorage.setItem(THEME_KEY, dark ? "dark" : "light");
        } catch {
            /* ignore — toggle still works for the session */
        }
    });

    async function run() {
        const clean = cleanUrls();
        if (!clean.length) return;
        loading = true;
        error = null;
        results = [];
        target = clean.length === 1 ? clean[0] : `${clean.length} links`;
        try {
            results = await fetchOgBatch(clean);
        } catch (e) {
            error = (e as Error).message;
        } finally {
            loading = false;
        }
    }

    // Skim from the inputs: record the URLs as a new history entry — one ?url= param
    // each, so Back/Forward step between skims and the address bar stays shareable —
    // then fetch. Skip the push when nothing changed, to avoid duplicate entries.
    function submit() {
        const clean = cleanUrls();
        if (!clean.length) return;
        const params = new URLSearchParams(location.search);
        const current = params.toString();
        params.delete("url");
        for (const u of clean) params.append("url", u);
        if (params.toString() !== current) {
            history.pushState({}, "", `?${params}`);
        }
        run();
    }

    // Back/Forward: restore the inputs from the URL and re-skim. No history write
    // here — we're moving through existing entries, not creating one.
    function onPopState() {
        const next = paramUrls();
        urls = next.length ? next : [""];
        // The entry we landed on carries its own selection; restore it before the
        // re-skim so the cards come back filtered as they were.
        selected = parsePlatformParam(location.search);
        if (next.length) {
            run();
        } else {
            results = [];
            error = null;
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
        if (presetUrls.length) run();
        return () => window.removeEventListener("popstate", onPopState);
    });
</script>

<div class="max-w-[1080px] mx-auto px-8 max-[620px]:px-[1.2rem]">
    <header class="flex items-center justify-between pt-6 pb-[1.4rem] border-b border-ink">
        <span class="text-[3.5rem] font-medium tracking-[-0.03em] skim-mark">skim</span>
        <div class="flex flex-col items-end gap-[0.55rem]">
            <span class="font-mono text-[0.72rem] tracking-[0.04em] uppercase text-ink-soft"
                >Social preview inspector</span
            >
            <button
                type="button"
                role="switch"
                aria-checked={dark}
                aria-label="Toggle preview dark mode"
                class="group font-mono text-[0.72rem] tracking-[0.04em] uppercase text-ink-soft bg-transparent border-none p-0 cursor-pointer inline-flex items-baseline gap-[0.5em]"
                onclick={() => (dark = !dark)}
            >
                previews
                <span
                    class="text-ink px-[0.25em] border-b border-lime transition-colors duration-[120ms] group-hover:bg-lime"
                    >{dark ? "dark" : "light"}</span
                >
            </button>
        </div>
    </header>

    <section class="pt-[2.75rem] pb-[1.75rem] max-w-[600px] mx-auto text-center max-[620px]:pt-12 max-[620px]:pb-8">
        <!-- The platform list doubles as the filter: a lit name carries the lime
             skim-mark (the wordmark's underline), clicking drops both the mark and
             the card. Rendered from PLATFORMS, so adding a platform lights up here
             for free rather than needing this line kept in step by hand.
             Tracking is tighter than this eyebrow line once wore: seven names at
             0.22em spilled iMessage onto a line of its own. It still wraps when the
             column gets narrow — the row just no longer starts out wrapped. -->
        <div
            role="group"
            aria-label="Platforms to preview"
            class="flex flex-wrap justify-center items-baseline gap-x-[0.2em] gap-y-[0.3em] font-mono text-[0.72rem] tracking-[0.13em] uppercase text-ink-soft mb-[0.85rem]"
        >
            {#each PLATFORMS as p, i (p.id)}
                <!-- The two hover states preview each other: a lit name floods lime
                     (the header toggle's idiom) to read as "click to drop this", and
                     an unlit one picks the mark back up to show what returns. Both
                     keep skim-mark's inline-block/1.45 metrics, which already match
                     the button's defaults, so lighting up shifts nothing. -->
                <button
                    type="button"
                    aria-pressed={selected.includes(p.id)}
                    class="bg-transparent border-none px-[0.25em] py-0 font-mono text-[0.72rem] tracking-[0.13em] uppercase cursor-pointer transition-colors duration-[120ms] {selected.includes(
                        p.id,
                    )
                        ? 'text-ink skim-mark hover:bg-lime'
                        : 'text-ink-soft hover:text-ink hover:skim-mark'}"
                    onclick={() => toggle(p.id)}>{p.label}</button
                >
                {#if i < PLATFORMS.length - 1}
                    <span aria-hidden="true">·</span>
                {/if}
            {/each}
        </div>
        <h1 class="text-[1.12rem] font-normal tracking-normal leading-normal text-ink-soft">
            See how a link looks when it&rsquo;s
            <em class="not-italic text-ink skim-mark">shared</em> &mdash; before it ships.
        </h1>
    </section>

    <section class="max-w-[524px] mx-auto pb-4">
        <UrlBar bind:urls {loading} onsubmit={submit} />
        <div
            class="font-mono text-[0.78rem] mt-[0.85rem] min-h-[1.1em] tracking-[0.01em] {error || single?.error
                ? 'text-warn'
                : 'text-ink-soft'}"
            role="status"
            aria-live="polite"
        >
            {#if loading}
                skimming {target}<span
                    class="inline-block w-[0.5em] h-[1em] bg-lime translate-y-[0.15em] ml-[0.15em] animate-tick"
                ></span>
            {:else if error}
                error &mdash; {error}
            {:else if single}
                {#if single.error}
                    error &mdash; {single.error}
                {:else}
                    done ·
                    {#if metaCount && metaCount > 0}
                        <button
                            type="button"
                            class="text-ink bg-transparent border-none px-[0.05em] cursor-pointer border-b border-lime transition-colors duration-[120ms] hover:bg-lime"
                            onclick={jumpToMeta}>{metaCount} meta tags</button
                        >
                    {:else}
                        {metaCount} meta tags
                    {/if}
                    read from {domain}
                {/if}
            {:else if results.length}
                done · skimmed {results.length} links · {okCount} ok{#if failCount}
                    · {failCount} failed{/if}
            {/if}
        </div>
    </section>
</div>

<!-- A labelled divider above each card set in a batch, so it's clear which URL the
     cards below belong to. Widths track the Results <main> so the bar lines up with
     the grid at every breakpoint. -->
{#snippet headerBar(r: FetchResult, i: number)}
    <div
        class="max-w-[calc(524px+4rem)] mx-auto px-8 pt-[3.25rem] max-[620px]:px-[1.2rem] lg:max-w-[calc(1096px+4rem)] ultra:max-w-[calc(1668px+4rem)]"
    >
        <div class="flex items-baseline gap-[0.85rem] pb-[0.55rem] border-b border-ink font-mono">
            <span class="text-[0.72rem] tracking-[0.1em] uppercase text-ink-soft shrink-0"
                >{i + 1} / {results.length}</span
            >
            <span class="text-[0.9rem] text-ink truncate">{r.url}</span>
            {#if r.error}
                <span class="text-warn text-[0.72rem] tracking-[0.1em] uppercase shrink-0 ml-auto">failed</span>
            {/if}
        </div>
    </div>
{/snippet}

{#snippet failNotice(r: FetchResult)}
    <div
        class="max-w-[calc(524px+4rem)] mx-auto px-8 pt-4 pb-2 max-[620px]:px-[1.2rem] lg:max-w-[calc(1096px+4rem)] ultra:max-w-[calc(1668px+4rem)]"
    >
        <div class="border border-line px-4 py-3 font-mono text-[0.8rem] text-warn [word-break:break-word]">
            couldn&rsquo;t skim {r.url} &mdash; {r.error}
        </div>
    </div>
{/snippet}

<!-- Results sit outside .wrap so the wide-screen grid can break past the 1080px
     reading column; the header/hero/console/footer stay centered and narrow. A lone
     URL renders exactly as before (with the meta-jump wiring); a batch stacks one
     labelled card set per URL. -->
{#if single}
    {#if single.data}
        <Results og={single.data} inputUrl={single.url} bind:metaEl flash={flashMeta} {dark} {selected} />
    {/if}
{:else if results.length}
    {#each results as r, i (r.url + "#" + i)}
        {@render headerBar(r, i)}
        {#if r.data}
            <Results og={r.data} inputUrl={r.url} {dark} {selected} />
        {:else}
            {@render failNotice(r)}
        {/if}
    {/each}
{/if}

<div class="max-w-[1080px] mx-auto px-8 max-[620px]:px-[1.2rem]">
    <footer
        class="border-t border-line pt-6 pb-10 font-mono text-[0.72rem] tracking-[0.03em] text-ink-soft flex justify-between flex-wrap gap-2"
    >
        <span>skim — a single-binary local tool. Fetches on your machine, so localhost works.</span>
    </footer>
</div>
