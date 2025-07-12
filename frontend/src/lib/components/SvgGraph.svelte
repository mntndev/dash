<script lang="ts">
    import type { DexcomReading } from "../types";
    
    interface Props {
        data: DexcomReading[];
        width?: number;
        height?: number;
        className?: string;
        lowThreshold?: number;
        highThreshold?: number;
        currentValue?: number;
        currentUnit?: string;
        timeAgo?: string;
        glucoseColor?: string;
    }
    
    let { data, width = 400, height = 200, className = "", lowThreshold = 70, highThreshold = 160, currentValue, currentUnit, timeAgo, glucoseColor }: Props = $props();
    
    // Graph dimensions and margins
    const margin = { top: 0, right: 5, bottom: 20, left: 30 };
    let graphWidth = $derived(width - margin.left - margin.right);
    let graphHeight = $derived(height - margin.top - margin.bottom);
    
    // Data processing
    let sortedData = $derived(() => {
        const filtered = [...data]
            .filter(d => d.value > 0)
            .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
        return filtered;
    });
    
    // Scales
    let timeExtent = $derived(() => {
        const dataArray = sortedData();
        if (dataArray.length === 0) return [new Date(), new Date()];
        const times = dataArray.map(d => new Date(d.timestamp).getTime());
        return [new Date(Math.min(...times)), new Date(Math.max(...times))];
    });
    
    let valueExtent = $derived(() => {
        const dataArray = sortedData();
        if (dataArray.length === 0) return [lowThreshold - 10, highThreshold + 10];
        const values = dataArray.map(d => d.value);
        const dataMin = Math.min(...values);
        const dataMax = Math.max(...values);
        
        // Include thresholds in the range to ensure they're always visible
        const min = Math.min(dataMin, lowThreshold);
        const max = Math.max(dataMax, highThreshold);
        
        const padding = (max - min) * 0.1;
        return [Math.max(40, min - padding), Math.min(400, max + padding)];
    });
    
    // Scale functions
    let xScale = $derived((timestamp: string) => {
        const time = new Date(timestamp).getTime();
        const [minTime, maxTime] = timeExtent();
        const range = maxTime.getTime() - minTime.getTime();
        
        if (range === 0 || range < 1000) {
            const dataArray = sortedData();
            const index = dataArray.findIndex(d => d.timestamp === timestamp);
            return (index / Math.max(1, dataArray.length - 1)) * graphWidth;
        }
        
        return ((time - minTime.getTime()) / range) * graphWidth;
    });
    
    let yScale = $derived((value: number) => {
        const [minValue, maxValue] = valueExtent();
        const range = maxValue - minValue;
        return range === 0 ? graphHeight / 2 : graphHeight - ((value - minValue) / range) * graphHeight;
    });
    
    // Generate SVG path for the line
    let linePath = $derived(() => {
        const dataArray = sortedData();
        if (dataArray.length === 0) return "";
        
        return dataArray.map((d, i) => {
            const x = xScale(d.timestamp);
            const y = yScale(d.value);
            return i === 0 ? `M ${x} ${y}` : `L ${x} ${y}`;
        }).join(" ");
    });

    // Calculate range boundaries - these should always be visible now
    let lowRangeY = $derived(() => yScale(lowThreshold));
    let highRangeY = $derived(() => yScale(highThreshold));

    // X-axis time labels
    let xTimeLabels = $derived(() => {
        const [minTime, maxTime] = timeExtent();
        const range = maxTime.getTime() - minTime.getTime();
        const hourInMs = 60 * 60 * 1000;
        const step = range > 6 * hourInMs ? 2 * hourInMs : hourInMs; // 1 or 2 hour steps
        
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
</script>

<svg class={className} viewBox="0 0 {width} {height}" preserveAspectRatio="xMidYMid meet">
    <defs>
        <!-- Low range hatch pattern -->
        <pattern
            id="lowHatch"
            patternUnits="userSpaceOnUse"
            width="4"
            height="4"
            patternTransform="scale(2 2)"
        >
            <path
                d="M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2"
                stroke="#ef4444"
                stroke-width="0.5"
                opacity="0.3"
            />
        </pattern>
        
        <!-- High range hatch pattern -->
        <pattern
            id="highHatch"
            patternUnits="userSpaceOnUse"
            width="4"
            height="4"
            patternTransform="scale(2 2)"
        >
            <path
                d="M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2"
                stroke="#f59e0b"
                stroke-width="0.5"
                opacity="0.3"
            />
        </pattern>
    </defs>
    
    <g transform="translate({margin.left}, {margin.top})">
        <!-- Low range background (below low threshold) -->
        <rect 
            x="0" 
            y="{lowRangeY()}" 
            width="{graphWidth}" 
            height="{Math.max(0, graphHeight - lowRangeY())}"
            fill="url(#lowHatch)"
        />
        
        <!-- High range background (above high threshold) - shortened to make room for glucose reading -->
        <rect 
            x="0" 
            y="0" 
            width="{graphWidth - 80}" 
            height="{Math.max(0, highRangeY())}"
            fill="url(#highHatch)"
        />
        <!-- Data line -->
        {#if linePath()}
            <path 
                d="{linePath()}" 
                fill="none" 
                stroke="#3b82f6" 
                stroke-width="2"
            ></path>
        {/if}
        
        <!-- Data points -->
        {#each sortedData() as point}
            {@const x = xScale(point.timestamp)}
            {@const y = yScale(point.value)}
            {@const color = point.value < lowThreshold ? "#ef4444" : 
                           point.value > highThreshold ? "#f59e0b" : 
                           "#3b82f6"}
            
            <circle 
                cx="{x}" 
                cy="{y}" 
                r="3" 
                fill="{color}"
            ></circle>
        {/each}
        
        <!-- Threshold lines and labels -->
        <line 
            x1="0" 
            y1="{lowRangeY()}" 
            x2="{graphWidth}" 
            y2="{lowRangeY()}" 
            stroke="#ef4444" 
            stroke-width="1" 
            stroke-dasharray="3,3"
        ></line>
        <text 
            x="-5" 
            y="{lowRangeY() + 4}" 
            text-anchor="end" 
            class="text-xs fill-red-500 font-semibold"
        >
            {lowThreshold}
        </text>
        
        <line 
            x1="0" 
            y1="{highRangeY()}" 
            x2="{graphWidth - 80}" 
            y2="{highRangeY()}" 
            stroke="#f59e0b" 
            stroke-width="1" 
            stroke-dasharray="3,3"
        ></line>
        <text 
            x="-5" 
            y="{highRangeY() + 4}" 
            text-anchor="end" 
            class="text-xs fill-orange-500 font-semibold"
        >
            {highThreshold}
        </text>
        
        <!-- X-axis time labels and ticks -->
        {#each xTimeLabels() as timeLabel}
            <!-- Tick mark -->
            <line 
                x1="{timeLabel.x}" 
                y1="{graphHeight}" 
                x2="{timeLabel.x}" 
                y2="{graphHeight + 5}" 
                stroke="#374151" 
                stroke-width="1"
            ></line>
            <!-- Time label -->
            <text 
                x="{timeLabel.x}" 
                y="{graphHeight + 15}" 
                text-anchor="middle" 
                class="text-xs fill-gray-500"
            >
                {timeLabel.label}
            </text>
        {/each}
        
        
        <!-- Axis lines -->
        <line x1="0" y1="0" x2="0" y2="{graphHeight}" stroke="#374151" stroke-width="1"></line>
        <line x1="0" y1="{graphHeight}" x2="{graphWidth}" y2="{graphHeight}" stroke="#374151" stroke-width="1"></line>
        
        <!-- Current glucose reading in top right -->
        {#if currentValue}
            <text 
                x="{graphWidth - 10}" 
                y="25" 
                text-anchor="end" 
                class="text-2xl font-bold font-mono {glucoseColor || 'fill-gray-600'}"
            >
                {currentValue}
            </text>
            {#if currentUnit}
                <text 
                    x="{graphWidth - 10}" 
                    y="40" 
                    text-anchor="end" 
                    class="text-xs fill-gray-500 font-medium"
                >
                    {currentUnit}
                </text>
            {/if}
            {#if timeAgo}
                <text 
                    x="{graphWidth - 10}" 
                    y="52" 
                    text-anchor="end" 
                    class="text-xs fill-gray-500"
                >
                    {timeAgo}
                </text>
            {/if}
        {/if}
    </g>
    
    <!-- Y-axis label -->
    <text x="12" y="{margin.top + graphHeight / 2}" text-anchor="middle" transform="rotate(-90, 12, {margin.top + graphHeight / 2})" class="text-xs fill-gray-600">mg/dL</text>
</svg>