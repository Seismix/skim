<script lang="ts">
  import { proxied } from "../lib/util";

  // Preview image with skim's placeholder/fallback states. `ratio` sets the
  // aspect-ratio of the placeholder so empty cards keep their shape; `class`
  // carries each platform's sizing (full-bleed, fixed thumb, capped, …).
  let { src, ratio = "", class: klass = "" }: { src: string; ratio?: string; class?: string } = $props();

  let failed = $state(false);
  // Reset the failed flag if the source changes between renders.
  $effect(() => {
    src;
    failed = false;
  });
</script>

{#if !src || failed}
  <div
    class="flex items-center justify-center text-[#b0b0b0] font-mono text-[0.78rem] tracking-[0.04em] bg-[repeating-linear-gradient(135deg,#fafafa_0_10px,#f2f2f2_10px_20px)] {klass}"
    style={ratio ? `aspect-ratio:${ratio};` : "min-height:120px;"}
  >
    {failed ? "image failed to load" : "no og:image"}
  </div>
{:else}
  <img class="block bg-mist {klass}" src={proxied(src)} alt="" onerror={() => (failed = true)} />
{/if}
