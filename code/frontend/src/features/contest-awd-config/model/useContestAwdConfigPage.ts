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
import {
  AWD_CHECKER_FIELD_ERROR_KEYS,
  AWD_HTTP_METHOD_OPTIONS,
  AWD_HTTP_STANDARD_PRESETS,
  buildCheckerConfigPreview,
  buildHTTPStandardCheckerConfig,
  buildLegacyProbeCheckerConfig,
  buildScriptCheckerConfig,
  buildTCPStandardCheckerConfig,
  createHTTPStandardDraft,
  createLegacyProbeDraft,
  createScriptCheckerDraft,
  createTCPStandardDraft,
  getHTTPStandardPresetDraft,
  type AWDCheckerFieldErrorKey,
  type AWDHTTPStandardDraft,
  type AWDLegacyProbeDraft,
  type AWDScriptCheckerDraft,
  type AWDTCPStandardDraft,
} from '@/components/platform/contest/awdCheckerConfigSupport'
import { useAwdCheckResultPresentation } from '@/features/awd-inspector'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useToast } from '@/composables/useToast'

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
  const expandedTCPCheckerStepIndex = ref<number | null>(null)

  const form = reactive({
    sla_score: 1,
    defense_score: 2,
  })
  const previewForm = reactive({
    access_url: '',
    preview_flag: 'flag{preview}',
  })
  const legacyProbeDraft = reactive<AWDLegacyProbeDraft>(createLegacyProbeDraft())
  const httpStandardDraft = reactive<AWDHTTPStandardDraft>(createHTTPStandardDraft())
  const tcpStandardDraft = reactive<AWDTCPStandardDraft>(createTCPStandardDraft())
  const scriptCheckerDraft = reactive<AWDScriptCheckerDraft>(createScriptCheckerDraft())
  const fieldErrors = reactive<
    Record<AWDCheckerFieldErrorKey | 'checker_type' | 'sla_score' | 'defense_score', string>
  >({
    checker_type: '',
    sla_score: '',
    defense_score: '',
    legacy_health_path: '',
    http_put_path: '',
    http_put_expected_status: '',
    http_put_headers_text: '',
    http_get_path: '',
    http_get_expected_status: '',
    http_get_headers_text: '',
    http_havoc_expected_status: '',
    http_havoc_headers_text: '',
    tcp_timeout: '',
    tcp_steps: '',
    script_entry: '',
    script_timeout: '',
    script_args_text: '',
    script_env_text: '',
  })

  let loadVersion = 0
  let syncingDraft = false

  const selectedService = computed(
    () => services.value.find((service) => service.id === selectedServiceId.value) || null
  )
  const selectedCheckerType = computed<AWDCheckerType | undefined>(
    () => selectedService.value?.checker_type
  )
  const httpActionSections = [
    {
      key: 'put_flag',
      title: 'PUT Flag',
      pathErrorKey: 'http_put_path',
      statusErrorKey: 'http_put_expected_status',
      headersErrorKey: 'http_put_headers_text',
    },
    {
      key: 'get_flag',
      title: 'GET Flag',
      pathErrorKey: 'http_get_path',
      statusErrorKey: 'http_get_expected_status',
      headersErrorKey: 'http_get_headers_text',
    },
    {
      key: 'havoc',
      title: 'Havoc',
      pathErrorKey: '',
      statusErrorKey: 'http_havoc_expected_status',
      headersErrorKey: 'http_havoc_headers_text',
    },
  ] as const
  const sortedServices = computed(() =>
    [...services.value].sort(
      (left, right) =>
        left.order - right.order || left.display_name.localeCompare(right.display_name)
    )
  )
  const checkerConfigJSON = computed(() =>
    JSON.stringify(
      selectedCheckerType.value
        ? buildCheckerConfigPreview(selectedCheckerType.value, {
            legacyProbeDraft,
            httpStandardDraft,
            tcpStandardDraft,
            scriptCheckerDraft,
          })
        : {},
      null,
      2
    )
  )
  const currentSignature = computed(() =>
    JSON.stringify({
      service_id: selectedService.value?.id || '',
      checker_type: selectedCheckerType.value || '',
      checker_config: selectedCheckerType.value ? buildCurrentCheckerConfig(false) : {},
    })
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

  function clearErrors() {
    fieldErrors.checker_type = ''
    fieldErrors.sla_score = ''
    fieldErrors.defense_score = ''
    for (const key of AWD_CHECKER_FIELD_ERROR_KEYS) {
      fieldErrors[key] = ''
    }
  }

  function clearPreviewState() {
    previewResult.value = null
    previewError.value = ''
    previewToken.value = ''
    previewSignature.value = ''
    previewForm.access_url = ''
    previewForm.preview_flag = 'flag{preview}'
  }

  function assignHTTPDraft(next: AWDHTTPStandardDraft) {
    httpStandardDraft.put_flag = { ...next.put_flag }
    httpStandardDraft.get_flag = { ...next.get_flag }
    httpStandardDraft.havoc = { ...next.havoc }
  }

  function assignTCPDraft(next: AWDTCPStandardDraft) {
    tcpStandardDraft.timeout_ms = next.timeout_ms
    tcpStandardDraft.steps.splice(
      0,
      tcpStandardDraft.steps.length,
      ...next.steps.map((step) => ({ ...step }))
    )
    expandedTCPCheckerStepIndex.value = null
  }

  function assignScriptDraft(next: AWDScriptCheckerDraft) {
    scriptCheckerDraft.runtime = next.runtime
    scriptCheckerDraft.entry = next.entry
    scriptCheckerDraft.timeout_sec = next.timeout_sec
    scriptCheckerDraft.args_text = next.args_text
    scriptCheckerDraft.env_text = next.env_text
    scriptCheckerDraft.output = next.output
  }

  function hydrateServiceDraft(service: AdminContestAWDServiceData | null) {
    syncingDraft = true
    clearErrors()
    clearPreviewState()
    form.sla_score = service?.sla_score ?? 1
    form.defense_score = service?.defense_score ?? 2
    legacyProbeDraft.health_path = createLegacyProbeDraft(service?.checker_config).health_path
    assignHTTPDraft(createHTTPStandardDraft(service?.checker_config))
    assignTCPDraft(createTCPStandardDraft(service?.checker_config))
    assignScriptDraft(createScriptCheckerDraft(service?.checker_config))
    syncingDraft = false
  }

  function applyHTTPPreset(presetId: string) {
    assignHTTPDraft(getHTTPStandardPresetDraft(presetId))
  }

  function addTCPCheckerStep() {
    tcpStandardDraft.steps.push({
      send: '',
      send_template: '',
      send_hex: '',
      expect_contains: '',
      expect_regex: '',
      timeout_ms: 3000,
    })
    expandedTCPCheckerStepIndex.value = tcpStandardDraft.steps.length - 1
  }

  function removeTCPCheckerStep(index: number) {
    if (tcpStandardDraft.steps.length <= 1) return
    tcpStandardDraft.steps.splice(index, 1)
    if (expandedTCPCheckerStepIndex.value === null) return
    expandedTCPCheckerStepIndex.value = Math.min(
      expandedTCPCheckerStepIndex.value,
      tcpStandardDraft.steps.length - 1
    )
  }

  function toggleTCPCheckerStep(index: number) {
    expandedTCPCheckerStepIndex.value =
      expandedTCPCheckerStepIndex.value === index ? null : index
  }

  function summarizeTCPCheckerStep(step: AWDTCPStandardDraft['steps'][number]): string {
    const send = step.send_template || step.send || step.send_hex
    const expect = step.expect_contains || step.expect_regex
    const parts = [
      send ? `发送 ${send}` : '',
      expect ? `期望 ${expect}` : '',
      Number.isInteger(step.timeout_ms) && step.timeout_ms > 0 ? `${step.timeout_ms}ms` : '',
    ].filter(Boolean)
    return parts.length > 0 ? parts.join(' · ') : '未配置收发规则'
  }

  function buildCurrentCheckerConfig(strict = true): Record<string, unknown> {
    switch (selectedCheckerType.value) {
      case 'http_standard':
        return buildHTTPStandardCheckerConfig(httpStandardDraft, strict).config
      case 'tcp_standard':
        return buildTCPStandardCheckerConfig(tcpStandardDraft, strict).config
      case 'script_checker':
        return buildScriptCheckerConfig(scriptCheckerDraft, strict).config
      case 'legacy_probe':
        return buildLegacyProbeCheckerConfig(legacyProbeDraft).config
      default:
        return {}
    }
  }

  function validateConfig(): boolean {
    clearErrors()
    if (!selectedCheckerType.value) {
      fieldErrors.checker_type = '当前题目包没有声明 checker 类型'
    }
    if (!Number.isInteger(form.sla_score) || form.sla_score < 0 || form.sla_score > 5) {
      fieldErrors.sla_score = 'SLA 分必须是 0-5 的整数'
    }
    if (
      !Number.isInteger(form.defense_score) ||
      form.defense_score < 0 ||
      form.defense_score > 5
    ) {
      fieldErrors.defense_score = '防守分必须是 0-5 的整数'
    }

    let result = { errors: {} as Partial<Record<AWDCheckerFieldErrorKey, string>> }
    switch (selectedCheckerType.value) {
      case 'http_standard':
        result = buildHTTPStandardCheckerConfig(httpStandardDraft, true)
        break
      case 'tcp_standard':
        result = buildTCPStandardCheckerConfig(tcpStandardDraft, true)
        break
      case 'script_checker':
        result = buildScriptCheckerConfig(scriptCheckerDraft, true)
        break
      case 'legacy_probe':
        result = buildLegacyProbeCheckerConfig(legacyProbeDraft)
        break
    }
    for (const [key, value] of Object.entries(result.errors)) {
      fieldErrors[key as AWDCheckerFieldErrorKey] = value || ''
    }
    return Object.values(fieldErrors).every((value) => !value)
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
    hydrateServiceDraft(service)
  })

  watch(currentSignature, (next, previous) => {
    if (!syncingDraft && previous && next !== previewSignature.value) {
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
