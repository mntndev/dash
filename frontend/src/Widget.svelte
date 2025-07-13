<script lang="ts">
  import type { WidgetType } from './types';
  import type { Component } from 'svelte';
  import ClockWidget from './widgets/ClockWidget.svelte';
  import HAEntityWidget from './widgets/HAEntityWidget.svelte';
  import HAButtonWidget from './widgets/HAButtonWidget.svelte';
  import HASwitchWidget from './widgets/HASwitchWidget.svelte';
  import HALightWidget from './widgets/HALightWidget.svelte';
  import DexcomWidget from './widgets/DexcomWidget.svelte';
  import HorizontalSplitWidget from './widgets/HorizontalSplitWidget.svelte';
  import VerticalSplitWidget from './widgets/VerticalSplitWidget.svelte';
  import GrowWidget from './widgets/GrowWidget.svelte';

  let { widget }: { widget: WidgetType } = $props();

  let currentWidget = $derived(widget);

  const widgets: Record<string, Component> = {
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

  let SelectedComponent = $derived(currentWidget ? widgets[currentWidget.Type] : null);

</script>

{#if SelectedComponent && currentWidget}
  {#key `${currentWidget.ID}-${currentWidget.Type}`}
    <SelectedComponent widget={currentWidget} />
  {/key}
{:else if currentWidget}
  <div class="flex items-center justify-center h-full text-gray-500 italic">
    <p>Unknown widget type: {currentWidget.Type}</p>
  </div>
{:else}
  <div class="flex items-center justify-center h-full text-gray-500 italic">
    <p>No widget provided</p>
  </div>
{/if}
