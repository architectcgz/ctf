<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'

import { runContestAWDCheckerPreview } from '@/api/admin'
import type {
  AdminChallengeListItem,
  AdminContestChallengeData,
  AWDCheckerPreviewData,
  AWDCheckerType,
  AWDTeamServiceData,
} from '@/api/contracts'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
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

type DialogMode = 'create' | 'edit'

const props = defineProps<{
  contestId?: string | null
  open: boolean
  mode: DialogMode
  challengeOptions: AdminChallengeListItem[]
  existingChallengeIds: string[]
  draft?: AdminContestChallengeData | null
  loadingChallengeCatalog: boolean
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      challenge_id: number
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

function createFieldErrorState() {
  return {
    challenge_id: '',
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
  props.mode === 'create' ? '新增 AWD 题目' : '编辑 AWD 题目配置'
)

const selectableChallenges = computed(() =>
  props.challengeOptions.filter(
    (item) => props.mode === 'edit' || !props.existingChallengeIds.includes(item.id)
  )
)

const activeChallengeLabel = computed(() => {
  if (props.mode === 'edit') {
    const title = props.draft?.title?.trim() || `Challenge #${props.draft?.challenge_id || ''}`
    return title
  }
  return (
    selectableChallenges.value.find((item) => item.id === form.challenge_id)?.title || '请选择题目'
  )
})

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

const currentCheckerSignature = computed(() =>
  JSON.stringify({
    challenge_id: form.challenge_id,
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
        : selectableChallenges.value[0]?.id || ''
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

function clearErrors() {
  for (const key of Object.keys(fieldErrors) as Array<keyof typeof fieldErrors>) {
    fieldErrors[key] = ''
  }
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

  if (props.mode === 'create' && !form.challenge_id) {
    fieldErrors.challenge_id = '请选择题目'
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

  if (!form.challenge_id) {
    fieldErrors.challenge_id = '请选择题目'
  }
  if (!previewForm.access_url.trim()) {
    fieldErrors.preview_access_url = '请输入目标访问地址'
  }

  const checkerResult = buildCheckerConfigResult(true)

  for (const [key, value] of Object.entries(checkerResult.errors)) {
    if (value) {
      fieldErrors[key as keyof typeof fieldErrors] = value
    }
  }

  return (
    !fieldErrors.challenge_id &&
    !fieldErrors.preview_access_url &&
    AWD_CHECKER_FIELD_ERROR_KEYS.every((key) => !fieldErrors[key])
  )
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

  try {
    const result = await runContestAWDCheckerPreview(props.contestId, {
      challenge_id: Number(form.challenge_id),
      checker_type: form.awd_checker_type,
      checker_config: buildCheckerConfig(),
      access_url: previewForm.access_url.trim(),
      preview_flag: previewForm.preview_flag.trim() || undefined,
    })
    previewResult.value = result
    previewToken.value = result.preview_token || ''
    previewSignature.value = currentCheckerSignature.value
    previewTokenInvalidated.value = false
  } catch (error) {
    previewError.value = error instanceof Error ? error.message : '试跑失败，请稍后重试。'
  } finally {
    previewing.value = false
  }
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
  if (!validate()) {
    return
  }

  const payload = {
    challenge_id: Number(form.challenge_id),
    points: form.points,
    order: form.order,
    is_visible: form.is_visible === 'true',
    awd_checker_type: form.awd_checker_type,
    awd_checker_config: buildCheckerConfig(),
    awd_sla_score: form.awd_sla_score,
    awd_defense_score: form.awd_defense_score,
  } as {
    challenge_id: number
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
  <AdminSurfaceModal
    :open="open"
    :title="dialogTitle"
    :subtitle="
      mode === 'create'
        ? '先完成题目关联和 Checker 草稿配置，保存后即可继续赛前试跑。'
        : '统一维护赛事题目的 Checker、分值、顺序和试跑结果。'
    "
    eyebrow="AWD Operations"
    width="57.5rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form class="space-y-6" @submit.prevent="handleSubmit">
      <div class="ui-field awd-config-field">
        <label class="ui-field__label" for="awd-challenge-config-challenge">题目</label>
        <template v-if="mode === 'create'">
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.challenge_id }">
            <select
              id="awd-challenge-config-challenge"
              v-model="form.challenge_id"
              class="ui-control"
            >
              <option value="" disabled>
                {{ loadingChallengeCatalog ? '正在加载题库...' : '请选择题目' }}
              </option>
              <option
                v-for="challenge in selectableChallenges"
                :key="challenge.id"
                :value="challenge.id"
              >
                {{ challenge.title }}
              </option>
            </select>
          </span>
        </template>
        <span v-else class="ui-control-wrap awd-config-readonly">
          <span class="ui-control awd-config-readonly__value">
            {{ activeChallengeLabel }}
          </span>
        </span>
        <p v-if="fieldErrors.challenge_id" class="ui-field__error">
          {{ fieldErrors.challenge_id }}
        </p>
      </div>

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
            <select
              id="awd-challenge-config-visible"
              v-model="form.is_visible"
              class="ui-control"
            >
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
        <header class="checker-config-block__head">
          <div>
            <div class="journal-note-label">Checker Config</div>
            <h3 class="checker-config-block__title">
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
            <header class="checker-action-section__head">
              <div class="journal-note-label">PUT Flag</div>
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
            <header class="checker-action-section__head">
              <div class="journal-note-label">GET Flag</div>
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
            <header class="checker-action-section__head">
              <div class="journal-note-label">Havoc</div>
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
                <label class="ui-field__label" for="awd-http-havoc-expected-status">
                  状态码
                </label>
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
        <header class="checker-config-block__head">
          <div>
            <div class="journal-note-label">Payload Preview</div>
            <h3 class="checker-config-block__title">最终 JSON 预览</h3>
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
        <header class="checker-config-block__head">
          <div>
            <div class="journal-note-label">Saved Validation</div>
            <h3 class="checker-config-block__title">最近一次已保存校验</h3>
          </div>
          <p class="checker-config-block__hint">
            这里显示已经写入赛事题目配置的校验状态；如果后续改了 Checker 草稿，需要重新试跑。
          </p>
        </header>

        <article class="checker-validation-card">
          <div class="checker-validation-card__top">
            <span :class="getValidationStateClass(props.draft?.awd_checker_validation_state)">
              {{ savedValidationStateLabel || '未验证' }}
            </span>
            <span
              v-if="props.draft?.awd_checker_last_preview_at"
              class="checker-validation-card__time"
            >
              {{ formatPreviewDateTime(props.draft?.awd_checker_last_preview_at) }}
            </span>
          </div>
          <p v-if="savedPreviewSummaryText" class="checker-validation-card__summary">
            {{ savedPreviewSummaryText }}
          </p>
          <p v-else class="checker-validation-card__summary">当前配置还没有保存过试跑结果。</p>
          <p v-if="savedPreviewAccessURL" class="checker-validation-card__meta">
            目标地址 {{ savedPreviewAccessURL }}
          </p>
        </article>
      </section>

      <section class="checker-config-block">
        <header class="checker-config-block__head">
          <div>
            <div class="journal-note-label">Checker Preview</div>
            <h3 class="checker-config-block__title">试跑 Checker</h3>
          </div>
          <p class="checker-config-block__hint">
            会按当前配置真实请求目标地址，但不会写入轮次、服务状态和排行榜数据。
          </p>
        </header>

        <div class="checker-preview-input-grid">
          <div class="ui-field awd-config-field">
            <label class="ui-field__label" for="awd-challenge-preview-access-url">
              目标访问地址
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
            <p class="ui-field__hint">
              未绑定正式轮次时，这个值会替代 FLAG 模板变量。
            </p>
          </div>
        </div>

        <div class="checker-preview-toolbar">
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
            {{ previewing ? '试跑中...' : '试跑 Checker' }}
          </button>
        </div>

        <p
          v-if="previewError"
          id="awd-challenge-preview-error"
          class="checker-preview-notice checker-preview-notice--error"
        >
          {{ previewError }}
        </p>

        <p
          v-else-if="previewTokenInvalidated"
          class="checker-preview-notice checker-preview-notice--warning"
        >
          需要重新试跑，当前 Checker 草稿已经不同于最近一次试跑配置。
        </p>
        <p
          v-else-if="canAttachPreviewToken"
          class="checker-preview-notice checker-preview-notice--success"
        >
          当前保存会附带最近一次试跑结果。
        </p>

        <section v-if="previewResult" class="checker-preview-result">
          <header class="checker-preview-result__head">
            <div class="journal-note-label">试跑结果</div>
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
              <div class="checker-preview-action-card__top">
                <div class="journal-note-label">{{ action.label }}</div>
                <span class="checker-preview-action-card__state">
                  {{ getPreviewActionStateText(action) }}
                </span>
              </div>
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
              <div class="checker-preview-target-card__top">
                <strong class="checker-preview-target-card__url">
                  {{ target.access_url || '未返回访问地址' }}
                </strong>
                <span class="checker-preview-target-card__state">
                  {{ getProbeStatusText(target.healthy, target.error_code, target.error) }}
                </span>
              </div>
              <p v-if="formatLatency(target.latency_ms)" class="checker-preview-target-card__hint">
                延迟 {{ formatLatency(target.latency_ms) }}
              </p>
            </article>
          </div>

          <div class="checker-preview-result__json">
            <div class="journal-note-label">原始结果</div>
            <pre id="awd-challenge-preview-result-json" class="checker-preview">{{
              previewResultJSONText
            }}</pre>
          </div>
        </section>
      </section>
    </form>

    <template #footer>
      <div class="awd-config-dialog__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="awd-challenge-config-submit"
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="saving || (mode === 'create' && loadingChallengeCatalog)"
          @click="handleSubmit"
        >
          {{ saving ? '保存中...' : mode === 'create' ? '新增题目' : '保存配置' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.awd-config-field {
  --ui-field-gap: var(--space-2);
}

.awd-config-readonly {
  background: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-elevated));
}

.awd-config-readonly__value {
  cursor: default;
}

.awd-config-control--mono {
  font-family: var(--font-family-mono);
}

.awd-config-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

.checker-config-block {
  display: grid;
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 74%, transparent);
}

.checker-config-block__head {
  display: grid;
  gap: 0.45rem;
}

.checker-config-block__title {
  margin: 0;
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
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 88%, var(--color-bg-surface-elevated));
  color: var(--color-text-primary);
  text-align: left;
  transition:
    border-color 0.2s ease,
    transform 0.2s ease;
}

.checker-preset-button:hover {
  border-color: var(--color-primary);
  transform: translateY(-1px);
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
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 70%, transparent);
}

.checker-action-section__head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
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
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 88%, var(--color-bg-surface-elevated));
  color: var(--color-text-primary);
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
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

.checker-preview-notice {
  margin: 0;
  padding: 0.85rem 1rem;
  border-radius: 1rem;
  font-size: 0.85rem;
}

.checker-preview-notice--error {
  border: 1px solid color-mix(in srgb, var(--color-danger) 30%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  color: var(--color-danger);
}

.checker-preview-notice--warning {
  border: 1px solid color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 8%, transparent);
  color: color-mix(in srgb, var(--color-warning) 82%, var(--color-text-primary));
}

.checker-preview-notice--success {
  border: 1px solid color-mix(in srgb, var(--color-success) 30%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, transparent);
  color: var(--color-success);
}

.checker-validation-card {
  display: grid;
  gap: 0.7rem;
  padding: 1rem 1.1rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 88%, var(--color-bg-surface-elevated));
}

.checker-validation-card__top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
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
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
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
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 70%, transparent);
}

.checker-preview-result__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
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
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 88%, var(--color-bg-surface-elevated));
}

.checker-preview-action-card__top,
.checker-preview-target-card__top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
}

.checker-preview-action-card__state,
.checker-preview-target-card__state {
  color: var(--color-text-secondary);
  font-size: 0.78rem;
}

.checker-preview-action-card__path,
.checker-preview-target-card__url {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
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
