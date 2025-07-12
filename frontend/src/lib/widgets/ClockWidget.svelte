<script lang="ts">
  import type { WidgetData, ClockData } from '../types';

  export let widget: WidgetData;
  
  $: clockData = widget.data as ClockData;
  $: displayTime = clockData?.display || new Date().toLocaleTimeString();
</script>

<div class="p-4 flex flex-col justify-center items-center flex-none">
  <div class="text-6xl md:text-7xl font-light font-mono text-blue-600">
    {displayTime}
  </div>
</div>