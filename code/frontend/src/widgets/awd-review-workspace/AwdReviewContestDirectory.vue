<script setup lang="ts">
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import AwdReviewContestHead from './AwdReviewContestHead.vue'
import AwdReviewDirectorySection from './AwdReviewDirectorySection.vue'
import AwdReviewDirectoryState from './AwdReviewDirectoryState.vue'
import AwdReviewIndexFilters from './AwdReviewIndexFilters.vue'
import TeacherAWDReviewContestRow from './TeacherAWDReviewContestRow.vue'

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
  <AwdReviewDirectorySection
    :total-count="total"
  >
    <template #filters>
      <AwdReviewIndexFilters
        :status-options="statusOptions"
        :status-filter="statusFilter"
        :keyword-filter="keywordFilter"
        @update-status-filter="emit('updateStatusFilter', $event)"
        @update-keyword-filter="emit('updateKeywordFilter', $event)"
      />
    </template>

    <AwdReviewDirectoryState
      :loading="loading"
      :error="error"
      :has-contests="hasContests"
      @reload="emit('reload')"
    >
      <section class="teacher-directory">
        <AwdReviewContestHead />

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
    </AwdReviewDirectoryState>
  </AwdReviewDirectorySection>
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
