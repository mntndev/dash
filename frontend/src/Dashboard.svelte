<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Events } from "@wailsio/runtime";
    import { DashboardService } from "../bindings/github.com/mntndev/dash/pkg/dashboard";
    import Widget from "./Widget.svelte";
    import { appState, setDashboardInfo } from "./stores.svelte";
    import type { DashboardUpdateEvent } from "./types";

    async function loadDashboard() {
        try {
            const dashboardInfo = await DashboardService.GetDashboardInfo();
            
            if (!dashboardInfo) {
                throw new Error("No dashboard info received");
            }

            setDashboardInfo(dashboardInfo);

            if (!dashboardInfo.status?.initializing) {
                appState.loading = false;
            }
        } catch (err) {
            appState.error = `Failed to load dashboard: ${err}`;
            appState.loading = false;
        }
    }

    async function handleDashboardUpdate(event: DashboardUpdateEvent) {
        if (!event.data?.length) return;
        
        const dashboardInfo = event.data[0];
        
        if (!dashboardInfo.status?.initializing) {
            await loadDashboard();
        } else {
            setDashboardInfo(dashboardInfo);
        }
    }

    onMount(async () => {
        await loadDashboard();
        Events.On("dashboard_info", handleDashboardUpdate);
        
        // Signal to backend that frontend is ready to receive events
        try {
            await DashboardService.SignalFrontendReady();
        } catch (err) {
            console.error("Failed to signal frontend ready:", err);
        }
    });

    onDestroy(() => {
        Events.Off("dashboard_info");
    });
</script>

<div class="w-full h-full flex flex-col font-sans">
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
        {:else if appState.dashboardInfo?.root_widget}
            <div class="w-full h-full min-h-96">
                <Widget widget={appState.dashboardInfo.root_widget} />
            </div>
        {/if}
    </div>

    {#if appState.dashboardInfo}
        {@const disconnectedServices = Object.entries(appState.dashboardInfo.status).filter(([, connected]) => !connected)}
        {#if disconnectedServices.length > 0}
            <div class="h-1 w-full flex">
                {#each disconnectedServices as [service] (service)}
                    <div class="flex-1 bg-gray-500 transition-colors duration-300" title="{service}: Disconnected"></div>
                {/each}
            </div>
        {/if}
    {/if}
</div>
