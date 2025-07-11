<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Events } from "@wailsio/runtime";
    import { DashboardAppService } from "../../bindings/github.com/mntndev/dash";
    import Widget from "./Widget.svelte";
    import type { DashboardData } from "./types";

    let dashboardData: DashboardData | null = null;
    let loading = true;
    let error = "";

    onMount(async () => {
        try {
            const data = await DashboardAppService.GetDashboardData();
            dashboardData = JSON.parse(data);
            loading = false;
        } catch (err) {
            error = `Failed to load dashboard: ${err}`;
            loading = false;
        }

        // Listen for dashboard updates
        Events.On("dashboard_update", (event: any) => {
            console.log("Dashboard update received:", event);
            if (event.data && event.data.length > 0) {
                dashboardData = event.data[0];
            }
        });
    });

    onDestroy(() => {
        Events.Off("dashboard_update");
    });

</script>

<!-- Dashboard container -->
<div class="w-full h-full flex flex-col font-sans">
    <!-- Content with padding -->
    <div class="flex-1 p-4 overflow-auto">
        {#if loading}
            <div class="flex flex-col items-center justify-center h-full">
                <div class="w-10 h-10 border-4 border-gray-300 border-t-blue-500 rounded-full animate-spin mb-4"></div>
                <p class="text-gray-600">Loading dashboard...</p>
            </div>
        {:else if error}
            <div class="flex flex-col items-center justify-center h-full">
                <p class="text-red-600">{error}</p>
            </div>
        {:else if dashboardData}
            <div class="w-full h-full min-h-96">
                <Widget widget={dashboardData.widget} />
            </div>
        {/if}
    </div>

    <!-- Connection status banner - outside padding -->
    {#if dashboardData}
        <div class="h-1 w-full flex">
            {#each Object.entries(dashboardData.status) as [service, connected]}
                <div 
                    class="flex-1 transition-colors duration-300"
                    class:bg-green-500={connected} 
                    class:bg-red-500={!connected}
                    title="{service}: {connected ? 'Connected' : 'Disconnected'}"
                ></div>
            {/each}
        </div>
    {/if}
</div>