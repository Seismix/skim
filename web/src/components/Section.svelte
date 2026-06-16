<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    name,
    spec,
    wide = false,
    highlight = false,
    el = $bindable(),
    children,
  }: {
    name: string;
    spec: string;
    wide?: boolean;
    highlight?: boolean;
    el?: HTMLElement;
    children: Snippet;
  } = $props();
</script>

<div class="sec" class:sec--wide={wide} bind:this={el}>
  <div class="sec-head">
    <span class="name" class:flash={highlight}>{name}</span>
    <span class="lead"></span>
    <span class="spec">{spec}</span>
  </div>
  {@render children()}
</div>
