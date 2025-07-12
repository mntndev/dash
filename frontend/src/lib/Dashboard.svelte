<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Events } from "@wailsio/runtime";
    import { DashboardAppService } from "../../bindings/github.com/mntndev/dash";
    import Widget from "./Widget.svelte";
    import { 
        appState, 
        setDashboardData
    } from "./stores.svelte";

    onMount(async () => {
        try {
            const data = await DashboardAppService.GetDashboardData();
            setDashboardData(JSON.parse(data));
            appState.loading = false;
        } catch (err) {
            appState.error = `Failed to load dashboard: ${err}`;
            appState.loading = false;
        }

        // Listen for dashboard updates
        Events.On("dashboard_update", (event: any) => {
            console.log("Dashboard update received:", event);
            if (event.data && event.data.length > 0) {
                setDashboardData(event.data[0]);
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
        {#if appState.loading}
            <div class="flex flex-col items-center justify-center h-full">
                <div class="w-10 h-10 border-4 border-gray-500 border-t-gray-200 animate-spin mb-4"></div>
                <p class="text-gray-500">Loading dashboard...</p>
            </div>
        {:else if appState.error}
            <div class="flex flex-col items-center justify-center h-full">
                <p class="text-gray-200">{appState.error}</p>
            </div>
        {:else if appState.dashboardData}
            <div class="w-full h-full min-h-96">
                <Widget widget={appState.dashboardData.widget} />
            </div>
        {/if}
    </div>

    <!-- Connection status banner - outside padding -->
    {#if appState.dashboardData}
        <div class="h-1 w-full flex">
            {#each Object.entries(appState.dashboardData.status) as [service, connected]}
                <div 
                    class="flex-1 transition-colors duration-300"
                    class:bg-gray-200={connected} 
                    class:bg-gray-500={!connected}
                    title="{service}: {connected ? 'Connected' : 'Disconnected'}"
                ></div>
            {/each}
        </div>
    {/if}
</div>