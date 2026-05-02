import { computed, reactive, ref, type ComputedRef } from 'vue'

import type { AdminContestAWDServiceData, AWDCheckerType } from '@/api/contracts'
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
} from './awdCheckerConfigSupport'

interface UseAwdCheckerConfigDraftOptions {
  selectedService: ComputedRef<AdminContestAWDServiceData | null>
  selectedCheckerType: ComputedRef<AWDCheckerType | undefined>
}

export function useAwdCheckerConfigDraft(options: UseAwdCheckerConfigDraftOptions) {
  const { selectedService, selectedCheckerType } = options

  const form = reactive({
    sla_score: 1,
    defense_score: 2,
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
  const expandedTCPCheckerStepIndex = ref<number | null>(null)
  const syncingDraft = ref(false)

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

  function clearErrors() {
    fieldErrors.checker_type = ''
    fieldErrors.sla_score = ''
    fieldErrors.defense_score = ''
    for (const key of AWD_CHECKER_FIELD_ERROR_KEYS) {
      fieldErrors[key] = ''
    }
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
    syncingDraft.value = true
    clearErrors()
    form.sla_score = service?.sla_score ?? 1
    form.defense_score = service?.defense_score ?? 2
    legacyProbeDraft.health_path = createLegacyProbeDraft(service?.checker_config).health_path
    assignHTTPDraft(createHTTPStandardDraft(service?.checker_config))
    assignTCPDraft(createTCPStandardDraft(service?.checker_config))
    assignScriptDraft(createScriptCheckerDraft(service?.checker_config))
    syncingDraft.value = false
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

  return {
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
    clearErrors,
    hydrateServiceDraft,
    removeTCPCheckerStep,
    summarizeTCPCheckerStep,
    toggleTCPCheckerStep,
    validateConfig,
  }
}
