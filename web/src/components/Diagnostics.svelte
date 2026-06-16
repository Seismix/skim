<script lang="ts">
  export interface Diag {
    tag: string;
    msg: string; // trusted, author-written HTML (contains <code>)
  }
  let { diags }: { diags: Diag[] } = $props();
</script>

<div class="diag">
  {#each diags as d}
    <div class="diag-row">
      <span class="tag">{d.tag}</span>
      <span class="msg">{@html d.msg}</span>
    </div>
  {/each}
</div>
