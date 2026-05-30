import { describe, expect, it } from 'vitest'
import { ref } from 'vue'

import type { AdminContestChallengeViewData, AdminContestTeamData } from '@/api/contracts'

import { useAwdAttackLogDialogForm } from './useAwdAttackLogDialogForm'
import { useAwdRoundCreateDialogForm } from './useAwdRoundCreateDialogForm'
import { useAwdServiceCheckDialogForm } from './useAwdServiceCheckDialogForm'

function buildTeam(overrides: Partial<AdminContestTeamData> = {}): AdminContestTeamData {
  return {
    id: '12',
    contest_id: 'awd-1',
    name: 'Red Team',
    captain_id: '1',
    max_members: 5,
    member_count: 1,
    created_at: '2026-03-18T09:00:00.000Z',
    ...overrides,
  }
}

function buildChallengeLink(
  overrides: Partial<AdminContestChallengeViewData> = {}
): AdminContestChallengeViewData {
  return {
    id: 'link-1',
    contest_id: 'awd-1',
    challenge_id: 'challenge-1',
    awd_service_id: '7009',
    title: 'Web Checker',
    category: 'web',
    difficulty: 'easy',
    points: 120,
    order: 1,
    is_visible: true,
    created_at: '2026-03-18T09:00:00.000Z',
    ...overrides,
  }
}

describe('AWD operations dialog form workflows', () => {
  it('应在打开创建轮次弹层时重置默认值并构造 payload', () => {
    const open = ref(false)
    const nextRoundNumber = ref(7)
    const dialog = useAwdRoundCreateDialogForm({ open, nextRoundNumber })

    open.value = true

    expect(dialog.dialogTitle.value).toBe('创建第 7 轮')
    expect(dialog.form.round_number).toBe(7)
    expect(dialog.buildPayload()).toEqual({
      round_number: 7,
      status: 'pending',
      attack_score: 50,
      defense_score: 50,
    })
  })

  it('服务检查表单应解析 JSON 对象并解析 service_id', () => {
    const dialog = useAwdServiceCheckDialogForm({
      open: ref(true),
      teams: ref([buildTeam()]),
      challengeLinks: ref([buildChallengeLink()]),
    })

    dialog.form.check_result_text = '{"latency_ms": 38}'

    expect(dialog.buildPayload()).toEqual({
      team_id: 12,
      service_id: 7009,
      service_status: 'up',
      check_result: { latency_ms: 38 },
    })
  })

  it('攻击日志表单应阻止同队互打并在合法时构造 payload', () => {
    const dialog = useAwdAttackLogDialogForm({
      open: ref(true),
      teams: ref([
        buildTeam(),
        buildTeam({ id: '13', name: 'Blue Team', captain_id: '2' }),
      ]),
      challengeLinks: ref([buildChallengeLink()]),
    })

    dialog.form.victim_team_id = '12'
    expect(dialog.buildPayload()).toBeNull()
    expect(dialog.fieldErrors.victim_team_id).toBe('攻击队伍和受害队伍不能相同')

    dialog.form.victim_team_id = '13'
    dialog.form.submitted_flag = ' flag{demo} '

    expect(dialog.buildPayload()).toEqual({
      attacker_team_id: 12,
      victim_team_id: 13,
      service_id: 7009,
      attack_type: 'flag_capture',
      submitted_flag: 'flag{demo}',
      is_success: true,
    })
  })
})
