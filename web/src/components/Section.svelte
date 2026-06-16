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

<!-- mt/mb give vertical rhythm on narrow screens; the wide-screen grid zeroes
     them (lg:my-0) and spaces rows with row-gap instead. wide → full-bleed row. -->
<div class="mt-[3.75rem] mb-[1.4rem] lg:my-0 {wide ? 'lg:col-span-full' : ''}" bind:this={el}>
    <div class="flex items-baseline gap-[0.85rem] font-mono text-[0.74rem] tracking-[0.16em] uppercase">
        <!-- brief lime flash matching the jump-link hover, so the target is clear even when no scroll happens -->
        <span
            class="text-ink font-semibold transition-[background-color,box-shadow] duration-500 ease-in-out {highlight
                ? 'bg-lime shadow-[0_0_0_3px_var(--color-lime)] transition-none'
                : ''}">{name}</span
        >
        <span class="flex-1 border-b border-dotted border-[#c4c4c4] -translate-y-[0.28em] min-w-6"></span>
        <span class="text-ink-soft">{spec}</span>
    </div>
    {@render children()}
</div>
