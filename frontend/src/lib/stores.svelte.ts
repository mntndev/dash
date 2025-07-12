import type { DashboardData, WidgetData } from './types';

// Create reactive state objects that can be exported and mutated
export const appState = $state({
    dashboardData: null as DashboardData | null,
    loading: true,
    error: "",
    widgetDataMap: new Map<string, WidgetData>()
});

// Helper to get widget data for a specific widget
export function getWidgetData(widgetId: string): WidgetData | undefined {
    return appState.widgetDataMap.get(widgetId);
}

// Helper to update a single widget's data
export function updateWidgetData(widgetId: string, data: any) {
    const widget = appState.widgetDataMap.get(widgetId);
    if (widget) {
        appState.widgetDataMap.set(widgetId, { ...widget, data, last_update: new Date().toISOString() });
    }
}

// Helper to build widget map from dashboard data
export function buildWidgetMap(widget: WidgetData, map = new Map<string, WidgetData>()): Map<string, WidgetData> {
    map.set(widget.id, widget);
    
    if (widget.children) {
        widget.children.forEach(child => buildWidgetMap(child, map));
    }
    
    return map;
}

// Helper to set dashboard data and update widget map
export function setDashboardData(data: DashboardData | null) {
    appState.dashboardData = data;
    if (data?.widget) {
        const map = buildWidgetMap(data.widget);
        appState.widgetDataMap = map;
    }
}