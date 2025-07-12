<script lang="ts">
  import type { WidgetData } from '../types';
  import Widget from '../Widget.svelte';

  export let widget: WidgetData;
  
  $: sizes = widget.data?.sizes || [];
  $: children = widget.children || [];
</script>

<div class="flex flex-row w-full h-full gap-2">
  {#each children as child}
    <Widget widget={child} />
  {/each}
</div>