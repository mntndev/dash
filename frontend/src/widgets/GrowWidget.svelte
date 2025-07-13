<script lang="ts">
    import type { WidgetType, WidgetDataUpdateEvent } from "../types";
    
    interface GrowData {
        grow_value: string;
    }
    import Widget from "../Widget.svelte";
    import { Events } from '@wailsio/runtime';
    import { onMount, onDestroy } from 'svelte';

    let { widget }: { widget: WidgetType } = $props();
    
    let growValue = $state("1");

    onMount(() => {
        Events.On("widget_data_update", (event: WidgetDataUpdateEvent) => {
            if (event.data && event.data.length > 0) {
                const updateInfo = event.data[0];
                if (updateInfo.widget_id === widget.ID && updateInfo.data) {
                    const data = updateInfo.data as GrowData;
                    growValue = data?.grow_value || "1";
                }
            }
        });
    });

    onDestroy(() => {
        Events.Off("widget_data_update");
    });
    
    let children = $derived(widget.Children || []);
    let hasChild = $derived(children.length > 0);
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
