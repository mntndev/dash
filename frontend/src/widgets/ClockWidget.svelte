<script lang="ts">
  import type { WidgetType, WidgetDataUpdateEvent } from '../types';
  
  interface ClockData {
    time: string;
    format: string;
    display: string;
  }
  import { Events } from '@wailsio/runtime';
  import { onMount, onDestroy } from 'svelte';

  let { widget }: { widget: WidgetType } = $props();
  
  let clockData = $state<ClockData | null>(null);
  let displayTime = $derived(clockData?.display || new Date().toLocaleTimeString());

  onMount(() => {
    Events.On("widget_data_update", (event: WidgetDataUpdateEvent) => {
      if (event.data && event.data.length > 0) {
        const updateInfo = event.data[0];
        if (updateInfo.widget_id === widget.ID && updateInfo.data) {
          clockData = updateInfo.data as ClockData;
        }
      }
    });
  });

  onDestroy(() => {
    Events.Off("widget_data_update");
  });
</script>

<div class="p-4 flex flex-col justify-center items-center flex-none">
  <div class="text-6xl md:text-7xl font-light font-mono text-gray-200">
    {displayTime}
  </div>
</div>