<script lang="ts">
    import Section from "./Section.svelte";
    import Img from "./Img.svelte";
    import Diagnostics, { type Diag } from "./Diagnostics.svelte";
    import type { OgData } from "../lib/types";
    import { getDomain, proxied } from "../lib/util";
    import { PLATFORMS, type PlatformId } from "../lib/platforms";

    let {
        og,
        inputUrl,
        metaEl = $bindable(),
        flash = false,
        dark = false,
    }: {
        og: OgData;
        inputUrl: string;
        metaEl?: HTMLElement;
        flash?: boolean;
        dark?: boolean;
    } = $props();

    let domain = $derived(getDomain(og.url || inputUrl));
    let isLarge = $derived(og.twitterCard === "summary_large_image");
    // The image classification comes from the server (imageKind in main.go) —
    // templates only pick layouts from the verdict, so the icon-vs-banner
    // heuristic has exactly one implementation.
    let icon = $derived(og.imageKind === "icon");
    let banner = $derived(og.imageKind === "banner");
    // Likewise each card's spec label reads from the API's platforms block; the
    // shape strings and their rules live only in main.go's renderPlatforms. A
    // plain function, not $derived — the template tracks the og.platforms read
    // at each call site.
    function shape(id: PlatformId): string {
        return og.platforms?.[id]?.shape ?? "";
    }
    // Discord's footer icon: the site favicon, falling back to an icon-sized
    // og:image (e.g. Reddit serves its logo as both favicon and og:image).
    let dcFootIcon = $derived(og.favicon || (icon ? og.image : ""));
    let metaEntries = $derived(Object.entries(og.allMeta || {}));

    let diags = $derived.by(() => {
        const list: Diag[] = [];
        if (!og.title)
            list.push({ tag: "no title", msg: "Missing <code>og:title</code> — most platforms require it." });
        if (!og.description)
            list.push({ tag: "no desc", msg: "Missing <code>og:description</code> — recommended everywhere." });
        if (!og.image)
            list.push({
                tag: "no image",
                msg: "Missing <code>og:image</code> — cards render without a preview image.",
            });
        if (og.image && !og.imageWidth)
            list.push({
                tag: "no dims",
                msg: "Missing <code>og:image:width/height</code> — Facebook may skip the image on first share.",
            });
        if (!og.twitterCard)
            list.push({
                tag: "no tw:card",
                msg: "Missing <code>twitter:card</code> — Twitter/X falls back to a basic summary.",
            });
        return list;
    });

    let titleLen = $derived((og.title || "").length);
    let descLen = $derived((og.description || "").length);
    let imgInfo = $derived(
        !og.image ? "none" : og.imageWidth && og.imageHeight ? `${og.imageWidth}×${og.imageHeight}` : "present",
    );

    function hideBroken(e: Event) {
        (e.currentTarget as HTMLElement).style.display = "none";
    }
</script>

{#snippet cap()}
    <div class="font-mono text-[0.72rem] text-ink-soft tracking-[0.03em] mt-[0.7rem] [&_b]:text-ink [&_b]:font-medium">
        title <b>{titleLen}</b>ch · desc <b>{descLen}</b>ch · image <b>{imgInfo}</b>
    </div>
{/snippet}

<!-- WhatsApp's unfurl text, identical in all three shapes — only its box moves
     (under the banner, or beside the thumb). -->
{#snippet waText()}
    <div class="text-[0.88rem] font-medium text-[#111b21] dark:text-[#e9edef] leading-[1.3] line-clamp-2">
        {og.title || "Untitled"}
    </div>
    {#if og.description}
        <div class="text-[0.8rem] text-[#667781] dark:text-[#8696a0] leading-[1.35] line-clamp-2 mt-[0.1rem]">
            {og.description}
        </div>
    {/if}
    <div class="font-mono text-[0.72rem] text-[#667781] dark:text-[#8696a0] mt-[0.15rem] truncate">{domain}</div>
{/snippet}

<!-- iMessage's caption, shared by the banner and thumb shapes. og:description is
     absent by design: the rich link renders title + domain only, whatever the
     page offers. #8e8e93 is Apple's secondary label, unchanged across themes. -->
{#snippet imText()}
    <div class="text-[0.88rem] font-semibold text-[#000000] dark:text-[#ffffff] leading-[1.3] line-clamp-2">
        {og.title || "Untitled"}
    </div>
    <div class="font-mono text-[0.72rem] text-[#8e8e93] mt-[0.1rem] truncate">{domain}</div>
{/snippet}

<!-- One snippet per platform card, holding just the card's markup — the shared
     Section scaffolding and the cap line live in the {#each} that renders them.
     Adding a platform = a registry entry in platforms.ts, a snippet named after
     its id here, and the matching case in main.go's renderPlatforms. -->

{#snippet facebook()}
    <div class="max-w-[524px] border border-line dark:border-[#393a3b] bg-paper dark:bg-[#242526]">
        <Img src={og.image} ratio="1.91/1" class="w-full aspect-[1.91/1] object-cover" />
        <div class="py-[0.7rem] px-4 bg-mist dark:bg-[#3a3b3c]">
            <div class="font-mono text-[0.7rem] text-[#65676b] dark:text-[#b0b3b8] uppercase tracking-[0.02em]">
                {domain}
            </div>
            <div
                class="text-base font-semibold text-[#1c1e21] dark:text-[#e4e6eb] my-[0.18rem] leading-[1.3] line-clamp-2"
            >
                {og.title || "Untitled"}
            </div>
            <div class="text-[0.86rem] text-[#65676b] dark:text-[#b0b3b8] line-clamp-1">{og.description}</div>
        </div>
    </div>
{/snippet}

{#snippet twitter()}
    {#if isLarge}
        <div
            class="max-w-[524px] border border-line dark:border-[#38444d] rounded-2xl overflow-hidden bg-paper dark:bg-[#15202b]"
        >
            <Img src={og.image} ratio="2/1" class="w-full aspect-[2/1] object-cover" />
            <div class="py-[0.7rem] px-[0.9rem]">
                <div class="text-[0.95rem] font-semibold text-[#0f1419] dark:text-[#e7e9ea] line-clamp-2">
                    {og.title || "Untitled"}
                </div>
                <div class="text-[0.9rem] text-[#536471] dark:text-[#8b98a5] line-clamp-2">
                    {og.description}
                </div>
                <div class="font-mono text-[0.8rem] text-[#536471] dark:text-[#8b98a5] mt-[0.15rem]">
                    {domain}
                </div>
            </div>
        </div>
    {:else}
        <div
            class="max-w-[524px] flex border border-line dark:border-[#38444d] rounded-2xl overflow-hidden bg-paper dark:bg-[#15202b]"
        >
            <Img src={og.image} class="w-[130px] min-h-[130px] object-cover shrink-0" />
            <div class="py-[0.7rem] px-[0.9rem] flex flex-col justify-center overflow-hidden">
                <div class="text-[0.92rem] font-semibold text-[#0f1419] dark:text-[#e7e9ea] line-clamp-2">
                    {og.title || "Untitled"}
                </div>
                <div class="text-[0.86rem] text-[#536471] dark:text-[#8b98a5] line-clamp-2">
                    {og.description}
                </div>
                <div class="font-mono text-[0.78rem] text-[#536471] dark:text-[#8b98a5]">{domain}</div>
            </div>
        </div>
    {/if}
{/snippet}

{#snippet linkedin()}
    <div class="max-w-[524px] border border-line dark:border-[#ffffff26] bg-paper dark:bg-[#1b1f23]">
        <Img src={og.image} ratio="1.91/1" class="w-full aspect-[1.91/1] object-cover" />
        <div class="py-[0.65rem] px-[0.95rem] bg-paper dark:bg-[#1b1f23] border-t border-line dark:border-[#ffffff26]">
            <div class="text-[0.92rem] font-semibold text-[#000000e6] dark:text-[#ffffffe6] truncate">
                {og.title || "Untitled"}
            </div>
            <div class="font-mono text-[0.72rem] text-[#00000099] dark:text-[#ffffff99] mt-[0.1rem]">
                {domain}
            </div>
        </div>
    </div>
{/snippet}

{#snippet discord()}
    <!-- Discord ships Light/Dark/Midnight themes; base classes are the light
         embed, dark: holds the dark embed. The toggle flips both. -->
    <div
        class="max-w-[432px] bg-[#f2f3f5] dark:bg-[#2b2d31] border-l-4 border-[#5865f2] rounded-[4px] py-[0.7rem] px-4"
    >
        <div class="text-base font-semibold text-[#006ce7] dark:text-[#00a8fc] mb-[0.3rem] leading-[1.3]">
            {og.title || "Untitled"}
        </div>
        {#if og.description}
            <!-- Discord shows several lines and keeps source paragraph breaks. -->
            <div
                class="text-[0.86rem] text-[#4e5058] dark:text-[#dbdee1] leading-[1.4] whitespace-pre-wrap line-clamp-[7]"
            >
                {og.description}
            </div>
        {/if}
        {#if banner}
            <Img
                src={og.image}
                class="max-w-[min(400px,100%)] max-h-[225px] w-auto rounded-[4px] bg-[#e3e5e8]! dark:bg-[#1e1f22]! mt-[0.6rem]"
            />
        {/if}
        {#if og.siteName}
            <!-- footer: site favicon + og:site_name, as Discord renders it -->
            <div class="flex items-center gap-2 mt-[0.65rem] text-[0.78rem] text-[#4e5058] dark:text-[#dbdee1]">
                {#if dcFootIcon}
                    <img
                        class="w-5 h-5 rounded-full object-cover shrink-0"
                        src={proxied(dcFootIcon)}
                        alt=""
                        onerror={hideBroken}
                    />
                {/if}
                <span>{og.siteName}</span>
            </div>
        {/if}
    </div>
{/snippet}

{#snippet slack()}
    <!-- dark:border-l re-asserts the accent: the dark all-side shorthand sorts
         after the base left longhand and would otherwise recolor the bar. -->
    <div
        class="max-w-[524px] border border-line dark:border-[#ffffff1a] border-l-4 border-l-[#1d9bd1] dark:border-l-[#1d9bd1] py-[0.7rem] px-4 bg-paper dark:bg-[#1a1d21]"
    >
        <div class="flex items-center gap-[0.4rem] mb-[0.22rem]">
            {#if og.favicon}
                <img
                    class="w-4 h-4 rounded-[3px] object-cover shrink-0"
                    src={proxied(og.favicon)}
                    alt=""
                    onerror={hideBroken}
                />
            {/if}
            <span class="text-[0.88rem] font-bold text-[#1d1c1d] dark:text-[#d1d2d3]">{og.siteName || domain}</span>
        </div>
        <div class="text-[0.92rem] font-semibold text-[#1264a3] dark:text-[#1d9bd1] mb-[0.2rem]">
            {og.title || "Untitled"}
        </div>
        <div class="text-[0.88rem] text-[#454245] dark:text-[#ababad] mb-2 leading-[1.4]">{og.description}</div>
        {#if banner}
            <Img src={og.image} class="w-auto max-w-full max-h-[280px] rounded-[8px]" />
        {/if}
    </div>
{/snippet}

{#snippet whatsapp()}
    <!-- WhatsApp unfurls *inside* the outgoing chat bubble (green on both
         themes), keeping the sent link as text underneath. The preview block
         is a translucent black wash rather than a fixed color, so it darkens
         whichever green sits behind it. -->
    <div class="max-w-[524px] bg-[#d9fdd3] dark:bg-[#005c4b] rounded-[7.5px] p-[3px] shadow-[0_1px_0.5px_#0b141a21]">
        <div class="bg-black/[0.05] dark:bg-black/[0.13] rounded-[6px] overflow-hidden">
            {#if banner}
                <Img src={og.image} ratio="1.91/1" class="w-full aspect-[1.91/1] object-cover" />
                <div class="py-[0.5rem] px-[0.6rem]">{@render waText()}</div>
            {:else}
                <!-- An icon-sized image demotes the banner to a square thumb;
                     with no image at all the row is text only. -->
                <div class="flex items-center gap-[0.6rem] py-[0.5rem] px-[0.6rem]">
                    <div class="flex-1 min-w-0">{@render waText()}</div>
                    {#if icon}
                        <Img src={og.image} class="w-[76px] h-[76px] rounded-[4px] object-cover shrink-0" />
                    {/if}
                </div>
            {/if}
        </div>
        <div
            class="py-[0.3rem] px-[0.35rem] text-[0.9rem] leading-[1.35] text-[#027eb5] dark:text-[#53bdeb] line-clamp-2 [word-break:break-all]"
        >
            {og.url || inputUrl}
        </div>
    </div>
{/snippet}

{#snippet imessage()}
    <!-- Where WhatsApp unfurls inside the bubble and keeps the sent link as
         text, iMessage *replaces* the link: the rich link becomes the bubble,
         so there's no URL underneath. A bare link unfurls into the gray
         bubble, not the blue one. A real iMessage bubble is far narrower, but
         the card holds the 524px column the other platforms use, so the set
         stays comparable side by side; the 18px corner is kept literal. -->
    <div class="max-w-[524px] rounded-[18px] overflow-hidden bg-[#f1f1f1] dark:bg-[#2c2c2e]">
        {#if banner}
            <Img src={og.image} ratio="1.91/1" class="w-full aspect-[1.91/1] object-cover" />
            <div class="py-[0.5rem] px-[0.75rem]">{@render imText()}</div>
        {:else}
            <!-- The thumb leads the row here — the mirror of WhatsApp, which
                 trails it. With no image the row is text only. -->
            <div class="flex items-center gap-[0.6rem] py-[0.5rem] px-[0.75rem]">
                {#if icon}
                    <Img src={og.image} class="w-[76px] h-[76px] rounded-[8px] object-cover shrink-0" />
                {/if}
                <div class="flex-1 min-w-0">{@render imText()}</div>
            </div>
        {/if}
    </div>
{/snippet}

<!-- .results lives outside .wrap (see App.svelte) so the grid can break past the
     1080px reading column; it carries its own 2rem gutter. From lg the platform
     cards flow into a 2-col (then 3-col at `ultra`) grid to cut scrolling, with
     each column held at the card's true 524px instead of being crushed narrow. -->
<main
    class="max-w-[calc(524px+4rem)] mx-auto pt-4 px-8 pb-24 max-[620px]:px-[1.2rem] lg:max-w-[calc(1096px+4rem)] lg:grid lg:grid-cols-[repeat(2,minmax(0,524px))] lg:justify-center lg:gap-x-12 lg:gap-y-[2.75rem] lg:items-start ultra:max-w-[calc(1668px+4rem)] ultra:grid-cols-[repeat(3,minmax(0,524px))] {dark
        ? 'theme-dark'
        : ''}"
>
    {#if diags.length}
        <Section name="Diagnostics" spec={`${diags.length} issue${diags.length > 1 ? "s" : ""}`} wide>
            <Diagnostics {diags} />
        </Section>
    {/if}

    <!-- The card set: one Section per platform, driven by the registry, so a
         card can't be forgotten when a platform is added — the body snippet
         is looked up by the platform's id. -->
    {#each PLATFORMS as p (p.id)}
        {@const body = { facebook, twitter, linkedin, discord, slack, whatsapp, imessage }[p.id]}
        <Section name={p.card} spec={shape(p.id)}>
            {@render body()}
            {@render cap()}
        </Section>
    {/each}

    {#if metaEntries.length}
        <Section name="Raw metadata" spec={`${metaEntries.length} tags`} wide bind:el={metaEl} highlight={flash}>
            <table class="w-full border-collapse font-mono text-[0.82rem] [&_tr:hover_td]:bg-[#fbfdf8]">
                <tbody>
                    <tr>
                        <th
                            class="text-left font-semibold py-2 px-[0.8rem] border-b border-ink uppercase tracking-[0.08em] text-[0.7rem]"
                            >Property</th
                        >
                        <th
                            class="text-left font-semibold py-2 px-[0.8rem] border-b border-ink uppercase tracking-[0.08em] text-[0.7rem]"
                            >Content</th
                        >
                    </tr>
                    <!-- Keyed by property name (unique in the map), so a re-skim
                         that returns different tags moves rows instead of
                         rewriting them index-by-index. -->
                    {#each metaEntries as [k, v] (k)}
                        <tr>
                            <td
                                class="py-2 px-[0.8rem] border-b border-line align-top text-ink-soft whitespace-nowrap w-60 max-[620px]:w-auto"
                                >{k}</td
                            >
                            <td class="py-2 px-[0.8rem] border-b border-line align-top text-ink [word-break:break-word]"
                                >{v}</td
                            >
                        </tr>
                    {/each}
                </tbody>
            </table>
        </Section>
    {/if}
</main>
