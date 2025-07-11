<script lang="ts">
  import type { WidgetData, ClockData } from '../types';

  export let widget: WidgetData;
  
  $: clockData = widget.data as ClockData;
  $: displayTime = clockData?.display || new Date().toLocaleTimeString();
</script>

<div class="clock-widget">
  <div class="clock-display">
    {displayTime}
  </div>
  <div class="clock-label">
    Current Time
  </div>
</div>

<style>
  .clock-widget {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100%;
  }

  .clock-display {
    font-size: 2rem;
    font-weight: bold;
    font-family: 'Courier New', monospace;
    color: #3b82f6;
  }

  .clock-label {
    font-size: 0.875rem;
    color: #6b7280;
    margin-top: 0.5rem;
  }

  :global(.dashboard.dark) .clock-display {
    color: #60a5fa;
  }

  :global(.dashboard.dark) .clock-label {
    color: #9ca3af;
  }

  @media (max-width: 768px) {
    .clock-display {
      font-size: 1.5rem;
    }
  }
</style>