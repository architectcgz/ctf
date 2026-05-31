interface TeamMemberInput {
  user_id: string
  username: string
}

interface TeamPresentationInput {
  captain_user_id: string
  invite_code?: string
  members: TeamMemberInput[]
}

export interface TeamMemberPresentationItem {
  userId: string
  username: string
  isCaptain: boolean
  roleLabel: string | null
}

export function getTeamCaptainLabel(): string {
  return '队长'
}

export function getTeamEmptyStateLabel(): string {
  return '当前账号尚未加入队伍。'
}

export function getTeamMemberCount(team: TeamPresentationInput | null | undefined): number {
  return team?.members.length ?? 0
}

export function isCurrentUserTeamCaptain(
  team: TeamPresentationInput | null | undefined,
  currentUserId: string | null | undefined
): boolean {
  return Boolean(team && currentUserId && team.captain_user_id === currentUserId)
}

export function getTeamInviteCodeLabel(inviteCode: string | null | undefined): string | null {
  if (!inviteCode) {
    return null
  }
  return `邀请码: ${inviteCode}`
}

export function buildTeamMemberPresentation(
  team: TeamPresentationInput | null | undefined
): TeamMemberPresentationItem[] {
  if (!team) {
    return []
  }

  return team.members.map((member) => {
    const isCaptain = member.user_id === team.captain_user_id
    return {
      userId: member.user_id,
      username: member.username,
      isCaptain,
      roleLabel: isCaptain ? getTeamCaptainLabel() : null,
    }
  })
}
