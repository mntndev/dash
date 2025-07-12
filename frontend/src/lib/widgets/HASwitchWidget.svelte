<script lang="ts">
    import type { WidgetData, HAEntityData } from "../types";
    import { DashboardService } from "../../../bindings/github.com/mntndev/dash/pkg/dashboard";

    let { widget }: { widget: WidgetData } = $props();

    let isLoading = false;

    let entityData = $derived(widget.data as HAEntityData);
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
            await DashboardService.TriggerWidget(widget.id);
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
    on:click={toggleSwitch}
    role="button"
    tabindex="0"
    on:keydown={(e) => e.key === 'Enter' && toggleSwitch()}
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
