<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import type {
  AdminContestChallengeData,
  AdminContestTeamData,
  AWDTeamServiceData,
} from '@/api/contracts'

const props = defineProps<{
  open: boolean
  teams: AdminContestTeamData[]
  challengeLinks: AdminContestChallengeData[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      team_id: number
      challenge_id: number
      service_status: AWDTeamServiceData['service_status']
      check_result?: Record<string, unknown>
    },
  ]
}>()

const form = reactive({
  team_id: '',
  challenge_id: '',
  service_status: 'up' as AWDTeamServiceData['service_status'],
  check_result_text: '{}',
})

const fieldErrors = reactive({
  team_id: '',
  challenge_id: '',
  check_result_text: '',
})

const challengeOptions = computed(() =>
  [...props.challengeLinks].sort(
    (a, b) => a.order - b.order || Number(a.challenge_id) - Number(b.challenge_id)
  )
)
const hasTargets = computed(() => props.teams.length > 0 && challengeOptions.value.length > 0)

function getChallengeLabel(challenge: AdminContestChallengeData): string {
  const prefix = challenge.title?.trim()
    ? challenge.title.trim()
    : `Challenge #${challenge.challenge_id}`
  return `${prefix} · ${challenge.is_visible ? '可见' : '隐藏'}`
}

watch(
  () => props.open,
  (open) => {
    if (!open) {
      return
    }
    form.team_id = props.teams[0]?.id || ''
    form.challenge_id = challengeOptions.value[0]?.challenge_id || ''
    form.service_status = 'up'
    form.check_result_text = '{}'
    clearErrors()
  },
  { immediate: true }
)

function clearErrors() {
  fieldErrors.team_id = ''
  fieldErrors.challenge_id = ''
  fieldErrors.check_result_text = ''
}

function closeDialog() {
  emit('update:open', false)
}

function parseCheckResult(): Record<string, unknown> | null {
  const trimmed = form.check_result_text.trim()
  if (!trimmed) {
    return {}
  }

  try {
    const parsed = JSON.parse(trimmed)
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>
    }
    fieldErrors.check_result_text = '检查结果必须是 JSON 对象'
    return null
  } catch {
    fieldErrors.check_result_text = '检查结果必须是合法 JSON'
    return null
  }
}

function handleSubmit() {
  clearErrors()

  if (!form.team_id) {
    fieldErrors.team_id = '请选择队伍'
  }
  if (!form.challenge_id) {
    fieldErrors.challenge_id = '请选择题目'
  }

  const checkResult = parseCheckResult()
  if (!form.team_id || !form.challenge_id || !checkResult) {
    return
  }

  emit('save', {
    team_id: Number(form.team_id),
    challenge_id: Number(form.challenge_id),
    service_status: form.service_status,
    check_result: checkResult,
  })
}
</script>

<template>
  <ElDialog
    :model-value="open"
    title="录入服务检查"
    width="560px"
    @close="closeDialog"
    @update:model-value="emit('update:open', $event)"
  >
    <form class="space-y-5" @submit.prevent="handleSubmit">
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-service-team"
            >队伍</label
          >
          <select
            id="awd-service-team"
            v-model="form.team_id"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option value="" disabled>请选择队伍</option>
            <option v-for="team in teams" :key="team.id" :value="team.id">
              {{ team.name }}
            </option>
          </select>
          <p v-if="fieldErrors.team_id" class="text-xs text-[var(--color-danger)]">
            {{ fieldErrors.team_id }}
          </p>
        </div>

        <div class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            for="awd-service-challenge"
            >题目</label
          >
          <select
            id="awd-service-challenge"
            v-model="form.challenge_id"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option value="" disabled>请选择题目</option>
            <option
              v-for="challenge in challengeOptions"
              :key="challenge.id"
              :value="challenge.challenge_id"
            >
              {{ getChallengeLabel(challenge) }}
            </option>
          </select>
          <p v-if="fieldErrors.challenge_id" class="text-xs text-[var(--color-danger)]">
            {{ fieldErrors.challenge_id }}
          </p>
        </div>
      </div>

      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-service-status"
          >服务状态</label
        >
        <select
          id="awd-service-status"
          v-model="form.service_status"
          class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
        >
          <option value="up">正常</option>
          <option value="down">下线</option>
          <option value="compromised">已失陷</option>
        </select>
      </div>

      <div class="space-y-2">
        <label
          class="text-sm font-medium text-[var(--color-text-primary)]"
          for="awd-service-check-result"
          >检查结果 JSON</label
        >
        <textarea
          id="awd-service-check-result"
          v-model="form.check_result_text"
          rows="6"
          class="w-full rounded-xl border border-border bg-surface px-4 py-3 font-mono text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          placeholder='{"http_status":200,"latency_ms":38}'
        />
        <p v-if="fieldErrors.check_result_text" class="text-xs text-[var(--color-danger)]">
          {{ fieldErrors.check_result_text }}
        </p>
        <p v-else-if="!hasTargets" class="text-xs text-[var(--color-warning)]">
          当前赛事缺少队伍或题目，暂时无法录入服务检查。
        </p>
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
          id="awd-service-check-submit"
          type="button"
          class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="saving || !hasTargets"
          @click="handleSubmit"
        >
          {{ saving ? '保存中...' : '保存检查结果' }}
        </button>
      </div>
    </template>
  </ElDialog>
</template>
