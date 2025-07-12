<script lang="ts">
    import type { WidgetData, HAButtonData } from "../types";
    import { DashboardAppService } from "../../../bindings/github.com/mntndev/dash";
    let { widget }: { widget: WidgetData } = $props();

    let isLoading = false;
    let lastTriggered = "";

    let buttonData = $derived(widget.data as HAButtonData);
    let buttonLabel = $derived(buttonData?.label || "Button");

    async function triggerWidget() {
        if (isLoading) return;

        isLoading = true;
        try {
            await DashboardAppService.TriggerWidget(widget.id);
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
        <h3 class="text-lg md:text-xl font-semibold m-0">{buttonLabel}</h3>
        <div class="text-xs text-gray-500 mt-1 font-mono">
            {buttonData?.domain}.{buttonData?.service}
        </div>
    </div>

    <div class="flex flex-1 items-center justify-center">
        <button
            class="w-full h-12 md:h-14 border-none text-white font-medium cursor-pointer transition-all duration-200 flex items-center justify-center min-h-11 hover:transform hover:-translate-y-px active:transform active:translate-y-0 disabled:opacity-60 disabled:cursor-not-allowed disabled:transform-none"
            class:bg-blue-600={!isLoading}
            class:hover:bg-blue-700={!isLoading}
            class:bg-gray-500={isLoading}
            on:click={triggerWidget}
            disabled={isLoading}
        >
            {#if isLoading}
                <div class="w-5 h-5 border-2 border-transparent border-t-white rounded-full animate-spin"></div>
            {:else}
                {buttonLabel}
            {/if}
        </button>
    </div>

    {#if lastTriggered}
        <div class="text-xs text-gray-400 mt-2 text-center">
            Triggered: {lastTriggered}
        </div>
    {/if}
</div>
