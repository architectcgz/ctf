import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { useBackofficeBreadcrumbDetail } from '@/shared/model/layout/useBackofficeBreadcrumbDetail'
import { useTheme } from '@/shared/model/theme/useTheme'
import { useLayoutSessionActionsBridge } from '@/shared/model/layout'
import { useWorkspaceShellNavigation } from '@/shared/model/layout/useWorkspaceShellNavigation'
import { resolveRouteTitle } from '@/utils/routeTitle'

function firstRouteParamValue(param: unknown): string {
  if (Array.isArray(param)) {
    return param[0] ?? ''
  }
  if (typeof param === 'string') {
    return param
  }
  return ''
}

function prefixedDetailLabel(prefix: string, value: string, fallback: string): string {
  const normalizedValue = value.trim()
  return normalizedValue ? `${prefix} #${normalizedValue}` : fallback
}

export function useTopNavViewState() {
  const route = useRoute()
  const router = useRouter()
  const authStore = useAuthStore()
  const isMobile = ref(typeof window !== 'undefined' ? window.innerWidth < 768 : false)
  const brandPickerRef = ref<HTMLElement | null>(null)
  const brandPickerOpen = ref(false)

  const { logout } = useLayoutSessionActionsBridge()
  const { availableBrands, brand, setBrand, theme, toggleTheme } = useTheme()
  const { breadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const pageTitle = computed(() => resolveRouteTitle(route))
  const backofficeDetailLabel = computed(() => {
    if (breadcrumbDetailTitle.value) {
      return breadcrumbDetailTitle.value
    }

    const routeName = String(route.name ?? '')
    const id = firstRouteParamValue(route.params.id)
    const className = firstRouteParamValue(route.params.className)
    const studentId = firstRouteParamValue(route.params.studentId)
    const contestId = firstRouteParamValue(route.params.contestId)
    const importId = firstRouteParamValue(route.params.importId)

    if (
      [
        'PlatformChallengeDetail',
        'PlatformChallengeTopologyStudio',
        'PlatformChallengeWriteup',
        'PlatformChallengeWriteupView',
      ].includes(routeName)
    ) {
      return prefixedDetailLabel('题目', id, '题目详情')
    }

    if (routeName === 'PlatformChallengeImportPreview') {
      return prefixedDetailLabel('导入', importId, '导入预览')
    }

    if (['ContestEdit', 'ContestAnnouncements', 'ContestOperations'].includes(routeName)) {
      return prefixedDetailLabel('竞赛', id, '竞赛详情')
    }

    if (['PlatformAwdReviewDetail', 'TeacherAWDReviewDetail'].includes(routeName)) {
      return prefixedDetailLabel('赛事', contestId, 'AWD复盘详情')
    }

    if (
      [
        'PlatformStudentAnalysis',
        'PlatformStudentReviewArchive',
        'TeacherStudentAnalysis',
        'TeacherStudentReviewArchive',
      ].includes(routeName)
    ) {
      return studentId.trim() ? `学生 ${studentId.trim()}` : '学生详情'
    }

    if (
      [
        'PlatformClassStudents',
        'PlatformClassTrend',
        'PlatformClassReview',
        'PlatformClassInsights',
        'PlatformClassIntervention',
        'TeacherClassStudents',
        'TeacherClassTrend',
        'TeacherClassReview',
        'TeacherClassInsights',
        'TeacherClassIntervention',
      ].includes(routeName)
    ) {
      return className.trim() || '班级详情'
    }

    return null
  })

  const shell = useWorkspaceShellNavigation(() => ({
    path: route.path,
    fullPath: route.fullPath,
    role: authStore.user?.role,
    routeName: String(route.name ?? ''),
    pageTitle: pageTitle.value,
    detailLabel: backofficeDetailLabel.value,
  }))

  const backofficeBreadcrumb = computed(() => shell.value.breadcrumb)
  const roleCaption = computed(() => shell.value.roleCaption)
  const currentBrandLabel = computed(
    () => availableBrands.find((option) => option.value === brand.value)?.label || '绿色'
  )
  const userDisplayName = computed(() => authStore.user?.name || authStore.user?.username || '未登录')
  const userInitial = computed(() => userDisplayName.value.slice(0, 1).toUpperCase())

  function onResize() {
    if (typeof window === 'undefined') return
    isMobile.value = window.innerWidth < 768
  }

  function toggleBrandPicker(): void {
    brandPickerOpen.value = !brandPickerOpen.value
  }

  function closeBrandPicker(): void {
    brandPickerOpen.value = false
  }

  function selectBrand(nextBrand: (typeof availableBrands)[number]['value']): void {
    setBrand(nextBrand)
    closeBrandPicker()
  }

  function navigateBreadcrumb(path: string): void {
    void router.push(path)
  }

  function handleDocumentPointerDown(event: MouseEvent): void {
    if (!brandPickerOpen.value) return

    const target = event.target
    if (!(target instanceof Node)) return
    if (brandPickerRef.value?.contains(target)) return

    closeBrandPicker()
  }

  function handleDocumentKeydown(event: KeyboardEvent): void {
    if (event.key !== 'Escape' || !brandPickerOpen.value) return
    closeBrandPicker()
  }

  onMounted(() => {
    if (typeof window === 'undefined') return
    window.addEventListener('resize', onResize)
    document.addEventListener('mousedown', handleDocumentPointerDown)
    document.addEventListener('keydown', handleDocumentKeydown)
  })

  onUnmounted(() => {
    if (typeof window === 'undefined') return
    window.removeEventListener('resize', onResize)
    document.removeEventListener('mousedown', handleDocumentPointerDown)
    document.removeEventListener('keydown', handleDocumentKeydown)
  })

  return {
    isMobile,
    brandPickerRef,
    brandPickerOpen,
    backofficeBreadcrumb,
    roleCaption,
    currentBrandLabel,
    userDisplayName,
    userInitial,
    logout,
    availableBrands,
    brand,
    theme,
    toggleTheme,
    toggleBrandPicker,
    closeBrandPicker,
    selectBrand,
    navigateBreadcrumb,
  }
}
