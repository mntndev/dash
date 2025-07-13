<script lang="ts">
  import type { WidgetType } from '../types';
  import { useHAEntity } from '../composables/useHAEntity.svelte';

  let { widget }: { widget: WidgetType } = $props();
  
  const { entity } = useHAEntity(widget);
  
  let entityName = $derived(entity.data?.entity_id?.split('.')[1]?.replace(/_/g, ' ') || 'Unknown');
  let lastUpdated = $derived(entity.data?.last_updated ? new Date(entity.data.last_updated).toLocaleString() : 'Never');
  
</script>

<div class="p-4 flex flex-col flex-none">
  <div class="mb-4">
    <h3 class="text-lg md:text-xl font-semibold m-0 capitalize text-gray-200">{entityName}</h3>
    <div class="text-xs text-gray-500 mt-1 font-mono">{entity.data?.entity_id || 'unknown'}</div>
  </div>
  
  <div class="flex items-baseline gap-2 mb-auto">
    <div class="text-2xl md:text-3xl font-bold text-gray-200">{entity.data?.state || 'unavailable'}</div>
    {#if entity.data?.attributes?.unit_of_measurement}
      <div class="text-sm text-gray-500">{entity.data.attributes.unit_of_measurement}</div>
    {/if}
  </div>
  
  <div class="mt-auto pt-4">
    <div class="text-xs text-gray-500">
      Updated: {lastUpdated}
    </div>
  </div>
</div>