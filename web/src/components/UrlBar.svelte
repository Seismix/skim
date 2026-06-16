<script lang="ts">
    let { url = $bindable(""), loading, onsubmit }: { url?: string; loading: boolean; onsubmit: () => void } = $props();

    function onkeydown(e: KeyboardEvent) {
        if (e.key === "Enter") onsubmit();
    }
</script>

<div
    class="flex items-stretch border border-ink bg-paper transition-shadow duration-150 focus-within:shadow-[inset_0_0_0_1px_var(--color-ink)] max-[620px]:flex-wrap"
>
    <span
        class="flex items-center pl-[1.1rem] pr-4 font-mono text-[0.95rem] text-ink-soft border-r border-line select-none max-[620px]:hidden"
        aria-hidden="true">▸</span
    >
    <input
        class="flex-1 min-w-0 border-none outline-none bg-transparent text-ink font-mono text-[1.05rem] px-4 py-[1.15rem] placeholder:text-[#aaa] max-[620px]:basis-full"
        type="text"
        bind:value={url}
        {onkeydown}
        spellcheck="false"
        autocomplete="off"
        placeholder="localhost:3000  ·  example.com"
        aria-label="URL to inspect"
    />
    <button
        class="border-none bg-lime text-ink font-mono text-[0.82rem] font-semibold tracking-[0.14em] uppercase px-8 cursor-pointer whitespace-nowrap transition-colors duration-150 hover:bg-ink hover:text-lime disabled:bg-mist disabled:text-[#999] disabled:cursor-progress focus-visible:[outline-offset:-4px] max-[620px]:flex-1 max-[620px]:p-4 max-[620px]:border-t max-[620px]:border-line"
        type="button"
        onclick={onsubmit}
        disabled={loading}
    >
        {loading ? "Skimming" : "Skim"}
    </button>
</div>
