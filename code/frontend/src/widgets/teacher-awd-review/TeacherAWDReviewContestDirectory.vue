<script setup lang="ts">
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import TeacherAWDReviewContestHead from './TeacherAWDReviewContestHead.vue'
import TeacherAWDReviewContestRow from './TeacherAWDReviewContestRow.vue'
import TeacherAWDReviewDirectorySection from './TeacherAWDReviewDirectorySection.vue'
import TeacherAWDReviewDirectoryState from './TeacherAWDReviewDirectoryState.vue'
import TeacherAWDReviewIndexFilters from './TeacherAWDReviewIndexFilters.vue'

type ContestStatusOption = {
  value: '' | TeacherAWDReviewContestItemData['status']
  label: string
}

defineProps<{
  loading: boolean
  error: string | null
  contests: TeacherAWDReviewContestItemData[]
  total: number
  page: number
  totalPages: number
  hasContests: boolean
  statusOptions: readonly ContestStatusOption[]
  statusFilter: '' | TeacherAWDReviewContestItemData['status']
  keywordFilter: string
  contestStatusLabel: (status: string) => string
}>()

const emit = defineEmits<{
  reload: []
  openContest: [contestId: string]
  changePage: [page: number]
  updateStatusFilter: [status: '' | TeacherAWDReviewContestItemData['status']]
  updateKeywordFilter: [keyword: string]
}>()
</script>

<template>
  <TeacherAWDReviewDirectorySection
    :total-count="total"
  >
    <template #filters>
      <TeacherAWDReviewIndexFilters
        :status-options="statusOptions"
        :status-filter="statusFilter"
        :keyword-filter="keywordFilter"
        @update-status-filter="emit('updateStatusFilter', $event)"
        @update-keyword-filter="emit('updateKeywordFilter', $event)"
      />
    </template>

    <TeacherAWDReviewDirectoryState
      :loading="loading"
      :error="error"
      :has-contests="hasContests"
      @reload="emit('reload')"
    >
      <section class="teacher-directory">
        <TeacherAWDReviewContestHead />

        <TeacherAWDReviewContestRow
          v-for="contest in contests"
          :key="contest.id"
          :contest="contest"
          :contest-status-label="contestStatusLabel"
          @open-contest="emit('openContest', $event)"
        />
      </section>

      <WorkspaceDirectoryPagination
        v-if="total > 0"
        class="teacher-directory-pagination"
        :page="page"
        :total-pages="totalPages"
        :total="total"
        :disabled="loading"
        :total-label="`共 ${total} 场赛事`"
        @change-page="emit('changePage', $event)"
      />
    </TeacherAWDReviewDirectoryState>
  </TeacherAWDReviewDirectorySection>
</template>

<style scoped>
.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory :deep(.workspace-directory-pagination-shell) {
  margin-top: var(--space-2);
}
</style>
