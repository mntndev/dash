<script lang="ts">
  import type { Widget as WidgetType, ClockData } from '../types';

  let { widget }: { widget: WidgetType } = $props();
  
  let clockData = $derived(widget.Data as ClockData);
  let displayTime = $derived(clockData?.display || new Date().toLocaleTimeString());
</script>

<div class="p-4 flex flex-col justify-center items-center flex-none">
  <div class="text-6xl md:text-7xl font-light font-mono text-gray-200">
    {displayTime}
  </div>
</div>