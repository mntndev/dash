import type { DashboardInfo, DashboardData, WidgetData, WidgetTreeNode } from './types';

// Create reactive state objects that can be exported and mutated
export const appState = $state({
    dashboardInfo: null as DashboardInfo | null,
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
        console.log(`[Widget Update] ${widgetId} (${widget.type}):`, data);
        appState.widgetDataMap.set(widgetId, { ...widget, data, last_update: new Date().toISOString() });
    }
}

// Helper to set dashboard info
export function setDashboardInfo(info: DashboardInfo | null) {
    appState.dashboardInfo = info;
}

// Helper to update a specific widget's data and log the change
export function setWidgetData(widgetData: WidgetData) {
    const existingWidget = appState.widgetDataMap.get(widgetData.id);
    if (existingWidget && JSON.stringify(existingWidget.data) !== JSON.stringify(widgetData.data)) {
        console.log(`[Widget Data Change] ${widgetData.id} (${widgetData.type}):`, {
            from: existingWidget.data,
            to: widgetData.data
        });
    }
    
    console.log(`[Widget Update] ${widgetData.id} (${widgetData.type}):`, widgetData.data);
    appState.widgetDataMap.set(widgetData.id, widgetData);
}

// Helper to build widget IDs from tree (for initialization)
export function getAllWidgetIds(node: WidgetTreeNode, ids: string[] = []): string[] {
    ids.push(node.id);
    if (node.children) {
        node.children.forEach(child => getAllWidgetIds(child, ids));
    }
    return ids;
}

// Legacy helper for backward compatibility (will be removed)
export function buildWidgetMap(widget: WidgetData & { children?: WidgetData[] }, map = new Map<string, WidgetData>()): Map<string, WidgetData> {
    const existingWidget = map.get(widget.id);
    if (existingWidget && JSON.stringify(existingWidget.data) !== JSON.stringify(widget.data)) {
        console.log(`[Widget Data Change] ${widget.id} (${widget.type}):`, {
            from: existingWidget.data,
            to: widget.data
        });
    }
    
    map.set(widget.id, { 
        id: widget.id, 
        type: widget.type, 
        data: widget.data, 
        last_update: widget.last_update 
    });
    
    if (widget.children) {
        widget.children.forEach(child => buildWidgetMap(child, map));
    }
    
    return map;
}

// Legacy helper for backward compatibility (will be removed)
export function setDashboardData(data: DashboardData | null) {
    if (data?.widget) {
        const map = buildWidgetMap(data.widget, appState.widgetDataMap);
        appState.widgetDataMap = map;
    }
}