<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  AlertTriangle,
  ArrowLeft,
  CheckCircle2,
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
}

function removeTCPCheckerStep(index: number) {
  if (tcpStandardDraft.steps.length <= 1) return
  tcpStandardDraft.steps.splice(index, 1)
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
  <section class="awd-config-page workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero">
    <div v-if="loading" class="awd-config-page__loading">
      <AppLoading>正在同步 AWD 配置...</AppLoading>
    </div>

    <header class="awd-config-page__topbar">
      <button type="button" class="ui-btn ui-btn--ghost" @click="goBackToStudio">
        <ArrowLeft class="h-4 w-4" />
        返回工作台
      </button>
      <div class="awd-config-page__title-block">
        <div class="workspace-overline">AWD Service Config</div>
        <h1 class="workspace-page-title">AWD 服务配置</h1>
        <p class="awd-config-page__subtitle">
          {{ contest?.title || 'AWD 赛事' }}
        </p>
      </div>
      <button type="button" class="ui-btn ui-btn--secondary" :disabled="refreshing" @click="loadPage(false)">
        <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': refreshing }" />
        刷新
      </button>
    </header>

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
      <aside class="awd-config-page__services" aria-label="AWD 服务目录">
        <header class="list-heading awd-config-section-head">
          <div>
            <div class="journal-note-label">Service Directory</div>
            <h2 class="list-heading__title">服务目录</h2>
          </div>
        </header>

        <AppEmpty
          v-if="!loading && sortedServices.length === 0"
          title="暂无 AWD 服务"
          description="请先回到题目编排关联 AWD 题目。"
          icon="ShieldCheck"
          class="awd-config-page__empty"
        />

        <div v-else class="awd-service-list" role="list">
          <button
            v-for="service in sortedServices"
            :key="service.id"
            type="button"
            class="awd-service-row"
            :class="{ 'is-active': selectedServiceId === service.id }"
            role="listitem"
            @click="selectService(service)"
          >
            <span class="awd-service-row__main">
              <strong :title="service.display_name">{{ service.display_name }}</strong>
              <small>{{ service.category || '通用' }} · RANK {{ service.order }}</small>
            </span>
            <span class="awd-service-row__meta">
              <span class="awd-service-row__checker">{{ getCheckerTypeLabel(service.checker_type) }}</span>
              <span class="validation-pill" :class="service.validation_state || 'pending'">
                {{ getValidationLabel(service.validation_state) }}
              </span>
            </span>
          </button>
        </div>
      </aside>

      <section class="awd-config-page__editor">
        <AppEmpty
          v-if="!selectedService"
          title="请选择服务"
          description="从左侧目录选择一个 AWD 服务后继续配置。"
          icon="ShieldCheck"
          class="awd-config-page__empty"
        />

        <template v-else>
          <header class="awd-config-editor-head">
            <div>
              <div class="workspace-overline">Current Service</div>
              <h2>{{ selectedService.display_name }}</h2>
              <p>{{ selectedService.title || selectedService.display_name }}</p>
            </div>
            <div class="awd-config-editor-head__meta">
              <span class="engine-lock">
                <ShieldCheck class="h-4 w-4" />
                {{ getProtocolLabel(selectedCheckerType) }}
              </span>
              <span class="engine-lock">
                <Code2 class="h-4 w-4" />
                {{ getCheckerTypeLabel(selectedCheckerType) }}
              </span>
            </div>
          </header>

          <div v-if="fieldErrors.checker_type" class="awd-config-alert">
            <AlertTriangle class="h-4 w-4" />
            <span>{{ fieldErrors.checker_type }}，请先在 AWD 题库修正题目包协议与 checker 契约。</span>
          </div>

          <section class="awd-config-form-section">
            <header class="list-heading awd-config-section-head">
              <div>
                <div class="journal-note-label">Score Weight</div>
                <h3 class="list-heading__title">权重设置</h3>
              </div>
            </header>
            <div class="awd-config-score-grid">
              <label class="ui-field">
                <span class="ui-field__label">SLA 分</span>
                <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.sla_score }">
                  <input v-model.number="form.sla_score" type="number" min="0" max="5" step="1" class="ui-control" />
                </span>
                <span v-if="fieldErrors.sla_score" class="ui-field__error">{{ fieldErrors.sla_score }}</span>
              </label>
              <label class="ui-field">
                <span class="ui-field__label">防守分</span>
                <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.defense_score }">
                  <input v-model.number="form.defense_score" type="number" min="0" max="5" step="1" class="ui-control" />
                </span>
                <span v-if="fieldErrors.defense_score" class="ui-field__error">{{ fieldErrors.defense_score }}</span>
              </label>
            </div>
          </section>

          <section class="awd-config-form-section">
            <header class="list-heading awd-config-section-head">
              <div>
                <div class="journal-note-label">Checker Parameters</div>
                <h3 class="list-heading__title">{{ getCheckerTypeLabel(selectedCheckerType) }}</h3>
              </div>
            </header>

            <label v-if="selectedCheckerType === 'legacy_probe'" class="ui-field">
              <span class="ui-field__label">健康检查路径</span>
              <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.legacy_health_path }">
                <input v-model="legacyProbeDraft.health_path" type="text" class="ui-control" placeholder="/healthz" />
              </span>
              <span v-if="fieldErrors.legacy_health_path" class="ui-field__error">{{ fieldErrors.legacy_health_path }}</span>
            </label>

            <template v-else-if="selectedCheckerType === 'http_standard'">
              <div class="checker-preset-strip">
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
                class="checker-action-section"
              >
                <header class="list-heading checker-action-section__head">
                  <h4 class="list-heading__title checker-action-section__title">{{ action.title }}</h4>
                </header>
                <div class="checker-action-grid">
                  <label class="ui-field">
                    <span class="ui-field__label">Method</span>
                    <span class="ui-control-wrap">
                      <select v-model="httpStandardDraft[action.key].method" class="ui-control">
                        <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">{{ method }}</option>
                      </select>
                    </span>
                  </label>
                  <label class="ui-field">
                    <span class="ui-field__label">Path</span>
                    <span class="ui-control-wrap" :class="{ 'is-error': action.pathErrorKey ? !!fieldErrors[action.pathErrorKey] : false }">
                      <input v-model="httpStandardDraft[action.key].path" type="text" class="ui-control" />
                    </span>
                    <span v-if="action.pathErrorKey && fieldErrors[action.pathErrorKey]" class="ui-field__error">{{ fieldErrors[action.pathErrorKey] }}</span>
                  </label>
                  <label class="ui-field">
                    <span class="ui-field__label">状态码</span>
                    <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors[action.statusErrorKey] }">
                      <input v-model.number="httpStandardDraft[action.key].expected_status" type="number" min="1" step="1" class="ui-control" />
                    </span>
                    <span v-if="fieldErrors[action.statusErrorKey]" class="ui-field__error">{{ fieldErrors[action.statusErrorKey] }}</span>
                  </label>
                </div>
                <div class="checker-action-extra-grid">
                  <label class="ui-field">
                    <span class="ui-field__label">Body Template</span>
                    <span class="ui-control-wrap">
                      <textarea v-model="httpStandardDraft[action.key].body_template" rows="3" class="ui-control awd-config-control--mono" />
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
                      <textarea v-model="httpStandardDraft[action.key].headers_text" rows="3" class="ui-control awd-config-control--mono" />
                    </span>
                    <span v-if="fieldErrors[action.headersErrorKey]" class="ui-field__error">{{ fieldErrors[action.headersErrorKey] }}</span>
                  </label>
                </div>
              </section>
            </template>

            <template v-else-if="selectedCheckerType === 'tcp_standard'">
              <label class="ui-field awd-config-small-field">
                <span class="ui-field__label">总超时</span>
                <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.tcp_timeout }">
                  <input v-model.number="tcpStandardDraft.timeout_ms" type="number" min="1" max="60000" step="100" class="ui-control" />
                </span>
                <span v-if="fieldErrors.tcp_timeout" class="ui-field__error">{{ fieldErrors.tcp_timeout }}</span>
              </label>
              <span v-if="fieldErrors.tcp_steps" class="ui-field__error">{{ fieldErrors.tcp_steps }}</span>
              <section v-for="(step, index) in tcpStandardDraft.steps" :key="index" class="checker-action-section">
                <header class="list-heading checker-action-section__head">
                  <h4 class="list-heading__title checker-action-section__title">Step {{ index + 1 }}</h4>
                  <button v-if="tcpStandardDraft.steps.length > 1" type="button" class="ui-btn ui-btn--secondary" @click="removeTCPCheckerStep(index)">删除</button>
                </header>
                <div class="checker-action-extra-grid">
                  <label class="ui-field"><span class="ui-field__label">Send</span><textarea v-model="step.send" rows="3" class="ui-control awd-config-control--mono" /></label>
                  <label class="ui-field"><span class="ui-field__label">Send Template</span><textarea v-model="step.send_template" rows="3" class="ui-control awd-config-control--mono" /></label>
                  <label class="ui-field"><span class="ui-field__label">Send Hex</span><textarea v-model="step.send_hex" rows="3" class="ui-control awd-config-control--mono" /></label>
                  <label class="ui-field"><span class="ui-field__label">Expect Contains</span><textarea v-model="step.expect_contains" rows="3" class="ui-control awd-config-control--mono" /></label>
                  <label class="ui-field"><span class="ui-field__label">Expect Regex</span><input v-model="step.expect_regex" type="text" class="ui-control awd-config-control--mono" /></label>
                  <label class="ui-field"><span class="ui-field__label">Step Timeout</span><input v-model.number="step.timeout_ms" type="number" min="0" max="60000" step="100" class="ui-control" /></label>
                </div>
              </section>
              <button type="button" class="ui-btn ui-btn--secondary" @click="addTCPCheckerStep">添加步骤</button>
            </template>

            <template v-else-if="selectedCheckerType === 'script_checker'">
              <div class="checker-action-grid">
                <label class="ui-field"><span class="ui-field__label">Runtime</span><select v-model="scriptCheckerDraft.runtime" class="ui-control"><option value="python3">python3</option></select></label>
                <label class="ui-field"><span class="ui-field__label">输出格式</span><select v-model="scriptCheckerDraft.output" class="ui-control"><option value="exit_code">Exit Code</option><option value="json">JSON</option></select></label>
                <label class="ui-field"><span class="ui-field__label">入口文件</span><span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_entry }"><input v-model="scriptCheckerDraft.entry" type="text" class="ui-control" /></span><span v-if="fieldErrors.script_entry" class="ui-field__error">{{ fieldErrors.script_entry }}</span></label>
                <label class="ui-field"><span class="ui-field__label">超时时间</span><span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_timeout }"><input v-model.number="scriptCheckerDraft.timeout_sec" type="number" min="1" max="60" step="1" class="ui-control" /></span><span v-if="fieldErrors.script_timeout" class="ui-field__error">{{ fieldErrors.script_timeout }}</span></label>
              </div>
              <label class="ui-field"><span class="ui-field__label">Args</span><textarea v-model="scriptCheckerDraft.args_text" rows="3" class="ui-control awd-config-control--mono" /></label>
              <label class="ui-field"><span class="ui-field__label">Env JSON</span><span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_env_text }"><textarea v-model="scriptCheckerDraft.env_text" rows="4" class="ui-control awd-config-control--mono" /></span><span v-if="fieldErrors.script_env_text" class="ui-field__error">{{ fieldErrors.script_env_text }}</span></label>
            </template>

            <pre class="checker-json-preview" id="awd-config-json-preview">{{ checkerConfigJSON }}</pre>
          </section>

          <section class="awd-config-form-section">
            <header class="list-heading awd-config-section-head">
              <div>
                <div class="journal-note-label">Checker Preview</div>
                <h3 class="list-heading__title">试跑</h3>
              </div>
            </header>
            <div class="checker-action-grid">
              <label class="ui-field">
                <span class="ui-field__label">目标访问地址</span>
                <input v-model="previewForm.access_url" type="text" class="ui-control" placeholder="留空时由平台启动预览实例" />
              </label>
              <label class="ui-field">
                <span class="ui-field__label">预览 Flag</span>
                <input v-model="previewForm.preview_flag" type="text" class="ui-control awd-config-control--mono" />
              </label>
            </div>
            <div v-if="previewError" class="awd-config-alert">
              <AlertTriangle class="h-4 w-4" />
              <span>{{ previewError }}</span>
            </div>
            <div v-if="previewResult" class="preview-result">
              <CheckCircle2 class="h-4 w-4" />
              <span>{{ previewSummary || getCheckStatusLabel(String(previewResult.service_status)) || previewResult.service_status }}</span>
              <small v-if="previewAccessURL">{{ previewAccessURL }}</small>
            </div>
          </section>

          <footer class="awd-config-page__footer">
            <button type="button" class="ui-btn ui-btn--secondary" :disabled="previewing || saving" @click="handlePreview">
              <Play class="h-4 w-4" />
              {{ previewing ? '试跑中...' : '试跑 Checker' }}
            </button>
            <button type="button" class="ui-btn ui-btn--primary" :disabled="saving || previewing" @click="handleSave">
              <Save class="h-4 w-4" />
              {{ saving ? '保存中...' : canAttachPreviewToken ? '保存并写入试跑结果' : '保存配置' }}
            </button>
          </footer>
        </template>
      </section>
    </main>
  </section>
</template>

<style scoped>
.awd-config-page {
  position: relative;
  min-height: calc(100vh - var(--app-header-height, 4rem));
  display: flex;
  flex-direction: column;
  background: var(--color-bg-base);
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

.awd-config-page__topbar {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-5);
  padding: var(--space-5) var(--space-7);
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: var(--color-bg-surface);
}

.awd-config-page__title-block {
  min-width: 0;
}

.awd-config-page__subtitle {
  margin: var(--space-1) 0 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
}

.awd-config-page__body {
  min-height: 0;
  flex: 1;
  display: grid;
  grid-template-columns: minmax(18rem, 22rem) minmax(0, 1fr);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 54%, transparent),
      transparent 42%
    ),
    var(--color-bg-base);
}

.awd-config-page__services,
.awd-config-page__editor {
  min-width: 0;
  min-height: 0;
  overflow: auto;
  padding: var(--space-6);
}

.awd-config-page__services {
  border-right: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
}

.awd-service-list {
  display: grid;
  gap: var(--space-2);
}

.awd-service-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-2);
  width: 100%;
  border: 1px solid transparent;
  border-radius: var(--ui-control-radius);
  background: transparent;
  padding: var(--space-3);
  text-align: left;
  color: var(--color-text-primary);
  cursor: pointer;
  transition:
    background var(--ui-motion-fast),
    border-color var(--ui-motion-fast);
}

.awd-service-row:hover,
.awd-service-row:focus-visible,
.awd-service-row.is-active {
  border-color: color-mix(in srgb, var(--color-primary) 34%, transparent);
  background: color-mix(in srgb, var(--color-primary-soft) 54%, var(--color-bg-surface));
  outline: none;
}

.awd-service-row__main,
.awd-service-row__meta {
  min-width: 0;
  display: grid;
  gap: var(--space-1);
}

.awd-service-row__main strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-14);
}

.awd-service-row__main small,
.awd-service-row__checker {
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
}

.validation-pill {
  width: fit-content;
  border-radius: var(--ui-badge-radius-soft);
  padding: var(--space-1) var(--space-2);
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.validation-pill.passed {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.validation-pill.failed,
.validation-pill.stale {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.awd-config-editor-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: var(--space-5);
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
}

.awd-config-editor-head h2 {
  margin: var(--space-1) 0 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-24);
  font-weight: 900;
}

.awd-config-editor-head p {
  margin: var(--space-2) 0 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
}

.awd-config-editor-head__meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.engine-lock,
.preview-result,
.awd-config-alert {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  border-radius: var(--ui-control-radius);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-13);
}

.engine-lock {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
}

.awd-config-alert {
  margin-top: var(--space-4);
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.awd-config-form-section {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-6) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 70%, transparent);
}

.awd-config-score-grid,
.checker-action-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.checker-action-section {
  display: grid;
  gap: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid color-mix(in srgb, var(--color-border-default) 70%, transparent);
}

.checker-action-section__head {
  align-items: center;
}

.checker-action-section__title {
  font-size: var(--font-size-14);
}

.checker-action-extra-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.checker-action-extra-grid__wide {
  grid-column: 1 / -1;
}

.checker-preset-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.awd-config-small-field {
  max-width: 18rem;
}

.awd-config-control--mono,
.checker-json-preview {
  font-family: var(--font-family-mono);
}

.checker-json-preview {
  margin: 0;
  overflow: auto;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  border-radius: var(--ui-control-radius);
  background: var(--color-bg-surface);
  padding: var(--space-4);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
}

.preview-result {
  width: fit-content;
  background: var(--color-success-soft);
  color: var(--color-success);
}

.preview-result small {
  color: var(--color-text-secondary);
}

.awd-config-page__footer {
  position: sticky;
  bottom: 0;
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding: var(--space-4) 0 0;
  background: linear-gradient(
    180deg,
    transparent,
    var(--color-bg-base) var(--space-4)
  );
}

.awd-config-page__empty {
  margin: var(--space-8);
}

@media (max-width: 1023px) {
  .awd-config-page__body {
    grid-template-columns: 1fr;
  }

  .awd-config-page__services {
    border-right: 0;
    border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  }
}

@media (max-width: 767px) {
  .awd-config-page__topbar,
  .awd-config-editor-head,
  .awd-config-score-grid,
  .checker-action-grid,
  .checker-action-extra-grid {
    grid-template-columns: 1fr;
  }

  .awd-config-page__topbar {
    align-items: stretch;
  }
}
</style>
