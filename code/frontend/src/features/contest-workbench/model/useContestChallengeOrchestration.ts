import { computed, onMounted, ref, watch, type Ref } from 'vue'

import {
  listAdminContestChallenges,
  listContestAWDServices,
} from '@/api/admin/contests'
import { getChallenges } from '@/api/admin/authoring'
import type {
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'
import { ApiError } from '@/api/request'
import { useToast } from '@/composables/useToast'
import { mapPlatformContestAwdServicesToChallengeLinks } from '@/utils/platformContestAwdChallengeLinks'
import { useContestAwdChallengePicker } from './useContestAwdChallengePicker'
import { useContestChallengeMutations } from './useContestChallengeMutations'
import { useContestChallengePool } from './useContestChallengePool'

const CHALLENGE_CATALOG_PAGE_SIZE = 100
const AWD_CHALLENGE_PAGE_SIZE = 20

export interface ContestOrchestrationSavePayload {
  challenge_id?: number
  awd_challenge_id?: number
  awd_challenge_ids?: number[]
  points: number
  order: number
  is_visible: boolean
  awd_checker_type?: AdminContestChallengeViewData['awd_checker_type']
  awd_checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}

interface UseContestChallengeOrchestrationOptions {
  contestId: Readonly<Ref<string>>
  contestMode: Readonly<Ref<ContestDetailData['mode']>>
  challengeLinks: Readonly<Ref<AdminContestChallengeViewData[] | undefined>>
  loadingExternal: Readonly<Ref<boolean | undefined>>
  loadErrorExternal: Readonly<Ref<string | undefined>>
  createDialogRequestKey: Readonly<Ref<number | undefined>>
  onUpdated: () => void
}

export function useContestChallengeOrchestration(options: UseContestChallengeOrchestrationOptions) {
  const toast = useToast()
  const loading = ref(true)
  const saving = ref(false)
  const loadingChallengeCatalog = ref(false)
  const localChallengeLinks = ref<AdminContestChallengeViewData[]>([])
  const localLoadError = ref('')
  const challengeCatalog = ref<AdminChallengeListItem[]>([])
  const dialogOpen = ref(false)
  const dialogMode = ref<'create' | 'edit'>('create')
  const editingChallenge = ref<AdminContestChallengeViewData | null>(null)
  const removingChallengeId = ref<string | null>(null)
  const openActionMenuId = ref<string | null>(null)

  const usingExternalChallengeLinks = computed(() => options.challengeLinks.value !== undefined)
  const currentChallengeLinks = computed(() => options.challengeLinks.value ?? localChallengeLinks.value)
  const panelLoading = computed(() =>
    usingExternalChallengeLinks.value ? Boolean(options.loadingExternal.value) : loading.value
  )
  const panelLoadError = computed(() =>
    usingExternalChallengeLinks.value
      ? options.loadErrorExternal.value?.trim() ?? ''
      : localLoadError.value
  )

  const { visibleItems, summaryItems, filterItems, activeFilter, isAwdContest, setFilter } =
    useContestChallengePool(currentChallengeLinks, options.contestMode)
  const showAwdChallengeFilters = false
  const showChallengeOverflowMenu = false

  const panelCopy = computed(() =>
    isAwdContest.value
      ? '维护统一题目池，从 AWD 题库选题并完成比赛级分值编排。'
      : '维护统一题目池，安排题目顺序、分值和可见状态。'
  )
  const emptyState = computed(() => ({
    title: '暂无关联题目',
    description: '先从题库里关联题目，再安排顺序。',
  }))

  const existingChallengeIdSet = computed(
    () => new Set(currentChallengeLinks.value.map((item) => String(item.challenge_id)))
  )
  const existingChallengeIds = computed(() => Array.from(existingChallengeIdSet.value))

  const {
    filters: awdChallengeFilters,
    list: awdChallengeCatalog,
    total: awdChallengeTotal,
    page: awdChallengePage,
    pageSize: awdChallengePageSize,
    loading: loadingAwdChallengeCatalog,
    loadError: awdChallengeLoadError,
    refresh: refreshAwdChallengeCatalog,
    changePage: changeAwdChallengePage,
    setKeyword: setAwdChallengeKeyword,
    setServiceType: setAwdChallengeServiceType,
    setDeploymentMode: setAwdChallengeDeploymentMode,
    setReadinessStatus: setAwdChallengeReadiness,
  } = useContestAwdChallengePicker({
    existingChallengeIds,
    pageSize: AWD_CHALLENGE_PAGE_SIZE,
  })

  const dialogChallengeOptions = computed(() =>
    dialogMode.value === 'edit'
      ? challengeCatalog.value
      : challengeCatalog.value.filter((item) => !existingChallengeIdSet.value.has(String(item.id)))
  )

  function getChallengeTitle(item: AdminContestChallengeViewData): string {
    return item.title?.trim() || `Challenge #${item.challenge_id}`
  }

  function humanizeRequestError(error: unknown, fallback: string): string {
    if (error instanceof ApiError && error.message.trim()) return error.message
    if (error instanceof Error && error.message.trim()) return error.message
    return fallback
  }

  async function refresh() {
    if (usingExternalChallengeLinks.value) {
      options.onUpdated()
      return
    }

    loading.value = true
    try {
      if (options.contestMode.value === 'awd') {
        const nextAwdServices = await listContestAWDServices(options.contestId.value)
        localChallengeLinks.value = mapPlatformContestAwdServicesToChallengeLinks(nextAwdServices)
      } else {
        localChallengeLinks.value = await listAdminContestChallenges(options.contestId.value)
      }
      localLoadError.value = ''
    } catch (error) {
      localLoadError.value = humanizeRequestError(error, '加载失败')
    } finally {
      loading.value = false
    }
  }

  async function ensureChallengeCatalogLoaded() {
    if (loadingChallengeCatalog.value || challengeCatalog.value.length > 0) return

    loadingChallengeCatalog.value = true
    try {
      const result = await getChallenges({
        page: 1,
        page_size: CHALLENGE_CATALOG_PAGE_SIZE,
        status: 'published',
      })
      challengeCatalog.value = result.list
    } catch (error) {
      toast.error(humanizeRequestError(error, '题库加载失败'))
    } finally {
      loadingChallengeCatalog.value = false
    }
  }

  function openCreateDialog() {
    dialogMode.value = 'create'
    editingChallenge.value = null
    dialogOpen.value = true
    if (isAwdContest.value) {
      void changeAwdChallengePage(1)
      return
    }
    void ensureChallengeCatalogLoaded()
  }

  function handleCreateAction() {
    openCreateDialog()
  }

  function openEditDialog(challenge: AdminContestChallengeViewData) {
    dialogMode.value = 'edit'
    editingChallenge.value = challenge
    dialogOpen.value = true
    if (isAwdContest.value) void refreshAwdChallengeCatalog()
  }

  function closeDialog() {
    dialogOpen.value = false
    editingChallenge.value = null
  }
  const { handleSave, handleRemove } = useContestChallengeMutations({
    contestId: options.contestId,
    contestMode: options.contestMode,
    usingExternalChallengeLinks,
    isAwdContest,
    dialogMode,
    editingChallenge,
    awdChallengeCatalog,
    saving,
    removingChallengeId,
    onUpdated: options.onUpdated,
    refresh,
    closeDialog,
    humanizeRequestError,
    getChallengeTitle,
  })

  onMounted(() => {
    if (!usingExternalChallengeLinks.value) void refresh()
  })

  watch(
    options.createDialogRequestKey,
    (requestKey, previousRequestKey) => {
      if (!requestKey || requestKey === previousRequestKey) return
      handleCreateAction()
    },
    { immediate: true }
  )

  watch(awdChallengeLoadError, (error, previousError) => {
    if (!error || error === previousError) return
    toast.error(error)
  })

  return {
    activeFilter,
    awdChallengeCatalog,
    awdChallengeFilters,
    awdChallengeLoadError,
    awdChallengePage,
    awdChallengePageSize,
    awdChallengeTotal,
    changeAwdChallengePage,
    closeDialog,
    currentChallengeLinks,
    dialogChallengeOptions,
    dialogMode,
    dialogOpen,
    editingChallenge,
    emptyState,
    existingChallengeIds,
    filterItems,
    handleCreateAction,
    handleRemove,
    handleSave,
    isAwdContest,
    loadingAwdChallengeCatalog,
    loadingChallengeCatalog,
    openActionMenuId,
    openEditDialog,
    panelCopy,
    panelLoadError,
    panelLoading,
    refresh,
    refreshAwdChallengeCatalog,
    removingChallengeId,
    saving,
    setAwdChallengeDeploymentMode,
    setAwdChallengeKeyword,
    setAwdChallengeReadiness,
    setAwdChallengeServiceType,
    setFilter,
    showAwdChallengeFilters,
    showChallengeOverflowMenu,
    summaryItems,
    visibleItems,
  }
}
