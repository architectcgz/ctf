export type StudentAwdPageKey = 'overview' | 'services' | 'targets' | 'attacks' | 'collab'

export type AdminAwdPageKey =
  | 'overview'
  | 'readiness'
  | 'rounds'
  | 'services'
  | 'attacks'
  | 'traffic'
  | 'alerts'
  | 'instances'
  | 'replay'

export type TeacherAwdPageKey = 'overview' | 'teams' | 'services' | 'replay' | 'export'

export type AwdPageKey = StudentAwdPageKey | AdminAwdPageKey | TeacherAwdPageKey

export interface AwdPageDefinition<TPageKey extends string> {
  key: TPageKey
  label: string
  description: string
}

export interface AwdHeroMetric {
  label: string
  value: string
  helper?: string
}

export interface AwdTimelineItem {
  id: string
  time: string
  title: string
  description: string
}
