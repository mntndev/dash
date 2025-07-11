<script lang="ts">
  import type { WidgetData } from './types';
  // System widgets
  import { ClockWidget } from './widgets/system';
  // Home Assistant widgets
  import { HAEntityWidget, HAButtonWidget } from './widgets/homeassistant';
  // Layout widgets
  import { HorizontalSplitWidget, VerticalSplitWidget, VStackWidget, HStackWidget } from './widgets/layout';

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
        return HorizontalSplitWidget;
      case 'vertical_split':
        return VerticalSplitWidget;
      case 'vstack':
        return VStackWidget;
      case 'hstack':
        return HStackWidget;
      default:
        return null;
    }
  }

  $: component = getWidgetComponent(widget.type);
</script>

<div class="p-4 flex flex-col w-full h-full">
  {#if component}
    <svelte:component this={component} {widget} />
  {:else}
    <div class="flex items-center justify-center h-full text-gray-500 italic">
      <p>Unknown widget type: {widget.type}</p>
    </div>
  {/if}
</div>