<script lang="ts">
    import type { WidgetType } from "../types";
    import { useHAEntity } from "../composables/useHAEntity.svelte";

    let { widget }: { widget: WidgetType } = $props();

    const { entity, triggerWidget } = useHAEntity(widget);
    
    let entityName = $derived(entity.data?.entity_id?.split('.')[1]?.replace(/_/g, ' ') || 'Unknown');
    let isOn = $derived(entity.data?.state === "on");
    
    // Log any changes in entityData
    $effect(() => {
        if (entity.data) {
            console.log(`[HA Switch] ${entity.data.entity_id} data updated:`, entity.data);
        }
    });

    const toggleSwitch = triggerWidget;
</script>

<div 
    class="p-4 flex flex-col cursor-pointer hover:bg-gray-800 transition-colors duration-200 flex-none"
    class:opacity-75={entity.isLoading}
    onclick={toggleSwitch}
    role="button"
    tabindex="0"
    onkeydown={(e) => e.key === 'Enter' && toggleSwitch()}
>
    <div class="flex items-center gap-3 mb-4">
        <div
            class="w-10 h-10 border-2 transition-all duration-300 flex items-center justify-center"
            class:bg-gray-200={isOn && !entity.isLoading}
            class:border-gray-200={isOn && !entity.isLoading}
            class:bg-transparent={!isOn || entity.isLoading}
            class:border-gray-500={!isOn || entity.isLoading}
        >
            {#if entity.isLoading}
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
                {entity.data?.entity_id || "unknown"}
            </div>
        </div>
    </div>
    
</div>
