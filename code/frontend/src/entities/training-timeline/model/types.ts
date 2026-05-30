export interface TimelineEvent {
  id: string
  type: string
  title: string
  created_at: string
  detail?: string
  points?: number
  meta?: Record<string, unknown>
}
