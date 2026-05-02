import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  getContest,
  listContestAWDServices,
  runContestAWDCheckerPreview,
  updateContestAWDService,
} from '@/api/admin/contests'
import type {
  AdminContestAWDServiceData,
  AWDCheckerPreviewData,
  AWDCheckerType,
  ContestDetailData,
} from '@/api/contracts'
import { useAwdCheckResultPresentation } from '@/features/awd-inspector'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useToast } from '@/composables/useToast'
import { useAwdCheckerConfigDraft } from './useAwdCheckerConfigDraft'

export function useContestAwdConfigPage() {
  const route = useRoute()
  const router = useRouter()
  const toast = useToast()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const contestId = computed(() => String(route.params.id ?? ''))
  const loading = ref(true)
  const refreshing = ref(false)
  const saving = ref(false)
  const previewing = ref(false)
  const loadError = ref('')
  const contest = ref<ContestDetailData | null>(null)
  const services = ref<AdminContestAWDServiceData[]>([])
  const selectedServiceId = ref('')
  const previewResult = ref<AWDCheckerPreviewData | null>(null)
  const previewError = ref('')
  const previewToken = ref('')
  const previewSignature = ref('')
  const previewForm = reactive({
    access_url: '',
    preview_flag: 'flag{preview}',
  })

  let loadVersion = 0

  const selectedService = computed(
    () => services.value.find((service) => service.id === selectedServiceId.value) || null
  )
  const selectedCheckerType = computed<AWDCheckerType | undefined>(
    () => selectedService.value?.checker_type
  )
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

  const sortedServices = computed(() =>
    [...services.value].sort(
      (left, right) =>
        left.order - right.order || left.display_name.localeCompare(right.display_name)
    )
  )
  const canAttachPreviewToken = computed(
    () => Boolean(previewToken.value) && previewSignature.value === currentSignature.value
  )

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

  function readNumericID(value: string): number | undefined {
    const next = Number(value)
    return Number.isFinite(next) && next > 0 ? next : undefined
  }

  function readServiceQuery(): string {
    const value = route.query.service
    if (Array.isArray(value)) {
      return String(value[0] ?? '')
    }
    return typeof value === 'string' ? value : ''
  }

  function syncServiceQuery(serviceId: string) {
    if (!serviceId || readServiceQuery() === serviceId) return
    void router.replace({
      name: 'ContestAWDConfig',
      params: { id: contestId.value },
      query: { ...route.query, service: serviceId },
    })
  }

  function clearPreviewState() {
    previewResult.value = null
    previewError.value = ''
    previewToken.value = ''
    previewSignature.value = ''
    previewForm.access_url = ''
    previewForm.preview_flag = 'flag{preview}'
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
      const requestedServiceId = readServiceQuery()
      const selectedServiceStillExists = serviceList.some(
        (service) => service.id === selectedServiceId.value
      )
      const requestedServiceExists = serviceList.some(
        (service) => service.id === requestedServiceId
      )
      if (!selectedServiceId.value || !selectedServiceStillExists) {
        selectedServiceId.value = requestedServiceExists
          ? requestedServiceId
          : serviceList[0]?.id || ''
      }
      syncServiceQuery(selectedServiceId.value)
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

  function selectService(service: AdminContestAWDServiceData) {
    selectedServiceId.value = service.id
    void router.replace({
      name: 'ContestAWDConfig',
      params: { id: contestId.value },
      query: { service: service.id },
    })
  }

  function goBackToStudio() {
    void router.push({
      name: 'ContestEdit',
      params: { id: contestId.value },
      query: { panel: 'awd-config' },
    })
  }

  async function handlePreview() {
    if (
      previewing.value ||
      !selectedService.value ||
      !selectedCheckerType.value ||
      !validateConfig()
    ) {
      return
    }
    previewing.value = true
    previewError.value = ''
    previewResult.value = null
    previewToken.value = ''
    previewSignature.value = ''
    try {
      const result = await runContestAWDCheckerPreview(contestId.value, {
        ...(readNumericID(selectedService.value.id)
          ? { service_id: readNumericID(selectedService.value.id) }
          : {}),
        awd_challenge_id: Number(selectedService.value.awd_challenge_id),
        checker_type: selectedCheckerType.value,
        checker_config: buildCurrentCheckerConfig(),
        ...(previewForm.access_url.trim()
          ? { access_url: previewForm.access_url.trim() }
          : {}),
        preview_flag: previewForm.preview_flag.trim() || undefined,
      })
      previewResult.value = result
      previewToken.value = result.preview_token || ''
      previewSignature.value = currentSignature.value
    } catch (error) {
      previewError.value = error instanceof Error && error.message.trim() ? error.message : '试跑失败'
    } finally {
      previewing.value = false
    }
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
    if (!syncingDraft.value && previous && next !== previewSignature.value) {
      previewToken.value = ''
    }
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
