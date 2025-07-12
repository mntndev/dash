// Dashboard structure (static information)
export interface DashboardInfo {
  title: string;
  theme: string;
  root_widget: Widget;
  status: Record<string, boolean>;
}

// Widget interface that matches the actual JSON serialization from Go  
export interface Widget {
  ID: string;
  Type: string;
  Data: any;
  LastUpdate: string;
  Children?: Widget[];
}

// Widget tree node (structure only, no data) - DEPRECATED
export interface WidgetTreeNode {
  id: string;
  type: string;
  children?: WidgetTreeNode[];
}

// Widget data (dynamic content) - DEPRECATED, use Widget instead
export interface WidgetData {
  id: string;
  type: string;
  data: any;
  last_update: string;
}

// Legacy type for backward compatibility (will be removed)
export interface DashboardData {
  title: string;
  theme: string;
  widget: WidgetData & { children?: WidgetData[] };
  status: Record<string, boolean>;
}


export interface HAEntityData {
  entity_id: string;
  state: string;
  attributes: Record<string, any>;
  last_changed: string;
  last_updated: string;
}

export interface HAButtonData {
  entity_id: string;
  service: string;
  domain: string;
  label: string;
}

export interface ClockData {
  time: string;
  format: string;
  display: string;
}

export interface DexcomData {
  value: number;
  trend: string;
  timestamp: string;
  unit: string;
  historical?: DexcomReading[];
  low_threshold: number;
  high_threshold: number;
}

export interface DexcomReading {
  value: number;
  trend: string;
  timestamp: string;
}