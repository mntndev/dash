<script lang="ts">
  import type { WidgetTreeNode, WidgetData, Widget } from './types';
  import ClockWidget from './widgets/ClockWidget.svelte';
  import HAEntityWidget from './widgets/HAEntityWidget.svelte';
  import HAButtonWidget from './widgets/HAButtonWidget.svelte';
  import HASwitchWidget from './widgets/HASwitchWidget.svelte';
  import HALightWidget from './widgets/HALightWidget.svelte';
  import DexcomWidget from './widgets/DexcomWidget.svelte';
  import HorizontalSplitWidget from './widgets/HorizontalSplitWidget.svelte';
  import VerticalSplitWidget from './widgets/VerticalSplitWidget.svelte';
  import GrowWidget from './widgets/GrowWidget.svelte';

  // Accept the new Widget structure or legacy formats for backward compatibility
  let { widgetTreeNode, widget }: { widgetTreeNode?: WidgetTreeNode, widget?: Widget | WidgetData } = $props();

  // Use the widget directly if it's the new format, otherwise convert from legacy formats
  let currentWidget = $derived(
    widget || (widgetTreeNode ? {
      ID: widgetTreeNode.id,
      Type: widgetTreeNode.type,
      Data: null,
      LastUpdate: new Date().toISOString(),
      Children: widgetTreeNode.children?.map(child => ({
        ID: child.id,
        Type: child.type,
        Data: null,
        LastUpdate: new Date().toISOString(),
        Children: []
      }))
    } : null)
  );

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

  let component = $derived(currentWidget ? widgets[currentWidget.Type] : null);
  
  // Debug log for all widgets
  $effect(() => {
    console.log('[Widget Debug]', {
      currentWidget,
      hasWidget: !!currentWidget,
      widgetType: currentWidget?.Type,
      component: !!component
    });
  });
  
  // Debug log for container widgets
  $effect(() => {
    if (currentWidget && currentWidget.Type && (currentWidget.Type.includes('split') || currentWidget.Type === 'grow')) {
      console.log(`[Container Widget] ${currentWidget.ID} (${currentWidget.Type}):`, {
        hasChildren: !!currentWidget.Children?.length,
        childrenCount: currentWidget.Children?.length || 0,
        widget: currentWidget
      });
    }
  });
</script>

{#if component && currentWidget}
  {#key `${currentWidget.ID}-${currentWidget.Type}`}
    <svelte:component this={component} widget={currentWidget} />
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