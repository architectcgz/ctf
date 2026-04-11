<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import type { AdminChallengeListItem, AdminContestChallengeData, AWDCheckerType } from '@/api/contracts'

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
  awd_checker_config_text: '{}',
})

const fieldErrors = reactive({
  challenge_id: '',
  points: '',
  order: '',
  awd_sla_score: '',
  awd_defense_score: '',
  awd_checker_config_text: '',
})

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
    form.awd_checker_config_text = JSON.stringify(props.draft?.awd_checker_config || {}, null, 2)
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
  fieldErrors.challenge_id = ''
  fieldErrors.points = ''
  fieldErrors.order = ''
  fieldErrors.awd_sla_score = ''
  fieldErrors.awd_defense_score = ''
  fieldErrors.awd_checker_config_text = ''
}

function closeDialog() {
  emit('update:open', false)
}

function parseCheckerConfig(): Record<string, unknown> | null {
  const trimmed = form.awd_checker_config_text.trim()
  if (!trimmed) {
    return {}
  }

  try {
    const parsed = JSON.parse(trimmed)
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>
    }
    fieldErrors.awd_checker_config_text = 'Checker 配置必须是 JSON 对象'
    return null
  } catch {
    fieldErrors.awd_checker_config_text = 'Checker 配置必须是合法 JSON'
    return null
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

  const checkerConfig = parseCheckerConfig()
  if (!checkerConfig) {
    return false
  }

  return (
    !fieldErrors.challenge_id &&
    !fieldErrors.points &&
    !fieldErrors.order &&
    !fieldErrors.awd_sla_score &&
    !fieldErrors.awd_defense_score &&
    !fieldErrors.awd_checker_config_text
  )
}

function handleSubmit() {
  if (!validate()) {
    return
  }

  const checkerConfig = parseCheckerConfig()
  if (!checkerConfig) {
    return
  }

  emit('save', {
    challenge_id: Number(form.challenge_id),
    points: form.points,
    order: form.order,
    is_visible: form.is_visible === 'true',
    awd_checker_type: form.awd_checker_type,
    awd_checker_config: checkerConfig,
    awd_sla_score: form.awd_sla_score,
    awd_defense_score: form.awd_defense_score,
  })
}
</script>

<template>
  <ElDialog
    :model-value="open"
    :title="dialogTitle"
    width="620px"
    @close="closeDialog"
    @update:model-value="emit('update:open', $event)"
  >
    <form class="space-y-5" @submit.prevent="handleSubmit">
      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-challenge-config-challenge">题目</label>
        <template v-if="mode === 'create'">
          <select
            id="awd-challenge-config-challenge"
            v-model="form.challenge_id"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option value="" disabled>{{ loadingChallengeCatalog ? '正在加载题库...' : '请选择题目' }}</option>
            <option v-for="challenge in selectableChallenges" :key="challenge.id" :value="challenge.id">
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
        <p v-if="fieldErrors.challenge_id" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.challenge_id }}</p>
      </div>

      <div class="grid gap-4 sm:grid-cols-3">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-challenge-config-points">分值</label>
          <input
            id="awd-challenge-config-points"
            v-model.number="form.points"
            type="number"
            min="1"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.points" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.points }}</p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-challenge-config-order">顺序</label>
          <input
            id="awd-challenge-config-order"
            v-model.number="form.order"
            type="number"
            min="0"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.order" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.order }}</p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-challenge-config-visible">可见性</label>
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
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-challenge-config-checker-type">Checker 类型</label>
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
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-challenge-config-sla-score">SLA 分</label>
          <input
            id="awd-challenge-config-sla-score"
            v-model.number="form.awd_sla_score"
            type="number"
            min="0"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.awd_sla_score" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.awd_sla_score }}</p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-challenge-config-defense-score">防守分</label>
          <input
            id="awd-challenge-config-defense-score"
            v-model.number="form.awd_defense_score"
            type="number"
            min="0"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.awd_defense_score" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.awd_defense_score }}</p>
        </div>
      </div>

      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-challenge-config-json">Checker 配置 JSON</label>
        <textarea
          id="awd-challenge-config-json"
          v-model="form.awd_checker_config_text"
          rows="8"
          class="w-full rounded-xl border border-border bg-surface px-4 py-3 font-mono text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          placeholder='{"get_flag":{"method":"GET","path":"/flag"}}'
        />
        <p v-if="fieldErrors.awd_checker_config_text" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.awd_checker_config_text }}</p>
      </div>
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
