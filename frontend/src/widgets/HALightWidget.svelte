<script lang="ts">
    import type { WidgetType, HAEntityData } from "../types";
    import { DashboardService } from "../../bindings/github.com/mntndev/dash/pkg/dashboard";
    import { Events } from '@wailsio/runtime';
    import { onMount, onDestroy } from 'svelte';

    let { widget }: { widget: WidgetType } = $props();

    let isLoading = $state(false);
    let localBrightnessPercent = $state(0);
    let userInteracting = $state(false);
    let initialized = $state(false);

    // Derived values using runes
    let entityData = $state<HAEntityData | null>(null);

    onMount(() => {
        Events.On("widget_data_update", (event: any) => {
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
        entityData?.entity_id?.split(".")[1]?.replace(/_/g, " ") || "Unknown",
    );
    let isOn = $derived(entityData?.state === "on");
    let brightness = $derived(entityData?.attributes?.brightness || 0);
    let maxBrightness = $derived(255); // Home Assistant brightness range is 0-255
    let brightnessPercent = $derived(
        Math.round((brightness / maxBrightness) * 100),
    );
    
    // Log any changes in entityData
    $effect(() => {
        if (entityData) {
            console.log(`[HA Light] ${entityData.entity_id} data updated:`, entityData);
        }
    });

    // Effect to initialize and update local brightness
    $effect(() => {
        if (!initialized && brightnessPercent > 0) {
            localBrightnessPercent = brightnessPercent;
            initialized = true;
            console.log("Initialized brightness to:", localBrightnessPercent);
        } else if (initialized && !userInteracting && !isLoading) {
            // Only update if there's a significant difference to avoid micro-updates
            const diff = Math.abs(brightnessPercent - localBrightnessPercent);
            if (diff > 5) {
                // Increased threshold to 5%
                console.log(
                    "Updating brightness from",
                    localBrightnessPercent,
                    "to",
                    brightnessPercent,
                );
                localBrightnessPercent = brightnessPercent;
            }
        } else if (userInteracting) {
            console.log(
                "Blocking brightness update because user is interacting",
            );
        }
    });

    async function toggleLight() {
        if (isLoading) return;

        isLoading = true;
        try {
            await DashboardService.TriggerWidget(widget.ID);
        } catch (error) {
            console.error("Failed to toggle light:", error);
        } finally {
            isLoading = false;
        }
    }

    function onBrightnessStart() {
        console.log("User started interacting with brightness slider");
        userInteracting = true;
    }

    async function onBrightnessEnd() {
        console.log("User finished interacting with brightness slider");

        // Send the brightness value when user releases the slider
        await setBrightness(localBrightnessPercent);

        // Allow server updates again after a short delay
        setTimeout(() => {
            userInteracting = false;
            console.log("Allowing server updates again");
        }, 500);
    }

    function onBrightnessInput(event: Event) {
        if (!isOn) return;

        const target = event.target as HTMLInputElement;
        localBrightnessPercent = parseInt(target.value);
        // No API call here - just update the UI immediately
    }

    async function setBrightness(brightnessPercent: number) {
        if (isLoading || !isOn) return;

        const newBrightness = Math.round(
            (brightnessPercent / 100) * maxBrightness,
        );

        isLoading = true;
        try {
            await DashboardService.SetLightBrightness(
                widget.ID,
                newBrightness,
            );
        } catch (error) {
            console.error("Failed to set brightness:", error);
            // Revert to server value on error
            localBrightnessPercent = Math.round(
                (brightness / maxBrightness) * 100,
            );
        } finally {
            isLoading = false;
        }
    }
</script>

<div
    class="p-4 flex flex-col transition-colors duration-200 flex-none"
    class:opacity-75={isLoading}
>
    <!-- Light Name -->
    <div
        class="mb-4 cursor-pointer hover:bg-gray-800 transition-colors duration-200 p-2"
        on:click={toggleLight}
        role="button"
        tabindex="0"
        on:keydown={(e) => e.key === "Enter" && toggleLight()}
    ></div>

    <!-- Icon and Brightness Column -->
    <div class="flex items-start gap-3">
        <div
            class="h-full mr-2 min-w-12 aspect-square border-2 transition-all duration-300 flex items-center justify-center cursor-pointer"
            class:bg-gray-200={isOn && !isLoading}
            class:border-gray-200={isOn && !isLoading}
            class:bg-transparent={!isOn || isLoading}
            class:border-gray-500={!isOn || isLoading}
            on:click={toggleLight}
            role="button"
            tabindex="0"
            on:keydown={(e) => e.key === "Enter" && toggleLight()}
        >
            {#if isLoading}
                <div
                    class="w-4 h-4 border-2 border-transparent border-t-gray-500 animate-spin"
                ></div>
            {:else if isOn}
                <!-- Light bulb icon -->
                <svg
                    class="w-6 h-6 text-gray-800"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                >
                    <path
                        d="M11 3a1 1 0 10-2 0v1a1 1 0 102 0V3zM15.657 5.757a1 1 0 00-1.414-1.414l-.707.707a1 1 0 001.414 1.414l.707-.707zM18 10a1 1 0 01-1 1h-1a1 1 0 110-2h1a1 1 0 011 1zM5.05 6.464A1 1 0 106.464 5.05l-.707-.707a1 1 0 00-1.414 1.414l.707.707zM5 10a1 1 0 01-1 1H3a1 1 0 110-2h1a1 1 0 011 1zM8 16v-1h4v1a2 2 0 11-4 0zM12 14c.015-.34.208-.646.477-.859a4 4 0 10-4.954 0c.27.213.462.519.477.859h4z"
                    />
                </svg>
            {/if}
        </div>
        <div class="flex-1 flex flex-col gap-2">
            {#if isOn && localBrightnessPercent > 0}
                <div class="text-lg text-gray-200 font-semibold">
                    {entityName} -
                    <span class="text-gray-500">{localBrightnessPercent}%</span>
                </div>
                <input
                    id={`brightness-${widget.ID}`}
                    type="range"
                    min="1"
                    max="100"
                    value={localBrightnessPercent}
                    on:mousedown={onBrightnessStart}
                    on:mouseup={onBrightnessEnd}
                    on:touchstart={onBrightnessStart}
                    on:touchend={onBrightnessEnd}
                    on:input={onBrightnessInput}
                    class="w-full h-5 bg-gray-800 appearance-none cursor-pointer slider"
                    disabled={isLoading}
                />
            {:else}
                <div class="text-lg text-gray-500">{entityName} - Off</div>
            {/if}
        </div>
    </div>
</div>

<style>
    .slider::-webkit-slider-thumb {
        appearance: none;
        height: 20px;
        width: 20px;
        background: #e5e7eb;
        cursor: pointer;
    }

    .slider::-moz-range-thumb {
        height: 20px;
        width: 20px;
        background: #e5e7eb;
        cursor: pointer;
        border: none;
    }
</style>
