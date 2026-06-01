<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <ContestManageOverviewPanel
        v-show="activePanel === 'overview'"
        :active="activePanel === 'overview'"
        :list="list"
        :total="total"
        :summary="summary"
        :page="page"
        :page-size="pageSize"
        :loading="loading"
        :status-filter="statusFilter"
        :awd-contests="awdContests"
        :build-edit-route="buildEditRoute"
        :build-workbench-route="buildWorkbenchRoute"
        @refresh="emit('refresh')"
        @open-create="openCreatePanel"
        @update-status-filter="emit('updateStatusFilter', $event)"
        @announce="emit('announce', $event)"
        @change-page="emit('changePage', $event)"
      />

      <ContestManageCreatePanel
        v-show="activePanel === 'create'"
        :active="activePanel === 'create'"
        :draft="createDraft"
        :saving="createSaving"
        :field-locks="createFieldLocks"
        @back="emit('switchPanel', 'overview')"
        @save="emit('saveCreateContest', $event)"
      />
    </main>
  </section>
</template>

<script setup lang="ts">
import type { ContestDetailData, ContestListSummaryData } from '@/api/contracts'
import type {
  ContestManagePanelKey,
  ContestEditRouteTarget,
  ContestFieldLocks,
  ContestFormDraft,
  ContestOperationsRouteTarget,
} from '../model'
import ContestManageCreatePanel from './ContestManageCreatePanel.vue'
import ContestManageOverviewPanel from './ContestManageOverviewPanel.vue'
import './contestOrchestrationPage.css'
import type { ContestManageStatusFilter } from './contestOrchestrationPage.types'

const props = defineProps<{
  list: ContestDetailData[]
  total: number
  summary: ContestListSummaryData
  page: number
  pageSize: number
  loading: boolean
  statusFilter: ContestManageStatusFilter
  awdContests: ContestDetailData[]
  createDraft: ContestFormDraft
  createSaving: boolean
  createFieldLocks: ContestFieldLocks
  activePanel: ContestManagePanelKey
  buildEditRoute: (contestId: string) => ContestEditRouteTarget
  buildWorkbenchRoute: (contestId: string) => ContestOperationsRouteTarget
}>()

const emit = defineEmits<{
  refresh: []
  prepareCreateContest: []
  saveCreateContest: [value: ContestFormDraft]
  switchPanel: [panel: ContestManagePanelKey]
  updateStatusFilter: [value: ContestManageStatusFilter]
  announce: [contest: ContestDetailData]
  changePage: [page: number]
}>()

function openCreatePanel() {
  emit('prepareCreateContest')
  emit('switchPanel', 'create')
}
</script>
