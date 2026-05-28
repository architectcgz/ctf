export type AWDOperationsPanelKey = 'inspector' | 'instances'

export interface AWDOperationsTabItem {
  key: AWDOperationsPanelKey
  label: string
  tabId: string
  panelId: string
}
