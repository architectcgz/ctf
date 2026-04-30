<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  AlertTriangle,
  ArrowLeft,
  CheckCircle2,
  ChevronDown,
  Code2,
  Play,
  RefreshCw,
  Save,
  ShieldCheck,
} from 'lucide-vue-next'

import {
  getContest,
  listContestAWDServices,
  runContestAWDCheckerPreview,
  updateContestAWDService,
} from '@/api/admin'
import type {
  AdminContestAWDServiceData,
  AWDCheckerPreviewData,
  AWDCheckerType,
  ContestDetailData,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import ContestAwdConfigFooter from '@/components/platform/contest/ContestAwdConfigFooter.vue'
import ContestAwdConfigTopbar from '@/components/platform/contest/ContestAwdConfigTopbar.vue'
import ContestAwdDebugStation from '@/components/platform/contest/ContestAwdDebugStation.vue'
import ContestAwdEditorHeader from '@/components/platform/contest/ContestAwdEditorHeader.vue'
import ContestAwdScoreWeights from '@/components/platform/contest/ContestAwdScoreWeights.vue'
import ContestAwdServiceDirectory from '@/components/platform/contest/ContestAwdServiceDirectory.vue'
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
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useToast } from '@/composables/useToast'

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
const fieldErrors = reactive<Record<AWDCheckerFieldErrorKey | 'checker_type' | 'sla_score' | 'defense_score', string>>({
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
const selectedCheckerType = computed<AWDCheckerType | undefined>(() => selectedService.value?.checker_type)
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
  [...services.value].sort((left, right) => left.order - right.order || left.display_name.localeCompare(right.display_name))
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

const { summarizeCheckResult, getCheckStatusLabel, getPrimaryAccessURL } = useAwdCheckResultPresentation({
  formatDateTime: formatDateTime,
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
  tcpStandardDraft.steps.splice(0, tcpStandardDraft.steps.length, ...next.steps.map((step) => ({ ...step })))
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
  expandedTCPCheckerStepIndex.value = expandedTCPCheckerStepIndex.value === index ? null : index
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
  if (!Number.isInteger(form.defense_score) || form.defense_score < 0 || form.defense_score > 5) {
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
    const selectedServiceStillExists = serviceList.some((service) => service.id === selectedServiceId.value)
    const requestedServiceExists = serviceList.some((service) => service.id === requestedServiceId)
    if (!selectedServiceId.value || !selectedServiceStillExists) {
      selectedServiceId.value = requestedServiceExists ? requestedServiceId : serviceList[0]?.id || ''
    }
    syncServiceQuery(selectedServiceId.value)
    loadError.value = ''
  } catch (error) {
    if (version !== loadVersion) return
    loadError.value = error instanceof Error && error.message.trim() ? error.message : 'AWD 配置加载失败'
    toast.error(loadError.value)
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
  void router.push({ name: 'ContestEdit', params: { id: contestId.value }, query: { panel: 'awd-config' } })
}

async function handlePreview() {
  if (previewing.value || !selectedService.value || !selectedCheckerType.value || !validateConfig()) return
  previewing.value = true
  previewError.value = ''
  previewResult.value = null
  previewToken.value = ''
  previewSignature.value = ''
  try {
    const result = await runContestAWDCheckerPreview(contestId.value, {
      ...(readNumericID(selectedService.value.id) ? { service_id: readNumericID(selectedService.value.id) } : {}),
      awd_challenge_id: Number(selectedService.value.awd_challenge_id),
      checker_type: selectedCheckerType.value,
      checker_config: buildCurrentCheckerConfig(),
      ...(previewForm.access_url.trim() ? { access_url: previewForm.access_url.trim() } : {}),
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
  if (saving.value || !selectedService.value || !selectedCheckerType.value || !validateConfig()) return
  saving.value = true
  try {
    await updateContestAWDService(contestId.value, selectedService.value.id, {
      checker_type: selectedCheckerType.value,
      checker_config: buildCurrentCheckerConfig(),
      awd_sla_score: form.sla_score,
      awd_defense_score: form.defense_score,
      ...(canAttachPreviewToken.value ? { awd_checker_preview_token: previewToken.value } : {}),
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
</script>

<template>
  <section class="awd-config-page workspace-shell journal-shell journal-shell-admin">
    <div v-if="loading" class="awd-config-page__loading">
      <AppLoading>正在同步 AWD 配置...</AppLoading>
    </div>

    <ContestAwdConfigTopbar
      :contest-title="contest?.title || 'AWD 赛事'"
      :service-name="selectedService?.display_name || '请选择服务'"
      :refreshing="refreshing"
      @back="goBackToStudio"
      @refresh="loadPage(false)"
    />

    <AppEmpty
      v-if="loadError && !contest"
      title="AWD 配置加载失败"
      :description="loadError"
      icon="AlertTriangle"
      class="awd-config-page__empty"
    >
      <template #action>
        <button type="button" class="ui-btn ui-btn--primary" @click="loadPage(true)">
          重试
        </button>
      </template>
    </AppEmpty>

    <main v-else class="awd-config-page__body">
      <ContestAwdServiceDirectory
        :loading="loading"
        :services="sortedServices"
        :selected-service-id="selectedServiceId"
        :get-checker-type-label="getCheckerTypeLabel"
        :get-validation-label="getValidationLabel"
        @select="selectService"
      />

      <section class="awd-config-page__editor">
        <AppEmpty
          v-if="!selectedService"
          title="请选择服务"
          description="从左侧目录选择一个 AWD 服务后继续配置。"
          icon="ShieldCheck"
          class="awd-config-page__empty"
        />

        <template v-else>
          <ContestAwdEditorHeader
            :display-name="selectedService.display_name"
            :title="selectedService.title || selectedService.display_name"
            :protocol-label="getProtocolLabel(selectedCheckerType)"
            :checker-type-label="getCheckerTypeLabel(selectedCheckerType)"
          />

          <div v-if="fieldErrors.checker_type" class="awd-config-alert">
            <AlertTriangle class="h-4 w-4" />
            <span>{{ fieldErrors.checker_type }}，请先在 AWD 题库修正题目包协议与 checker 契约。</span>
          </div>

          <ContestAwdScoreWeights
            v-model:sla-score="form.sla_score"
            v-model:defense-score="form.defense_score"
            :sla-error="fieldErrors.sla_score"
            :defense-error="fieldErrors.defense_score"
          />

          <section class="awd-config-form-section awd-config-card awd-config-card--canvas">
            <header class="list-heading awd-config-section-head">
              <div>
                <div class="journal-note-label">Checker Parameters</div>
                <h3 class="list-heading__title">{{ getCheckerTypeLabel(selectedCheckerType) }}</h3>
              </div>
              <span class="awd-config-section-tag">配置画布</span>
            </header>

            <label v-if="selectedCheckerType === 'legacy_probe'" class="ui-field">
              <span class="ui-field__label">健康检查路径</span>
              <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.legacy_health_path }">
                <input v-model="legacyProbeDraft.health_path" type="text" class="ui-control" placeholder="/healthz" />
              </span>
              <span v-if="fieldErrors.legacy_health_path" class="ui-field__error">{{ fieldErrors.legacy_health_path }}</span>
            </label>

            <template v-else-if="selectedCheckerType === 'http_standard'">
              <div class="checker-preset-strip checker-preset-strip--compact">
                <button
                  v-for="preset in AWD_HTTP_STANDARD_PRESETS"
                  :key="preset.id"
                  type="button"
                  class="ui-btn ui-btn--secondary checker-preset-button"
                  @click="applyHTTPPreset(preset.id)"
                >
                  {{ preset.label }}
                </button>
              </div>

              <section
                v-for="action in httpActionSections"
                :key="action.key"
                class="checker-action-section checker-action-section--panel"
              >
                <header class="list-heading checker-action-section__head">
                  <div class="checker-action-section__heading">
                    <h4 class="list-heading__title checker-action-section__title">{{ action.title }}</h4>
                    <span class="checker-action-section__hint">动作配置</span>
                  </div>
                </header>
                <div class="checker-action-grid checker-action-grid--http">
                  <label class="ui-field checker-field checker-field--method">
                    <span class="ui-field__label">Method</span>
                    <span class="ui-control-wrap">
                      <select v-model="httpStandardDraft[action.key].method" class="ui-control">
                        <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">{{ method }}</option>
                      </select>
                    </span>
                  </label>
                  <label class="ui-field checker-field checker-field--path">
                    <span class="ui-field__label">Path</span>
                    <span class="ui-control-wrap" :class="{ 'is-error': action.pathErrorKey ? !!fieldErrors[action.pathErrorKey] : false }">
                      <input v-model="httpStandardDraft[action.key].path" type="text" class="ui-control" />
                    </span>
                    <span v-if="action.pathErrorKey && fieldErrors[action.pathErrorKey]" class="ui-field__error">{{ fieldErrors[action.pathErrorKey] }}</span>
                  </label>
                  <label class="ui-field checker-field checker-field--status">
                    <span class="ui-field__label">状态码</span>
                    <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors[action.statusErrorKey] }">
                      <input v-model.number="httpStandardDraft[action.key].expected_status" type="number" min="1" step="1" class="ui-control" />
                    </span>
                    <span v-if="fieldErrors[action.statusErrorKey]" class="ui-field__error">{{ fieldErrors[action.statusErrorKey] }}</span>
                  </label>
                </div>
                <div class="checker-action-extra-grid checker-action-extra-grid--http">
                  <label class="ui-field checker-field checker-field--wide">
                    <span class="ui-field__label">Body Template</span>
                    <span class="ui-control-wrap">
                      <textarea v-model="httpStandardDraft[action.key].body_template" rows="2" class="ui-control awd-config-control--mono" />
                    </span>
                  </label>
                  <label class="ui-field">
                    <span class="ui-field__label">Expected Substring</span>
                    <span class="ui-control-wrap">
                      <input v-model="httpStandardDraft[action.key].expected_substring" type="text" class="ui-control awd-config-control--mono" />
                    </span>
                  </label>
                  <label class="ui-field checker-action-extra-grid__wide">
                    <span class="ui-field__label">Headers JSON</span>
                    <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors[action.headersErrorKey] }">
                      <textarea v-model="httpStandardDraft[action.key].headers_text" rows="2" class="ui-control awd-config-control--mono" />
                    </span>
                    <span v-if="fieldErrors[action.headersErrorKey]" class="ui-field__error">{{ fieldErrors[action.headersErrorKey] }}</span>
                  </label>
                </div>
              </section>
            </template>

            <template v-else-if="selectedCheckerType === 'tcp_standard'">
              <div class="checker-toolbar">
                <label class="ui-field awd-config-small-field">
                  <span class="ui-field__label">总超时</span>
                  <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.tcp_timeout }">
                    <input v-model.number="tcpStandardDraft.timeout_ms" type="number" min="1" max="60000" step="100" class="ui-control" />
                  </span>
                  <span v-if="fieldErrors.tcp_timeout" class="ui-field__error">{{ fieldErrors.tcp_timeout }}</span>
                </label>
                <button type="button" class="ui-btn ui-btn--secondary" @click="addTCPCheckerStep">添加步骤</button>
              </div>
              <span v-if="fieldErrors.tcp_steps" class="ui-field__error">{{ fieldErrors.tcp_steps }}</span>
              <section
                v-for="(step, index) in tcpStandardDraft.steps"
                :key="index"
                class="checker-action-section checker-action-section--panel checker-action-section--tcp"
                :class="{ 'is-collapsed': expandedTCPCheckerStepIndex !== index }"
              >
                <header class="list-heading checker-action-section__head">
                  <button
                    type="button"
                    class="checker-step-toggle"
                    :aria-expanded="expandedTCPCheckerStepIndex === index"
                    @click="toggleTCPCheckerStep(index)"
                  >
                    <span class="checker-action-section__heading">
                      <span class="list-heading__title checker-action-section__title">Step {{ index + 1 }}</span>
                      <span class="checker-action-section__hint">{{ summarizeTCPCheckerStep(step) }}</span>
                    </span>
                    <ChevronDown class="h-4 w-4 checker-step-toggle__icon" />
                  </button>
                  <button v-if="tcpStandardDraft.steps.length > 1" type="button" class="ui-btn ui-btn--secondary" @click="removeTCPCheckerStep(index)">删除</button>
                </header>
                <div v-show="expandedTCPCheckerStepIndex === index" class="checker-action-extra-grid checker-action-extra-grid--tcp">
                  <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Send</span><span class="ui-control-wrap"><textarea v-model="step.send" rows="2" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Send Template</span><span class="ui-control-wrap"><textarea v-model="step.send_template" rows="2" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field"><span class="ui-field__label">Send Hex</span><span class="ui-control-wrap"><textarea v-model="step.send_hex" rows="2" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Expect Contains</span><span class="ui-control-wrap"><textarea v-model="step.expect_contains" rows="2" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field"><span class="ui-field__label">Expect Regex</span><span class="ui-control-wrap"><input v-model="step.expect_regex" type="text" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field"><span class="ui-field__label">Step Timeout</span><span class="ui-control-wrap"><input v-model.number="step.timeout_ms" type="number" min="0" max="60000" step="100" class="ui-control" /></span></label>
                </div>
              </section>
            </template>

            <template v-else-if="selectedCheckerType === 'script_checker'">
              <div class="checker-action-grid checker-action-grid--script-meta">
                <label class="ui-field"><span class="ui-field__label">Runtime</span><span class="ui-control-wrap"><select v-model="scriptCheckerDraft.runtime" class="ui-control"><option value="python3">python3</option></select></span></label>
                <label class="ui-field"><span class="ui-field__label">输出格式</span><span class="ui-control-wrap"><select v-model="scriptCheckerDraft.output" class="ui-control"><option value="exit_code">Exit Code</option><option value="json">JSON</option></select></span></label>
                <label class="ui-field"><span class="ui-field__label">超时时间</span><span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_timeout }"><input v-model.number="scriptCheckerDraft.timeout_sec" type="number" min="1" max="60" step="1" class="ui-control" /></span><span v-if="fieldErrors.script_timeout" class="ui-field__error">{{ fieldErrors.script_timeout }}</span></label>
              </div>
              <label class="ui-field">
                <span class="ui-field__label">入口文件</span>
                <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_entry }"><input v-model="scriptCheckerDraft.entry" type="text" class="ui-control" /></span>
                <span v-if="fieldErrors.script_entry" class="ui-field__error">{{ fieldErrors.script_entry }}</span>
              </label>
              <div class="checker-action-extra-grid checker-action-extra-grid--script">
                <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Args</span><span class="ui-control-wrap"><textarea v-model="scriptCheckerDraft.args_text" rows="3" class="ui-control awd-config-control--mono" /></span></label>
                <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Env JSON</span><span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_env_text }"><textarea v-model="scriptCheckerDraft.env_text" rows="3" class="ui-control awd-config-control--mono" /></span><span v-if="fieldErrors.script_env_text" class="ui-field__error">{{ fieldErrors.script_env_text }}</span></label>
              </div>
            </template>
          </section>

          <ContestAwdDebugStation
            v-model:access-url="previewForm.access_url"
            v-model:preview-flag="previewForm.preview_flag"
            :checker-config-json="checkerConfigJSON"
            :previewing="previewing"
            :preview-result="previewResult"
            :preview-error="previewError"
            :preview-access-url="previewAccessURL"
            :preview-summary="previewSummary"
            :get-check-status-label="getCheckStatusLabel"
          />

          <ContestAwdConfigFooter
            :previewing="previewing"
            :saving="saving"
            :preview-error="previewError"
            :preview-result="previewResult"
            :can-attach-preview-token="canAttachPreviewToken"
            @preview="handlePreview"
            @save="handleSave"
          />
        </template>
      </section>
    </main>
  </section>
</template>

<style scoped>
.awd-config-page {
  --awd-card-radius: 0.75rem;
  --awd-card-border: color-mix(in srgb, var(--color-border-default) 80%, transparent);
  --awd-card-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --awd-card-subtle: color-mix(in srgb, var(--color-bg-surface) 72%, var(--color-bg-base));
  --awd-card-shadow: 0 0.85rem 2rem color-mix(in srgb, var(--color-shadow-soft) 22%, transparent);
  --ui-control-background: color-mix(
    in srgb,
    var(--color-bg-elevated) 62%,
    var(--color-bg-surface)
  );
  --ui-control-border: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --ui-control-color: var(--color-text-primary);
  --ui-control-placeholder: color-mix(in srgb, var(--color-text-muted) 86%, transparent);
  --ui-control-focus-border: color-mix(in srgb, var(--color-primary) 58%, var(--color-border-default));
  --ui-control-focus-background: color-mix(
    in srgb,
    var(--color-bg-surface) 76%,
    var(--color-bg-elevated)
  );
  --ui-control-focus-shadow: 0 0 0 0.2rem color-mix(in srgb, var(--color-primary) 16%, transparent);
  position: relative;
  min-height: calc(100vh - var(--app-header-height, 4rem));
  max-height: calc(100vh - var(--app-header-height, 4rem));
  display: flex;
  flex-direction: column;
  background: var(--color-bg-base);
  overflow: hidden;
}

.awd-config-page__loading {
  position: absolute;
  inset: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, var(--color-bg-base) 82%, transparent);
}

.awd-config-page__body {
  min-height: 0;
  height: calc(100vh - var(--app-header-height, 4rem) - 3.5rem);
  flex: 1;
  display: grid;
  grid-template-columns: minmax(17rem, 20rem) minmax(0, 1fr);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 54%, transparent),
      transparent 42%
    ),
    var(--color-bg-base);
}

.awd-config-page__editor {
  min-width: 0;
  min-height: 0;
  overflow: auto;
  padding: var(--space-5);
}

.awd-config-page__editor {
  display: grid;
  align-content: start;
  gap: var(--space-5);
}

.awd-config-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.awd-config-alert {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  border-radius: var(--ui-control-radius);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-13);
}

.awd-config-alert {
  margin-top: var(--space-4);
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.awd-config-form-section {
  display: grid;
  gap: var(--space-3);
}

.awd-config-card {
  padding: var(--space-4);
  border: 1px solid var(--awd-card-border);
  border-radius: var(--awd-card-radius);
  background: var(--awd-card-surface);
  box-shadow: var(--awd-card-shadow);
}

.awd-config-card--compact {
  gap: var(--space-2);
}

.awd-config-card--canvas {
  gap: var(--space-4);
}

.awd-config-section-tag {
  flex: none;
  border-radius: var(--ui-badge-radius-soft);
  padding: var(--space-1) var(--space-2);
  background: color-mix(in srgb, var(--color-primary-soft) 55%, var(--color-bg-surface));
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 700;
}

.checker-action-grid {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-toolbar {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.checker-action-section {
  display: grid;
  gap: var(--space-3);
  padding-top: var(--space-3);
  border-top: 1px solid color-mix(in srgb, var(--color-border-default) 70%, transparent);
}

.checker-action-section--panel {
  padding: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: calc(var(--awd-card-radius) - 0.125rem);
  background: var(--awd-card-subtle);
  box-shadow: 0 0.45rem 1rem color-mix(in srgb, var(--color-shadow-soft) 12%, transparent);
}

.checker-action-section--tcp.is-collapsed {
  gap: 0;
  padding-block: var(--space-2);
}

.checker-action-section--panel :deep(.ui-control-wrap) {
  border: 1px solid var(--ui-control-border);
  background: var(--ui-control-background);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--color-text-primary) 5%, transparent);
}

.checker-action-section--panel :deep(.ui-control-wrap:focus-within) {
  border-color: var(--ui-control-focus-border);
  background: var(--ui-control-focus-background);
  box-shadow:
    var(--ui-control-focus-shadow),
    inset 0 1px 0 color-mix(in srgb, var(--color-text-primary) 7%, transparent);
}

.checker-action-section--panel :deep(.ui-control) {
  background: transparent;
}

.checker-action-section--panel :deep(.ui-control) {
  min-height: 2.25rem;
}

.checker-action-section--panel textarea.ui-control {
  min-height: 3.5rem;
  line-height: 1.4;
}

.checker-action-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.checker-step-toggle {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  border: 0;
  background: transparent;
  padding: 0;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.checker-step-toggle:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 42%, transparent);
  outline-offset: var(--space-1);
}

.checker-step-toggle__icon {
  flex: none;
  color: var(--color-text-secondary);
  transition: transform var(--ui-motion-fast);
}

.checker-step-toggle[aria-expanded='true'] .checker-step-toggle__icon {
  transform: rotate(180deg);
}

.checker-action-section__heading {
  min-width: 0;
  display: grid;
  gap: 0;
}

.checker-action-section__title {
  font-size: var(--font-size-14);
}

.checker-action-section__hint {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
}

.checker-action-extra-grid {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-action-extra-grid__wide {
  grid-column: span 2;
}

.checker-action-grid--http {
  grid-template-columns: minmax(6.5rem, 8rem) minmax(0, 1fr) minmax(7rem, 8.5rem);
}

.checker-action-extra-grid--http {
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.85fr) minmax(0, 1fr);
}

.checker-action-extra-grid--tcp {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.checker-action-grid--script-meta,
.checker-action-grid--preview {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-action-extra-grid--script {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.checker-field--method,
.checker-field--status {
  min-width: 0;
}

.checker-field--path {
  min-width: 0;
}

.checker-field--wide {
  grid-column: span 2;
}

.checker-preset-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.checker-preset-strip--compact {
  margin-bottom: var(--space-1);
}

.awd-config-small-field {
  max-width: 18rem;
}

.awd-config-page :deep(.ui-field) {
  gap: var(--space-1);
}

.awd-config-page :deep(.ui-field__label) {
  font-size: var(--font-size-12);
}

.awd-config-page :deep(.ui-control) {
  min-height: 2.5rem;
}

.awd-config-page textarea.ui-control {
  min-height: 4.5rem;
  resize: vertical;
}

.awd-config-control--mono {
  font-family: var(--font-family-mono);
}

.awd-config-page__empty {
  margin: var(--space-8);
}

@media (max-width: 1023px) {
  .awd-config-page__body {
    grid-template-columns: 1fr;
  }

  .checker-action-grid,
  .checker-action-extra-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .checker-action-grid--http,
  .checker-action-grid--preview,
  .checker-action-extra-grid--http,
  .checker-action-extra-grid--tcp,
  .checker-action-extra-grid--script {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .checker-action-extra-grid__wide {
    grid-column: 1 / -1;
  }

  .checker-field--wide {
    grid-column: 1 / -1;
  }
}

@media (max-width: 767px) {
  .checker-action-grid,
  .checker-action-extra-grid {
    grid-template-columns: 1fr;
  }

  .checker-action-extra-grid__wide {
    grid-column: auto;
  }

  .checker-field--wide {
    grid-column: auto;
  }

  .checker-action-grid--http,
  .checker-action-grid--script-meta,
  .checker-action-grid--preview,
  .checker-action-extra-grid--http,
  .checker-action-extra-grid--tcp,
  .checker-action-extra-grid--script {
    grid-template-columns: 1fr;
  }
}
</style>
