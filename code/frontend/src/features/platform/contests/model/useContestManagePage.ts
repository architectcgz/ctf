import { computed, onMounted, ref } from 'vue'

import type { ContestDetailData } from '@/api/contracts'
import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'
import {
  buildContestAnnouncementsRoute,
  buildContestEditRoute,
  buildContestOperationsRoute,
} from './contestManageRoutes'
import {
  buildContestManagePanelQuery,
  resolveContestManagePanel,
  type ContestManagePanelKey,
} from './useContestManagePanelRoute'
import { usePlatformContests } from './usePlatformContests'

export function useContestManagePage() {
  const { query, replaceQuery } = useRouteQueryTransport()
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
  const activePanel = computed<ContestManagePanelKey>(() => resolveContestManagePanel(query.value.panel))
  const announcementDrawerOpen = ref(false)
  const activeAnnouncementContest = ref<ContestDetailData | null>(null)

  onMounted(() => {
    void refresh()
  })

  function updateStatusFilter(value: typeof statusFilter.value) {
    statusFilter.value = value
  }

  async function switchPanel(panel: ContestManagePanelKey): Promise<void> {
    if (activePanel.value === panel) {
      return
    }

    await replaceQuery(buildContestManagePanelQuery(query.value, panel))
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
      await switchPanel('overview')
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
    activePanel,
    announcementDrawerOpen,
    activeAnnouncementContest,
    prepareCreateContest,
    openEditDialog,
    closeDialog,
    closeAWDStartOverrideDialog,
    confirmAWDStartOverride,
    saveContest,
    updateStatusFilter,
    switchPanel,
    handleDialogOpenChange,
    handleAwdStartOverrideDialogOpenChange,
    openAnnouncementDrawer,
    closeAnnouncementDrawer,
    handleCreateContestSave,
    buildContestEditRoute,
    buildContestOperationsRoute,
    buildContestAnnouncementsRoute,
  }
}
