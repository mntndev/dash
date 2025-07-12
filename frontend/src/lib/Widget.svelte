<script lang="ts">
  import type { WidgetData } from './types';
  import ClockWidget from './widgets/ClockWidget.svelte';
  import HAEntityWidget from './widgets/HAEntityWidget.svelte';
  import HAButtonWidget from './widgets/HAButtonWidget.svelte';
  import HASwitchWidget from './widgets/HASwitchWidget.svelte';
  import HALightWidget from './widgets/HALightWidget.svelte';
  import DexcomWidget from './widgets/DexcomWidget.svelte';
  import HorizontalSplitWidget from './widgets/HorizontalSplitWidget.svelte';
  import VerticalSplitWidget from './widgets/VerticalSplitWidget.svelte';
  import GrowWidget from './widgets/GrowWidget.svelte';

  let { widget }: { widget: WidgetData } = $props();

  const widgets: Record<string, any> = {
    'clock': ClockWidget,
    'home_assistant.entity': HAEntityWidget,
    'home_assistant.button': HAButtonWidget,
    'home_assistant.switch': HASwitchWidget,
    'home_assistant.light': HALightWidget,
    'dexcom': DexcomWidget,
    'horizontal_split': HorizontalSplitWidget,
    'vertical_split': VerticalSplitWidget,
    'grow': GrowWidget
  };

  let component = $derived(widgets[widget.type]);
  
  // Extract stable widget identity from changing data
  let widgetId = $derived(widget.id);
  let widgetType = $derived(widget.type);
</script>

{#if component}
  {#key `${widgetId}-${widgetType}`}
    <svelte:component this={component} {widget} />
  {/key}
{:else}
  <div class="flex items-center justify-center h-full text-gray-500 italic">
    <p>Unknown widget type: {widget.type}</p>
  </div>
{/if}