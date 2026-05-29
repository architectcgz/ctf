import type { AWDTrafficSummaryData } from '@/api/contracts'

export interface AWDTrafficServiceOptionView {
  serviceId: string
  title: string
}

export interface AWDTrafficSummaryStatView {
  key: string
  label: string
  value: string
  hint: string
}

export interface AWDTrafficTrendRowView {
  bucket_start_at: string
  request_count: number
  error_count: number
  ratio: number
  label: string
}

export interface AWDTrafficIntelligenceGridProps {
  summary: AWDTrafficSummaryData
  trendRows: AWDTrafficTrendRowView[]
  trendNarrative: string
}
