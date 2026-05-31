import { describe, expect, it } from 'vitest'

import type { TeamData } from '@/api/contracts'

import {
  buildTeamMemberPresentation,
  getTeamCaptainLabel,
  getTeamEmptyStateLabel,
  getTeamInviteCodeLabel,
  getTeamMemberCount,
  isCurrentUserTeamCaptain,
} from './presentation'

function buildTeam(overrides: Partial<TeamData> = {}): TeamData {
  return {
    id: 'team-1',
    name: 'Red',
    invite_code: 'RED-CTF',
    captain_user_id: 'user-1',
    members: [
      {
        user_id: 'user-1',
        username: 'captain',
        joined_at: '2024-03-15T09:00:00Z',
      },
      {
        user_id: 'user-2',
        username: 'member',
        joined_at: '2024-03-15T09:01:00Z',
      },
    ],
    ...overrides,
  }
}

describe('team presentation', () => {
  it('maps stable member count, captain relation, and labels through entity owner', () => {
    const team = buildTeam()

    expect(getTeamCaptainLabel()).toBe('队长')
    expect(getTeamEmptyStateLabel()).toBe('当前账号尚未加入队伍。')
    expect(getTeamMemberCount(team)).toBe(2)
    expect(getTeamMemberCount(null)).toBe(0)
    expect(isCurrentUserTeamCaptain(team, 'user-1')).toBe(true)
    expect(isCurrentUserTeamCaptain(team, 'user-2')).toBe(false)
    expect(isCurrentUserTeamCaptain(null, 'user-1')).toBe(false)
  })

  it('builds stable member presentation items for team views', () => {
    expect(buildTeamMemberPresentation(null)).toEqual([])

    expect(buildTeamMemberPresentation(buildTeam())).toEqual([
      {
        userId: 'user-1',
        username: 'captain',
        isCaptain: true,
        roleLabel: '队长',
      },
      {
        userId: 'user-2',
        username: 'member',
        isCaptain: false,
        roleLabel: null,
      },
    ])
  })

  it('formats invite code labels through entity presentation owner', () => {
    expect(getTeamInviteCodeLabel('RED-CTF')).toBe('邀请码: RED-CTF')
    expect(getTeamInviteCodeLabel(undefined)).toBeNull()
    expect(getTeamInviteCodeLabel('')).toBeNull()
  })
})
