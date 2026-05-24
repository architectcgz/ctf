import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const targetByRouteName = {
  TeacherClassTrend: {
    panel: 'trend',
    routeName: 'TeacherClassStudents',
  },
  TeacherClassReview: {
    panel: 'review',
    routeName: 'TeacherClassStudents',
  },
  TeacherClassInsights: {
    panel: 'insight',
    routeName: 'TeacherClassStudents',
  },
  TeacherClassIntervention: {
    panel: 'action',
    routeName: 'TeacherClassStudents',
  },
  PlatformClassTrend: {
    panel: 'trend',
    routeName: 'PlatformClassStudents',
  },
  PlatformClassReview: {
    panel: 'review',
    routeName: 'PlatformClassStudents',
  },
  PlatformClassInsights: {
    panel: 'insight',
    routeName: 'PlatformClassStudents',
  },
  PlatformClassIntervention: {
    panel: 'action',
    routeName: 'PlatformClassStudents',
  },
} as const

export function useClassWorkspaceSection() {
  const route = useRoute()
  const router = useRouter()

  const target = computed(() => {
    const routeName = route.name as keyof typeof targetByRouteName | undefined
    return routeName ? targetByRouteName[routeName] : null
  })

  async function redirectToCanonicalWorkspace(): Promise<void> {
    if (!target.value) return

    await router.replace({
      name: target.value.routeName,
      params: {
        className: route.params.className,
      },
      query: {
        ...route.query,
        panel: target.value.panel,
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
