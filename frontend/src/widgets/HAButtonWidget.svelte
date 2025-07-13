<script lang="ts">
    import type { WidgetType, WidgetDataUpdateEvent } from "../types";
    
    interface HAButtonData {
        entity_id: string;
        service: string;
        domain: string;
        label: string;
    }
    import { DashboardService } from "../../bindings/github.com/mntndev/dash/pkg/dashboard";
    import { Events } from '@wailsio/runtime';
    import { onMount, onDestroy } from 'svelte';
    
    let { widget }: { widget: WidgetType } = $props();

    let isLoading = $state(false);
    let lastTriggered = $state("");

    let buttonData = $state<HAButtonData | null>(null);

    onMount(() => {
        Events.On("widget_data_update", (event: WidgetDataUpdateEvent) => {
            if (event.data && event.data.length > 0) {
                const updateInfo = event.data[0];
                if (updateInfo.widget_id === widget.ID && updateInfo.data) {
                    buttonData = updateInfo.data as HAButtonData;
                }
            }
        });
    });

    onDestroy(() => {
        Events.Off("widget_data_update");
    });
    let buttonLabel = $derived(buttonData?.label || "Button");

    async function triggerWidget() {
        if (isLoading) return;

        isLoading = true;
        try {
            await DashboardService.TriggerWidget(widget.ID);
            lastTriggered = new Date().toLocaleTimeString();
        } catch (error) {
            console.error("Failed to trigger widget:", error);
        } finally {
            isLoading = false;
        }
    }
</script>

<div class="p-4 flex flex-col flex-none">
    <div class="mb-4">
        <h3 class="text-lg md:text-xl font-semibold m-0 text-gray-200">{buttonLabel}</h3>
        <div class="text-xs text-gray-500 mt-1 font-mono">
            {buttonData?.domain}.{buttonData?.service}
        </div>
    </div>

    <div class="flex flex-1 items-center justify-center">
        <button
            class="w-full h-12 md:h-14 border border-gray-500 text-gray-200 font-medium cursor-pointer transition-all duration-200 flex items-center justify-center min-h-11 hover:bg-gray-800 disabled:opacity-60 disabled:cursor-not-allowed"
            onclick={triggerWidget}
            disabled={isLoading}
        >
            {#if isLoading}
                <div class="w-5 h-5 border-2 border-transparent border-t-gray-200 animate-spin"></div>
            {:else}
                {buttonLabel}
            {/if}
        </button>
    </div>

    {#if lastTriggered}
        <div class="text-xs text-gray-500 mt-2 text-center">
            Triggered: {lastTriggered}
        </div>
    {/if}
</div>
