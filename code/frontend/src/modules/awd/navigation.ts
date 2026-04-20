import type {
  AdminAwdPageKey,
  AwdPageDefinition,
  StudentAwdPageKey,
  TeacherAwdPageKey,
} from './types'

export const STUDENT_AWD_PAGES: AwdPageDefinition<StudentAwdPageKey>[] = [
  { key: 'overview', label: '战场总览', description: '查看当前轮次、分数变化与关键事件。' },
  { key: 'services', label: '我的服务', description: '查看本队服务状态、flag 与恢复动作。' },
  { key: 'targets', label: '目标目录', description: '浏览目标队伍、服务与攻击面。' },
  { key: 'attacks', label: '攻击记录', description: '查看攻击提交结果与近期命中情况。' },
  { key: 'collab', label: '队伍协作', description: '查看分工、协作提醒与战术摘要。' },
]

export const ADMIN_AWD_PAGES: AwdPageDefinition<AdminAwdPageKey>[] = [
  { key: 'overview', label: 'AWD 总览', description: '汇总赛事健康度、轮次状态与核心指标。' },
  { key: 'readiness', label: 'Readiness', description: '查看赛前检查结果与阻塞项。' },
  { key: 'rounds', label: '轮次控制台', description: '查看轮次节奏、调度和干预动作。' },
  { key: 'services', label: '服务矩阵', description: '查看队伍与服务维度的健康状态。' },
  { key: 'attacks', label: '攻击日志', description: '查看攻击流水与命中结果。' },
  { key: 'traffic', label: '流量态势', description: '查看请求态势、异常路径与热点来源。' },
  { key: 'alerts', label: '告警中心', description: '查看高风险异常与待处理告警。' },
  { key: 'instances', label: '实例运维', description: '查看实例状态、恢复动作与运维摘要。' },
  { key: 'replay', label: '赛后复盘与导出', description: '查看赛后回放摘要与导出入口。' },
]

export const TEACHER_AWD_PAGES: AwdPageDefinition<TeacherAwdPageKey>[] = [
  { key: 'overview', label: '教学总览', description: '查看整场复盘摘要与轮次线索。' },
  { key: 'teams', label: '队伍复盘', description: '查看队伍表现、案例与下钻入口。' },
  { key: 'services', label: 'Service 复盘', description: '按服务聚合教学案例与共性问题。' },
  { key: 'replay', label: '轮次回放', description: '按时间线回看轮次关键证据。' },
  { key: 'export', label: '报告导出', description: '导出归档与教师报告。' },
]

export function buildStudentAwdPath(contestId: string, page: StudentAwdPageKey): string {
  return `/contests/${encodeURIComponent(contestId)}/awd/${page}`
}

export function buildAdminAwdPath(contestId: string, page: AdminAwdPageKey): string {
  return `/platform/contests/${encodeURIComponent(contestId)}/awd/${page}`
}

export function buildTeacherAwdPath(contestId: string, page: TeacherAwdPageKey): string {
  return `/academy/awd-reviews/${encodeURIComponent(contestId)}/${page}`
}
