<script lang="ts">
    import { tick } from "svelte";

    let {
        urls = $bindable([""]),
        loading,
        onsubmit,
    }: { urls?: string[]; loading: boolean; onsubmit: () => void } = $props();

    // Element refs per row, so a freshly added box can take focus immediately.
    let inputs = $state<HTMLInputElement[]>([]);

    function onkeydown(e: KeyboardEvent) {
        if (e.key === "Enter") onsubmit();
    }

    async function add() {
        urls.push("");
        await tick();
        inputs[urls.length - 1]?.focus();
    }

    function remove(i: number) {
        urls.splice(i, 1);
    }
</script>

<div class="flex flex-col gap-[0.6rem]">
    {#each urls as _, i (i)}
        <div
            class="flex items-stretch border border-ink bg-paper transition-shadow duration-150 focus-within:shadow-[inset_0_0_0_1px_var(--color-ink)]"
        >
            <span
                class="flex items-center pl-[1.1rem] pr-4 font-mono text-[0.95rem] text-ink-soft border-r border-line select-none max-[620px]:hidden"
                aria-hidden="true">▸</span
            >
            <input
                bind:this={inputs[i]}
                class="flex-1 min-w-0 border-none outline-none bg-transparent text-ink font-mono text-[1.05rem] px-4 py-[1.15rem] placeholder:text-[#aaa]"
                type="text"
                bind:value={urls[i]}
                {onkeydown}
                spellcheck="false"
                autocomplete="off"
                placeholder="localhost:3000, example.com, …"
                aria-label={`URL ${i + 1} to inspect`}
            />
            {#if urls.length > 1}
                <button
                    type="button"
                    class="border-none border-l border-line bg-transparent text-ink-soft font-mono text-[1.2rem] leading-none px-[1.05rem] cursor-pointer transition-colors duration-150 hover:bg-mist hover:text-warn"
                    onclick={() => remove(i)}
                    aria-label={`Remove URL ${i + 1}`}
                    title="Remove">×</button
                >
            {/if}
        </div>
    {/each}

    <div class="flex items-stretch gap-[0.6rem]">
        <button
            type="button"
            class="border border-ink bg-paper text-ink font-mono text-[0.82rem] font-semibold tracking-[0.14em] uppercase px-6 cursor-pointer whitespace-nowrap transition-colors duration-150 hover:bg-ink hover:text-lime"
            onclick={add}
            title="Add another URL">+ url</button
        >
        <span class="flex-1"></span>
        <button
            class="border-none bg-lime text-ink font-mono text-[0.82rem] font-semibold tracking-[0.14em] uppercase px-8 py-[0.95rem] cursor-pointer whitespace-nowrap transition-colors duration-150 hover:bg-ink hover:text-lime disabled:bg-mist disabled:text-[#999] disabled:cursor-progress focus-visible:[outline-offset:-4px]"
            type="button"
            onclick={onsubmit}
            disabled={loading}
        >
            {loading ? "Skimming" : "Skim"}
        </button>
    </div>
</div>
