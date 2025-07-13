import type { WidgetType, WidgetDataUpdateEvent } from '../types';
import { Events } from '@wailsio/runtime';
import { onMount, onDestroy } from 'svelte';
import { DashboardService } from '../../bindings/github.com/mntndev/dash/pkg/dashboard';

export interface HAEntityData {
  entity_id: string;
  state: string;
  attributes: Record<string, unknown>;
  last_changed: string;
  last_updated: string;
}

export function useHAEntity(widget: WidgetType) {
  const entity = $state({
    data: null as HAEntityData | null,
    isLoading: false
  });

  onMount(() => {
    Events.On("widget_data_update", (event: WidgetDataUpdateEvent) => {
      if (event.data && event.data.length > 0) {
        const updateInfo = event.data[0];
        if (updateInfo.widget_id === widget.ID && updateInfo.data) {
          entity.data = updateInfo.data as HAEntityData;
        }
      }
    });
  });

  onDestroy(() => {
    Events.Off("widget_data_update");
  });

  // Common actions
  async function triggerWidget() {
    if (entity.isLoading) return;

    entity.isLoading = true;
    try {
      await DashboardService.TriggerWidget(widget.ID);
    } catch (error) {
      console.error("Failed to trigger widget:", error);
    } finally {
      entity.isLoading = false;
    }
  }

  return {
    entity,
    triggerWidget
  };
}
