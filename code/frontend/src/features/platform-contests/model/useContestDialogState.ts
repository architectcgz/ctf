import { ref } from 'vue'

import type { ContestDetailData, ContestStatus } from '@/api/contracts'

import type { ContestFormDraft, PlatformContestStatus } from './contestFormSupport'

interface UseContestDialogStateOptions {
  createEmptyDraft: () => ContestFormDraft
  createDraftFromContest: (contest: ContestDetailData) => ContestFormDraft
  normalizeEditableStatus: (status: ContestStatus) => PlatformContestStatus
}

export function useContestDialogState(options: UseContestDialogStateOptions) {
  const dialogOpen = ref(false)
  const editingContestId = ref<string | null>(null)
  const editingBaseStatus = ref<PlatformContestStatus | null>(null)
  const formDraft = ref<ContestFormDraft>(options.createEmptyDraft())

  function prepareCreateContest() {
    editingContestId.value = null
    editingBaseStatus.value = null
    formDraft.value = options.createEmptyDraft()
    dialogOpen.value = false
  }

  function openCreateDialog() {
    prepareCreateContest()
    dialogOpen.value = true
  }

  function openEditDialog(contest: ContestDetailData) {
    editingContestId.value = contest.id
    editingBaseStatus.value = options.normalizeEditableStatus(contest.status)
    formDraft.value = options.createDraftFromContest(contest)
    dialogOpen.value = true
  }

  function closeDialog() {
    dialogOpen.value = false
  }

  return {
    dialogOpen,
    editingContestId,
    editingBaseStatus,
    formDraft,
    prepareCreateContest,
    openCreateDialog,
    openEditDialog,
    closeDialog,
  }
}
