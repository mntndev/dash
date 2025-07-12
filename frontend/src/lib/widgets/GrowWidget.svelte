<script lang="ts">
    import type { WidgetData } from "../types";
    import Widget from "../Widget.svelte";

    let { widget }: { widget: WidgetData } = $props();
    
    let children = $derived(widget.children || []);
    let hasChild = $derived(children.length > 0);
</script>

{#if hasChild}
    <div class="flex-1">
        {#each children as child (child.id)}
            <Widget widget={child} />
        {/each}
    </div>
{:else}
    <div class="flex-1"></div>
{/if}
