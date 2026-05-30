import { computed, reactive, watch, type Ref } from 'vue'

import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
  AWDAttackLogData,
} from '@/api/contracts'

import type { AwdCreateAttackLogPayload } from './awdOperationsDialogContracts'
import {
  resolveAwdServiceId,
  sortAwdChallengeLinks,
} from './awdOperationsDialogOptions'

interface UseAwdAttackLogDialogFormOptions {
  open: Readonly<Ref<boolean>>
  teams: Readonly<Ref<AdminContestTeamData[]>>
  challengeLinks: Readonly<Ref<AdminContestChallengeViewData[]>>
}

function createDefaultForm(
  teams: AdminContestTeamData[],
  challengeOptions: AdminContestChallengeViewData[]
) {
  return {
    attacker_team_id: teams[0]?.id || '',
    victim_team_id: teams[1]?.id || teams[0]?.id || '',
    challenge_id: challengeOptions[0]?.challenge_id || '',
    attack_type: 'flag_capture' as AWDAttackLogData['attack_type'],
    submitted_flag: '',
    is_success: true,
  }
}

export function useAwdAttackLogDialogForm({
  open,
  teams,
  challengeLinks,
}: UseAwdAttackLogDialogFormOptions) {
  const form = reactive(createDefaultForm(teams.value, sortAwdChallengeLinks(challengeLinks.value)))
  const fieldErrors = reactive({
    attacker_team_id: '',
    victim_team_id: '',
    challenge_id: '',
  })

  const challengeOptions = computed(() => sortAwdChallengeLinks(challengeLinks.value))
  const hasTargets = computed(() => teams.value.length >= 2 && challengeOptions.value.length > 0)

  function clearErrors() {
    fieldErrors.attacker_team_id = ''
    fieldErrors.victim_team_id = ''
    fieldErrors.challenge_id = ''
  }

  function resetForm() {
    Object.assign(form, createDefaultForm(teams.value, challengeOptions.value))
    clearErrors()
  }

  function buildPayload(): AwdCreateAttackLogPayload | null {
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

    const selectedServiceId = resolveAwdServiceId(challengeOptions.value, form.challenge_id)
    if (form.challenge_id && selectedServiceId == null) {
      fieldErrors.challenge_id = '当前题目缺少 AWD 服务标识'
    }

    if (fieldErrors.attacker_team_id || fieldErrors.victim_team_id || fieldErrors.challenge_id) {
      return null
    }
    if (selectedServiceId == null) {
      return null
    }

    return {
      attacker_team_id: Number(form.attacker_team_id),
      victim_team_id: Number(form.victim_team_id),
      service_id: selectedServiceId,
      attack_type: form.attack_type,
      submitted_flag: form.submitted_flag.trim() || undefined,
      is_success: form.is_success,
    }
  }

  watch(
    () => open.value,
    (isOpen) => {
      if (!isOpen) {
        return
      }
      resetForm()
    },
    { immediate: true }
  )

  return {
    form,
    fieldErrors,
    challengeOptions,
    hasTargets,
    buildPayload,
  }
}
