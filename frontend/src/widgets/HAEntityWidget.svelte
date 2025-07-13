<script lang="ts">
  import type { WidgetType, WidgetDataUpdateEvent } from '../types';
  
  interface HAEntityData {
    entity_id: string;
    state: string;
    attributes: Record<string, unknown>;
    last_changed: string;
    last_updated: string;
  }
  import { Events } from '@wailsio/runtime';
  import { onMount, onDestroy } from 'svelte';

  let { widget }: { widget: WidgetType } = $props();
  
  let entityData = $state<HAEntityData | null>(null);

  onMount(() => {
    Events.On("widget_data_update", (event: WidgetDataUpdateEvent) => {
      if (event.data && event.data.length > 0) {
        const updateInfo = event.data[0];
        if (updateInfo.widget_id === widget.ID && updateInfo.data) {
          entityData = updateInfo.data as HAEntityData;
        }
      }
    });
  });

  onDestroy(() => {
    Events.Off("widget_data_update");
  });
  let entityName = $derived(entityData?.entity_id?.split('.')[1]?.replace(/_/g, ' ') || 'Unknown');
  let lastUpdated = $derived(entityData?.last_updated ? new Date(entityData.last_updated).toLocaleString() : 'Never');
  
</script>

<div class="p-4 flex flex-col flex-none">
  <div class="mb-4">
    <h3 class="text-lg md:text-xl font-semibold m-0 capitalize text-gray-200">{entityName}</h3>
    <div class="text-xs text-gray-500 mt-1 font-mono">{entityData?.entity_id || 'unknown'}</div>
  </div>
  
  <div class="flex items-baseline gap-2 mb-auto">
    <div class="text-2xl md:text-3xl font-bold text-gray-200">{entityData?.state || 'unavailable'}</div>
    {#if entityData?.attributes?.unit_of_measurement}
      <div class="text-sm text-gray-500">{entityData.attributes.unit_of_measurement}</div>
    {/if}
  </div>
  
  <div class="mt-auto pt-4">
    <div class="text-xs text-gray-500">
      Updated: {lastUpdated}
    </div>
  </div>
</div>