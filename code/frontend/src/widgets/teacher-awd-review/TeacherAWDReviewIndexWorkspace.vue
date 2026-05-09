<script setup lang="ts">
import { computed } from 'vue'
import { FolderKanban, RefreshCcw } from 'lucide-vue-next'

import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import TeacherAWDReviewContestDirectory from './TeacherAWDReviewContestDirectory.vue'
import TeacherAWDReviewSummaryPanel from './TeacherAWDReviewSummaryPanel.vue'
import TeacherAWDReviewSurfaceShell from './TeacherAWDReviewSurfaceShell.vue'
import TeacherAWDReviewWorkspaceHeader from './TeacherAWDReviewWorkspaceHeader.vue'
import {
  buildTeacherAwdReviewIndexSummaryItems,
  TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY,
} from './model/presentation'

interface ContestSummary {
  totalCount: number
  runningCount: number
  exportReadyCount: number
}

type ContestStatusOption = {
  value: '' | TeacherAWDReviewContestItemData['status']
  label: string
}

const props = defineProps<{
  loading: boolean
  error: string | null
  contests: TeacherAWDReviewContestItemData[]
  total: number
  page: number
  totalPages: number
  hasContests: boolean
  statusOptions: readonly ContestStatusOption[]
  contestSummary: ContestSummary
  statusFilter: '' | TeacherAWDReviewContestItemData['status']
  keywordFilter: string
  contestStatusLabel: (status: string) => string
}>()

const emit = defineEmits<{
  openDashboard: []
  refresh: []
  reload: []
  openContest: [contestId: string]
  changePage: [page: number]
  updateStatusFilter: [status: '' | TeacherAWDReviewContestItemData['status']]
  updateKeywordFilter: [keyword: string]
}>()

const summaryItems = computed(() =>
  buildTeacherAwdReviewIndexSummaryItems(props.contestSummary)
)
</script>

<template>
  <TeacherAWDReviewSurfaceShell>
    <div class="teacher-page">
      <TeacherAWDReviewWorkspaceHeader
        :overline="TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.overline"
        :title="TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.title"
        header-class="awd-review-index-header"
        overline-class="awd-review-index-overline"
      >
        <template #description>
          {{ TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.description }}
        </template>

        <template #actions>
          <button
            type="button"
            class="teacher-btn teacher-btn--ghost"
            @click="emit('openDashboard')"
          >
            {{ TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.openDashboardAction }}
          </button>
          <button
            type="button"
            class="teacher-btn teacher-btn--primary"
            @click="emit('refresh')"
          >
            <RefreshCcw class="h-4 w-4" />
            {{ TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.refreshDirectoryAction }}
          </button>
        </template>
      </TeacherAWDReviewWorkspaceHeader>

      <TeacherAWDReviewSummaryPanel
        :title="TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.summaryTitle"
        :items="summaryItems"
      >
        <template #title-prefix>
          <FolderKanban class="h-4 w-4" />
        </template>
      </TeacherAWDReviewSummaryPanel>

      <TeacherAWDReviewContestDirectory
        :loading="loading"
        :error="error"
        :contests="contests"
        :total="total"
        :page="page"
        :total-pages="totalPages"
        :has-contests="hasContests"
        :status-options="statusOptions"
        :status-filter="statusFilter"
        :keyword-filter="keywordFilter"
        :contest-status-label="contestStatusLabel"
        @reload="emit('reload')"
        @open-contest="emit('openContest', $event)"
        @change-page="emit('changePage', $event)"
        @update-status-filter="emit('updateStatusFilter', $event)"
        @update-keyword-filter="emit('updateKeywordFilter', $event)"
      />
    </div>
  </TeacherAWDReviewSurfaceShell>
</template>

<style scoped>
.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.awd-review-index-overline {
  font-size: var(--journal-overline-font-size, var(--font-size-0-70));
  font-weight: 700;
  letter-spacing: var(--journal-overline-letter-spacing, 0.2em);
  text-transform: uppercase;
  color: var(--journal-accent, var(--color-primary));
}

@media (max-width: 1080px) {
  .teacher-page {
    min-height: auto;
  }
}
</style>
