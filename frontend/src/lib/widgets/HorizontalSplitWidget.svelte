<script lang="ts">
  import type { Widget as WidgetType } from '../types';
  import Widget from '../Widget.svelte';

  let { widget }: { widget: WidgetType } = $props();
  
  let sizes = $derived(widget.Data?.sizes || []);
  let children = $derived(widget.Children || []);
</script>

<div class="flex flex-row w-full h-full gap-2">
  {#each children as child (child.ID)}
    <Widget widget={child} />
  {/each}
</div>