<script setup lang="ts">
import type { AwdReviewContestItemData } from '@/api/contracts'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import AwdReviewContestHead from './AwdReviewContestHead.vue'
import AwdReviewContestRow from './AwdReviewContestRow.vue'
import AwdReviewDirectorySection from './AwdReviewDirectorySection.vue'
import AwdReviewDirectoryState from './AwdReviewDirectoryState.vue'
import AwdReviewIndexFilters from './AwdReviewIndexFilters.vue'

type ContestStatusOption = {
  value: '' | AwdReviewContestItemData['status']
  label: string
}

interface AwdReviewDetailRoute {
  name: 'TeacherAWDReviewDetail' | 'PlatformAwdReviewDetail'
  params: {
    contestId: string
  }
}

defineProps<{
  loading: boolean
  error: string | null
  contests: AwdReviewContestItemData[]
  total: number
  page: number
  totalPages: number
  hasContests: boolean
  statusOptions: readonly ContestStatusOption[]
  buildContestRoute: (contestId: string) => AwdReviewDetailRoute
  statusFilter: '' | AwdReviewContestItemData['status']
  keywordFilter: string
  contestStatusLabel: (status: string) => string
}>()

const emit = defineEmits<{
  reload: []
  changePage: [page: number]
  updateStatusFilter: [status: '' | AwdReviewContestItemData['status']]
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

        <AwdReviewContestRow
          v-for="contest in contests"
          :key="contest.id"
          :contest="contest"
          :build-contest-route="buildContestRoute"
          :contest-status-label="contestStatusLabel"
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
