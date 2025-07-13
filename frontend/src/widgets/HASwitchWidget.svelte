<script lang="ts">
    import type { WidgetType, WidgetDataUpdateEvent } from "../types";
    
    interface HAEntityData {
        entity_id: string;
        state: string;
        attributes: Record<string, unknown>;
        last_changed: string;
        last_updated: string;
    }
    import { DashboardService } from "../../bindings/github.com/mntndev/dash/pkg/dashboard";
    import { Events } from '@wailsio/runtime';
    import { onMount, onDestroy } from 'svelte';

    let { widget }: { widget: WidgetType } = $props();

    let isLoading = $state(false);

    let entityData = $state<HAEntityData | null>(null);

    onMount(() => {
        Events.On("widget_data_update", (event: WidgetDataUpdateEvent) => {
            if (event.data && event.data.length > 0) {
                const updateInfo = event.data[0];
                if (updateInfo.widget_id === widget.ID && updateInfo.data) {
                    entityData = updateInfo.data as HAEntityData;
                }
            }
        });
    });

    onDestroy(() => {
        Events.Off("widget_data_update");
    });
    let entityName = $derived(
        entityData?.entity_id?.split(".")[1]?.replace(/_/g, " ") || "Unknown"
    );
    let isOn = $derived(entityData?.state === "on");
    
    // Log any changes in entityData
    $effect(() => {
        if (entityData) {
            console.log(`[HA Switch] ${entityData.entity_id} data updated:`, entityData);
        }
    });

    async function toggleSwitch() {
        if (isLoading) return;

        isLoading = true;
        try {
            await DashboardService.TriggerWidget(widget.ID);
        } catch (error) {
            console.error("Failed to toggle switch:", error);
        } finally {
            isLoading = false;
        }
    }
</script>

<div 
    class="p-4 flex flex-col cursor-pointer hover:bg-gray-800 transition-colors duration-200 flex-none"
    class:opacity-75={isLoading}
    onclick={toggleSwitch}
    role="button"
    tabindex="0"
    onkeydown={(e) => e.key === 'Enter' && toggleSwitch()}
>
    <div class="flex items-center gap-3 mb-4">
        <div
            class="w-10 h-10 border-2 transition-all duration-300 flex items-center justify-center"
            class:bg-gray-200={isOn && !isLoading}
            class:border-gray-200={isOn && !isLoading}
            class:bg-transparent={!isOn || isLoading}
            class:border-gray-500={!isOn || isLoading}
        >
            {#if isLoading}
                <div class="w-4 h-4 border-2 border-transparent border-t-gray-500 animate-spin"></div>
            {/if}
        </div>
        <div class="flex-1">
            <h3
                class="text-lg md:text-xl font-semibold m-0 capitalize text-gray-200"
            >
                {entityName}
            </h3>
            <div class="text-xs text-gray-500 mt-1 font-mono">
                {entityData?.entity_id || "unknown"}
            </div>
        </div>
    </div>
    
</div>
