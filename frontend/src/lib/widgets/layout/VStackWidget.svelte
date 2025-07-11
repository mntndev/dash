<script lang="ts">
  import type { WidgetData } from '../../types';
  import Widget from '../../Widget.svelte';

  export let widget: WidgetData;
  
  $: children = widget.children || [];
</script>

<div class="flex flex-col w-full h-full gap-2">
  {#each children as child}
    <div class="flex min-w-0 min-h-0">
      <Widget widget={child} />
    </div>
  {/each}
</div>