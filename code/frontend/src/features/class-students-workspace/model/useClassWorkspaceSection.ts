import { computed } from 'vue'

const panelByRouteName = {
  TeacherClassTrend: 'trend',
  TeacherClassReview: 'review',
  TeacherClassInsights: 'insight',
  TeacherClassIntervention: 'action',
  PlatformClassTrend: 'trend',
  PlatformClassReview: 'review',
  PlatformClassInsights: 'insight',
  PlatformClassIntervention: 'action',
} as const

const canonicalRouteNameByAliasRouteName = {
  TeacherClassTrend: 'TeacherClassStudents',
  TeacherClassReview: 'TeacherClassStudents',
  TeacherClassInsights: 'TeacherClassStudents',
  TeacherClassIntervention: 'TeacherClassStudents',
  PlatformClassTrend: 'PlatformClassStudents',
  PlatformClassReview: 'PlatformClassStudents',
  PlatformClassInsights: 'PlatformClassStudents',
  PlatformClassIntervention: 'PlatformClassStudents',
} as const

type ClassWorkspacePanel = (typeof panelByRouteName)[keyof typeof panelByRouteName]

interface ClassWorkspaceRouteLike {
  name?: string | symbol | null
  params: {
    className?: string | string[] | null
  }
  query: Record<string, unknown>
}

interface UseClassWorkspaceSectionOptions {
  route: ClassWorkspaceRouteLike
}

export function useClassWorkspaceSection(options: UseClassWorkspaceSectionOptions) {
  const { route } = options

  const panel = computed(() => {
    if (typeof route.name !== 'string') return null

    return (panelByRouteName[route.name as keyof typeof panelByRouteName] ?? null) as
      | ClassWorkspacePanel
      | null
  })

  const canonicalWorkspaceTarget = computed(() => {
    if (typeof route.name !== 'string') return null

    const workspaceRouteName =
      canonicalRouteNameByAliasRouteName[
        route.name as keyof typeof canonicalRouteNameByAliasRouteName
      ] ?? null
    const className = normalizeClassName(route.params.className)
    if (!workspaceRouteName || !panel.value || !className) return null

    return {
      name: workspaceRouteName,
      params: {
        className,
      },
      query: {
        ...route.query,
        panel: panel.value,
      },
    }
  })

  return {
    canonicalWorkspaceTarget,
    panel,
  }
}

function normalizeClassName(className: ClassWorkspaceRouteLike['params']['className']): string | null {
  if (Array.isArray(className)) {
    return className[0] ?? null
  }

  return className ?? null
}
