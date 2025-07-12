<script lang="ts">
    import type { WidgetData } from "../types";
    import Widget from "../Widget.svelte";

    let { widget }: { widget: WidgetData } = $props();
    
    let children = $derived(widget.children || []);
    let hasChild = $derived(children.length > 0);
    let growValue = $derived(widget.data?.grow_value || "1");
    let growStyle = $derived(`flex-grow: ${growValue};`);
</script>

{#if hasChild}
    <div class="h-full" style={growStyle}>
        {#each children as child (child.id)}
            <Widget widget={child} />
        {/each}
    </div>
{:else}
    <div class="h-full" style={growStyle}></div>
{/if}
