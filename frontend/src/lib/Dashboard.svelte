<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Events } from "@wailsio/runtime";
    import { DashboardService } from "../../bindings/github.com/mntndev/dash/pkg/dashboard";
    import Widget from "./Widget.svelte";
    import { 
        appState, 
        setDashboardInfo,
        setWidgetData,
        getAllWidgetIds
    } from "./stores.svelte";

    async function loadDashboard() {
        try {
            console.log("[Dashboard] Loading dashboard...");
            const dashboardInfo = await DashboardService.GetDashboardInfo();
            console.log("[Dashboard] Got dashboard info:", dashboardInfo);
            
            if (!dashboardInfo) {
                throw new Error("No dashboard info received");
            }
            
            // Check if still initializing
            if (dashboardInfo.status?.initializing) {
                console.log("[Dashboard] Dashboard still initializing, will wait for event...");
                setDashboardInfo(dashboardInfo);
                return; // Wait for dashboard_info event
            }
            
            setDashboardInfo(dashboardInfo);
            
            // Get initial data for all widgets - but we don't need to since the root_widget contains all data
            if (dashboardInfo.root_widget) {
                console.log("[Dashboard] Root widget loaded:", dashboardInfo.root_widget);
            } else {
                console.warn("[Dashboard] No root_widget in dashboard info");
            }
            
            console.log("[Dashboard] Setting loading to false");
            appState.loading = false;
        } catch (err) {
            console.error("[Dashboard] Error loading dashboard:", err);
            appState.error = `Failed to load dashboard: ${err}`;
            appState.loading = false;
        }
    }

    onMount(async () => {
        // Wait for Wails runtime to be ready
        const waitForWails = () => {
            return new Promise((resolve) => {
                if (typeof window !== 'undefined' && (window as any)._wails) {
                    resolve(true);
                } else {
                    setTimeout(() => waitForWails().then(resolve), 100);
                }
            });
        };

        console.log("[Dashboard] Waiting for Wails runtime...");
        await waitForWails();
        console.log("[Dashboard] Wails runtime ready, loading dashboard...");
        
        await loadDashboard();

        // Listen for dashboard info updates (structure changes)
        Events.On("dashboard_info", async (event: any) => {
            console.log("[Dashboard Info] Structure update received");
            if (event.data && event.data.length > 0) {
                const dashboardInfo = event.data[0];
                console.log("[Dashboard Info] Received dashboard info:", dashboardInfo);
                
                // If this is the completion of initialization, load the full dashboard
                if (!dashboardInfo.status?.initializing) {
                    console.log("[Dashboard Info] Initialization complete, loading full dashboard");
                    await loadDashboard();
                } else {
                    setDashboardInfo(dashboardInfo);
                }
            }
        });

        // Listen for widget-specific data updates
        Events.On("widget_data_update", (event: any) => {
            console.log("[Widget Data Update] Received:", event.data);
            if (event.data && event.data.length > 0) {
                const updateInfo = event.data[0];
                if (updateInfo.widget_id && updateInfo.data) {
                    setWidgetData(updateInfo.data);
                }
            }
        });
    });

    onDestroy(() => {
        Events.Off("dashboard_info");
        Events.Off("widget_data_update");
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
        {:else if appState.dashboardInfo}
            <div class="w-full h-full min-h-96">
                <Widget widget={appState.dashboardInfo.root_widget} />
            </div>
        {/if}
    </div>

    <!-- Connection status banner - outside padding -->
    {#if appState.dashboardInfo}
        {@const disconnectedServices = Object.entries(appState.dashboardInfo.status).filter(([service, connected]) => !connected)}
        {#if disconnectedServices.length > 0}
            <div class="h-1 w-full flex">
                {#each disconnectedServices as [service, connected]}
                    <div 
                        class="flex-1 transition-colors duration-300 bg-gray-500"
                        title="{service}: Disconnected"
                    ></div>
                {/each}
            </div>
        {/if}
    {/if}
</div>