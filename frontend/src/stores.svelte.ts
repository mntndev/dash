import type { DashboardInfo } from './types';

// Create reactive state objects that can be exported and mutated
export const appState = $state({
    dashboardInfo: null as DashboardInfo | null,
    loading: true,
    error: ""
});

// Helper to set dashboard info
export function setDashboardInfo(info: DashboardInfo | null) {
    appState.dashboardInfo = info;
}