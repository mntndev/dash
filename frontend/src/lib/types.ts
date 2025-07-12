export interface DashboardData {
  title: string;
  theme: string;
  widget: WidgetData;
  status: Record<string, boolean>;
}

export interface WidgetData {
  id: string;
  type: string;
  data: any;
  last_update: string;
  children?: WidgetData[];
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
}