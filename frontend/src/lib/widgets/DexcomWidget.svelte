<script lang="ts">
    import type { WidgetData, DexcomData } from "../types";
    import { getWidgetData } from "../stores.svelte";
    import SvgGraph from "../components/SvgGraph.svelte";

    let { widget }: { widget: WidgetData } = $props();
    
    // Get reactive widget data using runes
    let currentWidget = $derived(getWidgetData(widget.id) || widget);
    
    // Derived values using runes
    let glucoseData = $derived(currentWidget.data as DexcomData);
    let glucoseValue = $derived(glucoseData?.value || 0);
    let trend = $derived(glucoseData?.trend || "?");
    let unit = $derived(glucoseData?.unit || "mg/dL");
    let timestamp = $derived(glucoseData?.timestamp ? new Date(glucoseData.timestamp) : null);
    let historicalData = $derived(glucoseData?.historical || []);
    
    // Determine glucose level color based on value (SVG fill classes)
    let glucoseColor = $derived(() => {
        const lowThreshold = glucoseData?.low_threshold || 70;
        const highThreshold = glucoseData?.high_threshold || 160;
        
        if (glucoseValue < lowThreshold) return "fill-red-500"; // Low
        if (glucoseValue > highThreshold) return "fill-orange-500"; // High
        return "fill-green-500"; // Normal
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

<div class="flex flex-col flex-1 min-h-0">
    <!-- Historical graph -->
    <div class="flex-1 min-h-0 mb-4 min-h-[120px] relative">
        {#if historicalData && historicalData.length > 0}
            <SvgGraph 
                data={historicalData} 
                width={400} 
                height={180} 
                className="w-full h-full"
                lowThreshold={glucoseData?.low_threshold || 70}
                highThreshold={glucoseData?.high_threshold || 160}
                currentValue={glucoseValue}
                currentUnit={unit}
                timeAgo={timeAgo()}
                glucoseColor={glucoseColor()}
            />
        {:else}
            <div class="flex-1 flex items-center justify-center text-gray-500 min-h-[120px]">
                <div class="text-center">
                    <div class="text-lg mb-2">📊</div>
                    <div class="text-sm">Loading historical data...</div>
                </div>
            </div>
        {/if}
    </div>
</div>