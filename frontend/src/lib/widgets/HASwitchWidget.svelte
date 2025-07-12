<script lang="ts">
    import type { WidgetData, HAEntityData } from "../types";
    import { DashboardAppService } from "../../../bindings/github.com/mntndev/dash";

    export let widget: WidgetData;

    let isLoading = false;

    $: entityData = widget.data as HAEntityData;
    $: entityName =
        entityData?.entity_id?.split(".")[1]?.replace(/_/g, " ") || "Unknown";
    $: isOn = entityData?.state === "on";

    async function toggleSwitch() {
        if (isLoading) return;

        isLoading = true;
        try {
            await DashboardAppService.TriggerWidget(widget.id);
        } catch (error) {
            console.error("Failed to toggle switch:", error);
        } finally {
            isLoading = false;
        }
    }
</script>

<div 
    class="p-4 flex flex-col cursor-pointer hover:bg-gray-50 hover:bg-opacity-5 transition-colors duration-200 flex-none"
    class:opacity-75={isLoading}
    on:click={toggleSwitch}
    role="button"
    tabindex="0"
    on:keydown={(e) => e.key === 'Enter' && toggleSwitch()}
>
    <div class="flex items-center gap-3 mb-4">
        <div
            class="w-10 h-10 border-2 transition-all duration-300 flex items-center justify-center"
            class:bg-green-500={isOn && !isLoading}
            class:border-green-500={isOn && !isLoading}
            class:bg-transparent={!isOn || isLoading}
            class:border-gray-400={!isOn || isLoading}
        >
            {#if isLoading}
                <div class="w-4 h-4 border-2 border-transparent border-t-gray-400 rounded-full animate-spin"></div>
            {/if}
        </div>
        <div class="flex-1">
            <h3
                class="text-lg md:text-xl font-semibold m-0 capitalize text-white"
            >
                {entityName}
            </h3>
            <div class="text-xs text-gray-500 mt-1 font-mono">
                {entityData?.entity_id || "unknown"}
            </div>
        </div>
    </div>
    
</div>
