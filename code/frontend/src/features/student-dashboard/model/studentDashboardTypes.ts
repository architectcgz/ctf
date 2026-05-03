import type { Component } from 'vue'

export type DashboardPanelKey =
  | 'overview'
  | 'category'
  | 'recommendation'
  | 'timeline'
  | 'difficulty'

export interface DashboardHighlightItem {
  label: string
  value: string
  description: string
  icon: Component
}

export interface DashboardPanelTab {
  key: DashboardPanelKey
  label: string
  panelId: string
  tabId: string
}
