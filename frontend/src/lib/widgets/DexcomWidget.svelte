<script lang="ts">
    import type { WidgetData, DexcomData } from "../types";
    import { getWidgetData } from "../stores.svelte";

    let { widget }: { widget: WidgetData } = $props();
    
    // Get reactive widget data using runes
    let currentWidget = $derived(getWidgetData(widget.id) || widget);
    
    // Derived values using runes
    let glucoseData = $derived(currentWidget.data as DexcomData);
    let glucoseValue = $derived(glucoseData?.value || 0);
    let trend = $derived(glucoseData?.trend || "?");
    let unit = $derived(glucoseData?.unit || "mg/dL");
    let timestamp = $derived(glucoseData?.timestamp ? new Date(glucoseData.timestamp) : null);
    
    // Determine glucose level color based on value
    let glucoseColor = $derived(() => {
        if (glucoseValue < 70) return "text-red-500"; // Low
        if (glucoseValue > 180) return "text-orange-500"; // High
        if (glucoseValue > 250) return "text-red-500"; // Very high
        return "text-green-500"; // Normal
    });
    
    // Format timestamp
    let timeAgo = $derived(() => {
        if (!timestamp) return "Unknown";
        const now = new Date();
        const diffMs = now.getTime() - timestamp.getTime();
        const diffMins = Math.floor(diffMs / 60000);
        
        if (diffMins < 1) return "Just now";
        if (diffMins === 1) return "1 min ago";
        if (diffMins < 60) return `${diffMins} mins ago`;
        
        const diffHours = Math.floor(diffMins / 60);
        if (diffHours === 1) return "1 hour ago";
        return `${diffHours} hours ago`;
    });
</script>

<div class="p-6 flex flex-col items-center justify-center bg-gradient-to-br from-blue-50 to-blue-100 rounded-lg border border-blue-200 flex-none">
    <!-- Main glucose display -->
    <div class="flex items-center gap-3 mb-4">
        <div class="text-5xl md:text-6xl font-bold font-mono {glucoseColor()}">
            {glucoseValue}
        </div>
        <div class="flex flex-col items-start">
            <div class="text-2xl md:text-3xl font-semibold text-gray-600">
                {trend}
            </div>
            <div class="text-sm text-gray-500 font-medium">
                {unit}
            </div>
        </div>
    </div>
    
    <!-- Status information -->
    <div class="text-center">
        <div class="text-lg font-semibold text-gray-700 mb-1">
            Blood Glucose
        </div>
        <div class="text-sm text-gray-500">
            {timeAgo()}
        </div>
        {#if timestamp}
            <div class="text-xs text-gray-400 mt-1">
                {timestamp.toLocaleTimeString()}
            </div>
        {/if}
    </div>
    
    <!-- Range indicator -->
    <div class="mt-4 w-full max-w-xs">
        <div class="flex justify-between text-xs text-gray-500 mb-1">
            <span>Low</span>
            <span>Normal</span>
            <span>High</span>
        </div>
        <div class="h-2 bg-gradient-to-r from-red-400 via-green-400 to-red-400 rounded-full relative">
            <!-- Current value indicator -->
            <div 
                class="absolute w-3 h-3 bg-white border-2 border-gray-600 rounded-full transform -translate-y-0.5"
                style="left: {Math.min(Math.max((glucoseValue - 40) / 260 * 100, 0), 100)}%"
            ></div>
        </div>
        <div class="flex justify-between text-xs text-gray-400 mt-1">
            <span>40</span>
            <span>70-180</span>
            <span>300</span>
        </div>
    </div>
</div>