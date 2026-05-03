export type AwdReviewDirectoryColumnKey =
  | 'code'
  | 'name'
  | 'rounds'
  | 'teams'
  | 'status'
  | 'action'

export interface AwdReviewDirectoryColumnSchema {
  key: AwdReviewDirectoryColumnKey
  label: string
  headClass?: string
  rowClass: string
}

export const AWD_REVIEW_DIRECTORY_COLUMN_SCHEMA: readonly AwdReviewDirectoryColumnSchema[] = [
  {
    key: 'code',
    label: '代号',
    headClass: 'teacher-directory-head-cell-code',
    rowClass: 'teacher-directory-cell teacher-directory-cell-code',
  },
  {
    key: 'name',
    label: '赛事',
    headClass: 'teacher-directory-head-cell-name',
    rowClass: 'teacher-directory-cell teacher-directory-cell-name',
  },
  {
    key: 'rounds',
    label: '轮次',
    rowClass: 'teacher-directory-row-metrics',
  },
  {
    key: 'teams',
    label: '队伍',
    rowClass: 'teacher-directory-row-metrics',
  },
  {
    key: 'status',
    label: '状态',
    rowClass: 'teacher-directory-row-tags',
  },
  {
    key: 'action',
    label: '操作',
    rowClass: 'teacher-directory-row-action',
  },
] as const

export const AWD_REVIEW_DIRECTORY_HEADERS = AWD_REVIEW_DIRECTORY_COLUMN_SCHEMA.map(
  (column) => column.label
) as readonly string[]

export const AWD_REVIEW_DIRECTORY_COLUMNS =
  'minmax(0, 7rem) minmax(0, 2.1fr) minmax(0, 1fr) minmax(0, 0.85fr) minmax(0, 1fr) auto'
