import { computed, onMounted, ref } from 'vue'

import type { ContestDetailData } from '@/api/contracts'
import { usePlatformContests } from './usePlatformContests'

export function useContestManagePage() {
  const {
    list,
    total,
    summary,
    page,
    pageSize,
    loading,
    refresh,
    changePage,
    statusFilter,
    dialogOpen,
    mode,
    saving,
    formDraft,
    fieldLocks,
    statusOptions,
    awdStartOverrideDialogState,
    prepareCreateContest,
    openEditDialog,
    closeDialog,
    closeAWDStartOverrideDialog,
    confirmAWDStartOverride,
    saveContest,
  } = usePlatformContests()

  const awdContests = computed(() => list.value.filter((item) => item.mode === 'awd'))
  const requestedPanel = ref<'overview' | 'list' | 'create' | null>(null)
  const requestedPanelVersion = ref(0)
  const announcementDrawerOpen = ref(false)
  const activeAnnouncementContest = ref<ContestDetailData | null>(null)

  onMounted(() => {
    void refresh()
  })

  function updateStatusFilter(value: typeof statusFilter.value) {
    statusFilter.value = value
  }

  function requestContestPanel(panel: 'overview' | 'list' | 'create') {
    requestedPanel.value = panel
    requestedPanelVersion.value += 1
  }

  function handleDialogOpenChange(value: boolean) {
    if (!value) {
      closeDialog()
    }
  }

  function handleAwdStartOverrideDialogOpenChange(value: boolean) {
    if (!value) {
      closeAWDStartOverrideDialog()
    }
  }

  function openAnnouncementDrawer(contest: ContestDetailData): void {
    activeAnnouncementContest.value = contest
    announcementDrawerOpen.value = true
  }

  function closeAnnouncementDrawer(): void {
    announcementDrawerOpen.value = false
  }

  async function handleCreateContestSave(draft: Parameters<typeof saveContest>[0]): Promise<void> {
    const result = await saveContest(draft)
    if (result === 'create') {
      requestContestPanel('list')
    }
  }

  return {
    list,
    total,
    summary,
    page,
    pageSize,
    loading,
    refresh,
    changePage,
    statusFilter,
    dialogOpen,
    mode,
    saving,
    formDraft,
    fieldLocks,
    statusOptions,
    awdStartOverrideDialogState,
    awdContests,
    requestedPanel,
    requestedPanelVersion,
    announcementDrawerOpen,
    activeAnnouncementContest,
    prepareCreateContest,
    openEditDialog,
    closeDialog,
    closeAWDStartOverrideDialog,
    confirmAWDStartOverride,
    saveContest,
    updateStatusFilter,
    handleDialogOpenChange,
    handleAwdStartOverrideDialogOpenChange,
    openAnnouncementDrawer,
    closeAnnouncementDrawer,
    handleCreateContestSave,
  }
}
