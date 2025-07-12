<script lang="ts">
    import type { Widget as WidgetType } from "../types";
    import Widget from "../Widget.svelte";

    let { widget }: { widget: WidgetType } = $props();
    
    let children = $derived(widget.Children || []);
    let hasChild = $derived(children.length > 0);
    let growValue = $derived(widget.Data?.grow_value || "1");
    let growStyle = $derived(`flex-grow: ${growValue};`);
</script>

{#if hasChild}
    <div class="h-full" style={growStyle}>
        {#each children as child (child.ID)}
            <Widget widget={child} />
        {/each}
    </div>
{:else}
    <div class="h-full" style={growStyle}></div>
{/if}
