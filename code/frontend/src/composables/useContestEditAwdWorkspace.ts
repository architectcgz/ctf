import { computed, ref, type Ref } from 'vue'

import {
  createContestAWDService,
  getContestAWDReadiness,
  listAdminAwdServiceTemplates,
  listContestAWDServices,
  updateContestAWDService,
  type AdminContestAWDServiceCreatePayload,
} from '@/api/admin'
import type {
  AdminAwdServiceTemplateData,
  AdminContestChallengeViewData,
  AWDReadinessData,
  ContestDetailData,
} from '@/api/contracts'
import { ApiError } from '@/api/request'
import type { ContestWorkbenchStageKey } from '@/composables/useContestWorkbench'
import { useToast } from '@/composables/useToast'
import { mapPlatformContestAwdServicesToChallengeLinks } from '@/utils/platformContestAwdChallengeLinks'

export interface ContestAwdChallengeConfigPayload {
  challenge_id?: number
  template_id: number
  points: number
  order: number
  is_visible: boolean
  awd_checker_type: AdminContestChallengeViewData['awd_checker_type']
  awd_checker_config: Record<string, unknown>
  awd_sla_score: number
  awd_defense_score: number
  awd_checker_preview_token?: string
}

interface UseContestEditAwdWorkspaceOptions {
  contest: Ref<ContestDetailData | null>
  contestId: Ref<string>
  selectTab: (tab: ContestWorkbenchStageKey) => void
}

type AwdConfigFocusSource = 'pool' | 'preflight' | null

export function useContestEditAwdWorkspace(options: UseContestEditAwdWorkspaceOptions) {
  const { contest, contestId, selectTab } = options
  const toast = useToast()

  const loadingAwdStageData = ref(false)
  const savingChallengeConfig = ref(false)
  const awdConfigLoadError = ref('')
  const awdPreflightLoadError = ref('')
  const awdChallengeLinks = ref<AdminContestChallengeViewData[]>([])
  const awdChallengeLinksLoaded = ref(false)
  const awdReadiness = ref<AWDReadinessData | null>(null)
  const awdServiceTemplateCatalog = ref<AdminAwdServiceTemplateData[]>([])
  const awdChallengeConfigDialogOpen = ref(false)
  const awdChallengeConfigMode = ref<'create' | 'edit'>('create')
  const editingAwdChallengeLink = ref<AdminContestChallengeViewData | null>(null)
  const activeAwdChallengeId = ref<string | null>(null)
  const awdConfigFocusSource = ref<AwdConfigFocusSource>(null)
  const loadingAwdServiceTemplateCatalog = ref(false)

  const sortedAwdChallengeLinks = computed(() =>
    [...awdChallengeLinks.value].sort(
      (left, right) => left.order - right.order || left.challenge_id.localeCompare(right.challenge_id)
    )
  )
  const activeAwdChallengeIndex = computed(() =>
    sortedAwdChallengeLinks.value.findIndex((item) => item.challenge_id === activeAwdChallengeId.value)
  )
  const canNavigatePreviousAwdChallenge = computed(() => activeAwdChallengeIndex.value > 0)
  const canNavigateNextAwdChallenge = computed(
    () => activeAwdChallengeIndex.value >= 0 && activeAwdChallengeIndex.value < sortedAwdChallengeLinks.value.length - 1
  )
  const existingAwdChallengeIds = computed(() => awdChallengeLinks.value.map((item) => item.challenge_id))

  function humanizeRequestError(error: unknown, fallback: string): string {
    if (error instanceof ApiError && error.message.trim()) return error.message
    if (error instanceof Error && error.message.trim()) return error.message
    return fallback
  }

  function resetAwdWorkbenchState() {
    awdConfigLoadError.value = ''
    awdPreflightLoadError.value = ''
    awdChallengeLinks.value = []
    awdChallengeLinksLoaded.value = false
    awdReadiness.value = null
    awdServiceTemplateCatalog.value = []
    awdChallengeConfigDialogOpen.value = false
  }

  async function refreshAwdWorkbenchData(nextContestId = contestId.value): Promise<void> {
    if (!contest.value || contest.value.mode !== 'awd' || !nextContestId) {
      resetAwdWorkbenchState()
      return
    }

    loadingAwdStageData.value = true
    try {
      awdConfigLoadError.value = ''
      awdPreflightLoadError.value = ''
      const [awdServicesResult, readinessResult] = await Promise.allSettled([
        listContestAWDServices(nextContestId),
        getContestAWDReadiness(nextContestId),
      ])

      if (awdServicesResult.status === 'fulfilled') {
        awdChallengeLinks.value = mapPlatformContestAwdServicesToChallengeLinks(
          awdServicesResult.value
        )
        awdChallengeLinksLoaded.value = true
      } else {
        awdConfigLoadError.value = humanizeRequestError(awdServicesResult.reason, 'AWD 配置同步失败')
        toast.error(awdConfigLoadError.value)
      }

      if (readinessResult.status === 'fulfilled') {
        awdReadiness.value = readinessResult.value
      } else {
        awdPreflightLoadError.value = humanizeRequestError(
          readinessResult.reason,
          '赛前检查同步失败'
        )
        toast.error(awdPreflightLoadError.value)
      }
    } finally {
      loadingAwdStageData.value = false
    }
  }

  async function loadAwdServiceTemplateCatalog(): Promise<void> {
    if (loadingAwdServiceTemplateCatalog.value || awdServiceTemplateCatalog.value.length > 0) return

    loadingAwdServiceTemplateCatalog.value = true
    try {
      const result = await listAdminAwdServiceTemplates({ page: 1, page_size: 100, status: 'published' })
      awdServiceTemplateCatalog.value = result.list
    } catch (error) {
      toast.error(humanizeRequestError(error, '服务模板加载失败'))
    } finally {
      loadingAwdServiceTemplateCatalog.value = false
    }
  }

  function setActiveAwdChallenge(challengeId: string | null, source: AwdConfigFocusSource) {
    activeAwdChallengeId.value = challengeId
    awdConfigFocusSource.value = challengeId ? source : null
  }

  function buildAwdServicePayload(
    payload: ContestAwdChallengeConfigPayload
  ): AdminContestAWDServiceCreatePayload {
    return {
      template_id: payload.template_id,
      points: payload.points,
      order: payload.order,
      is_visible: payload.is_visible,
      checker_type: payload.awd_checker_type ?? undefined,
      checker_config: payload.awd_checker_config,
      awd_sla_score: payload.awd_sla_score,
      awd_defense_score: payload.awd_defense_score,
      awd_checker_preview_token: payload.awd_checker_preview_token,
    }
  }

  function focusAwdChallengeByOffset(offset: -1 | 1) {
    if (activeAwdChallengeIndex.value < 0) return

    const nextChallenge = sortedAwdChallengeLinks.value[activeAwdChallengeIndex.value + offset]
    if (!nextChallenge) return

    setActiveAwdChallenge(nextChallenge.challenge_id, awdConfigFocusSource.value)
  }

  function openAwdChallengeCreateDialog() {
    awdChallengeConfigMode.value = 'create'
    editingAwdChallengeLink.value = null
    awdChallengeConfigDialogOpen.value = true
    void loadAwdServiceTemplateCatalog()
  }

  function openAwdChallengeEditDialog(challenge: AdminContestChallengeViewData) {
    setActiveAwdChallenge(challenge.challenge_id, awdConfigFocusSource.value)
    awdChallengeConfigMode.value = 'edit'
    editingAwdChallengeLink.value = challenge
    awdChallengeConfigDialogOpen.value = true
    void loadAwdServiceTemplateCatalog()
  }

  async function handleSaveAwdChallengeConfig(payload: ContestAwdChallengeConfigPayload) {
    if (!contest.value || savingChallengeConfig.value) return

    savingChallengeConfig.value = true
    try {
      const servicePayload = buildAwdServicePayload(payload)
      if (awdChallengeConfigMode.value === 'create') {
        await createContestAWDService(contest.value.id, servicePayload)
      } else if (editingAwdChallengeLink.value?.awd_service_id) {
        await updateContestAWDService(
          contest.value.id,
          editingAwdChallengeLink.value.awd_service_id,
          servicePayload
        )
      }

      awdChallengeConfigDialogOpen.value = false
      await refreshAwdWorkbenchData(contest.value.id)
    } catch (error) {
      toast.error(humanizeRequestError(error, '保存 AWD 配置失败'))
    } finally {
      savingChallengeConfig.value = false
    }
  }

  function handleOpenAwdConfigFromPool(challenge: AdminContestChallengeViewData) {
    activeAwdChallengeId.value = challenge.challenge_id
    awdConfigFocusSource.value = 'pool'
    selectTab('awd-config')
  }

  function handleNavigateAwdChallengeFromPreflight(challengeId: string) {
    setActiveAwdChallenge(challengeId, 'preflight')
    selectTab('awd-config')
  }

  return {
    activeAwdChallengeId,
    awdChallengeConfigDialogOpen,
    awdChallengeConfigMode,
    awdChallengeLinks,
    awdChallengeLinksLoaded,
    awdConfigFocusSource,
    awdConfigLoadError,
    awdPreflightLoadError,
    awdReadiness,
    awdServiceTemplateCatalog,
    canNavigateNextAwdChallenge,
    canNavigatePreviousAwdChallenge,
    editingAwdChallengeLink,
    existingAwdChallengeIds,
    focusAwdChallengeByOffset,
    handleNavigateAwdChallengeFromPreflight,
    handleOpenAwdConfigFromPool,
    handleSaveAwdChallengeConfig,
    loadAwdServiceTemplateCatalog,
    loadingAwdServiceTemplateCatalog,
    loadingAwdStageData,
    openAwdChallengeCreateDialog,
    openAwdChallengeEditDialog,
    refreshAwdWorkbenchData,
    savingChallengeConfig,
  }
}
