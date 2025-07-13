// Dashboard structure (static information)
export interface DashboardInfo {
  title: string;
  theme: string;
  root_widget: WidgetType;
  status: Record<string, boolean>;
}

// Widget interface that matches the actual JSON serialization from Go  
export interface WidgetType {
  ID: string;
  Type: string;
  LastUpdate: string;
  Children?: WidgetType[];
}

// Event interfaces for Wails runtime events
export interface DashboardUpdateEvent {
  data: DashboardInfo[];
}

export interface WidgetUpdateInfo {
  widget_id: string;
  data: unknown;
}

export interface WidgetDataUpdateEvent {
  data: WidgetUpdateInfo[];
}

