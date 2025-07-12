<script lang="ts">
  import type { WidgetData, HAEntityData } from '../types';

  let { widget }: { widget: WidgetData } = $props();
  
  let entityData = $derived(widget.data as HAEntityData);
  let entityName = $derived(entityData?.entity_id?.split('.')[1]?.replace(/_/g, ' ') || 'Unknown');
  let lastUpdated = $derived(entityData?.last_updated ? new Date(entityData.last_updated).toLocaleString() : 'Never');
  
  // Log any changes in entityData
  $effect(() => {
    if (entityData) {
      console.log(`[HA Entity] ${entityData.entity_id} data updated:`, entityData);
    }
  });
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