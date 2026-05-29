<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
  AWDTeamServiceData,
} from '@/api/contracts'

import AWDOperationsDialogFooter from './AWDOperationsDialogFooter.vue'
import AWDServiceCheckResultSection from './AWDServiceCheckResultSection.vue'
import AWDServiceCheckTargetSection from './AWDServiceCheckTargetSection.vue'
import type { AwdCreateServiceCheckPayload } from './awdOperationsDialogContracts'
import {
  resolveAwdServiceId,
  sortAwdChallengeLinks,
} from './awdOperationsDialogOptions'
import './awdOperationsDialogs.css'

const props = defineProps<{
  open: boolean
  teams: AdminContestTeamData[]
  challengeLinks: AdminContestChallengeViewData[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: AwdCreateServiceCheckPayload]
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

const challengeOptions = computed(() => sortAwdChallengeLinks(props.challengeLinks))
const hasTargets = computed(() => props.teams.length > 0 && challengeOptions.value.length > 0)

function getSelectedServiceId(): number | null {
  return resolveAwdServiceId(challengeOptions.value, form.challenge_id)
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
  if (props.saving) {
    return
  }

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
      <AWDServiceCheckTargetSection
        :teams="teams"
        :challenge-options="challengeOptions"
        :form="form"
        :field-errors="fieldErrors"
      />

      <AWDServiceCheckResultSection
        :form="form"
        :field-errors="fieldErrors"
        :has-targets="hasTargets"
      />
    </form>

    <template #footer>
      <AWDOperationsDialogFooter
        submit-id="awd-service-check-submit"
        submit-text="保存检查结果"
        saving-text="保存中..."
        :saving="saving"
        :disabled="saving || !hasTargets"
        @cancel="closeDialog"
        @submit="handleSubmit"
      />
    </template>
  </AdminSurfaceModal>
</template>
