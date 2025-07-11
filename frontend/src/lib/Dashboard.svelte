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
        Events.On("dashboard_update", (data: any) => {
            dashboardData = data.data;
        });
    });

    onDestroy(() => {
        Events.Off("dashboard_update");
    });

    $: isDark = dashboardData?.theme === "dark";
</script>

<div class="dashboard" class:dark={isDark}>
    {#if loading}
        <div class="loading">
            <div class="spinner"></div>
            <p class="text-gray-600">Loading dashboard...</p>
        </div>
    {:else if error}
        <div class="error">
            <p class="text-red-600">{error}</p>
        </div>
    {:else if dashboardData}
        <header class="dashboard-header">
            <h1 class="text-2xl font-bold">{dashboardData.title}</h1>
            <div class="status-indicators">
                {#each Object.entries(dashboardData.status) as [service, connected]}
                    <div class="status-indicator" class:connected>
                        <div class="status-dot"></div>
                        <span class="text-sm">{service}</span>
                    </div>
                {/each}
            </div>
        </header>

        <main class="dashboard-content">
            <div class="widget-container">
                <Widget widget={dashboardData.widget} />
            </div>
        </main>
    {/if}
</div>

<style>
    .dashboard {
        width: 100vw;
        height: 100vh;
        display: flex;
        flex-direction: column;
        background: white;
        color: black;
        font-family:
            -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    }

    .dashboard.dark {
        background: #1a1a1a;
        color: white;
    }

    .loading,
    .error {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        flex: 1;
    }

    .spinner {
        width: 40px;
        height: 40px;
        border: 4px solid #f3f3f3;
        border-top: 4px solid #3498db;
        border-radius: 50%;
        animation: spin 1s linear infinite;
        margin-bottom: 1rem;
    }

    @keyframes spin {
        0% {
            transform: rotate(0deg);
        }
        100% {
            transform: rotate(360deg);
        }
    }

    .dashboard-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 1rem 2rem;
        border-bottom: 1px solid #e5e5e5;
    }

    .dashboard.dark .dashboard-header {
        border-bottom-color: #333;
    }

    .status-indicators {
        display: flex;
        gap: 1rem;
    }

    .status-indicator {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #ef4444;
    }

    .status-indicator.connected .status-dot {
        background: #22c55e;
    }


    .dashboard-content {
        flex: 1;
        padding: 2rem;
        overflow: auto;
    }

    .widget-container {
        width: 100%;
        height: 100%;
        min-height: 600px;
    }


    @media (max-width: 768px) {
        .dashboard-header {
            padding: 1rem;
        }

        .dashboard-content {
            padding: 1rem;
        }

        .widget-container {
            min-height: 400px;
        }
    }
</style>
