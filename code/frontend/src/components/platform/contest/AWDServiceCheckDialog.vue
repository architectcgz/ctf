<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
  AWDTeamServiceData,
} from '@/api/contracts'

const props = defineProps<{
  open: boolean
  teams: AdminContestTeamData[]
  challengeLinks: AdminContestChallengeViewData[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      team_id: number
      service_id: number
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

function getChallengeLabel(challenge: AdminContestChallengeViewData): string {
  const prefix = challenge.title?.trim()
    ? challenge.title.trim()
    : `Challenge #${challenge.challenge_id}`
  return `${prefix} · ${challenge.is_visible ? '可见' : '隐藏'}`
}

function getSelectedServiceId(): number | null {
  const challenge = challengeOptions.value.find((item) => item.challenge_id === form.challenge_id)
  if (!challenge?.awd_service_id) {
    return null
  }
  return Number(challenge.awd_service_id)
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
  const selectedServiceId = getSelectedServiceId()
  if (form.challenge_id && selectedServiceId == null) {
    fieldErrors.challenge_id = '当前题目缺少 AWD 服务标识'
  }

  const checkResult = parseCheckResult()
  if (!form.team_id || !form.challenge_id || selectedServiceId == null || !checkResult) {
    return
  }

  emit('save', {
    team_id: Number(form.team_id),
    service_id: selectedServiceId,
    service_status: form.service_status,
    check_result: checkResult,
  })
}
</script>

<template>
  <AdminSurfaceModal
    :open="open"
    title="录入服务检查"
    subtitle="针对当前轮的队伍服务状态补录检查结果，便于赛后复盘和运维对账。"
    eyebrow="AWD Operations"
    width="35rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="space-y-5"
      @submit.prevent="handleSubmit"
    >
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="ui-field awd-service-field">
          <label
            class="ui-field__label"
            for="awd-service-team"
          >队伍</label>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.team_id }"
          >
            <select
              id="awd-service-team"
              v-model="form.team_id"
              class="ui-control"
            >
              <option
                value=""
                disabled
              >请选择队伍</option>
              <option
                v-for="team in teams"
                :key="team.id"
                :value="team.id"
              >
                {{ team.name }}
              </option>
            </select>
          </span>
          <p
            v-if="fieldErrors.team_id"
            class="ui-field__error"
          >
            {{ fieldErrors.team_id }}
          </p>
        </div>

        <div class="ui-field awd-service-field">
          <label
            class="ui-field__label"
            for="awd-service-challenge"
          >题目</label>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.challenge_id }"
          >
            <select
              id="awd-service-challenge"
              v-model="form.challenge_id"
              class="ui-control"
            >
              <option
                value=""
                disabled
              >请选择题目</option>
              <option
                v-for="challenge in challengeOptions"
                :key="challenge.id"
                :value="challenge.challenge_id"
              >
                {{ getChallengeLabel(challenge) }}
              </option>
            </select>
          </span>
          <p
            v-if="fieldErrors.challenge_id"
            class="ui-field__error"
          >
            {{ fieldErrors.challenge_id }}
          </p>
        </div>
      </div>

      <div class="ui-field awd-service-field">
        <label
          class="ui-field__label"
          for="awd-service-status"
        >服务状态</label>
        <span class="ui-control-wrap">
          <select
            id="awd-service-status"
            v-model="form.service_status"
            class="ui-control"
          >
            <option value="up">正常</option>
            <option value="down">下线</option>
            <option value="compromised">已失陷</option>
          </select>
        </span>
      </div>

      <div class="ui-field awd-service-field">
        <label
          class="ui-field__label"
          for="awd-service-check-result"
        >检查结果 JSON</label>
        <span
          class="ui-control-wrap"
          :class="{ 'is-error': !!fieldErrors.check_result_text }"
        >
          <textarea
            id="awd-service-check-result"
            v-model="form.check_result_text"
            rows="6"
            class="ui-control awd-service-field__textarea"
            placeholder="{&quot;http_status&quot;:200,&quot;latency_ms&quot;:38}"
          />
        </span>
        <p
          v-if="fieldErrors.check_result_text"
          class="ui-field__error"
        >
          {{ fieldErrors.check_result_text }}
        </p>
        <p
          v-else-if="!hasTargets"
          class="ui-field__hint awd-service-field__warning"
        >
          当前赛事缺少队伍或题目，暂时无法录入服务检查。
        </p>
      </div>
    </form>

    <template #footer>
      <div class="awd-service-dialog__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="awd-service-check-submit"
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="saving || !hasTargets"
          @click="handleSubmit"
        >
          {{ saving ? '保存中...' : '保存检查结果' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.awd-service-field {
  --ui-field-gap: var(--space-2);
}

.awd-service-field__textarea {
  min-height: 8.75rem;
  font-family: var(--font-family-mono);
}

.awd-service-field__warning {
  color: var(--color-warning);
}

.awd-service-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

@media (max-width: 767px) {
  .awd-service-dialog__footer {
    flex-direction: column-reverse;
  }
}
</style>
