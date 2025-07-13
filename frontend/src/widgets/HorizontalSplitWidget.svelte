<script lang="ts">
  import type { WidgetType, WidgetDataUpdateEvent } from '../types';
  import Widget from '../Widget.svelte';
  import { Events } from '@wailsio/runtime';
  import { onMount, onDestroy } from 'svelte';

  let { widget }: { widget: WidgetType } = $props();
  
  onMount(() => {
    Events.On("widget_data_update", (event: WidgetDataUpdateEvent) => {
      if (event.data && event.data.length > 0) {
        const updateInfo = event.data[0];
        if (updateInfo.widget_id === widget.ID && updateInfo.data) {
          // Handle size data if needed in the future
          // const data = updateInfo.data as SplitData;
        }
      }
    });
  });

  onDestroy(() => {
    Events.Off("widget_data_update");
  });
  
  let children = $derived(widget.Children || []);
</script>

<div class="flex flex-row w-full h-full gap-2">
  {#each children as child (child.ID)}
    <Widget widget={child} />
  {/each}
</div>