<template>
  <ContestDetailWorkspace
    :contest="contest"
    :team="team"
    :challenges="challenges"
    :announcements="announcements"
    :announcements-error="announcementsError"
    :loading="loading"
    :countdown="countdown"
    :selected-challenge="selectedChallenge"
    :flag-input="flagInput"
    :submitting="submitting"
    :submit-result="submitResult"
    :show-create-team="showCreateTeam"
    :show-join-team="showJoinTeam"
    :team-name="teamName"
    :team-id-input="teamIdInput"
    :creating-team="creatingTeam"
    :joining-team="joiningTeam"
    :is-captain="isCaptain"
    :active-workspace-tab="activeWorkspaceTab"
    :workspace-tabs="workspaceTabs"
    :solved-count="solvedCount"
    :total-points="totalPoints"
    :member-count="memberCount"
    :contest-accessible="contestAccessible"
    :set-tab-button-ref="setTabButtonRef"
    :select-workspace-tab="selectWorkspaceTab"
    :handle-workspace-tab-keydown="handleWorkspaceTabKeydown"
    :select-challenge="selectChallenge"
    :set-flag-input="(value) => (flagInput = value)"
    :submit-flag-action="submitFlagAction"
    :set-show-create-team="(value) => (showCreateTeam = value)"
    :set-show-join-team="(value) => (showJoinTeam = value)"
    :set-team-name="(value) => (teamName = value)"
    :set-team-id-input="(value) => (teamIdInput = value)"
    :open-create-team="openCreateTeam"
    :close-create-team="closeCreateTeam"
    :create-team-action="createTeamAction"
    :open-join-team="openJoinTeam"
    :close-join-team="closeJoinTeam"
    :join-team-action="joinTeamAction"
    :kick-member="kickMember"
    :refresh-announcements="refreshAnnouncements"
    :sync-announcements-incrementally="syncAnnouncementsIncrementally"
  />
</template>

<script setup lang="ts">
import { useContestDetailRoutePage } from '@/features/contest-detail'
import { ContestDetailWorkspace } from '@/widgets/contest-detail-workspace'

const {
  contest,
  team,
  challenges,
  announcements,
  announcementsError,
  loading,
  countdown,
  selectedChallenge,
  flagInput,
  submitting,
  submitResult,
  showCreateTeam,
  showJoinTeam,
  teamName,
  teamIdInput,
  creatingTeam,
  joiningTeam,
  isCaptain,
  activeWorkspaceTab,
  setTabButtonRef,
  selectWorkspaceTab,
  handleWorkspaceTabKeydown,
  workspaceTabs,
  solvedCount,
  totalPoints,
  memberCount,
  contestAccessible,
  selectChallenge,
  submitFlagAction,
  openCreateTeam,
  closeCreateTeam,
  createTeamAction,
  openJoinTeam,
  closeJoinTeam,
  joinTeamAction,
  kickMember,
  refreshAnnouncements,
  syncAnnouncementsIncrementally,
} = useContestDetailRoutePage()
</script>
