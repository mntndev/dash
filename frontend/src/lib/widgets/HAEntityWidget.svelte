<script lang="ts">
  import type { WidgetData, HAEntityData } from '../types';

  export let widget: WidgetData;
  
  $: entityData = widget.data as HAEntityData;
  $: entityName = entityData?.entity_id?.split('.')[1]?.replace(/_/g, ' ') || 'Unknown';
  $: lastUpdated = entityData?.last_updated ? new Date(entityData.last_updated).toLocaleString() : 'Never';
</script>

<div class="ha-entity-widget">
  <div class="entity-header">
    <h3 class="entity-name">{entityName}</h3>
    <div class="entity-id">{entityData?.entity_id || 'unknown'}</div>
  </div>
  
  <div class="entity-state">
    <div class="state-value">{entityData?.state || 'unavailable'}</div>
    {#if entityData?.attributes?.unit_of_measurement}
      <div class="state-unit">{entityData.attributes.unit_of_measurement}</div>
    {/if}
  </div>
  
  <div class="entity-footer">
    <div class="last-updated">
      Updated: {lastUpdated}
    </div>
  </div>
</div>

<style>
  .ha-entity-widget {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .entity-header {
    margin-bottom: 1rem;
  }

  .entity-name {
    font-size: 1.1rem;
    font-weight: 600;
    margin: 0;
    text-transform: capitalize;
  }

  .entity-id {
    font-size: 0.75rem;
    color: #6b7280;
    margin-top: 0.25rem;
    font-family: monospace;
  }

  .entity-state {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin-bottom: auto;
  }

  .state-value {
    font-size: 1.5rem;
    font-weight: bold;
    color: #059669;
  }

  .state-unit {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .entity-footer {
    margin-top: auto;
    padding-top: 1rem;
  }

  .last-updated {
    font-size: 0.75rem;
    color: #9ca3af;
  }

  :global(.dashboard.dark) .entity-id {
    color: #9ca3af;
  }

  :global(.dashboard.dark) .state-value {
    color: #10b981;
  }

  :global(.dashboard.dark) .state-unit {
    color: #9ca3af;
  }

  :global(.dashboard.dark) .last-updated {
    color: #6b7280;
  }

  @media (max-width: 768px) {
    .entity-name {
      font-size: 1rem;
    }
    
    .state-value {
      font-size: 1.25rem;
    }
  }
</style>