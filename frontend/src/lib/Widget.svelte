<script lang="ts">
  import type { WidgetData } from './types';
  import ClockWidget from './widgets/ClockWidget.svelte';
  import HAEntityWidget from './widgets/HAEntityWidget.svelte';
  import HAButtonWidget from './widgets/HAButtonWidget.svelte';

  export let widget: WidgetData;

  function getWidgetComponent(type: string) {
    switch (type) {
      case 'clock':
        return ClockWidget;
      case 'home_assistant.entity':
        return HAEntityWidget;
      case 'home_assistant.button':
        return HAButtonWidget;
      case 'horizontal_split':
      case 'vertical_split':
        return null; // Handle layout widgets inline
      default:
        return null;
    }
  }

  $: component = getWidgetComponent(widget.type);
  $: isLayoutWidget = widget.type === 'horizontal_split' || widget.type === 'vertical_split';
  $: layoutDirection = widget.type === 'horizontal_split' ? 'row' : 'column';
  $: sizes = widget.data?.sizes || [];
  $: children = widget.children || [];
</script>

<div class="widget">
  {#if isLayoutWidget}
    <div class="layout-widget" style="flex-direction: {layoutDirection}">
      {#each children as child, index}
        <div 
          class="layout-panel" 
          style="flex: {sizes[index] || 1};"
        >
          <svelte:self widget={child} />
        </div>
      {/each}
    </div>
  {:else if component}
    <svelte:component this={component} {widget} />
  {:else}
    <div class="unknown-widget">
      <p>Unknown widget type: {widget.type}</p>
    </div>
  {/if}
</div>

<style>
  .widget {
    background: white;
    border: 1px solid #e5e5e5;
    border-radius: 0.5rem;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    transition: all 0.2s;
    width: 100%;
    height: 100%;
  }

  .layout-widget {
    display: flex;
    width: 100%;
    height: 100%;
    gap: 0.5rem;
  }

  .layout-panel {
    display: flex;
    flex: 1;
    min-width: 0;
    min-height: 0;
  }

  .widget:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  :global(.dashboard.dark) .widget {
    background: #2d2d2d;
    border-color: #404040;
    color: white;
  }

  .unknown-widget {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #6b7280;
    font-style: italic;
  }

  @media (max-width: 768px) {
    .widget {
      padding: 0.75rem;
    }
  }
</style>