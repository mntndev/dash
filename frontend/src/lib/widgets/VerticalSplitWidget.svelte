<script lang="ts">
  import type { WidgetData } from '../types';
  import Widget from '../Widget.svelte';

  export let widget: WidgetData;
  
  $: sizes = widget.data?.sizes || [];
  $: children = widget.children || [];
</script>

<div class="flex flex-col h-full gap-2 flex-1">
  {#each children as child}
    <Widget widget={child} />
  {/each}
</div>