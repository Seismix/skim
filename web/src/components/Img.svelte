<script lang="ts">
  import { proxied } from "../lib/util";

  // Preview image with skim's placeholder/fallback states. `ratio` sets the
  // aspect-ratio of the placeholder so empty cards keep their shape.
  let { src, ratio = "" }: { src: string; ratio?: string } = $props();

  let failed = $state(false);
  // Reset the failed flag if the source changes between renders.
  $effect(() => {
    src;
    failed = false;
  });
</script>

{#if !src || failed}
  <div
    class="card-img no-img"
    style={ratio ? `aspect-ratio:${ratio};` : "min-height:120px;"}
  >
    {failed ? "image failed to load" : "no og:image"}
  </div>
{:else}
  <img class="card-img" src={proxied(src)} alt="" onerror={() => (failed = true)} />
{/if}
