export type BackofficeRole = 'teacher' | 'admin'

export type BackofficeModuleKey =
  | 'overview'
  | 'operations'
  | 'resources'
  | 'contestOps'
  | 'governance'

export interface BackofficeSecondaryItem {
  routeName: string
  label: string
  path: string
  roles: BackofficeRole[]
  isMatch: (path: string) => boolean
}

export interface BackofficeModule {
  key: BackofficeModuleKey
  label: string
  roles: BackofficeRole[]
  secondaryItems: BackofficeSecondaryItem[]
}

function matchPrefix(path: string, prefix: string): boolean {
  return path === prefix || path.startsWith(`${prefix}/`)
}

function matchAny(path: string, prefixes: string[]): boolean {
  return prefixes.some((prefix) => matchPrefix(path, prefix))
}

function matchExact(path: string, candidates: string[]): boolean {
  return candidates.includes(path)
}

const backofficeModules: BackofficeModule[] = [
  {
    key: 'overview',
    label: '总览',
    roles: ['teacher', 'admin'],
    secondaryItems: [
      {
        routeName: 'TeacherDashboard',
        label: '教学概览',
        path: '/academy/overview',
        roles: ['teacher'],
        isMatch: (path) => matchAny(path, ['/academy/overview']),
      },
      {
        routeName: 'AdminDashboard',
        label: '系统概览',
        path: '/admin/dashboard',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/dashboard', '/platform/overview']),
      },
    ],
  },
  {
    key: 'operations',
    label: '教学运营',
    roles: ['teacher', 'admin'],
    secondaryItems: [
      {
        routeName: 'ClassManagement',
        label: '班级管理',
        path: '/academy/classes',
        roles: ['teacher'],
        isMatch: (path) => matchExact(path, ['/academy/classes', '/teacher/classes']),
      },
      {
        routeName: 'AdminClassManagement',
        label: '班级管理',
        path: '/platform/classes',
        roles: ['admin'],
        isMatch: (path) => matchExact(path, ['/platform/classes', '/admin/classes']),
      },
      {
        routeName: 'TeacherStudentManagement',
        label: '学生管理',
        path: '/academy/students',
        roles: ['teacher', 'admin'],
        isMatch: (path) =>
          matchAny(path, [
            '/academy/students',
            '/teacher/students',
            '/academy/classes',
            '/teacher/classes',
          ]),
      },
      {
        routeName: 'TeacherAWDReviewIndex',
        label: 'AWD复盘',
        path: '/academy/awd-reviews',
        roles: ['teacher', 'admin'],
        isMatch: (path) => matchAny(path, ['/academy/awd-reviews', '/teacher/awd-reviews']),
      },
      {
        routeName: 'TeacherInstanceManagement',
        label: '实例管理',
        path: '/academy/instances',
        roles: ['teacher', 'admin'],
        isMatch: (path) => matchAny(path, ['/academy/instances', '/teacher/instances']),
      },
    ],
  },
  {
    key: 'resources',
    label: '题库与资源',
    roles: ['teacher', 'admin'],
    secondaryItems: [
      {
        routeName: 'ChallengeManage',
        label: '题目管理',
        path: '/platform/challenges',
        roles: ['teacher', 'admin'],
        isMatch: (path) =>
          matchAny(path, ['/platform/challenges', '/admin/challenges']) &&
          !matchAny(path, ['/platform/challenges/package-format', '/admin/challenges/package-format']),
      },
      {
        routeName: 'AdminEnvironmentTemplateLibrary',
        label: '环境模板',
        path: '/platform/environment-templates',
        roles: ['teacher', 'admin'],
        isMatch: (path) => matchAny(path, ['/platform/environment-templates', '/admin/environment-templates']),
      },
      {
        routeName: 'ImageManage',
        label: '镜像管理',
        path: '/platform/images',
        roles: ['teacher', 'admin'],
        isMatch: (path) => matchAny(path, ['/platform/images', '/admin/images']),
      },
    ],
  },
  {
    key: 'contestOps',
    label: '赛事运维',
    roles: ['admin'],
    secondaryItems: [
      {
        routeName: 'AdminContestOpsEnvironment',
        label: '环境管理',
        path: '/admin/contest-ops/environment',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/contest-ops/environment']),
      },
      {
        routeName: 'AdminContestOpsTraffic',
        label: '流量监控',
        path: '/admin/contest-ops/traffic',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/contest-ops/traffic']),
      },
      {
        routeName: 'AdminContestOpsProjector',
        label: '大屏投射',
        path: '/admin/contest-ops/projector',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/contest-ops/projector']),
      },
      {
        routeName: 'AdminContestOpsScoreboard',
        label: '排行榜',
        path: '/admin/contest-ops/scoreboard',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/contest-ops/scoreboard']),
      },
    ],
  },
  {
    key: 'governance',
    label: '系统治理',
    roles: ['admin'],
    secondaryItems: [
      {
        routeName: 'ContestManage',
        label: '竞赛管理',
        path: '/admin/contests',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/contests', '/platform/contests']),
      },
      {
        routeName: 'UserManage',
        label: '用户管理',
        path: '/admin/users',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/users', '/platform/users']),
      },
      {
        routeName: 'CheatDetection',
        label: '作弊检测',
        path: '/admin/integrity',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/integrity', '/platform/integrity']),
      },
      {
        routeName: 'AuditLog',
        label: '审计日志',
        path: '/admin/audit',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/admin/audit', '/platform/audit']),
      },
    ],
  },
]

export function isBackofficePath(path: string): boolean {
  return path.startsWith('/academy/') || path.startsWith('/platform/') || path.startsWith('/admin/')
}

export function getVisibleBackofficeModules(role?: string | null): BackofficeModule[] {
  if (role !== 'teacher' && role !== 'admin') {
    return []
  }

  return backofficeModules
    .filter((module) => module.roles.includes(role))
    .map((module) => ({
      ...module,
      secondaryItems: module.secondaryItems.filter((item) => item.roles.includes(role)),
    }))
}

export function getBackofficeModuleByPath(path: string): BackofficeModule | null {
  return backofficeModules.find((module) => module.secondaryItems.some((item) => item.isMatch(path))) ?? null
}

export function getBackofficeActiveSecondaryRouteName(path: string): string | null {
  const module = getBackofficeModuleByPath(path)
  const item = module?.secondaryItems.find((secondaryItem) => secondaryItem.isMatch(path))
  return item?.routeName ?? null
}

export function getVisibleBackofficeSecondaryItems(path: string, role?: string | null) {
  const module = getBackofficeModuleByPath(path)
  if (!module || (role !== 'teacher' && role !== 'admin')) {
    return []
  }

  return module.secondaryItems
    .filter((item) => item.roles.includes(role))
    .map((item) => ({
      ...item,
      active: item.isMatch(path),
    }))
}

export { backofficeModules }
