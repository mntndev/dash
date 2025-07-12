<script lang="ts">
    import type { WidgetData } from "../types";
    import Widget from "../Widget.svelte";

    export let widget: WidgetData;
    
    $: children = widget.children || [];
    $: hasChild = children.length > 0;
</script>

{#if hasChild}
    <div class="flex-1">
        {#each children as child}
            <Widget widget={child} />
        {/each}
    </div>
{:else}
    <div class="flex-1"></div>
{/if}
