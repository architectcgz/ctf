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

function matchContestOperationsManagePath(path: string): boolean {
  return /^\/platform\/contests\/[^/]+\/(?:manage|awd-config)(?:\/.*)?$/.test(path)
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
        routeName: 'PlatformOverview',
        label: '系统概览',
        path: '/platform/overview',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/platform/overview']),
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
        isMatch: (path) => matchExact(path, ['/academy/classes']),
      },
      {
        routeName: 'PlatformClassManagement',
        label: '班级管理',
        path: '/platform/classes',
        roles: ['admin'],
        isMatch: (path) => matchExact(path, ['/platform/classes']),
      },
      {
        routeName: 'TeacherStudentManagement',
        label: '学生管理',
        path: '/academy/students',
        roles: ['teacher'],
        isMatch: (path) =>
          matchAny(path, ['/academy/students', '/academy/classes']),
      },
      {
        routeName: 'PlatformStudentManagement',
        label: '学生管理',
        path: '/platform/students',
        roles: ['admin'],
        isMatch: (path) =>
          matchAny(path, ['/platform/students', '/platform/classes']),
      },
      {
        routeName: 'TeacherAWDReviewIndex',
        label: 'AWD复盘',
        path: '/academy/awd-reviews',
        roles: ['teacher'],
        isMatch: (path) => matchAny(path, ['/academy/awd-reviews']),
      },
      {
        routeName: 'PlatformAwdReviewIndex',
        label: 'AWD复盘',
        path: '/platform/awd-reviews',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/platform/awd-reviews']),
      },
      {
        routeName: 'TeacherInstanceManagement',
        label: '实例管理',
        path: '/academy/instances',
        roles: ['teacher'],
        isMatch: (path) => matchAny(path, ['/academy/instances']),
      },
      {
        routeName: 'PlatformInstanceManagement',
        label: '实例管理',
        path: '/platform/instances',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/platform/instances']),
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
          matchAny(path, ['/platform/challenges']) &&
          !matchAny(path, ['/platform/challenges/package-format']),
      },
      {
        routeName: 'PlatformAwdChallengeLibrary',
        label: 'AWD题库',
        path: '/platform/awd-challenges',
        roles: ['teacher', 'admin'],
        isMatch: (path) => matchAny(path, ['/platform/awd-challenges']),
      },
      {
        routeName: 'ImageManage',
        label: '镜像管理',
        path: '/platform/images',
        roles: ['teacher', 'admin'],
        isMatch: (path) => matchAny(path, ['/platform/images']),
      },
    ],
  },
  {
    key: 'contestOps',
    label: '赛事运维',
    roles: ['admin'],
    secondaryItems: [
      {
        routeName: 'PlatformContestOpsIndex',
        label: '竞赛列表',
        path: '/platform/contest-ops/contests',
        roles: ['admin'],
        isMatch: (path) =>
          matchAny(path, [
            '/platform/contest-ops/contests',
            '/platform/contest-ops/environment',
            '/platform/contest-ops/traffic',
            '/platform/contest-ops/scoreboard',
          ]) || matchContestOperationsManagePath(path),
      },
      {
        routeName: 'PlatformContestProjector',
        label: '大屏展示',
        path: '/platform/contest-ops/projector',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/platform/contest-ops/projector']),
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
        label: '竞赛目录',
        path: '/platform/contests',
        roles: ['admin'],
        isMatch: (path) =>
          matchAny(path, ['/platform/contests']) && !matchContestOperationsManagePath(path),
      },
      {
        routeName: 'UserManage',
        label: '用户管理',
        path: '/platform/users',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/platform/users']),
      },
      {
        routeName: 'CheatDetection',
        label: '作弊检测',
        path: '/platform/integrity',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/platform/integrity']),
      },
      {
        routeName: 'AuditLog',
        label: '审计日志',
        path: '/platform/audit',
        roles: ['admin'],
        isMatch: (path) => matchAny(path, ['/platform/audit']),
      },
    ],
  },
]

export function isBackofficePath(path: string): boolean {
  return path.startsWith('/academy/') || path.startsWith('/platform/')
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
