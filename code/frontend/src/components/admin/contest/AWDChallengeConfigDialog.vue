<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import type {
  AdminChallengeListItem,
  AdminContestChallengeData,
  AWDCheckerType,
} from '@/api/contracts'
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
    selectableChallenges.value.find((item) => item.id === form.challenge_id)?.title ||
    '请选择题目'
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

    form.challenge_id =
      props.mode === 'edit' ? props.draft?.challenge_id || '' : selectableChallenges.value[0]?.id || ''
    form.points = props.draft?.points ?? 100
    form.order = props.draft?.order ?? 0
    form.is_visible = props.draft?.is_visible === false ? 'false' : 'true'
    form.awd_checker_type = props.draft?.awd_checker_type || 'legacy_probe'
    form.awd_sla_score = props.draft?.awd_sla_score ?? 0
    form.awd_defense_score = props.draft?.awd_defense_score ?? 0
    assignLegacyProbeDraft(createLegacyProbeDraft(props.draft?.awd_checker_config))
    assignHTTPStandardDraft(createHTTPStandardDraft(props.draft?.awd_checker_config))
    clearErrors()
  },
  { immediate: true }
)

watch(
  () => [props.open, props.mode, selectableChallenges.value.map((item) => item.id).join(',')] as const,
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

  const checkerResult =
    form.awd_checker_type === 'http_standard'
      ? buildHTTPStandardCheckerConfig(httpStandardDraft, true)
      : buildLegacyProbeCheckerConfig(legacyProbeDraft)

  for (const [key, value] of Object.entries(checkerResult.errors)) {
    if (value) {
      fieldErrors[key as keyof typeof fieldErrors] = value
    }
  }

  return Object.values(fieldErrors).every((value) => !value)
}

function buildCheckerConfig() {
  return form.awd_checker_type === 'http_standard'
    ? buildHTTPStandardCheckerConfig(httpStandardDraft, true).config
    : buildLegacyProbeCheckerConfig(legacyProbeDraft).config
}

function handleSubmit() {
  if (!validate()) {
    return
  }

  emit('save', {
    challenge_id: Number(form.challenge_id),
    points: form.points,
    order: form.order,
    is_visible: form.is_visible === 'true',
    awd_checker_type: form.awd_checker_type,
    awd_checker_config: buildCheckerConfig(),
    awd_sla_score: form.awd_sla_score,
    awd_defense_score: form.awd_defense_score,
  })
}
</script>

<template>
  <ElDialog
    :model-value="open"
    :title="dialogTitle"
    width="920px"
    @close="closeDialog"
    @update:model-value="emit('update:open', $event)"
  >
    <form class="space-y-6" @submit.prevent="handleSubmit">
      <div class="space-y-2">
        <label
          class="text-sm font-medium text-[var(--color-text-primary)]"
          for="awd-challenge-config-challenge"
        >
          题目
        </label>
        <template v-if="mode === 'create'">
          <select
            id="awd-challenge-config-challenge"
            v-model="form.challenge_id"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option value="" disabled>{{ loadingChallengeCatalog ? '正在加载题库...' : '请选择题目' }}</option>
            <option
              v-for="challenge in selectableChallenges"
              :key="challenge.id"
              :value="challenge.id"
            >
              {{ challenge.title }}
            </option>
          </select>
        </template>
        <div
          v-else
          class="rounded-xl border border-border bg-surface-alt/40 px-4 py-3 text-sm text-[var(--color-text-primary)]"
        >
          {{ activeChallengeLabel }}
        </div>
        <p v-if="fieldErrors.challenge_id" class="text-xs text-[var(--color-danger)]">
          {{ fieldErrors.challenge_id }}
        </p>
      </div>

      <div class="grid gap-4 sm:grid-cols-3">
        <div class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            for="awd-challenge-config-points"
          >
            分值
          </label>
          <input
            id="awd-challenge-config-points"
            v-model.number="form.points"
            type="number"
            min="1"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.points" class="text-xs text-[var(--color-danger)]">
            {{ fieldErrors.points }}
          </p>
        </div>

        <div class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            for="awd-challenge-config-order"
          >
            顺序
          </label>
          <input
            id="awd-challenge-config-order"
            v-model.number="form.order"
            type="number"
            min="0"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.order" class="text-xs text-[var(--color-danger)]">
            {{ fieldErrors.order }}
          </p>
        </div>

        <div class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            for="awd-challenge-config-visible"
          >
            可见性
          </label>
          <select
            id="awd-challenge-config-visible"
            v-model="form.is_visible"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option value="true">可见</option>
            <option value="false">隐藏</option>
          </select>
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-3">
        <div class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            for="awd-challenge-config-checker-type"
          >
            Checker 类型
          </label>
          <select
            id="awd-challenge-config-checker-type"
            v-model="form.awd_checker_type"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option value="legacy_probe">基础探活</option>
            <option value="http_standard">HTTP 标准 Checker</option>
          </select>
        </div>

        <div class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            for="awd-challenge-config-sla-score"
          >
            SLA 分
          </label>
          <input
            id="awd-challenge-config-sla-score"
            v-model.number="form.awd_sla_score"
            type="number"
            min="0"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.awd_sla_score" class="text-xs text-[var(--color-danger)]">
            {{ fieldErrors.awd_sla_score }}
          </p>
        </div>

        <div class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            for="awd-challenge-config-defense-score"
          >
            防守分
          </label>
          <input
            id="awd-challenge-config-defense-score"
            v-model.number="form.awd_defense_score"
            type="number"
            min="0"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p
            v-if="fieldErrors.awd_defense_score"
            class="text-xs text-[var(--color-danger)]"
          >
            {{ fieldErrors.awd_defense_score }}
          </p>
        </div>
      </div>

      <section class="checker-config-block">
        <header class="checker-config-block__head">
          <div>
            <div class="journal-note-label">Checker Config</div>
            <h3 class="workspace-tab-heading__title checker-config-block__title">
              {{ form.awd_checker_type === 'http_standard' ? 'HTTP Standard 配置' : 'Legacy Probe 配置' }}
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

        <div v-if="form.awd_checker_type === 'legacy_probe'" class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            for="awd-challenge-config-legacy-health-path"
          >
            健康检查路径
          </label>
          <input
            id="awd-challenge-config-legacy-health-path"
            v-model="legacyProbeDraft.health_path"
            type="text"
            placeholder="/healthz"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p class="text-xs text-[var(--color-text-secondary)]">
            留空时使用当前全局健康检查路径。
          </p>
          <p
            v-if="fieldErrors.legacy_health_path"
            class="text-xs text-[var(--color-danger)]"
          >
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
              class="checker-preset-button"
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
              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-put-method">Method</label>
                <select
                  id="awd-http-put-method"
                  v-model="httpStandardDraft.put_flag.method"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
                  <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">
                    {{ method }}
                  </option>
                </select>
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-put-path">Path</label>
                <input
                  id="awd-http-put-path"
                  v-model="httpStandardDraft.put_flag.path"
                  type="text"
                  placeholder="/api/flag"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
                <p v-if="fieldErrors.http_put_path" class="text-xs text-[var(--color-danger)]">
                  {{ fieldErrors.http_put_path }}
                </p>
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-put-expected-status">状态码</label>
                <input
                  id="awd-http-put-expected-status"
                  v-model.number="httpStandardDraft.put_flag.expected_status"
                  type="number"
                  min="1"
                  step="1"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
                <p
                  v-if="fieldErrors.http_put_expected_status"
                  class="text-xs text-[var(--color-danger)]"
                >
                  {{ fieldErrors.http_put_expected_status }}
                </p>
              </div>
            </div>

            <div class="checker-action-extra-grid">
              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-put-body-template">Body Template</label>
                <textarea
                  id="awd-http-put-body-template"
                  v-model="httpStandardDraft.put_flag.body_template"
                  rows="4"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 font-mono text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                />
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-put-headers">Headers JSON</label>
                <textarea
                  id="awd-http-put-headers"
                  v-model="httpStandardDraft.put_flag.headers_text"
                  rows="4"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 font-mono text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  placeholder='{"Content-Type":"application/json"}'
                />
                <p v-if="fieldErrors.http_put_headers_text" class="text-xs text-[var(--color-danger)]">
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
              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-get-method">Method</label>
                <select
                  id="awd-http-get-method"
                  v-model="httpStandardDraft.get_flag.method"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
                  <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">
                    {{ method }}
                  </option>
                </select>
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-get-path">Path</label>
                <input
                  id="awd-http-get-path"
                  v-model="httpStandardDraft.get_flag.path"
                  type="text"
                  placeholder="/api/flag"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
                <p v-if="fieldErrors.http_get_path" class="text-xs text-[var(--color-danger)]">
                  {{ fieldErrors.http_get_path }}
                </p>
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-get-expected-status">状态码</label>
                <input
                  id="awd-http-get-expected-status"
                  v-model.number="httpStandardDraft.get_flag.expected_status"
                  type="number"
                  min="1"
                  step="1"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
                <p
                  v-if="fieldErrors.http_get_expected_status"
                  class="text-xs text-[var(--color-danger)]"
                >
                  {{ fieldErrors.http_get_expected_status }}
                </p>
              </div>
            </div>

            <div class="checker-action-extra-grid">
              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-get-expected-substring">预期片段</label>
                <input
                  id="awd-http-get-expected-substring"
                  v-model="httpStandardDraft.get_flag.expected_substring"
                  type="text"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 font-mono text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  placeholder="{{FLAG}}"
                >
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-get-headers">Headers JSON</label>
                <textarea
                  id="awd-http-get-headers"
                  v-model="httpStandardDraft.get_flag.headers_text"
                  rows="4"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 font-mono text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  placeholder='{"Accept":"application/json"}'
                />
                <p v-if="fieldErrors.http_get_headers_text" class="text-xs text-[var(--color-danger)]">
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
              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-havoc-method">Method</label>
                <select
                  id="awd-http-havoc-method"
                  v-model="httpStandardDraft.havoc.method"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
                  <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">
                    {{ method }}
                  </option>
                </select>
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-havoc-path">Path</label>
                <input
                  id="awd-http-havoc-path"
                  v-model="httpStandardDraft.havoc.path"
                  type="text"
                  placeholder="/healthz"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-havoc-expected-status">状态码</label>
                <input
                  id="awd-http-havoc-expected-status"
                  v-model.number="httpStandardDraft.havoc.expected_status"
                  type="number"
                  min="1"
                  step="1"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                >
                <p
                  v-if="fieldErrors.http_havoc_expected_status"
                  class="text-xs text-[var(--color-danger)]"
                >
                  {{ fieldErrors.http_havoc_expected_status }}
                </p>
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-http-havoc-headers">Headers JSON</label>
              <textarea
                id="awd-http-havoc-headers"
                v-model="httpStandardDraft.havoc.headers_text"
                rows="4"
                class="w-full rounded-xl border border-border bg-surface px-4 py-3 font-mono text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                placeholder='{"X-Checker":"havoc"}'
              />
              <p v-if="fieldErrors.http_havoc_headers_text" class="text-xs text-[var(--color-danger)]">
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
            <h3 class="workspace-tab-heading__title checker-config-block__title">最终 JSON 预览</h3>
          </div>
          <p class="checker-config-block__hint">保存时会按下面的结构写入 `awd_checker_config`。</p>
        </header>

        <pre id="awd-challenge-config-preview" class="checker-preview">{{ checkerPreviewText }}</pre>
      </section>
    </form>

    <template #footer>
      <div class="flex items-center justify-end gap-2">
        <button
          type="button"
          class="rounded-xl border border-border px-4 py-2 text-sm text-[var(--color-text-primary)] transition hover:border-primary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="awd-challenge-config-submit"
          type="button"
          class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="saving || (mode === 'create' && loadingChallengeCatalog)"
          @click="handleSubmit"
        >
          {{ saving ? '保存中...' : mode === 'create' ? '新增题目' : '保存配置' }}
        </button>
      </div>
    </template>
  </ElDialog>
</template>

<style scoped>
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
  gap: 0.35rem;
  justify-items: start;
  padding: 0.9rem 1rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 88%, var(--color-bg-surface-elevated));
  color: var(--color-text-primary);
  text-align: left;
  transition: border-color 0.2s ease, transform 0.2s ease;
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

@media (max-width: 960px) {
  .checker-preset-strip,
  .checker-action-grid,
  .checker-action-extra-grid {
    grid-template-columns: 1fr;
  }
}
</style>
