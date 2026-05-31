export type NotificationFilter = 'all' | 'unread' | 'read'

export interface NotificationFilterOption {
  value: NotificationFilter
  label: string
}
