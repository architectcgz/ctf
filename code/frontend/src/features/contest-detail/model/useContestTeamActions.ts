import { ref, type Ref } from 'vue'

import { createTeam, joinTeam, kickTeamMember } from '@/api/contest'
import type { ContestDetailData, TeamData } from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

interface UseContestTeamActionsOptions {
  contest: Ref<ContestDetailData | null>
  team: Ref<TeamData | null>
  refreshTeam: () => Promise<void>
}

export function useContestTeamActions(options: UseContestTeamActionsOptions) {
  const { contest, team, refreshTeam } = options
  const toast = useToast()

  const showCreateTeam = ref(false)
  const showJoinTeam = ref(false)
  const teamName = ref('')
  const teamIdInput = ref('')
  const creatingTeam = ref(false)
  const joiningTeam = ref(false)

  function openCreateTeam() {
    showCreateTeam.value = true
  }

  function closeCreateTeam() {
    showCreateTeam.value = false
    teamName.value = ''
  }

  async function createTeamAction() {
    const name = teamName.value.trim()
    if (!name) {
      toast.warning('请输入队伍名称')
      return
    }
    if (name.length < 2 || name.length > 50) {
      toast.warning('队伍名称长度应在 2-50 字符之间')
      return
    }
    if (!contest.value || creatingTeam.value) {
      return
    }

    creatingTeam.value = true
    try {
      await createTeam(contest.value.id, { name })
      await refreshTeam()
      closeCreateTeam()
      toast.success('创建队伍成功')
    } catch (error) {
      console.error(error)
      toast.error(error instanceof Error ? error.message : '创建队伍失败')
    } finally {
      creatingTeam.value = false
    }
  }

  function openJoinTeam() {
    showJoinTeam.value = true
  }

  function closeJoinTeam() {
    showJoinTeam.value = false
    teamIdInput.value = ''
  }

  async function joinTeamAction() {
    const teamId = teamIdInput.value.trim()
    if (!teamId) {
      toast.warning('请输入队伍 ID')
      return
    }
    if (!contest.value || joiningTeam.value) {
      return
    }

    joiningTeam.value = true
    try {
      await joinTeam(contest.value.id, teamId)
      await refreshTeam()
      closeJoinTeam()
      toast.success('加入队伍成功')
    } catch (error) {
      console.error(error)
      toast.error(error instanceof Error ? error.message : '加入队伍失败')
    } finally {
      joiningTeam.value = false
    }
  }

  async function kickMember(userId: string) {
    if (!contest.value || !team.value) {
      return
    }

    const confirmed = await confirmDestructiveAction({
      title: '踢出成员',
      message: '确定踢出该成员？',
      confirmButtonText: '确认踢出',
    })
    if (!confirmed) {
      return
    }

    try {
      await kickTeamMember(contest.value.id, team.value.id, userId)
      await refreshTeam()
      toast.success('已踢出成员')
    } catch (error) {
      console.error(error)
      toast.error(error instanceof Error ? error.message : '踢出成员失败')
    }
  }

  return {
    showCreateTeam,
    showJoinTeam,
    teamName,
    teamIdInput,
    creatingTeam,
    joiningTeam,
    openCreateTeam,
    closeCreateTeam,
    createTeamAction,
    openJoinTeam,
    closeJoinTeam,
    joinTeamAction,
    kickMember,
  }
}
