<script lang="ts">
  import Section from "./Section.svelte";
  import Img from "./Img.svelte";
  import Diagnostics, { type Diag } from "./Diagnostics.svelte";
  import type { OgData } from "../lib/types";
  import { getDomain, isIconImage, proxied } from "../lib/util";

  let {
    og,
    inputUrl,
    metaEl = $bindable(),
    flash = false,
  }: { og: OgData; inputUrl: string; metaEl?: HTMLElement; flash?: boolean } = $props();

  let domain = $derived(getDomain(og.url || inputUrl));
  let isLarge = $derived(og.twitterCard === "summary_large_image");
  let icon = $derived(isIconImage(og));
  // Discord's footer icon: the site favicon, falling back to an icon-sized
  // og:image (e.g. Reddit serves its logo as both favicon and og:image).
  let dcFootIcon = $derived(og.favicon || (icon ? og.image : ""));
  let metaEntries = $derived(Object.entries(og.allMeta || {}));

  let diags = $derived.by(() => {
    const list: Diag[] = [];
    if (!og.title) list.push({ tag: "no title", msg: "Missing <code>og:title</code> — most platforms require it." });
    if (!og.description) list.push({ tag: "no desc", msg: "Missing <code>og:description</code> — recommended everywhere." });
    if (!og.image) list.push({ tag: "no image", msg: "Missing <code>og:image</code> — cards render without a preview image." });
    if (og.image && !og.imageWidth) list.push({ tag: "no dims", msg: "Missing <code>og:image:width/height</code> — Facebook may skip the image on first share." });
    if (!og.twitterCard) list.push({ tag: "no tw:card", msg: "Missing <code>twitter:card</code> — Twitter/X falls back to a basic summary." });
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
  <div class="cap">title <b>{titleLen}</b>ch · desc <b>{descLen}</b>ch · image <b>{imgInfo}</b></div>
{/snippet}

<main class="results">
  {#if diags.length}
    <Section name="Diagnostics" spec={`${diags.length} issue${diags.length > 1 ? "s" : ""}`} wide>
      <Diagnostics {diags} />
    </Section>
  {/if}

  <Section name="Facebook / Open Graph" spec="1.91 : 1">
    <div class="fb">
      <Img src={og.image} ratio="1.91/1" />
      <div class="body">
        <div class="site">{domain}</div>
        <div class="title">{og.title || "Untitled"}</div>
        <div class="desc">{og.description}</div>
      </div>
    </div>
    {@render cap()}
  </Section>

  {#if isLarge}
    <Section name="Twitter / X" spec="summary_large_image · 2 : 1">
      <div class="tw">
        <Img src={og.image} ratio="2/1" />
        <div class="body">
          <div class="title">{og.title || "Untitled"}</div>
          <div class="desc">{og.description}</div>
          <div class="site">{domain}</div>
        </div>
      </div>
      {@render cap()}
    </Section>
  {:else}
    <Section name="Twitter / X" spec="summary · 1 : 1">
      <div class="tws">
        <Img src={og.image} />
        <div class="body">
          <div class="title">{og.title || "Untitled"}</div>
          <div class="desc">{og.description}</div>
          <div class="site">{domain}</div>
        </div>
      </div>
      {@render cap()}
    </Section>
  {/if}

  <Section name="LinkedIn" spec="1.91 : 1">
    <div class="li">
      <Img src={og.image} ratio="1.91/1" />
      <div class="body">
        <div class="title">{og.title || "Untitled"}</div>
        <div class="site">{domain}</div>
      </div>
    </div>
    {@render cap()}
  </Section>

  <Section name="Discord" spec="embed">
    <div class="dc">
      <div class="title">{og.title || "Untitled"}</div>
      {#if og.description}
        <div class="desc">{og.description}</div>
      {/if}
      {#if og.image && !icon}
        <Img src={og.image} />
      {/if}
      {#if og.siteName}
        <div class="dc-foot">
          {#if dcFootIcon}
            <img class="dc-foot-icon" src={proxied(dcFootIcon)} alt="" onerror={hideBroken} />
          {/if}
          <span>{og.siteName}</span>
        </div>
      {/if}
    </div>
    {@render cap()}
  </Section>

  <Section name="Slack" spec="unfurl">
    <div class="sl">
      <div class="sl-head">
        {#if og.favicon}
          <img class="sl-fav" src={proxied(og.favicon)} alt="" onerror={hideBroken} />
        {/if}
        <span class="site">{og.siteName || domain}</span>
      </div>
      <div class="title">{og.title || "Untitled"}</div>
      <div class="desc">{og.description}</div>
      {#if og.image && !icon}
        <Img src={og.image} />
      {/if}
    </div>
    {@render cap()}
  </Section>

  {#if metaEntries.length}
    <Section name="Raw metadata" spec={`${metaEntries.length} tags`} wide bind:el={metaEl} highlight={flash}>
      <table class="meta">
        <tbody>
          <tr><th>Property</th><th>Content</th></tr>
          {#each metaEntries as [k, v]}
            <tr><td class="k">{k}</td><td class="v">{v}</td></tr>
          {/each}
        </tbody>
      </table>
    </Section>
  {/if}
</main>
