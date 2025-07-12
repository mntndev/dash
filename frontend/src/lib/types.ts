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