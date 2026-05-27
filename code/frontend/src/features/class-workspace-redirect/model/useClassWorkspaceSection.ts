import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

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

type ClassWorkspaceCanonicalRouteName = 'TeacherClassStudents' | 'PlatformClassStudents'

interface UseClassWorkspaceSectionOptions {
  workspaceRouteName: ClassWorkspaceCanonicalRouteName
}

export function useClassWorkspaceSection(options: UseClassWorkspaceSectionOptions) {
  const { workspaceRouteName } = options
  const route = useRoute()
  const router = useRouter()

  const panel = computed(() => {
    const routeName = route.name as keyof typeof panelByRouteName | undefined
    return routeName ? panelByRouteName[routeName] : null
  })

  async function redirectToCanonicalWorkspace(): Promise<void> {
    if (!panel.value) return

    await router.replace({
      name: workspaceRouteName,
      params: {
        className: route.params.className,
      },
      query: {
        ...route.query,
        panel: panel.value,
      },
    })
  }

  watch(
    () => [route.name, route.params.className, route.query.panel] as const,
    () => {
      void redirectToCanonicalWorkspace()
    },
    { immediate: true }
  )

  return {
    redirectToCanonicalWorkspace,
  }
}
