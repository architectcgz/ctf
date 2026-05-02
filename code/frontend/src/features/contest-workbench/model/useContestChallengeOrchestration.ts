import { computed, onMounted, ref, watch, type Ref } from 'vue'

import {
  createContestAWDService,
  createAdminContestChallenge,
  listAdminContestChallenges,
  listContestAWDServices,
  deleteContestAWDService,
  deleteAdminContestChallenge,
  updateContestAWDService,
  updateAdminContestChallenge,
} from '@/api/admin/contests'
import { getChallenges } from '@/api/admin/authoring'
import type {
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'
import { ApiError } from '@/api/request'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'
import { mapPlatformContestAwdServicesToChallengeLinks } from '@/utils/platformContestAwdChallengeLinks'
import { useContestAwdChallengePicker } from './useContestAwdChallengePicker'
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

  function summarizeAwdChallengeFailures(awdChallengeIds: number[]): string {
    const failedNames = awdChallengeIds.map(
      (awdChallengeId) =>
        awdChallengeCatalog.value.find((item) => Number(item.id) === awdChallengeId)?.name ||
        `AWD #${awdChallengeId}`
    )
    return `部分 AWD 题目关联失败：${failedNames.join('、')}`
  }

  function buildAwdServiceCreatePayload(
    awdChallengeId: number,
    payload: ContestOrchestrationSavePayload,
    order: number
  ) {
    const awdChallenge = awdChallengeCatalog.value.find(
      (item) => Number(item.id) === awdChallengeId
    )
    const checkerConfig =
      awdChallenge?.checker_config && typeof awdChallenge.checker_config === 'object'
        ? awdChallenge.checker_config
        : undefined

    return {
      awd_challenge_id: awdChallengeId,
      points: payload.points,
      order,
      is_visible: payload.is_visible,
      ...(awdChallenge?.checker_type ? { checker_type: awdChallenge.checker_type } : {}),
      ...(checkerConfig ? { checker_config: checkerConfig } : {}),
    }
  }

  async function handleSave(payload: ContestOrchestrationSavePayload) {
    saving.value = true
    try {
      if (isAwdContest.value) {
        const awdChallengeIds =
          dialogMode.value === 'create' && payload.awd_challenge_ids?.length
            ? payload.awd_challenge_ids
            : payload.awd_challenge_id
              ? [payload.awd_challenge_id]
              : []

        if (awdChallengeIds.length === 0) {
          toast.error('请选择 AWD 题目')
          return
        }

        if (dialogMode.value === 'create') {
          const results = await Promise.allSettled(
            awdChallengeIds.map((awdChallengeId, index) =>
              createContestAWDService(
                options.contestId.value,
                buildAwdServiceCreatePayload(awdChallengeId, payload, payload.order + index)
              )
            )
          )
          const failedResults = results.flatMap((result, index) =>
            result.status === 'rejected'
              ? [{ awdChallengeId: awdChallengeIds[index], error: result.reason }]
              : []
          )

          if (failedResults.length > 0) {
            const failedIds = failedResults.map(({ awdChallengeId }) => awdChallengeId)
            const failureMessage = summarizeAwdChallengeFailures(failedIds)

            if (failedResults.length === awdChallengeIds.length) {
              toast.error(failureMessage)
              return
            }

            toast.warning(failureMessage)
            options.onUpdated()
            if (!usingExternalChallengeLinks.value) {
              await refresh()
            }
            return
          }
        } else if (editingChallenge.value) {
          await updateContestAWDService(options.contestId.value, editingChallenge.value.awd_service_id!, {
            awd_challenge_id: awdChallengeIds[0],
            points: payload.points,
            order: payload.order,
            is_visible: payload.is_visible,
          })
        }
      } else if (dialogMode.value === 'create') {
        await createAdminContestChallenge(options.contestId.value, {
          challenge_id: payload.challenge_id!,
          points: payload.points,
          order: payload.order,
          is_visible: payload.is_visible,
        })
      } else if (editingChallenge.value) {
        await updateAdminContestChallenge(options.contestId.value, editingChallenge.value.challenge_id, {
          points: payload.points,
          order: payload.order,
          is_visible: payload.is_visible,
        })
      }

      toast.success('题目已保存')
      closeDialog()
      options.onUpdated()
      if (!usingExternalChallengeLinks.value) await refresh()
    } catch (error) {
      toast.error(humanizeRequestError(error, '保存失败'))
    } finally {
      saving.value = false
    }
  }

  async function handleRemove(challenge: AdminContestChallengeViewData) {
    const confirmed = await confirmDestructiveAction({
      title: '移除题目',
      message: `确认将“${getChallengeTitle(challenge)}”从竞赛中移除吗？`,
    })
    if (!confirmed) return

    removingChallengeId.value = challenge.id
    try {
      if (options.contestMode.value === 'awd') {
        await deleteContestAWDService(options.contestId.value, challenge.awd_service_id!)
      } else {
        await deleteAdminContestChallenge(options.contestId.value, challenge.challenge_id)
      }
      toast.success('题目已移除')
      options.onUpdated()
      if (!usingExternalChallengeLinks.value) await refresh()
    } catch (error) {
      toast.error(humanizeRequestError(error, '移除失败'))
    } finally {
      removingChallengeId.value = null
    }
  }

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
