<script lang="ts">
  import type { WidgetData } from '../types';
  import Widget from '../Widget.svelte';

  let { widget }: { widget: WidgetData } = $props();
  
  let sizes = $derived(widget.data?.sizes || []);
  let children = $derived(widget.children || []);
</script>

<div class="flex flex-col h-full gap-2 flex-1">
  {#each children as child (child.id)}
    <Widget widget={child} />
  {/each}
</div>