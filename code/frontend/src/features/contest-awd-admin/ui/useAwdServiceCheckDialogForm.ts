import { computed, reactive, watch, type Ref } from 'vue'

import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
  AWDTeamServiceData,
} from '@/api/contracts'

import type { AwdCreateServiceCheckPayload } from './awdOperationsDialogContracts'
import {
  resolveAwdServiceId,
  sortAwdChallengeLinks,
} from './awdOperationsDialogOptions'

interface UseAwdServiceCheckDialogFormOptions {
  open: Readonly<Ref<boolean>>
  teams: Readonly<Ref<AdminContestTeamData[]>>
  challengeLinks: Readonly<Ref<AdminContestChallengeViewData[]>>
}

function createDefaultForm(
  teams: AdminContestTeamData[],
  challengeOptions: AdminContestChallengeViewData[]
) {
  return {
    team_id: teams[0]?.id || '',
    challenge_id: challengeOptions[0]?.challenge_id || '',
    service_status: 'up' as AWDTeamServiceData['service_status'],
    check_result_text: '{}',
  }
}

export function useAwdServiceCheckDialogForm({
  open,
  teams,
  challengeLinks,
}: UseAwdServiceCheckDialogFormOptions) {
  const form = reactive(createDefaultForm(teams.value, sortAwdChallengeLinks(challengeLinks.value)))
  const fieldErrors = reactive({
    team_id: '',
    challenge_id: '',
    check_result_text: '',
  })

  const challengeOptions = computed(() => sortAwdChallengeLinks(challengeLinks.value))
  const hasTargets = computed(() => teams.value.length > 0 && challengeOptions.value.length > 0)

  function clearErrors() {
    fieldErrors.team_id = ''
    fieldErrors.challenge_id = ''
    fieldErrors.check_result_text = ''
  }

  function resetForm() {
    Object.assign(form, createDefaultForm(teams.value, challengeOptions.value))
    clearErrors()
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

  function buildPayload(): AwdCreateServiceCheckPayload | null {
    clearErrors()

    if (!form.team_id) {
      fieldErrors.team_id = '请选择队伍'
    }
    if (!form.challenge_id) {
      fieldErrors.challenge_id = '请选择题目'
    }

    const selectedServiceId = resolveAwdServiceId(challengeOptions.value, form.challenge_id)
    if (form.challenge_id && selectedServiceId == null) {
      fieldErrors.challenge_id = '当前题目缺少 AWD 服务标识'
    }

    const checkResult = parseCheckResult()
    if (!form.team_id || !form.challenge_id || selectedServiceId == null || !checkResult) {
      return null
    }

    return {
      team_id: Number(form.team_id),
      service_id: selectedServiceId,
      service_status: form.service_status,
      check_result: checkResult,
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
