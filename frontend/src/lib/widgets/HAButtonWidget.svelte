<script lang="ts">
    import type { WidgetData, HAButtonData } from "../types";
    import { DashboardAppService } from "../../../bindings/github.com/mntndev/dash";
    export let widget: WidgetData;

    let isLoading = false;
    let lastTriggered = "";

    $: buttonData = widget.data as HAButtonData;
    $: buttonLabel = buttonData?.label || "Button";

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

<div class="ha-button-widget">
    <div class="button-header">
        <h3 class="button-name">{buttonLabel}</h3>
        <div class="button-service">
            {buttonData?.domain}.{buttonData?.service}
        </div>
    </div>

    <div class="button-container">
        <button
            class="trigger-button"
            class:loading={isLoading}
            on:click={triggerWidget}
            disabled={isLoading}
        >
            {#if isLoading}
                <div class="spinner"></div>
            {:else}
                {buttonLabel}
            {/if}
        </button>
    </div>

    {#if lastTriggered}
        <div class="last-triggered">
            Triggered: {lastTriggered}
        </div>
    {/if}
</div>

<style>
    .ha-button-widget {
        display: flex;
        flex-direction: column;
        height: 100%;
    }

    .button-header {
        margin-bottom: 1rem;
    }

    .button-name {
        font-size: 1.1rem;
        font-weight: 600;
        margin: 0;
    }

    .button-service {
        font-size: 0.75rem;
        color: #6b7280;
        margin-top: 0.25rem;
        font-family: monospace;
    }

    .button-container {
        display: flex;
        flex: 1;
        align-items: center;
        justify-content: center;
    }

    .trigger-button {
        width: 100%;
        height: 48px;
        border: none;
        border-radius: 0.5rem;
        background: #3b82f6;
        color: white;
        font-size: 1rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 44px; /* Touch-friendly minimum */
    }

    .trigger-button:hover:not(:disabled) {
        background: #2563eb;
        transform: translateY(-1px);
    }

    .trigger-button:active:not(:disabled) {
        transform: translateY(0);
    }

    .trigger-button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
        transform: none;
    }

    .trigger-button.loading {
        background: #6b7280;
    }

    .spinner {
        width: 20px;
        height: 20px;
        border: 2px solid transparent;
        border-top: 2px solid white;
        border-radius: 50%;
        animation: spin 1s linear infinite;
    }

    @keyframes spin {
        0% {
            transform: rotate(0deg);
        }
        100% {
            transform: rotate(360deg);
        }
    }

    .last-triggered {
        font-size: 0.75rem;
        color: #9ca3af;
        margin-top: 0.5rem;
        text-align: center;
    }

    :global(.dashboard.dark) .button-service {
        color: #9ca3af;
    }

    :global(.dashboard.dark) .last-triggered {
        color: #6b7280;
    }

    @media (max-width: 768px) {
        .button-name {
            font-size: 1rem;
        }

        .trigger-button {
            height: 44px;
            font-size: 0.875rem;
        }
    }
</style>
