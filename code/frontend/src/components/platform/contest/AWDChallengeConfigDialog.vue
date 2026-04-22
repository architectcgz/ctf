<script setup lang="ts">
import { computed, onUnmounted, reactive, ref, watch } from 'vue'

import { runContestAWDCheckerPreview } from '@/api/admin'
import type {
  AdminAwdServiceTemplateData,
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  AWDCheckerPreviewData,
  AWDCheckerType,
  AWDTeamServiceData,
} from '@/api/contracts'
import SlideOverDrawer from '@/components/common/modal-templates/SlideOverDrawer.vue'
import {
  extractAwdRuntimeImageRef,
  formatAwdPreviewRuntimeError,
} from '@/components/platform/awdRuntimeImageSupport'
import {
  useContestAwdPreviewRealtime,
  type ContestAwdPreviewProgressEvent,
} from '@/composables/useContestAwdPreviewRealtime'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'
import {
  AWD_CHECKER_FIELD_ERROR_KEYS,
  AWD_HTTP_METHOD_OPTIONS,
  AWD_HTTP_STANDARD_PRESETS,
  buildCheckerConfigPreview,
  buildHTTPStandardCheckerConfig,
  buildLegacyProbeCheckerConfig,
  createHTTPStandardDraft,
  createLegacyProbeDraft,
  getHTTPStandardPresetDraft,
  type AWDHTTPStandardDraft,
  type AWDLegacyProbeDraft,
} from './awdCheckerConfigSupport'
import {
  AWD_CHECKER_PREVIEW_ATTEMPT_TOTAL,
  AWD_CHECKER_PREVIEW_PROGRESS_PHASES,
  formatAwdCheckerPreviewElapsed,
  resolveAwdCheckerPreviewProgressPhaseIndex,
  resolveAwdCheckerPreviewProgressPhaseIndexByKey,
} from './awdCheckerPreviewProgress'

type DialogMode = 'create' | 'edit'

const props = withDefaults(
  defineProps<{
    contestId?: string | null
    open: boolean
    mode: DialogMode
    challengeOptions: AdminChallengeListItem[]
    templateOptions?: AdminAwdServiceTemplateData[]
    existingChallengeIds: string[]
    draft?: AdminContestChallengeViewData | null
    loadingChallengeCatalog: boolean
    loadingTemplateCatalog?: boolean
    saving: boolean
  }>(),
  {
    templateOptions: () => [],
    loadingTemplateCatalog: false,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      challenge_id?: number
      template_id: number
      points: number
      order: number
      is_visible: boolean
      awd_checker_type: AWDCheckerType
      awd_checker_config: Record<string, unknown>
      awd_sla_score: number
      awd_defense_score: number
      awd_checker_preview_token?: string
    },
  ]
}>()

const form = reactive({
  challenge_id: '',
  template_id: '',
  points: 100,
  order: 0,
  is_visible: 'true',
  awd_checker_type: 'legacy_probe' as AWDCheckerType,
  awd_sla_score: 0,
  awd_defense_score: 0,
})

const legacyProbeDraft = reactive<AWDLegacyProbeDraft>(createLegacyProbeDraft())
const httpStandardDraft = reactive<AWDHTTPStandardDraft>(createHTTPStandardDraft())
const previewForm = reactive({
  access_url: '',
  preview_flag: 'flag{preview}',
})
const previewing = ref(false)
const previewResult = ref<AWDCheckerPreviewData | null>(null)
const previewError = ref('')
const previewToken = ref('')
const previewSignature = ref('')
const previewTokenInvalidated = ref(false)
const syncingDialogState = ref(false)
const previewProgressElapsedMs = ref(0)
const previewProgressPhaseIndex = ref(0)
const previewRequestId = ref('')
const previewRealtimePhaseLabel = ref('')
const previewRealtimePhaseDetail = ref('')
const previewRealtimeAttempt = ref<number | null>(null)
const previewRealtimeTotalAttempts = ref<number | null>(null)
const previewRealtimeStatus = ref('')

let previewProgressTimer: ReturnType<typeof globalThis.setInterval> | null = null
let previewProgressStartedAt = 0

function createFieldErrorState() {
  return {
    challenge_id: '',
    template_id: '',
    points: '',
    order: '',
    awd_sla_score: '',
    awd_defense_score: '',
    legacy_health_path: '',
    http_put_path: '',
    http_put_expected_status: '',
    http_put_headers_text: '',
    http_get_path: '',
    http_get_expected_status: '',
    http_get_headers_text: '',
    http_havoc_expected_status: '',
    http_havoc_headers_text: '',
    preview_access_url: '',
  }
}

const fieldErrors = reactive(createFieldErrorState())

const dialogTitle = computed(() =>
  props.mode === 'create' ? '新增 AWD 题库题目' : '编辑 AWD 题目配置'
)

const selectableChallenges = computed(() =>
  props.challengeOptions.filter(
    (item) => props.mode === 'edit' || !props.existingChallengeIds.includes(item.id)
  )
)
const selectableTemplates = computed(() => props.templateOptions)
const selectedTemplate = computed<AdminAwdServiceTemplateData | null>(() => {
  const templateId = form.template_id || props.draft?.awd_template_id || ''
  return selectableTemplates.value.find((item) => item.id === templateId) || null
})
const selectedTemplateName = computed(() => selectedTemplate.value?.name || '未选择模板')
const selectedTemplateCheckerTypeLabel = computed(() => {
  switch (selectedTemplate.value?.checker_type) {
    case 'http_standard':
      return 'HTTP Standard'
    case 'legacy_probe':
      return 'Legacy Probe'
    default:
      return '未定义'
  }
})
const selectedTemplateSnapshotSections = computed(() => {
  if (!selectedTemplate.value) {
    return []
  }
  return [
    {
      key: 'checker',
      label: 'Checker',
      value: selectedTemplate.value.checker_config,
      empty: '题库模板未写入 checker_config',
    },
    {
      key: 'flag',
      label: 'Flag 策略',
      value: selectedTemplate.value.flag_config,
      empty: '题库模板未写入 flag_config',
    },
    {
      key: 'access',
      label: '访问与入口',
      value: selectedTemplate.value.access_config,
      empty: '题库模板未写入 access_config',
    },
    {
      key: 'runtime',
      label: '运行参数',
      value: selectedTemplate.value.runtime_config,
      empty: '题库模板未写入 runtime_config',
    },
  ]
})

const activeChallengeLabel = computed(() => {
  if (props.mode === 'edit') {
    const title = props.draft?.title?.trim() || `Challenge #${props.draft?.challenge_id || ''}`
    return title
  }
  return (
    selectableTemplates.value.find((item) => item.id === form.template_id)?.name || '请选择服务模板'
  )
})

function resolvePreviewChallengeID(): number {
  if (props.mode === 'edit') {
    return Number(props.draft?.challenge_id || form.challenge_id || form.template_id || 0)
  }
  return Number(form.template_id || form.challenge_id || 0)
}

function resolvePreviewServiceID(): number {
  if (props.mode !== 'edit') {
    return 0
  }
  return Number(props.draft?.awd_service_id || 0)
}

const checkerPreviewText = computed(() =>
  JSON.stringify(
    buildCheckerConfigPreview(form.awd_checker_type, {
      legacyProbeDraft,
      httpStandardDraft,
    }),
    null,
    2
  )
)
const previewResultJSONText = computed(() =>
  JSON.stringify(previewResult.value?.check_result || {}, null, 2)
)
const selectedTemplateCategorySummary = computed(() => {
  if (!selectedTemplate.value) {
    return ''
  }
  return [
    selectedTemplate.value.category.toUpperCase(),
    selectedTemplate.value.difficulty,
    selectedTemplate.value.service_type,
    selectedTemplate.value.deployment_mode,
  ].join(' · ')
})
const selectedTemplateRuntimeImageRef = computed(() =>
  extractAwdRuntimeImageRef(selectedTemplate.value?.runtime_config)
)

function formatPreviewDateTime(value?: string): string {
  if (!value) {
    return '未记录'
  }
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

const {
  summarizeCheckResult,
  getCheckActions,
  getCheckTargets,
  getPrimaryAccessURL,
  getTargetProbeSummary,
  getProbeStatusText,
  getCheckStatusLabel,
  getValidationStateLabel,
  formatLatency,
} = useAwdCheckResultPresentation({
  formatDateTime: formatPreviewDateTime,
})

function buildPresentationResult(value: AWDCheckerPreviewData | null): Record<string, unknown> {
  if (!value) {
    return {}
  }
  return {
    ...value.check_result,
    preview_context: value.preview_context,
  }
}

const previewCheckResult = computed(() => previewResult.value?.check_result || {})
const previewPresentationResult = computed(() => buildPresentationResult(previewResult.value))
const previewSummaryText = computed(() =>
  previewResult.value ? summarizeCheckResult(previewPresentationResult.value) : ''
)
const previewTargetSummaryText = computed(() =>
  previewResult.value ? getTargetProbeSummary(previewCheckResult.value) : ''
)
const previewAccessURL = computed(() =>
  previewResult.value ? getPrimaryAccessURL(previewPresentationResult.value) : ''
)
const previewActions = computed(() => getCheckActions(previewCheckResult.value))
const previewTargets = computed(() => getCheckTargets(previewCheckResult.value))
const previewPendingSaveNotice = computed(() =>
  canAttachPreviewToken.value
    ? '试跑已完成，这还是临时结果。点击下方保存按钮后，才会写入“最近一次已保存校验”。'
    : ''
)
const previewRealtime = useContestAwdPreviewRealtime(
  props.contestId || '',
  handlePreviewRealtimeProgress
)
const previewProgressPhase = computed(
  () =>
    AWD_CHECKER_PREVIEW_PROGRESS_PHASES[previewProgressPhaseIndex.value] ||
    AWD_CHECKER_PREVIEW_PROGRESS_PHASES[0]
)
const previewProgressSteps = computed(() =>
  AWD_CHECKER_PREVIEW_PROGRESS_PHASES.map((phase, index) => ({
    ...phase,
    state:
      index < previewProgressPhaseIndex.value
        ? 'done'
        : index === previewProgressPhaseIndex.value
          ? 'current'
          : 'pending',
  }))
)
const previewProgressAttemptText = computed(() => {
  const totalAttempts = previewRealtimeTotalAttempts.value || AWD_CHECKER_PREVIEW_ATTEMPT_TOTAL
  if (previewRealtimeAttempt.value) {
    return `第 ${previewRealtimeAttempt.value} / ${totalAttempts} 轮`
  }
  const currentPhase = previewProgressPhase.value
  if (currentPhase.attempt) {
    return `第 ${currentPhase.attempt} / ${totalAttempts} 轮`
  }
  return `共 ${totalAttempts} 轮`
})
const previewProgressStatusText = computed(() => {
  if (previewRealtimeStatus.value === 'failed') {
    return '试跑失败'
  }
  if (previewRealtimePhaseLabel.value) {
    return previewRealtimePhaseLabel.value
  }
  const currentPhase = previewProgressPhase.value
  if (currentPhase.attempt) {
    return `正在执行第 ${currentPhase.attempt} / ${AWD_CHECKER_PREVIEW_ATTEMPT_TOTAL} 轮`
  }
  return currentPhase.label
})
const previewProgressDetailText = computed(
  () => previewRealtimePhaseDetail.value || previewProgressPhase.value.detail
)
const previewProgressElapsedText = computed(() =>
  formatAwdCheckerPreviewElapsed(previewProgressElapsedMs.value)
)
const previewProgressPercent = computed(() => {
  const phaseCount = AWD_CHECKER_PREVIEW_PROGRESS_PHASES.length
  return Math.round(((previewProgressPhaseIndex.value + 1) / phaseCount) * 100)
})
const previewButtonText = computed(() => {
  if (!previewing.value) {
    return '试跑 Checker'
  }
  const currentPhase = previewProgressPhase.value
  if (currentPhase.attempt) {
    return `试跑中 · 第 ${currentPhase.attempt}/${AWD_CHECKER_PREVIEW_ATTEMPT_TOTAL} 轮`
  }
  if (currentPhase.key === 'summary') {
    return '试跑中 · 汇总结果'
  }
  return '试跑中 · 准备环境'
})
const previewProgressConnectionHint = computed(() => {
  if (previewRealtime.status.value === 'open') {
    return '已接入实时进度事件'
  }
  if (previewRealtime.status.value === 'connecting') {
    return '正在连接实时进度事件'
  }
  return '实时事件不可用时，将回退为本地进度提示'
})
const submitButtonText = computed(() => {
  if (props.saving) {
    return '保存中...'
  }
  if (previewing.value) {
    return '试跑进行中，暂不能保存'
  }
  if (canAttachPreviewToken.value) {
    return props.mode === 'create' ? '新增题目并写入试跑结果' : '保存配置并写入试跑结果'
  }
  return props.mode === 'create' ? '新增题目' : '保存配置'
})

const savedPreviewResult = computed(() => props.draft?.awd_checker_last_preview_result || null)
const savedPreviewPresentationResult = computed(() =>
  buildPresentationResult(savedPreviewResult.value)
)
const savedPreviewSummaryText = computed(() =>
  savedPreviewResult.value ? summarizeCheckResult(savedPreviewPresentationResult.value) : ''
)
const savedPreviewAccessURL = computed(() =>
  savedPreviewResult.value ? getPrimaryAccessURL(savedPreviewPresentationResult.value) : ''
)
const savedValidationStateLabel = computed(() =>
  getValidationStateLabel(props.draft?.awd_checker_validation_state)
)
const savedPreviewEmptyText = computed(() => {
  if (previewPendingSaveNotice.value) {
    return '当前试跑结果尚未保存。点击下方保存按钮后，这里会显示最新一次已保存校验。'
  }
  return '当前配置还没有保存过试跑结果。'
})

const currentCheckerSignature = computed(() =>
  JSON.stringify({
    challenge_id: resolvePreviewChallengeID(),
    checker_type: form.awd_checker_type,
    checker_config: buildCheckerConfig(false),
  })
)

const canAttachPreviewToken = computed(
  () => Boolean(previewToken.value) && previewSignature.value === currentCheckerSignature.value
)

function assignLegacyProbeDraft(next: AWDLegacyProbeDraft) {
  legacyProbeDraft.health_path = next.health_path
}

function assignHTTPStandardDraft(next: AWDHTTPStandardDraft) {
  httpStandardDraft.put_flag.method = next.put_flag.method
  httpStandardDraft.put_flag.path = next.put_flag.path
  httpStandardDraft.put_flag.expected_status = next.put_flag.expected_status
  httpStandardDraft.put_flag.headers_text = next.put_flag.headers_text
  httpStandardDraft.put_flag.body_template = next.put_flag.body_template
  httpStandardDraft.put_flag.expected_substring = next.put_flag.expected_substring

  httpStandardDraft.get_flag.method = next.get_flag.method
  httpStandardDraft.get_flag.path = next.get_flag.path
  httpStandardDraft.get_flag.expected_status = next.get_flag.expected_status
  httpStandardDraft.get_flag.headers_text = next.get_flag.headers_text
  httpStandardDraft.get_flag.body_template = next.get_flag.body_template
  httpStandardDraft.get_flag.expected_substring = next.get_flag.expected_substring

  httpStandardDraft.havoc.method = next.havoc.method
  httpStandardDraft.havoc.path = next.havoc.path
  httpStandardDraft.havoc.expected_status = next.havoc.expected_status
  httpStandardDraft.havoc.headers_text = next.havoc.headers_text
  httpStandardDraft.havoc.body_template = next.havoc.body_template
  httpStandardDraft.havoc.expected_substring = next.havoc.expected_substring
}

watch(
  () => [props.open, props.mode, props.draft] as const,
  ([open]) => {
    if (!open) {
      return
    }

    syncingDialogState.value = true
    form.challenge_id =
      props.mode === 'edit'
        ? props.draft?.challenge_id || ''
        : selectableTemplates.value[0]?.id || ''
    form.template_id = props.draft?.awd_template_id || selectableTemplates.value[0]?.id || ''
    form.points = props.draft?.points ?? 100
    form.order = props.draft?.order ?? 0
    form.is_visible = props.draft?.is_visible === false ? 'false' : 'true'
    form.awd_checker_type = props.draft?.awd_checker_type || 'legacy_probe'
    form.awd_sla_score = props.draft?.awd_sla_score ?? 0
    form.awd_defense_score = props.draft?.awd_defense_score ?? 0
    assignLegacyProbeDraft(createLegacyProbeDraft(props.draft?.awd_checker_config))
    assignHTTPStandardDraft(createHTTPStandardDraft(props.draft?.awd_checker_config))
    previewForm.access_url = ''
    previewForm.preview_flag = 'flag{preview}'
    previewResult.value = null
    previewError.value = ''
    previewToken.value = ''
    previewSignature.value = ''
    previewTokenInvalidated.value = false
    resetPreviewProgress()
    clearErrors()
    syncingDialogState.value = false
  },
  { immediate: true }
)

watch(
  () =>
    [props.open, props.mode, selectableChallenges.value.map((item) => item.id).join(',')] as const,
  ([open, mode]) => {
    if (!open || mode !== 'create') {
      return
    }
    if (selectableChallenges.value.length === 0) {
      form.challenge_id = form.template_id
      return
    }
    const hasSelectedChallenge = selectableChallenges.value.some(
      (item) => item.id === form.challenge_id
    )
    if (!hasSelectedChallenge) {
      form.challenge_id = selectableChallenges.value[0]?.id || ''
    }
  },
  { immediate: true }
)

watch(
  () =>
    [props.open, props.mode, selectableTemplates.value.map((item) => item.id).join(',')] as const,
  ([open]) => {
    if (!open) {
      return
    }
    const hasSelectedTemplate = selectableTemplates.value.some(
      (item) => item.id === form.template_id
    )
    if (!hasSelectedTemplate) {
      form.template_id = props.draft?.awd_template_id || selectableTemplates.value[0]?.id || ''
    }
    if (props.mode === 'create') {
      form.challenge_id = form.template_id
    }
  },
  { immediate: true }
)

watch(
  () => currentCheckerSignature.value,
  (nextSignature, previousSignature) => {
    if (!props.open || syncingDialogState.value || !previousSignature || !previewToken.value) {
      return
    }
    if (previewSignature.value && nextSignature !== previewSignature.value) {
      previewToken.value = ''
      previewTokenInvalidated.value = true
    }
  }
)

onUnmounted(() => {
  resetPreviewProgress()
  previewRealtime.stop()
})

function clearErrors() {
  for (const key of Object.keys(fieldErrors) as Array<keyof typeof fieldErrors>) {
    fieldErrors[key] = ''
  }
}

function createPreviewRequestId(): string {
  return `awd-preview-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

function closeDialog() {
  emit('update:open', false)
}

function applyHTTPPreset(presetId: string) {
  assignHTTPStandardDraft(getHTTPStandardPresetDraft(presetId))
  clearCheckerErrors()
}

function clearCheckerErrors() {
  for (const key of AWD_CHECKER_FIELD_ERROR_KEYS) {
    fieldErrors[key] = ''
  }
}

function clearPreviewErrors() {
  fieldErrors.preview_access_url = ''
}

function validate(): boolean {
  clearErrors()

  if (props.mode === 'create' && selectableChallenges.value.length > 0 && !form.challenge_id) {
    fieldErrors.challenge_id = '请选择题目'
  }
  if (!form.template_id) {
    fieldErrors.template_id = '请选择服务模板'
  }
  if (!Number.isInteger(form.points) || form.points <= 0) {
    fieldErrors.points = '分值必须是大于 0 的整数'
  }
  if (!Number.isInteger(form.order) || form.order < 0) {
    fieldErrors.order = '顺序必须是大于等于 0 的整数'
  }
  if (!Number.isInteger(form.awd_sla_score) || form.awd_sla_score < 0) {
    fieldErrors.awd_sla_score = 'SLA 分必须是大于等于 0 的整数'
  }
  if (!Number.isInteger(form.awd_defense_score) || form.awd_defense_score < 0) {
    fieldErrors.awd_defense_score = '防守分必须是大于等于 0 的整数'
  }

  const checkerResult = buildCheckerConfigResult(true)

  for (const [key, value] of Object.entries(checkerResult.errors)) {
    if (value) {
      fieldErrors[key as keyof typeof fieldErrors] = value
    }
  }

  return Object.values(fieldErrors).every((value) => !value)
}

function buildCheckerConfigResult(strict = true) {
  return form.awd_checker_type === 'http_standard'
    ? buildHTTPStandardCheckerConfig(httpStandardDraft, strict)
    : buildLegacyProbeCheckerConfig(legacyProbeDraft)
}

function buildCheckerConfig(strict = true) {
  return buildCheckerConfigResult(strict).config
}

function validatePreview(): boolean {
  clearCheckerErrors()
  clearPreviewErrors()
  fieldErrors.challenge_id = ''

  if (resolvePreviewChallengeID() <= 0) {
    fieldErrors.challenge_id = '请选择服务模板'
  }

  const checkerResult = buildCheckerConfigResult(true)

  for (const [key, value] of Object.entries(checkerResult.errors)) {
    if (value) {
      fieldErrors[key as keyof typeof fieldErrors] = value
    }
  }

  return !fieldErrors.challenge_id && AWD_CHECKER_FIELD_ERROR_KEYS.every((key) => !fieldErrors[key])
}

async function handlePreview() {
  if (!props.contestId) {
    previewError.value = '当前缺少赛事上下文，无法试跑 Checker。'
    return
  }
  if (!validatePreview()) {
    return
  }

  previewing.value = true
  previewError.value = ''
  previewResult.value = null
  previewToken.value = ''
  previewSignature.value = ''
  previewTokenInvalidated.value = false
  const nextPreviewRequestId = createPreviewRequestId()
  startPreviewProgress()
  previewRequestId.value = nextPreviewRequestId

  try {
    await previewRealtime.start().catch(() => undefined)
    const accessURL = previewForm.access_url.trim()
    const result = await runContestAWDCheckerPreview(props.contestId, {
      ...(resolvePreviewServiceID() > 0 ? { service_id: resolvePreviewServiceID() } : {}),
      challenge_id: resolvePreviewChallengeID(),
      checker_type: form.awd_checker_type,
      checker_config: buildCheckerConfig(),
      ...(accessURL ? { access_url: accessURL } : {}),
      preview_flag: previewForm.preview_flag.trim() || undefined,
      preview_request_id: nextPreviewRequestId,
    })
    previewResult.value = result
    previewToken.value = result.preview_token || ''
    previewSignature.value = currentCheckerSignature.value
    previewTokenInvalidated.value = false
  } catch (error) {
    const message = error instanceof Error ? error.message : '试跑失败，请稍后重试。'
    previewError.value = formatAwdPreviewRuntimeError(
      message,
      selectedTemplateRuntimeImageRef.value
    )
  } finally {
    stopPreviewProgress()
    previewRealtime.stop()
    previewing.value = false
  }
}

function startPreviewProgress() {
  resetPreviewProgress()
  previewProgressStartedAt = Date.now()
  syncPreviewProgress()
  previewProgressTimer = globalThis.setInterval(() => {
    syncPreviewProgress()
  }, 240)
}

function stopPreviewProgress() {
  syncPreviewProgress()
  if (previewProgressTimer !== null) {
    globalThis.clearInterval(previewProgressTimer)
    previewProgressTimer = null
  }
}

function resetPreviewProgress() {
  if (previewProgressTimer !== null) {
    globalThis.clearInterval(previewProgressTimer)
    previewProgressTimer = null
  }
  previewProgressStartedAt = 0
  previewProgressElapsedMs.value = 0
  previewProgressPhaseIndex.value = 0
  previewRequestId.value = ''
  previewRealtimePhaseLabel.value = ''
  previewRealtimePhaseDetail.value = ''
  previewRealtimeAttempt.value = null
  previewRealtimeTotalAttempts.value = null
  previewRealtimeStatus.value = ''
}

function syncPreviewProgress() {
  if (!previewProgressStartedAt) {
    previewProgressElapsedMs.value = 0
    previewProgressPhaseIndex.value = 0
    return
  }
  const elapsedMs = Math.max(0, Date.now() - previewProgressStartedAt)
  previewProgressElapsedMs.value = elapsedMs
  previewProgressPhaseIndex.value = resolveAwdCheckerPreviewProgressPhaseIndex(elapsedMs)
}

function handlePreviewRealtimeProgress(payload: ContestAwdPreviewProgressEvent) {
  if (!previewing.value) {
    return
  }
  const nextRequestId = payload.preview_request_id?.trim()
  if (previewRequestId.value && nextRequestId && nextRequestId !== previewRequestId.value) {
    return
  }
  if (payload.phase_key) {
    previewProgressPhaseIndex.value = resolveAwdCheckerPreviewProgressPhaseIndexByKey(
      payload.phase_key
    )
  }
  previewRealtimePhaseLabel.value = payload.phase_label?.trim() || ''
  previewRealtimePhaseDetail.value = payload.detail?.trim() || ''
  previewRealtimeStatus.value = payload.status?.trim() || ''
  previewRealtimeAttempt.value =
    typeof payload.attempt === 'number' && payload.attempt > 0 ? payload.attempt : null
  previewRealtimeTotalAttempts.value =
    typeof payload.total_attempts === 'number' && payload.total_attempts > 0
      ? payload.total_attempts
      : null
}

function getPreviewStatusLabel(status: AWDTeamServiceData['service_status']): string {
  const labels: Record<AWDTeamServiceData['service_status'], string> = {
    up: '正常',
    down: '下线',
    compromised: '已失陷',
  }
  return labels[status]
}

function getPreviewStatusClass(status: AWDTeamServiceData['service_status']): string {
  const classes: Record<AWDTeamServiceData['service_status'], string> = {
    up: 'ui-badge ui-badge--pill ui-badge--soft checker-preview-status checker-preview-status--up',
    down: 'ui-badge ui-badge--pill ui-badge--soft checker-preview-status checker-preview-status--down',
    compromised:
      'ui-badge ui-badge--pill ui-badge--soft checker-preview-status checker-preview-status--compromised',
  }
  return classes[status]
}

function getValidationStateClass(value?: string): string {
  const state = value || 'pending'
  return `ui-badge ui-badge--pill ui-badge--soft checker-validation-chip checker-validation-chip--${state}`
}

function formatTemplateSnapshotJSON(value?: Record<string, unknown>): string {
  if (!value || Object.keys(value).length === 0) {
    return '{}'
  }
  return JSON.stringify(value, null, 2)
}

function getPreviewActionStateText(action: {
  healthy: boolean
  error_code?: string
  error?: string
}): string {
  if (action.healthy) {
    return '动作成功'
  }
  return getCheckStatusLabel(action.error_code) || action.error || '动作失败'
}

function handleSubmit() {
  if (props.saving) {
    return
  }

  if (!validate()) {
    return
  }

  const payload = {
    template_id: Number(form.template_id),
    points: form.points,
    order: form.order,
    is_visible: form.is_visible === 'true',
    awd_checker_type: form.awd_checker_type,
    awd_checker_config: buildCheckerConfig(),
    awd_sla_score: form.awd_sla_score,
    awd_defense_score: form.awd_defense_score,
  } as {
    challenge_id?: number
    template_id: number
    points: number
    order: number
    is_visible: boolean
    awd_checker_type: AWDCheckerType
    awd_checker_config: Record<string, unknown>
    awd_sla_score: number
    awd_defense_score: number
    awd_checker_preview_token?: string
  }

  if (canAttachPreviewToken.value && previewToken.value) {
    payload.awd_checker_preview_token = previewToken.value
  }

  emit('save', payload)
}
</script>

<template>
  <SlideOverDrawer
    :open="open"
    :title="dialogTitle"
    :subtitle="
      mode === 'create'
        ? '先从 AWD 题库模板选题，再整理赛事级 Checker 草稿，保存后即可继续赛前试跑。'
        : '统一维护赛事服务的 Checker、分值、顺序和试跑结果。'
    "
    eyebrow="AWD Operations"
    width="52rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form class="space-y-6" @submit.prevent="handleSubmit">
      <div v-if="mode === 'edit'" class="ui-field awd-config-field">
        <label class="ui-field__label" for="awd-challenge-config-challenge">赛事题目</label>
        <span class="ui-control-wrap awd-config-readonly">
          <span class="ui-control awd-config-readonly__value">
            {{ activeChallengeLabel }}
          </span>
        </span>
        <p v-if="fieldErrors.challenge_id" class="ui-field__error">
          {{ fieldErrors.challenge_id }}
        </p>
      </div>

      <div class="ui-field awd-config-field">
        <label class="ui-field__label" for="awd-challenge-config-template">AWD 题库模板</label>
        <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.template_id }">
          <select
            id="awd-challenge-config-template"
            v-model="form.template_id"
            class="ui-control"
            :disabled="loadingTemplateCatalog"
          >
            <option value="" disabled>
              {{ loadingTemplateCatalog ? '正在加载模板...' : '请选择服务模板' }}
            </option>
            <option v-for="template in selectableTemplates" :key="template.id" :value="template.id">
              {{ template.name }}
            </option>
          </select>
        </span>
        <p v-if="fieldErrors.template_id" class="ui-field__error">
          {{ fieldErrors.template_id }}
        </p>
        <p class="ui-field__hint">
          比赛服务会直接继承题库模板里的入口、端口、flag 与基础 checker
          定义，再叠加本页的分值和校验配置。
        </p>
      </div>

      <section v-if="selectedTemplate" class="template-snapshot-card">
        <header class="list-heading template-snapshot-card__head">
          <div>
            <div class="journal-note-label">Template Snapshot</div>
            <h3 class="list-heading__title template-snapshot-card__title">题库模板快照</h3>
          </div>
          <p class="template-snapshot-card__hint">
            新增题目时以题库模板为准，防守入口、可攻击端口和 flag 策略都应优先来自这里。
          </p>
        </header>

        <div class="template-snapshot-card__summary">
          <div class="template-snapshot-chip">
            <span class="template-snapshot-chip__label">模板</span>
            <span class="template-snapshot-chip__value">{{ selectedTemplateName }}</span>
          </div>
          <div class="template-snapshot-chip">
            <span class="template-snapshot-chip__label">分类</span>
            <span class="template-snapshot-chip__value">{{ selectedTemplateCategorySummary }}</span>
          </div>
          <div class="template-snapshot-chip">
            <span class="template-snapshot-chip__label">默认 Checker</span>
            <span class="template-snapshot-chip__value">{{
              selectedTemplateCheckerTypeLabel
            }}</span>
          </div>
          <div class="template-snapshot-chip">
            <span class="template-snapshot-chip__label">防守入口</span>
            <span class="template-snapshot-chip__value">{{
              selectedTemplate.defense_entry_mode || '未定义'
            }}</span>
          </div>
          <div class="template-snapshot-chip">
            <span class="template-snapshot-chip__label">Flag 模式</span>
            <span class="template-snapshot-chip__value">{{
              selectedTemplate.flag_mode || '未定义'
            }}</span>
          </div>
        </div>

        <div class="template-snapshot-card__grid">
          <article
            v-for="section in selectedTemplateSnapshotSections"
            :key="section.key"
            class="template-snapshot-panel"
          >
            <header class="template-snapshot-panel__head">
              <h4 class="template-snapshot-panel__title">
                {{ section.label }}
              </h4>
              <p class="template-snapshot-panel__hint">
                {{ section.empty }}
              </p>
            </header>
            <pre
              :id="`awd-template-snapshot-${section.key}`"
              class="checker-preview template-snapshot-panel__json"
              >{{ formatTemplateSnapshotJSON(section.value) }}</pre
            >
          </article>
        </div>
      </section>

      <div class="grid gap-4 sm:grid-cols-3">
        <div class="ui-field awd-config-field">
          <label class="ui-field__label" for="awd-challenge-config-points">分值</label>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.points }">
            <input
              id="awd-challenge-config-points"
              v-model.number="form.points"
              type="number"
              min="1"
              step="1"
              class="ui-control"
            />
          </span>
          <p v-if="fieldErrors.points" class="ui-field__error">
            {{ fieldErrors.points }}
          </p>
        </div>

        <div class="ui-field awd-config-field">
          <label class="ui-field__label" for="awd-challenge-config-order">顺序</label>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.order }">
            <input
              id="awd-challenge-config-order"
              v-model.number="form.order"
              type="number"
              min="0"
              step="1"
              class="ui-control"
            />
          </span>
          <p v-if="fieldErrors.order" class="ui-field__error">
            {{ fieldErrors.order }}
          </p>
        </div>

        <div class="ui-field awd-config-field">
          <label class="ui-field__label" for="awd-challenge-config-visible">可见性</label>
          <span class="ui-control-wrap">
            <select id="awd-challenge-config-visible" v-model="form.is_visible" class="ui-control">
              <option value="true">可见</option>
              <option value="false">隐藏</option>
            </select>
          </span>
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-3">
        <div class="ui-field awd-config-field">
          <label class="ui-field__label" for="awd-challenge-config-checker-type">
            Checker 类型
          </label>
          <span class="ui-control-wrap">
            <select
              id="awd-challenge-config-checker-type"
              v-model="form.awd_checker_type"
              class="ui-control"
            >
              <option value="legacy_probe">基础探活</option>
              <option value="http_standard">HTTP 标准 Checker</option>
            </select>
          </span>
        </div>

        <div class="ui-field awd-config-field">
          <label class="ui-field__label" for="awd-challenge-config-sla-score">SLA 分</label>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.awd_sla_score }">
            <input
              id="awd-challenge-config-sla-score"
              v-model.number="form.awd_sla_score"
              type="number"
              min="0"
              step="1"
              class="ui-control"
            />
          </span>
          <p v-if="fieldErrors.awd_sla_score" class="ui-field__error">
            {{ fieldErrors.awd_sla_score }}
          </p>
        </div>

        <div class="ui-field awd-config-field">
          <label class="ui-field__label" for="awd-challenge-config-defense-score">防守分</label>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.awd_defense_score }">
            <input
              id="awd-challenge-config-defense-score"
              v-model.number="form.awd_defense_score"
              type="number"
              min="0"
              step="1"
              class="ui-control"
            />
          </span>
          <p v-if="fieldErrors.awd_defense_score" class="ui-field__error">
            {{ fieldErrors.awd_defense_score }}
          </p>
        </div>
      </div>

      <section class="checker-config-block">
        <header class="list-heading checker-config-block__head">
          <div>
            <div class="journal-note-label">Checker Config</div>
            <h3 class="list-heading__title checker-config-block__title">
              {{
                form.awd_checker_type === 'http_standard'
                  ? 'HTTP Standard 配置'
                  : 'Legacy Probe 配置'
              }}
            </h3>
          </div>
          <p class="checker-config-block__hint">
            {{
              form.awd_checker_type === 'http_standard'
                ? '按动作填写巡检规则，保存时会自动构造成 awd_checker_config。'
                : '配置基础探活路径；留空则回退全局健康检查路径。'
            }}
          </p>
        </header>

        <div v-if="form.awd_checker_type === 'legacy_probe'" class="ui-field awd-config-field">
          <label class="ui-field__label" for="awd-challenge-config-legacy-health-path">
            健康检查路径
          </label>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.legacy_health_path }">
            <input
              id="awd-challenge-config-legacy-health-path"
              v-model="legacyProbeDraft.health_path"
              type="text"
              placeholder="/healthz"
              class="ui-control"
            />
          </span>
          <p class="ui-field__hint">留空时使用当前全局健康检查路径。</p>
          <p v-if="fieldErrors.legacy_health_path" class="ui-field__error">
            {{ fieldErrors.legacy_health_path }}
          </p>
        </div>

        <template v-else>
          <div class="checker-preset-strip">
            <button
              v-for="preset in AWD_HTTP_STANDARD_PRESETS"
              :id="`awd-http-preset-${preset.id}`"
              :key="preset.id"
              type="button"
              class="ui-btn ui-btn--secondary checker-preset-button"
              @click="applyHTTPPreset(preset.id)"
            >
              <span class="checker-preset-button__label">{{ preset.label }}</span>
              <span class="checker-preset-button__hint">{{ preset.description }}</span>
            </button>
          </div>

          <section class="checker-action-section">
            <header class="list-heading checker-action-section__head">
              <h4 class="list-heading__title checker-action-section__title">PUT Flag</h4>
              <p class="checker-action-section__hint">写入当前轮 flag。</p>
            </header>
            <div class="checker-action-grid">
              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-put-method">Method</label>
                <span class="ui-control-wrap awd-http-action-control">
                  <select
                    id="awd-http-put-method"
                    v-model="httpStandardDraft.put_flag.method"
                    class="ui-control"
                  >
                    <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">
                      {{ method }}
                    </option>
                  </select>
                </span>
              </div>

              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-put-path">Path</label>
                <span
                  class="ui-control-wrap awd-http-action-control"
                  :class="{ 'is-error': !!fieldErrors.http_put_path }"
                >
                  <input
                    id="awd-http-put-path"
                    v-model="httpStandardDraft.put_flag.path"
                    type="text"
                    placeholder="/api/flag"
                    class="ui-control"
                  />
                </span>
                <p v-if="fieldErrors.http_put_path" class="ui-field__error awd-http-action-error">
                  {{ fieldErrors.http_put_path }}
                </p>
              </div>

              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-put-expected-status">状态码</label>
                <span
                  class="ui-control-wrap awd-http-action-control"
                  :class="{ 'is-error': !!fieldErrors.http_put_expected_status }"
                >
                  <input
                    id="awd-http-put-expected-status"
                    v-model.number="httpStandardDraft.put_flag.expected_status"
                    type="number"
                    min="1"
                    step="1"
                    class="ui-control"
                  />
                </span>
                <p
                  v-if="fieldErrors.http_put_expected_status"
                  class="ui-field__error awd-http-action-error"
                >
                  {{ fieldErrors.http_put_expected_status }}
                </p>
              </div>
            </div>

            <div class="checker-action-extra-grid">
              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-put-body-template">
                  Body Template
                </label>
                <span class="ui-control-wrap awd-http-action-control">
                  <textarea
                    id="awd-http-put-body-template"
                    v-model="httpStandardDraft.put_flag.body_template"
                    rows="4"
                    class="ui-control awd-config-control--mono"
                  />
                </span>
              </div>

              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-put-headers">Headers JSON</label>
                <span
                  class="ui-control-wrap awd-http-action-control"
                  :class="{ 'is-error': !!fieldErrors.http_put_headers_text }"
                >
                  <textarea
                    id="awd-http-put-headers"
                    v-model="httpStandardDraft.put_flag.headers_text"
                    rows="4"
                    class="ui-control awd-config-control--mono"
                    placeholder='{"Content-Type":"application/json"}'
                  />
                </span>
                <p
                  v-if="fieldErrors.http_put_headers_text"
                  class="ui-field__error awd-http-action-error"
                >
                  {{ fieldErrors.http_put_headers_text }}
                </p>
              </div>
            </div>
          </section>

          <section class="checker-action-section">
            <header class="list-heading checker-action-section__head">
              <h4 class="list-heading__title checker-action-section__title">GET Flag</h4>
              <p class="checker-action-section__hint">回读并校验当前轮 flag。</p>
            </header>
            <div class="checker-action-grid">
              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-get-method">Method</label>
                <span class="ui-control-wrap awd-http-action-control">
                  <select
                    id="awd-http-get-method"
                    v-model="httpStandardDraft.get_flag.method"
                    class="ui-control"
                  >
                    <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">
                      {{ method }}
                    </option>
                  </select>
                </span>
              </div>

              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-get-path">Path</label>
                <span
                  class="ui-control-wrap awd-http-action-control"
                  :class="{ 'is-error': !!fieldErrors.http_get_path }"
                >
                  <input
                    id="awd-http-get-path"
                    v-model="httpStandardDraft.get_flag.path"
                    type="text"
                    placeholder="/api/flag"
                    class="ui-control"
                  />
                </span>
                <p v-if="fieldErrors.http_get_path" class="ui-field__error awd-http-action-error">
                  {{ fieldErrors.http_get_path }}
                </p>
              </div>

              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-get-expected-status">状态码</label>
                <span
                  class="ui-control-wrap awd-http-action-control"
                  :class="{ 'is-error': !!fieldErrors.http_get_expected_status }"
                >
                  <input
                    id="awd-http-get-expected-status"
                    v-model.number="httpStandardDraft.get_flag.expected_status"
                    type="number"
                    min="1"
                    step="1"
                    class="ui-control"
                  />
                </span>
                <p
                  v-if="fieldErrors.http_get_expected_status"
                  class="ui-field__error awd-http-action-error"
                >
                  {{ fieldErrors.http_get_expected_status }}
                </p>
              </div>
            </div>

            <div class="checker-action-extra-grid">
              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-get-expected-substring">
                  预期片段
                </label>
                <span class="ui-control-wrap awd-http-action-control">
                  <input
                    id="awd-http-get-expected-substring"
                    v-model="httpStandardDraft.get_flag.expected_substring"
                    type="text"
                    class="ui-control awd-config-control--mono"
                    placeholder="{{FLAG}}"
                  />
                </span>
              </div>

              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-get-headers">Headers JSON</label>
                <span
                  class="ui-control-wrap awd-http-action-control"
                  :class="{ 'is-error': !!fieldErrors.http_get_headers_text }"
                >
                  <textarea
                    id="awd-http-get-headers"
                    v-model="httpStandardDraft.get_flag.headers_text"
                    rows="4"
                    class="ui-control awd-config-control--mono"
                    placeholder='{"Accept":"application/json"}'
                  />
                </span>
                <p
                  v-if="fieldErrors.http_get_headers_text"
                  class="ui-field__error awd-http-action-error"
                >
                  {{ fieldErrors.http_get_headers_text }}
                </p>
              </div>
            </div>
          </section>

          <section class="checker-action-section">
            <header class="list-heading checker-action-section__head">
              <h4 class="list-heading__title checker-action-section__title">Havoc</h4>
              <p class="checker-action-section__hint">可选动作，路径留空时视为未启用。</p>
            </header>
            <div class="checker-action-grid">
              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-havoc-method">Method</label>
                <span class="ui-control-wrap awd-http-action-control">
                  <select
                    id="awd-http-havoc-method"
                    v-model="httpStandardDraft.havoc.method"
                    class="ui-control"
                  >
                    <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">
                      {{ method }}
                    </option>
                  </select>
                </span>
              </div>

              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-havoc-path">Path</label>
                <span class="ui-control-wrap awd-http-action-control">
                  <input
                    id="awd-http-havoc-path"
                    v-model="httpStandardDraft.havoc.path"
                    type="text"
                    placeholder="/healthz"
                    class="ui-control"
                  />
                </span>
              </div>

              <div class="ui-field awd-http-action-field">
                <label class="ui-field__label" for="awd-http-havoc-expected-status"> 状态码 </label>
                <span
                  class="ui-control-wrap awd-http-action-control"
                  :class="{ 'is-error': !!fieldErrors.http_havoc_expected_status }"
                >
                  <input
                    id="awd-http-havoc-expected-status"
                    v-model.number="httpStandardDraft.havoc.expected_status"
                    type="number"
                    min="1"
                    step="1"
                    class="ui-control"
                  />
                </span>
                <p
                  v-if="fieldErrors.http_havoc_expected_status"
                  class="ui-field__error awd-http-action-error"
                >
                  {{ fieldErrors.http_havoc_expected_status }}
                </p>
              </div>
            </div>

            <div class="ui-field awd-http-action-field">
              <label class="ui-field__label" for="awd-http-havoc-headers">Headers JSON</label>
              <span
                class="ui-control-wrap awd-http-action-control"
                :class="{ 'is-error': !!fieldErrors.http_havoc_headers_text }"
              >
                <textarea
                  id="awd-http-havoc-headers"
                  v-model="httpStandardDraft.havoc.headers_text"
                  rows="4"
                  class="ui-control awd-config-control--mono"
                  placeholder='{"X-Checker":"havoc"}'
                />
              </span>
              <p
                v-if="fieldErrors.http_havoc_headers_text"
                class="ui-field__error awd-http-action-error"
              >
                {{ fieldErrors.http_havoc_headers_text }}
              </p>
            </div>
          </section>
        </template>
      </section>

      <section class="checker-config-block">
        <header class="list-heading checker-config-block__head">
          <div>
            <div class="journal-note-label">Payload Preview</div>
            <h3 class="list-heading__title checker-config-block__title">最终 JSON 预览</h3>
          </div>
          <p class="checker-config-block__hint">保存时会按下面的结构写入 `awd_checker_config`。</p>
        </header>

        <pre id="awd-challenge-config-preview" class="checker-preview">{{
          checkerPreviewText
        }}</pre>
      </section>

      <section
        v-if="mode === 'edit' && (savedValidationStateLabel || savedPreviewResult)"
        class="checker-config-block"
      >
        <header class="list-heading checker-config-block__head">
          <div>
            <div class="journal-note-label">Saved Validation</div>
            <h3 class="list-heading__title checker-config-block__title">最近一次已保存校验</h3>
          </div>
          <p class="checker-config-block__hint">
            这里显示已经写入赛事题目配置的校验状态；如果后续改了 Checker 草稿，需要重新试跑。
          </p>
        </header>

        <article class="journal-note checker-validation-card">
          <header class="list-heading checker-validation-card__top">
            <span :class="getValidationStateClass(props.draft?.awd_checker_validation_state)">
              {{ savedValidationStateLabel || '未验证' }}
            </span>
            <span
              v-if="props.draft?.awd_checker_last_preview_at"
              class="checker-validation-card__time"
            >
              {{ formatPreviewDateTime(props.draft?.awd_checker_last_preview_at) }}
            </span>
          </header>
          <p v-if="savedPreviewSummaryText" class="checker-validation-card__summary">
            {{ savedPreviewSummaryText }}
          </p>
          <p v-else class="checker-validation-card__summary">
            {{ savedPreviewEmptyText }}
          </p>
          <p v-if="savedPreviewAccessURL" class="checker-validation-card__meta">
            目标地址 {{ savedPreviewAccessURL }}
          </p>
        </article>
      </section>

      <section class="checker-config-block">
        <header class="list-heading checker-config-block__head">
          <div>
            <div class="journal-note-label">Checker Preview</div>
            <h3 class="list-heading__title checker-config-block__title">试跑 Checker</h3>
          </div>
          <p class="checker-config-block__hint">
            会按当前配置真实请求目标地址，但不会写入轮次、服务状态和排行榜数据。
          </p>
        </header>

        <div class="checker-preview-input-grid">
          <div class="ui-field awd-config-field">
            <label class="ui-field__label" for="awd-challenge-preview-access-url">
              目标访问地址（可选）
            </label>
            <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.preview_access_url }">
              <input
                id="awd-challenge-preview-access-url"
                v-model="previewForm.access_url"
                type="text"
                placeholder="http://team1.example.com:8080"
                class="ui-control"
              />
            </span>
            <p v-if="fieldErrors.preview_access_url" class="ui-field__error">
              {{ fieldErrors.preview_access_url }}
            </p>
            <p class="ui-field__hint">
              留空时系统会自动拉起预览实例并在试跑后回收；需要指定外部目标时也可以手动填写。
            </p>
          </div>

          <div class="ui-field awd-config-field">
            <label class="ui-field__label" for="awd-challenge-preview-flag">预览 Flag</label>
            <span class="ui-control-wrap">
              <input
                id="awd-challenge-preview-flag"
                v-model="previewForm.preview_flag"
                type="text"
                placeholder="flag{preview}"
                class="ui-control awd-config-control--mono"
              />
            </span>
            <p class="ui-field__hint">未绑定正式轮次时，这个值会替代 FLAG 模板变量。</p>
          </div>
        </div>

        <header class="list-heading checker-preview-toolbar">
          <p class="checker-preview-toolbar__hint">
            如果当前 checker 使用 ROUND / TEAM_ID 模板变量，试跑会固定使用 `0` 作为占位值。
          </p>
          <button
            id="awd-challenge-preview-submit"
            type="button"
            class="ui-btn ui-btn--primary"
            :disabled="previewing || !contestId"
            @click="handlePreview"
          >
            {{ previewButtonText }}
          </button>
        </header>

        <p
          v-if="previewError"
          id="awd-challenge-preview-error"
          class="journal-note checker-preview-notice checker-preview-notice--error"
        >
          {{ previewError }}
        </p>

        <p
          v-else-if="previewTokenInvalidated"
          class="journal-note checker-preview-notice checker-preview-notice--warning"
        >
          需要重新试跑，当前 Checker 草稿已经不同于最近一次试跑配置。
        </p>
        <p
          v-else-if="canAttachPreviewToken"
          class="journal-note checker-preview-notice checker-preview-notice--success"
        >
          {{ previewPendingSaveNotice }}
        </p>

        <section
          v-if="previewing"
          id="awd-challenge-preview-progress"
          class="journal-note checker-preview-progress"
          aria-live="polite"
        >
          <header class="list-heading checker-preview-progress__head">
            <div>
              <div class="journal-note-label">Preview Progress</div>
              <h4 class="list-heading__title checker-preview-progress__title">正在试跑 Checker</h4>
            </div>
            <span class="ui-badge ui-badge--pill ui-badge--soft checker-preview-progress__badge">
              {{ previewProgressAttemptText }}
            </span>
          </header>

          <p id="awd-challenge-preview-progress-status" class="checker-preview-progress__summary">
            {{ previewProgressStatusText }}
          </p>
          <p class="checker-preview-progress__hint">
            {{ previewProgressDetailText }}
          </p>

          <div class="checker-preview-progress__meta">
            <span>共 {{ AWD_CHECKER_PREVIEW_ATTEMPT_TOTAL }} 轮试跑</span>
            <span>当前耗时 {{ previewProgressElapsedText }}</span>
            <span>{{ previewProgressConnectionHint }}</span>
          </div>

          <div class="checker-preview-progress__bar" aria-hidden="true">
            <span
              class="checker-preview-progress__bar-fill"
              :style="{ width: `${previewProgressPercent}%` }"
            />
          </div>

          <ol class="checker-preview-progress__step-list">
            <li
              v-for="phase in previewProgressSteps"
              :key="phase.key"
              class="checker-preview-progress__step"
              :class="`checker-preview-progress__step--${phase.state}`"
            >
              <span class="checker-preview-progress__step-index">
                {{ phase.attempt || '·' }}
              </span>
              <div class="checker-preview-progress__step-body">
                <strong class="checker-preview-progress__step-title">
                  {{ phase.label }}
                </strong>
                <span class="checker-preview-progress__step-detail">
                  {{ phase.detail }}
                </span>
              </div>
            </li>
          </ol>
        </section>

        <section v-if="previewResult" class="checker-preview-result">
          <header class="list-heading checker-preview-result__head">
            <h4 class="list-heading__title checker-preview-result__title">试跑结果</h4>
            <div
              id="awd-challenge-preview-status"
              :class="getPreviewStatusClass(previewResult.service_status)"
            >
              {{ getPreviewStatusLabel(previewResult.service_status) }}
            </div>
          </header>

          <p id="awd-challenge-preview-summary" class="checker-preview-result__summary">
            {{ previewSummaryText }}
          </p>
          <p v-if="previewAccessURL" class="checker-preview-result__hint">
            目标地址 {{ previewAccessURL }}
          </p>
          <p v-if="previewTargetSummaryText" class="checker-preview-result__hint">
            {{ previewTargetSummaryText }}
          </p>

          <div v-if="previewActions.length > 0" class="checker-preview-action-list">
            <article
              v-for="action in previewActions"
              :id="`awd-checker-preview-action-${action.key}`"
              :key="action.key"
              class="journal-note checker-preview-action-card"
            >
              <header class="list-heading checker-preview-action-card__top">
                <h5 class="list-heading__title checker-preview-action-card__title">
                  {{ action.label }}
                </h5>
                <span class="checker-preview-action-card__state">
                  {{ getPreviewActionStateText(action) }}
                </span>
              </header>
              <strong class="checker-preview-action-card__path">
                {{ action.method || 'GET' }} {{ action.path || '/' }}
              </strong>
            </article>
          </div>

          <div v-if="previewTargets.length > 0" class="checker-preview-target-list">
            <article
              v-for="(target, index) in previewTargets"
              :key="target.access_url || index"
              class="journal-note checker-preview-target-card"
            >
              <header class="list-heading checker-preview-target-card__top">
                <strong
                  class="list-heading__title checker-preview-target-card__title checker-preview-target-card__url"
                >
                  {{ target.access_url || '未返回访问地址' }}
                </strong>
                <span class="checker-preview-target-card__state">
                  {{ getProbeStatusText(target.healthy, target.error_code, target.error) }}
                </span>
              </header>
              <p v-if="formatLatency(target.latency_ms)" class="checker-preview-target-card__hint">
                延迟 {{ formatLatency(target.latency_ms) }}
              </p>
            </article>
          </div>

          <div class="journal-note checker-preview-result__json">
            <header class="list-heading checker-preview-result__json-head">
              <h4 class="list-heading__title checker-preview-result__json-title">原始结果</h4>
            </header>
            <pre id="awd-challenge-preview-result-json" class="checker-preview">{{
              previewResultJSONText
            }}</pre>
          </div>
        </section>
      </section>
    </form>

    <template #footer>
      <div class="awd-config-drawer-footer">
        <button type="button" class="ui-btn ui-btn--secondary" @click="closeDialog">取消</button>
        <button
          id="awd-challenge-config-submit"
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="saving || previewing || (mode === 'create' && loadingChallengeCatalog)"
          @click="handleSubmit"
        >
          {{ submitButtonText }}
        </button>
      </div>
    </template>
  </SlideOverDrawer>
</template>

<style scoped>
.awd-config-field {
  --ui-field-gap: var(--space-2);
}

.awd-config-readonly {
  background: var(--color-bg-elevated);
}

.awd-config-readonly__value {
  cursor: default;
  color: var(--color-text-primary);
}

.awd-config-control--mono {
  font-family: var(--font-family-mono);
}

.awd-config-drawer-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
  width: 100%;
}

.checker-config-block {
  display: grid;
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border-default);
}

.template-snapshot-card {
  display: grid;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  background: var(--color-bg-surface);
}

.template-snapshot-card__head {
  align-items: flex-start;
}

.template-snapshot-card__title {
  color: inherit;
}

.template-snapshot-card__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.85rem;
  line-height: 1.6;
  max-width: 28rem;
}

.template-snapshot-card__summary {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.template-snapshot-chip {
  display: grid;
  gap: 0.2rem;
  min-width: 10rem;
  padding: 0.75rem 0.9rem;
  border-radius: 0.85rem;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-subtle);
}

.template-snapshot-chip__label {
  font-size: 0.72rem;
  color: var(--color-text-secondary);
}

.template-snapshot-chip__value {
  color: var(--color-text-primary);
  font-weight: 600;
  word-break: break-word;
}

.template-snapshot-card__grid {
  display: grid;
  gap: 0.75rem;
}

.template-snapshot-panel {
  display: grid;
  gap: 0.5rem;
  padding: 0.85rem;
  border-radius: 0.9rem;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-subtle);
}

.template-snapshot-panel__head {
  display: grid;
  gap: 0.2rem;
}

.template-snapshot-panel__title {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.template-snapshot-panel__hint {
  margin: 0;
  font-size: 0.78rem;
  color: var(--color-text-secondary);
}

.template-snapshot-panel__json {
  margin: 0;
  min-height: 5rem;
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--color-text-primary);
}

.checker-config-block__head {
  align-items: flex-start;
}

.checker-config-block__title {
  color: inherit;
}

.checker-config-block__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.85rem;
  line-height: 1.6;
}

.checker-preset-strip {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-preset-button {
  display: grid;
  align-items: start;
  justify-content: start;
  gap: 0.35rem;
  justify-items: start;
  min-height: auto;
  padding: 0.9rem 1rem;
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  background: var(--color-bg-surface);
  color: var(--color-text-primary);
  text-align: left;
  transition:
    border-color 0.2s ease,
    transform 0.2s ease;
}

.checker-preset-button:hover {
  border-color: var(--color-primary);
  transform: translateY(-1px);
  background: var(--color-bg-elevated);
}

.checker-preset-button__label {
  font-size: 0.92rem;
  font-weight: 600;
}

.checker-preset-button__hint {
  color: var(--color-text-secondary);
  font-size: 0.78rem;
  line-height: 1.55;
}

.checker-action-section {
  display: grid;
  gap: 1rem;
  padding: 1rem 0 0;
  border-top: 1px solid var(--color-border-subtle);
}

.checker-action-section__head {
  align-items: flex-start;
}

.checker-action-section__title {
  margin: 0;
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.checker-action-section__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.8rem;
}

.checker-action-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: 0.7fr minmax(0, 1.5fr) 0.7fr;
}

.awd-http-action-field {
  --ui-field-gap: var(--space-2);
  min-width: 0;
}

.awd-http-action-field .ui-field__label {
  font-size: 0.875rem;
  font-weight: 600;
}

.awd-http-action-control {
  width: 100%;
}

.awd-http-action-error {
  font-size: var(--font-size-12);
}

.checker-action-extra-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.checker-preview {
  margin: 0;
  padding: 1rem;
  border-radius: 1rem;
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: 0.8rem;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.checker-preview-input-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.checker-preview-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.9rem;
}

.checker-preview-toolbar__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.8rem;
  line-height: 1.6;
}

.checker-preview-notice--error {
  border: 1px solid color-mix(in srgb, var(--color-danger) 30%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  color: var(--color-danger);
}

.checker-preview-notice--warning {
  border: 1px solid color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 8%, transparent);
  color: var(--color-warning);
}

.checker-preview-notice--success {
  border: 1px solid color-mix(in srgb, var(--color-success) 30%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, transparent);
  color: var(--color-success);
}

.checker-preview-progress {
  display: grid;
  gap: 0.9rem;
  padding: 1rem 1.05rem;
  border: 1px solid color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default));
  border-radius: 1rem;
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--color-primary) 8%, transparent),
      transparent 44%
    ),
    var(--color-bg-surface);
}

.checker-preview-progress__head {
  align-items: center;
}

.checker-preview-progress__title {
  margin: 0;
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.checker-preview-progress__badge {
  --ui-badge-tone: var(--color-primary);
}

.checker-preview-progress__summary {
  margin: 0;
  color: var(--color-text-primary);
  font-size: 0.92rem;
  font-weight: 600;
  line-height: 1.6;
}

.checker-preview-progress__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.82rem;
  line-height: 1.65;
}

.checker-preview-progress__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  color: var(--color-text-secondary);
  font-size: 0.78rem;
}

.checker-preview-progress__bar {
  position: relative;
  overflow: hidden;
  height: 0.5rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-primary) 10%, var(--color-bg-subtle));
}

.checker-preview-progress__bar-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(
    90deg,
    var(--color-primary),
    color-mix(in srgb, var(--color-primary) 55%, white)
  );
  transition: width 240ms ease;
}

.checker-preview-progress__step-list {
  display: grid;
  gap: 0.7rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.checker-preview-progress__step {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0.75rem;
  align-items: start;
  padding: 0.72rem 0.8rem;
  border: 1px solid var(--color-border-default);
  border-radius: 0.9rem;
  background: var(--color-bg-elevated);
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    transform 180ms ease;
}

.checker-preview-progress__step--current {
  border-color: color-mix(in srgb, var(--color-primary) 42%, transparent);
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-elevated));
  transform: translateY(-1px);
}

.checker-preview-progress__step--done {
  border-color: color-mix(in srgb, var(--color-success) 26%, transparent);
  background: color-mix(in srgb, var(--color-success) 6%, var(--color-bg-elevated));
}

.checker-preview-progress__step-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2rem;
  height: 2rem;
  border-radius: 999px;
  background: var(--color-bg-subtle);
  color: var(--color-text-secondary);
  font-family: var(--font-family-mono);
  font-size: 0.8rem;
  font-weight: 700;
}

.checker-preview-progress__step--current .checker-preview-progress__step-index {
  background: color-mix(in srgb, var(--color-primary) 14%, transparent);
  color: var(--color-primary);
}

.checker-preview-progress__step--done .checker-preview-progress__step-index {
  background: color-mix(in srgb, var(--color-success) 14%, transparent);
  color: var(--color-success);
}

.checker-preview-progress__step-body {
  display: grid;
  gap: 0.22rem;
}

.checker-preview-progress__step-title {
  color: var(--color-text-primary);
  font-size: 0.86rem;
}

.checker-preview-progress__step-detail {
  color: var(--color-text-secondary);
  font-size: 0.78rem;
  line-height: 1.6;
}

.checker-validation-card {
  display: grid;
  gap: 0.7rem;
  padding: 1rem 1.1rem;
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  background: var(--color-bg-surface);
}

.checker-validation-card__top {
  align-items: center;
}

.checker-validation-card__time,
.checker-validation-card__meta {
  color: var(--color-text-secondary);
  font-size: 0.82rem;
}

.checker-validation-card__summary {
  margin: 0;
  color: var(--color-text-primary);
  font-size: 0.9rem;
  line-height: 1.65;
}

.checker-validation-card__meta {
  margin: 0;
  font-family: var(--font-family-mono);
  word-break: break-all;
}

.checker-validation-chip--pending {
  --ui-badge-tone: var(--color-text-muted);
}

.checker-validation-chip--passed {
  --ui-badge-tone: var(--color-success);
}

.checker-validation-chip--failed {
  --ui-badge-tone: var(--color-danger);
}

.checker-validation-chip--stale {
  --ui-badge-tone: var(--color-warning);
}

.checker-preview-result {
  display: grid;
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border-default);
}

.checker-preview-result__head {
  align-items: center;
}

.checker-preview-result__title {
  margin: 0;
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.checker-preview-result__summary {
  margin: 0;
  color: var(--color-text-primary);
  font-size: 0.92rem;
  line-height: 1.7;
}

.checker-preview-result__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.82rem;
}

.checker-preview-status--up {
  --ui-badge-tone: var(--color-success);
}

.checker-preview-status--down {
  --ui-badge-tone: var(--color-warning);
}

.checker-preview-status--compromised {
  --ui-badge-tone: var(--color-danger);
}

.checker-preview-action-list,
.checker-preview-target-list {
  display: grid;
  gap: 0.85rem;
}

.checker-preview-action-list {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-preview-action-card,
.checker-preview-target-card {
  display: grid;
  gap: 0.5rem;
  padding: 0.95rem 1rem;
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  background: var(--color-bg-surface);
}

.checker-preview-action-card__top,
.checker-preview-target-card__top {
  align-items: center;
}

.checker-preview-action-card__title {
  margin: 0;
  font-size: 0.9rem;
  color: var(--color-text-primary);
}

.checker-preview-target-card__title {
  margin: 0;
}

.checker-preview-action-card__state,
.checker-preview-target-card__state {
  color: var(--color-text-secondary);
  font-size: 0.78rem;
}

.checker-preview-action-card__path,
.checker-preview-target-card__url {
  font-family: var(--font-family-mono);
  font-size: 0.82rem;
  color: var(--color-text-primary);
  word-break: break-all;
}

.checker-preview-target-card__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.78rem;
}

.checker-preview-result__json {
  display: grid;
  gap: 0.65rem;
}

.checker-preview-result__json-head {
  align-items: center;
}

.checker-preview-result__json-title {
  margin: 0;
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

@media (max-width: 960px) {
  .checker-preview-input-grid,
  .checker-preset-strip,
  .checker-action-grid,
  .checker-action-extra-grid,
  .checker-preview-action-list {
    grid-template-columns: 1fr;
  }
}
</style>
