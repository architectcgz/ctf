<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
  AWDAttackLogData,
} from '@/api/contracts'

import AWDAttackLogDetailsSection from './AWDAttackLogDetailsSection.vue'
import AWDAttackLogTargetSection from './AWDAttackLogTargetSection.vue'
import AWDOperationsDialogFooter from './AWDOperationsDialogFooter.vue'
import type { AwdCreateAttackLogPayload } from './awdOperationsDialogContracts'
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
  save: [value: AwdCreateAttackLogPayload]
}>()

const form = reactive({
  attacker_team_id: '',
  victim_team_id: '',
  challenge_id: '',
  attack_type: 'flag_capture' as AWDAttackLogData['attack_type'],
  submitted_flag: '',
  is_success: true,
})

const fieldErrors = reactive({
  attacker_team_id: '',
  victim_team_id: '',
  challenge_id: '',
})

const challengeOptions = computed(() => sortAwdChallengeLinks(props.challengeLinks))
const hasTargets = computed(() => props.teams.length >= 2 && challengeOptions.value.length > 0)

function getSelectedServiceId(): number | null {
  return resolveAwdServiceId(challengeOptions.value, form.challenge_id)
}

watch(
  () => props.open,
  (open) => {
    if (!open) {
      return
    }
    form.attacker_team_id = props.teams[0]?.id || ''
    form.victim_team_id = props.teams[1]?.id || props.teams[0]?.id || ''
    form.challenge_id = challengeOptions.value[0]?.challenge_id || ''
    form.attack_type = 'flag_capture'
    form.submitted_flag = ''
    form.is_success = true
    clearErrors()
  },
  { immediate: true }
)

function clearErrors() {
  fieldErrors.attacker_team_id = ''
  fieldErrors.victim_team_id = ''
  fieldErrors.challenge_id = ''
}

function closeDialog() {
  emit('update:open', false)
}

function handleSubmit() {
  if (props.saving) {
    return
  }

  clearErrors()

  if (!form.attacker_team_id) {
    fieldErrors.attacker_team_id = '请选择攻击队伍'
  }
  if (!form.victim_team_id) {
    fieldErrors.victim_team_id = '请选择受害队伍'
  }
  if (!form.challenge_id) {
    fieldErrors.challenge_id = '请选择题目'
  }
  if (
    form.attacker_team_id &&
    form.victim_team_id &&
    form.attacker_team_id === form.victim_team_id
  ) {
    fieldErrors.victim_team_id = '攻击队伍和受害队伍不能相同'
  }
  const selectedServiceId = getSelectedServiceId()
  if (form.challenge_id && selectedServiceId == null) {
    fieldErrors.challenge_id = '当前题目缺少 AWD 服务标识'
  }

  if (fieldErrors.attacker_team_id || fieldErrors.victim_team_id || fieldErrors.challenge_id) {
    return
  }
  if (selectedServiceId == null) {
    return
  }

  emit('save', {
    attacker_team_id: Number(form.attacker_team_id),
    victim_team_id: Number(form.victim_team_id),
    service_id: selectedServiceId,
    attack_type: form.attack_type,
    submitted_flag: form.submitted_flag.trim() || undefined,
    is_success: form.is_success,
  })
}
</script>

<template>
  <AdminSurfaceModal
    :open="open"
    title="补录攻击日志"
    subtitle="将线下核实过的攻击事件补录进复盘记录，不直接改写正式排行榜。"
    eyebrow="AWD Operations"
    width="35rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="space-y-5"
      @submit.prevent="handleSubmit"
    >
      <AWDAttackLogTargetSection
        :teams="teams"
        :challenge-options="challengeOptions"
        :form="form"
        :field-errors="fieldErrors"
      />

      <AWDAttackLogDetailsSection
        :form="form"
        :has-targets="hasTargets"
      />
    </form>

    <template #footer>
      <AWDOperationsDialogFooter
        submit-id="awd-attack-log-submit"
        submit-text="保存攻击日志"
        saving-text="保存中..."
        :saving="saving"
        :disabled="saving || !hasTargets"
        @cancel="closeDialog"
        @submit="handleSubmit"
      />
    </template>
  </AdminSurfaceModal>
</template>
