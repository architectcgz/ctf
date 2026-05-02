<script setup lang="ts">
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
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
  hasContests: boolean
  statusOptions: readonly ContestStatusOption[]
  statusFilter: '' | TeacherAWDReviewContestItemData['status']
  keywordFilter: string
  contestStatusLabel: (status: string) => string
}>()

const emit = defineEmits<{
  reload: []
  openContest: [contestId: string]
  updateStatusFilter: [status: '' | TeacherAWDReviewContestItemData['status']]
  updateKeywordFilter: [keyword: string]
}>()
</script>

<template>
  <TeacherAWDReviewDirectorySection
    :total-count="contests.length"
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
    </TeacherAWDReviewDirectoryState>
  </TeacherAWDReviewDirectorySection>
</template>

<style scoped>
.teacher-directory {
  display: flex;
  flex-direction: column;
}
</style>
