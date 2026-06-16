<script lang="ts">
  let {
    url = $bindable(""),
    loading,
    onsubmit,
  }: { url?: string; loading: boolean; onsubmit: () => void } = $props();

  function onkeydown(e: KeyboardEvent) {
    if (e.key === "Enter") onsubmit();
  }
</script>

<div class="field">
  <span class="sigil" aria-hidden="true">▸</span>
  <input
    type="text"
    bind:value={url}
    {onkeydown}
    spellcheck="false"
    autocomplete="off"
    placeholder="localhost:3000  ·  example.com"
    aria-label="URL to inspect"
  />
  <button type="button" onclick={onsubmit} disabled={loading}>
    {loading ? "Skimming" : "Skim"}
  </button>
</div>
