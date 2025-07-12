<script lang="ts">
  import type { WidgetType } from '../types';
  import Widget from '../Widget.svelte';
  import { Events } from '@wailsio/runtime';
  import { onMount, onDestroy } from 'svelte';

  let { widget }: { widget: WidgetType } = $props();
  
  let sizes = $state<any[]>([]);

  onMount(() => {
    Events.On("widget_data_update", (event: any) => {
      if (event.data && event.data.length > 0) {
        const updateInfo = event.data[0];
        if (updateInfo.widget_id === widget.ID && updateInfo.data) {
          sizes = updateInfo.data?.sizes || [];
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