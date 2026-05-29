import { computed, type Ref } from 'vue'

import type { AdminContestChallengeViewData, AdminContestTeamData } from '@/api/contracts'

interface UseAwdOperationsDialogAvailabilityOptions {
  teams: Readonly<Ref<AdminContestTeamData[]>>
  challengeLinks: Readonly<Ref<AdminContestChallengeViewData[]>>
}

export function useAwdOperationsDialogAvailability({
  teams,
  challengeLinks,
}: UseAwdOperationsDialogAvailabilityOptions) {
  const canRecordServiceChecks = computed(
    () => teams.value.length > 0 && challengeLinks.value.length > 0
  )
  const canRecordAttackLogs = computed(
    () => teams.value.length >= 2 && challengeLinks.value.length > 0
  )
  const serviceCheckHint = computed(() => {
    if (teams.value.length === 0 && challengeLinks.value.length === 0) {
      return '当前赛事还没有队伍和题目，无法录入服务检查。'
    }
    if (teams.value.length === 0) {
      return '当前赛事还没有队伍，无法录入服务检查。'
    }
    if (challengeLinks.value.length === 0) {
      return '当前赛事还没有关联题目，无法录入服务检查。'
    }
    return ''
  })
  const attackLogHint = computed(() => {
    if (teams.value.length < 2 && challengeLinks.value.length === 0) {
      return '至少需要 2 支队伍且已关联题目后，才能补录攻击日志。'
    }
    if (teams.value.length < 2) {
      return '至少需要 2 支队伍后，才能补录攻击日志。'
    }
    if (challengeLinks.value.length === 0) {
      return '当前赛事还没有关联题目，无法补录攻击日志。'
    }
    return ''
  })

  return {
    canRecordServiceChecks,
    canRecordAttackLogs,
    serviceCheckHint,
    attackLogHint,
  }
}
