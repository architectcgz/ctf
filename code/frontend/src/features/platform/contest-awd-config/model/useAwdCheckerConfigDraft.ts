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
  type AWDCheckerFieldErrorKey,
  type AWDHTTPStandardDraft,
  type AWDLegacyProbeDraft,
  type AWDScriptCheckerDraft,
  type AWDTCPStandardDraft,
} from './awdCheckerConfigSupport'
import { useAwdCheckerDraftHydration } from './useAwdCheckerDraftHydration'
import { useAwdTcpStepActions } from './useAwdTcpStepActions'

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
  const { hydrateServiceDraft, applyHTTPPreset, applyFieldErrors } = useAwdCheckerDraftHydration({
    form,
    legacyProbeDraft,
    httpStandardDraft,
    tcpStandardDraft,
    scriptCheckerDraft,
    fieldErrors,
    expandedTCPCheckerStepIndex,
    syncingDraft,
    clearErrors,
  })
  const { addTCPCheckerStep, removeTCPCheckerStep, toggleTCPCheckerStep, summarizeTCPCheckerStep } =
    useAwdTcpStepActions({
      tcpStandardDraft,
      expandedTCPCheckerStepIndex,
    })

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
    applyFieldErrors(result.errors)
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
