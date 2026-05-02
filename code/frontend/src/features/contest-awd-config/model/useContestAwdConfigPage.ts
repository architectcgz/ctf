import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  getContest,
  listContestAWDServices,
  updateContestAWDService,
} from '@/api/admin/contests'
import type {
  AdminContestAWDServiceData,
  AWDCheckerType,
  ContestDetailData,
} from '@/api/contracts'
import { useAwdCheckResultPresentation } from '@/features/awd-inspector'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useToast } from '@/composables/useToast'
import { useAwdChallengeSelection } from './useAwdChallengeSelection'
import { useAwdCheckerConfigDraft } from './useAwdCheckerConfigDraft'
import { useAwdCheckerPreviewFlow } from './useAwdCheckerPreview'

export function useContestAwdConfigPage() {
  const route = useRoute()
  const router = useRouter()
  const toast = useToast()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const contestId = computed(() => String(route.params.id ?? ''))
  const loading = ref(true)
  const refreshing = ref(false)
  const saving = ref(false)
  const loadError = ref('')
  const contest = ref<ContestDetailData | null>(null)
  const services = ref<AdminContestAWDServiceData[]>([])

  let loadVersion = 0

  const {
    selectedServiceId,
    selectedService,
    selectedCheckerType,
    sortedServices,
    readServiceQuery,
    reconcileSelectedServiceId,
    selectService,
  } = useAwdChallengeSelection({
    contestId,
    route,
    router,
    services,
  })
  const {
    AWD_HTTP_METHOD_OPTIONS,
    AWD_HTTP_STANDARD_PRESETS,
    checkerConfigJSON,
    currentSignature,
    expandedTCPCheckerStepIndex,
    fieldErrors,
    form,
    httpActionSections,
    httpStandardDraft,
    legacyProbeDraft,
    scriptCheckerDraft,
    syncingDraft,
    tcpStandardDraft,
    addTCPCheckerStep,
    applyHTTPPreset,
    buildCurrentCheckerConfig,
    hydrateServiceDraft,
    removeTCPCheckerStep,
    summarizeTCPCheckerStep,
    toggleTCPCheckerStep,
    validateConfig,
  } = useAwdCheckerConfigDraft({
    selectedService,
    selectedCheckerType,
  })

  const {
    canAttachPreviewToken,
    clearPreviewState,
    handlePreview,
    handleSignatureChange,
    previewError,
    previewForm,
    previewResult,
    previewToken,
    previewing,
  } = useAwdCheckerPreviewFlow({
    contestId,
    selectedService,
    selectedCheckerType,
    currentSignature,
    syncingDraft,
    validateConfig,
    buildCurrentCheckerConfig,
  })

  const { summarizeCheckResult, getCheckStatusLabel, getPrimaryAccessURL } =
    useAwdCheckResultPresentation({
      formatDateTime,
    })
  const previewSummary = computed(() =>
    previewResult.value
      ? summarizeCheckResult({
          ...previewResult.value.check_result,
          preview_context: previewResult.value.preview_context,
        })
      : ''
  )
  const previewAccessURL = computed(() =>
    previewResult.value
      ? getPrimaryAccessURL({
          ...previewResult.value.check_result,
          preview_context: previewResult.value.preview_context,
        })
      : ''
  )

  function formatDateTime(value?: string): string {
    if (!value) return '未记录'
    return new Date(value).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  function getCheckerTypeLabel(value?: AWDCheckerType): string {
    switch (value) {
      case 'http_standard':
        return 'HTTP 标准 Checker'
      case 'tcp_standard':
        return 'TCP 标准 Checker'
      case 'script_checker':
        return '脚本 Checker'
      case 'legacy_probe':
        return '基础探活'
      default:
        return '未声明 Checker'
    }
  }

  function getProtocolLabel(value?: AWDCheckerType): string {
    switch (value) {
      case 'http_standard':
        return 'Web HTTP'
      case 'tcp_standard':
        return 'Binary TCP'
      case 'script_checker':
        return '题目包脚本'
      case 'legacy_probe':
        return '基础探活'
      default:
        return '题目包未声明'
    }
  }

  function getValidationLabel(value?: AdminContestAWDServiceData['validation_state']): string {
    switch (value) {
      case 'passed':
        return '已通过'
      case 'failed':
        return '未通过'
      case 'stale':
        return '待重验'
      case 'pending':
      default:
        return '待验证'
    }
  }

  async function loadPage(initial = false) {
    if (!contestId.value) return
    const version = ++loadVersion
    if (initial) loading.value = true
    refreshing.value = !initial
    try {
      const [contestDetail, serviceList] = await Promise.all([
        getContest(contestId.value),
        listContestAWDServices(contestId.value),
      ])
      if (version !== loadVersion) return
      contest.value = contestDetail
      services.value = serviceList
      setBreadcrumbDetailTitle(contestDetail.title)
      reconcileSelectedServiceId()
      loadError.value = ''
    } catch (error) {
      if (version !== loadVersion) return
      loadError.value =
        error instanceof Error && error.message.trim() ? error.message : 'AWD 配置加载失败'
    } finally {
      if (version === loadVersion) {
        loading.value = false
        refreshing.value = false
      }
    }
  }

  function goBackToStudio() {
    void router.push({
      name: 'ContestEdit',
      params: { id: contestId.value },
      query: { panel: 'awd-config' },
    })
  }

  async function handleSave() {
    if (
      saving.value ||
      !selectedService.value ||
      !selectedCheckerType.value ||
      !validateConfig()
    ) {
      return
    }
    saving.value = true
    try {
      await updateContestAWDService(contestId.value, selectedService.value.id, {
        checker_type: selectedCheckerType.value,
        checker_config: buildCurrentCheckerConfig(),
        awd_sla_score: form.sla_score,
        awd_defense_score: form.defense_score,
        ...(canAttachPreviewToken.value
          ? { awd_checker_preview_token: previewToken.value }
          : {}),
      })
      toast.success(canAttachPreviewToken.value ? '配置与试跑结果已保存' : '配置已保存')
      await loadPage(false)
    } catch (error) {
      toast.error(error instanceof Error && error.message.trim() ? error.message : '保存 AWD 配置失败')
    } finally {
      saving.value = false
    }
  }

  watch(selectedService, (service) => {
    clearPreviewState()
    hydrateServiceDraft(service)
  })

  watch(currentSignature, (next, previous) => {
    handleSignatureChange(next, previous)
  })

  onMounted(() => {
    selectedServiceId.value = readServiceQuery()
    void loadPage(true)
  })

  onUnmounted(() => {
    setBreadcrumbDetailTitle()
  })

  return {
    AWD_HTTP_METHOD_OPTIONS,
    AWD_HTTP_STANDARD_PRESETS,
    canAttachPreviewToken,
    checkerConfigJSON,
    contest,
    expandedTCPCheckerStepIndex,
    fieldErrors,
    form,
    getCheckStatusLabel,
    getCheckerTypeLabel,
    getProtocolLabel,
    getValidationLabel,
    goBackToStudio,
    handlePreview,
    handleSave,
    httpActionSections,
    httpStandardDraft,
    legacyProbeDraft,
    loadError,
    loading,
    loadPage,
    previewAccessURL,
    previewError,
    previewForm,
    previewResult,
    previewSummary,
    previewing,
    refreshing,
    saving,
    scriptCheckerDraft,
    selectService,
    selectedCheckerType,
    selectedService,
    selectedServiceId,
    sortedServices,
    tcpStandardDraft,
    addTCPCheckerStep,
    applyHTTPPreset,
    removeTCPCheckerStep,
    summarizeTCPCheckerStep,
    toggleTCPCheckerStep,
  }
}
