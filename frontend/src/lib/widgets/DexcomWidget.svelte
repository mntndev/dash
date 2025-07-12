<script lang="ts">
    import type { WidgetData, DexcomData } from "../types";
    import { getWidgetData } from "../stores.svelte";

    let { widget }: { widget: WidgetData } = $props();
    
    // Constants
    const width = 400;
    const height = 180;
    const margin = { top: 0, right: 5, bottom: 20, left: 30 };
    
    // Basic reactive data
    let currentWidget = $derived(getWidgetData(widget.id) || widget);
    let glucoseData = $derived(currentWidget.data as DexcomData);
    let glucoseValue = $derived(glucoseData?.value || 0);
    let unit = $derived(glucoseData?.unit || "mg/dL");
    let timestamp = $derived(glucoseData?.timestamp ? new Date(glucoseData.timestamp) : null);
    let historicalData = $derived(glucoseData?.historical || []);
    let lowThreshold = $derived(glucoseData?.low_threshold || 70);
    let highThreshold = $derived(glucoseData?.high_threshold || 160);
    
    // Graph dimensions
    let graphWidth = $derived(width - margin.left - margin.right);
    let graphHeight = $derived(height - margin.top - margin.bottom);
    
    // Data processing
    let sortedData = $derived.by(() => {
        return [...historicalData]
            .filter(d => d.value > 0)
            .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
    });
    
    // Scales and extents
    let timeExtent = $derived.by(() => {
        if (sortedData.length === 0) return [new Date(), new Date()];
        const times = sortedData.map(d => new Date(d.timestamp).getTime());
        return [new Date(Math.min(...times)), new Date(Math.max(...times))];
    });
    
    let valueExtent = $derived.by(() => {
        if (sortedData.length === 0) return [lowThreshold - 10, highThreshold + 10];
        const values = sortedData.map(d => d.value);
        const dataMin = Math.min(...values);
        const dataMax = Math.max(...values);
        const min = Math.min(dataMin, lowThreshold);
        const max = Math.max(dataMax, highThreshold);
        const padding = (max - min) * 0.1;
        return [Math.max(40, min - padding), Math.min(400, max + padding)];
    });
    
    // Scale functions
    let xScale = $derived((timestamp: string) => {
        const time = new Date(timestamp).getTime();
        const [minTime, maxTime] = timeExtent;
        const range = maxTime.getTime() - minTime.getTime();
        
        if (range === 0 || range < 1000) {
            const index = sortedData.findIndex(d => d.timestamp === timestamp);
            return (index / Math.max(1, sortedData.length - 1)) * graphWidth;
        }
        
        return ((time - minTime.getTime()) / range) * graphWidth;
    });
    
    let yScale = $derived((value: number) => {
        const [minValue, maxValue] = valueExtent;
        const range = maxValue - minValue;
        return range === 0 ? graphHeight / 2 : graphHeight - ((value - minValue) / range) * graphHeight;
    });
    
    // SVG path and positions
    let linePath = $derived.by(() => {
        if (sortedData.length === 0) return "";
        return sortedData.map((d, i) => {
            const x = xScale(d.timestamp);
            const y = yScale(d.value);
            return i === 0 ? `M ${x} ${y}` : `L ${x} ${y}`;
        }).join(" ");
    });
    
    let lowRangeY = $derived(yScale(lowThreshold));
    let highRangeY = $derived(yScale(highThreshold));
    
    // Time labels
    let xTimeLabels = $derived.by(() => {
        const [minTime, maxTime] = timeExtent;
        const range = maxTime.getTime() - minTime.getTime();
        const hourInMs = 60 * 60 * 1000;
        const step = range > 6 * hourInMs ? 2 * hourInMs : hourInMs;
        
        const labels = [];
        const startTime = Math.ceil(minTime.getTime() / step) * step;
        
        for (let time = startTime; time <= maxTime.getTime(); time += step) {
            const date = new Date(time);
            labels.push({
                time: date.toISOString(),
                x: xScale(date.toISOString()),
                label: date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
            });
        }
        return labels;
    });
    
    // Display values - simplified to just use off-white text
    let glucoseColor = $derived("fill-gray-200");
    
    let timeAgo = $derived.by(() => {
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

<div class="flex flex-col flex-1 min-h-[200px]">
    {#if historicalData && historicalData.length > 0}
        <div class="w-full h-full min-h-[200px] max-h-[400px]">
            <svg class="w-full h-full" viewBox="0 0 {width} {height}" preserveAspectRatio="xMidYMid meet">
            <defs>
                <pattern id="lowHatch" patternUnits="userSpaceOnUse" width="4" height="4" patternTransform="scale(2 2)">
                    <path d="M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2" stroke="#6b7280" stroke-width="0.5" opacity="0.3"/>
                </pattern>
                <pattern id="highHatch" patternUnits="userSpaceOnUse" width="4" height="4" patternTransform="scale(2 2)">
                    <path d="M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2" stroke="#6b7280" stroke-width="0.5" opacity="0.3"/>
                </pattern>
            </defs>
            
            <g transform="translate({margin.left}, {margin.top})">
                <!-- Low range background -->
                <rect x="0" y="{lowRangeY}" width="{graphWidth}" height="{Math.max(0, graphHeight - lowRangeY)}" fill="url(#lowHatch)"/>
                
                <!-- High range background -->
                <rect x="0" y="0" width="{graphWidth - 80}" height="{Math.max(0, highRangeY)}" fill="url(#highHatch)"/>
                
                <!-- Data line -->
                {#if linePath}
                    <path d="{linePath}" fill="none" stroke="#e5e7eb" stroke-width="2"/>
                {/if}
                
                <!-- Data points -->
                {#each sortedData as point}
                    {@const x = xScale(point.timestamp)}
                    {@const y = yScale(point.value)}
                    <circle cx="{x}" cy="{y}" r="3" fill="#e5e7eb"/>
                {/each}
                
                <!-- Threshold lines -->
                <line x1="0" y1="{lowRangeY}" x2="{graphWidth}" y2="{lowRangeY}" stroke="#6b7280" stroke-width="1" stroke-dasharray="3,3"/>
                <text x="-5" y="{lowRangeY + 4}" text-anchor="end" class="text-xs fill-gray-500 font-semibold">{lowThreshold}</text>
                
                <line x1="0" y1="{highRangeY}" x2="{graphWidth - 80}" y2="{highRangeY}" stroke="#6b7280" stroke-width="1" stroke-dasharray="3,3"/>
                <text x="-5" y="{highRangeY + 4}" text-anchor="end" class="text-xs fill-gray-500 font-semibold">{highThreshold}</text>
                
                <!-- Time labels -->
                {#each xTimeLabels as timeLabel}
                    <line x1="{timeLabel.x}" y1="{graphHeight}" x2="{timeLabel.x}" y2="{graphHeight + 5}" stroke="#6b7280" stroke-width="1"/>
                    <text x="{timeLabel.x}" y="{graphHeight + 15}" text-anchor="middle" class="text-xs fill-gray-500">{timeLabel.label}</text>
                {/each}
                
                <!-- Axis lines -->
                <line x1="0" y1="0" x2="0" y2="{graphHeight}" stroke="#6b7280" stroke-width="1"/>
                <line x1="0" y1="{graphHeight}" x2="{graphWidth}" y2="{graphHeight}" stroke="#6b7280" stroke-width="1"/>
                
                <!-- Current glucose reading -->
                <text x="{graphWidth - 10}" y="15" text-anchor="end" class="text-2xl font-bold font-mono {glucoseColor}">{glucoseValue}</text>
                <text x="{graphWidth - 10}" y="30" text-anchor="end" class="text-xs fill-gray-500 font-medium">{unit}</text>
                <text x="{graphWidth - 10}" y="42" text-anchor="end" class="text-xs fill-gray-500">{timeAgo}</text>
            </g>
            
            <!-- Y-axis label -->
            <text x="12" y="{margin.top + graphHeight / 2}" text-anchor="middle" transform="rotate(-90, 12, {margin.top + graphHeight / 2})" class="text-xs fill-gray-500">mg/dL</text>
        </svg>
        </div>
    {:else}
        <div class="flex-1 flex items-center justify-center text-gray-500 min-h-[120px]">
            <div class="text-center">
                <div class="text-lg mb-2">📊</div>
                <div class="text-sm">Loading historical data...</div>
            </div>
        </div>
    {/if}
</div>