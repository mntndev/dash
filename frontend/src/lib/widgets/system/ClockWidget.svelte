<script lang="ts">
  import type { WidgetData, ClockData } from '../../types';

  export let widget: WidgetData;
  
  $: clockData = widget.data as ClockData;
  $: displayTime = clockData?.display || new Date().toLocaleTimeString();
</script>

<div class="flex flex-col justify-center items-center h-full">
  <div class="text-3xl md:text-4xl font-light font-mono text-blue-600">
    {displayTime}
  </div>
  <div class="text-sm text-gray-500 mt-2">
    Current Time
  </div>
</div>