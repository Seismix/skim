<script lang="ts">
    export interface Diag {
        tag: string;
        msg: string; // trusted, author-written HTML (contains <code>)
    }
    let { diags }: { diags: Diag[] } = $props();
</script>

<div class="border-t border-ink">
    {#each diags as d}
        <div class="flex gap-4 py-[0.85rem] border-b border-line text-[0.92rem]">
            <span
                class="font-mono text-[0.7rem] tracking-[0.1em] text-warn whitespace-nowrap pt-[0.18em] min-w-[5.5rem]"
                >{d.tag}</span
            >
            <span
                class="text-ink [&_code]:font-mono [&_code]:text-[0.86em] [&_code]:bg-mist [&_code]:px-[0.3em] [&_code]:py-[0.05em]"
                >{@html d.msg}</span
            >
        </div>
    {/each}
</div>
