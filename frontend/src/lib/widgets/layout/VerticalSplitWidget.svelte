<script lang="ts">
  import type { WidgetData } from '../../types';
  import Widget from '../../Widget.svelte';

  export let widget: WidgetData;
  
  $: sizes = widget.data?.sizes || [];
  $: children = widget.children || [];
</script>

<div class="flex flex-col w-full h-full gap-2">
  {#each children as child, index}
    <div 
      class="flex flex-1 min-w-0 min-h-0" 
      style="flex: {sizes[index] || 1};"
    >
      <Widget widget={child} />
    </div>
  {/each}
</div>