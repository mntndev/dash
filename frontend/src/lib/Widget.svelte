<script lang="ts">
  import type { WidgetData } from './types';
  import ClockWidget from './widgets/ClockWidget.svelte';
  import HAEntityWidget from './widgets/HAEntityWidget.svelte';
  import HAButtonWidget from './widgets/HAButtonWidget.svelte';
  import HASwitchWidget from './widgets/HASwitchWidget.svelte';
  import HorizontalSplitWidget from './widgets/HorizontalSplitWidget.svelte';
  import VerticalSplitWidget from './widgets/VerticalSplitWidget.svelte';
  import GrowWidget from './widgets/GrowWidget.svelte';

  export let widget: WidgetData;

  const widgets = {
    'clock': ClockWidget,
    'home_assistant.entity': HAEntityWidget,
    'home_assistant.button': HAButtonWidget,
    'home_assistant.switch': HASwitchWidget,
    'horizontal_split': HorizontalSplitWidget,
    'vertical_split': VerticalSplitWidget,
    'grow': GrowWidget
  };

  $: component = widgets[widget.type];
</script>

{#if component}
  <svelte:component this={component} {widget} />
{:else}
  <div class="flex items-center justify-center h-full text-gray-500 italic">
    <p>Unknown widget type: {widget.type}</p>
  </div>
{/if}