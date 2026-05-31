import { computed, type ComputedRef } from 'vue'

import {
  type BackofficeModule,
  getBackofficeModuleByPath,
  getVisibleBackofficeModules,
  getVisibleBackofficeSecondaryItems,
  isBackofficePath,
} from '@/config/backofficeNavigation'
import { resolveRouteTitle } from '@/utils/routeTitle'

export type WorkspaceRole = 'student' | 'teacher' | 'admin' | string | null | undefined

export interface WorkspaceShellNavigationInput {
  path: string
  fullPath?: string
  role?: WorkspaceRole
  routeName?: string
  pageTitle?: string
  detailLabel?: string | null
}

export interface WorkspaceShellSecondaryItem {
  routeName: string
  label: string
  path: string
  roles: string[]
  isMatch: (path: string) => boolean
}

export interface WorkspaceShellModule {
  key: string
  label: string
  roles: string[]
  secondaryItems: WorkspaceShellSecondaryItem[]
}

export interface WorkspaceShellBreadcrumb {
  workspacePath: string
  moduleLabel: string
  modulePath: string
  secondaryLabel: string
  secondaryPath: string
  detailLabel: string | null
  detailPath: string
}

export interface WorkspaceShellNavigation {
  isBackoffice: boolean
  brandMark: string
  brandKicker: string
  brandTitle: string
  brandTooltip: string
  roleBadge: string
  roleCaption: string
  modules: WorkspaceShellModule[]
  activeModuleKey: string | null
  activeSecondaryRouteName: string | null
  breadcrumb: WorkspaceShellBreadcrumb
}

function matchPrefix(path: string, prefix: string): boolean {
  return path === prefix || path.startsWith(`${prefix}/`)
}

function matchAny(path: string, prefixes: string[]): boolean {
  return prefixes.some((prefix) => matchPrefix(path, prefix))
}

function normalizeBackofficeModule(module: BackofficeModule): WorkspaceShellModule {
  return {
    ...module,
    key: module.key,
    roles: [...module.roles],
    secondaryItems: module.secondaryItems.map((item) => ({
      ...item,
      roles: [...item.roles],
    })),
  }
}

const studentModules: WorkspaceShellModule[] = [
  {
    key: 'training',
    label: '训练',
    roles: ['student'],
    secondaryItems: [
      {
        routeName: 'Dashboard',
        label: '仪表盘',
        path: '/student/dashboard',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/student/dashboard', '/dashboard']),
      },
      {
        routeName: 'Challenges',
        label: '题目',
        path: '/challenges',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/challenges']),
      },
      {
        routeName: 'Instances',
        label: '我的实例',
        path: '/student/instances',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/student/instances', '/instances']),
      },
      {
        routeName: 'SkillProfile',
        label: '能力画像',
        path: '/student/skill-profile',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/student/skill-profile', '/skill-profile']),
      },
    ],
  },
  {
    key: 'events',
    label: '赛事',
    roles: ['student'],
    secondaryItems: [
      {
        routeName: 'Contests',
        label: '竞赛',
        path: '/contests',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/contests']),
      },
      {
        routeName: 'Scoreboard',
        label: '排行榜',
        path: '/scoreboard',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/scoreboard']),
      },
    ],
  },
  {
    key: 'account',
    label: '账户',
    roles: ['student'],
    secondaryItems: [
      {
        routeName: 'Notifications',
        label: '通知',
        path: '/notifications',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/notifications']),
      },
      {
        routeName: 'Profile',
        label: '个人资料',
        path: '/profile',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/profile', '/student/profile']),
      },
      {
        routeName: 'SecuritySettings',
        label: '安全设置',
        path: '/settings/security',
        roles: ['student'],
        isMatch: (path) => matchAny(path, ['/settings/security', '/student/settings/security']),
      },
    ],
  },
]

function getStudentModules(role?: WorkspaceRole): WorkspaceShellModule[] {
  if (role && role !== 'student') return []
  return studentModules
}

function getRoleBadge(role: WorkspaceRole, isBackoffice: boolean): string {
  if (isBackoffice) {
    if (role === 'admin') return 'Administrator Panel'
    if (role === 'teacher') return 'Instructor Access'
    return 'Platform Console'
  }
  if (role === 'admin') return 'Administrator Panel'
  if (role === 'teacher') return 'Instructor Workspace'
  return 'Student Console'
}

function getRoleCaption(role: WorkspaceRole): string {
  if (role === 'admin') return '系统管理'
  if (role === 'teacher') return '教学空间'
  return '学生空间'
}

function getWorkspacePath(role: WorkspaceRole, isBackoffice: boolean): string {
  if (!isBackoffice) return '/student/dashboard'
  return role === 'teacher' ? '/academy/overview' : '/platform/overview'
}

function inferStudentDetailLabel(
  input: WorkspaceShellNavigationInput,
  secondaryPath: string
): string | null {
  if (input.detailLabel) return input.detailLabel
  if (input.path === secondaryPath) return null

  if (/^\/challenges\/[^/]+(?:\/.*)?$/.test(input.path)) return '题目详情'
  if (/^\/contests\/[^/]+(?:\/.*)?$/.test(input.path)) return '竞赛详情'
  if (/^\/scoreboard\/[^/]+(?:\/.*)?$/.test(input.path)) return '排行详情'
  if (/^\/notifications\/[^/]+(?:\/.*)?$/.test(input.path)) return '通知详情'

  const routeTitle = input.pageTitle || resolveRouteTitle({ path: input.path })
  return routeTitle || null
}

export function createWorkspaceShellNavigation(
  input: WorkspaceShellNavigationInput
): WorkspaceShellNavigation {
  const isBackoffice = isBackofficePath(input.path)
  const role = input.role
  const modules = isBackoffice
    ? getVisibleBackofficeModules(role).map(normalizeBackofficeModule)
    : getStudentModules(role)

  const activeModule = isBackoffice
    ? getBackofficeModuleByPath(input.path)
    : (modules.find((module) => module.secondaryItems.some((item) => item.isMatch(input.path))) ??
      null)

  const visibleSecondaryItems = isBackoffice
    ? getVisibleBackofficeSecondaryItems(input.path, role).map((item) => ({
        ...item,
        roles: [...item.roles],
      }))
    : (activeModule?.secondaryItems.map((item) => ({
        ...item,
        active: item.isMatch(input.path),
      })) ?? [])

  const activeSecondaryItem = visibleSecondaryItems.find((item) => item.active) ?? null
  const workspacePath = getWorkspacePath(role, isBackoffice)
  const modulePath = visibleSecondaryItems[0]?.path ?? workspacePath
  const secondaryPath = activeSecondaryItem?.path ?? modulePath
  const pageTitle = input.pageTitle || resolveRouteTitle({ path: input.path })
  const detailLabel = isBackoffice
    ? (input.detailLabel ?? null)
    : inferStudentDetailLabel(input, secondaryPath)

  return {
    isBackoffice,
    brandMark: isBackoffice ? 'OPS' : 'CTF',
    brandKicker: isBackoffice ? 'ChallengeOps' : 'Student Space',
    brandTitle: isBackoffice ? '后台工作台' : '攻防实训平台',
    brandTooltip: isBackoffice ? 'ChallengeOps 后台' : 'CTF 靶场平台',
    roleBadge: getRoleBadge(role, isBackoffice),
    roleCaption: getRoleCaption(role),
    modules,
    activeModuleKey: activeModule?.key ?? null,
    activeSecondaryRouteName: activeSecondaryItem?.routeName ?? null,
    breadcrumb: {
      workspacePath,
      moduleLabel: activeModule?.label ?? (isBackoffice ? '后台' : '训练'),
      modulePath,
      secondaryLabel: activeSecondaryItem?.label ?? pageTitle ?? '工作区',
      secondaryPath,
      detailLabel,
      detailPath: input.fullPath ?? input.path,
    },
  }
}

export function useWorkspaceShellNavigation(
  input: () => WorkspaceShellNavigationInput
): ComputedRef<WorkspaceShellNavigation> {
  return computed(() => createWorkspaceShellNavigation(input()))
}
