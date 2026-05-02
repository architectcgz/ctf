<script setup lang="ts">
import { computed } from 'vue'
import { FolderKanban, RefreshCcw } from 'lucide-vue-next'

import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import TeacherAWDReviewContestDirectory from './TeacherAWDReviewContestDirectory.vue'
import TeacherAWDReviewSummaryPanel from './TeacherAWDReviewSummaryPanel.vue'
import TeacherAWDReviewSurfaceShell from './TeacherAWDReviewSurfaceShell.vue'
import TeacherAWDReviewWorkspaceHeader from './TeacherAWDReviewWorkspaceHeader.vue'

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
  updateStatusFilter: [status: '' | TeacherAWDReviewContestItemData['status']]
  updateKeywordFilter: [keyword: string]
}>()

const summaryItems = computed(() => [
  {
    label: '赛事数量',
    value: props.contestSummary.totalCount,
    hint: '当前可进入 AWD 复盘的赛事总数',
  },
  {
    label: '进行中',
    value: props.contestSummary.runningCount,
    hint: '仍在持续产出实时攻防信号的赛事',
  },
  {
    label: '可导出教师报告',
    value: props.contestSummary.exportReadyCount,
    hint: '已结束并允许生成教师复盘报告的赛事',
  },
])
</script>

<template>
  <TeacherAWDReviewSurfaceShell>
    <div class="teacher-page">
      <TeacherAWDReviewWorkspaceHeader
        overline="AWD Review"
        title="AWD复盘"
        header-class="awd-review-index-header"
        overline-class="awd-review-index-overline"
      >
        <template #description>
          集中查看赛事轮次、状态与导出就绪度，从统一入口进入整场或单轮复盘。
        </template>

        <template #actions>
          <button
            type="button"
            class="teacher-btn teacher-btn--ghost"
            @click="emit('openDashboard')"
          >
            教学概览
          </button>
          <button
            type="button"
            class="teacher-btn teacher-btn--primary"
            @click="emit('refresh')"
          >
            <RefreshCcw class="h-4 w-4" />
            刷新目录
          </button>
        </template>
      </TeacherAWDReviewWorkspaceHeader>

      <TeacherAWDReviewSummaryPanel
        title="Review Snapshot"
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
        :has-contests="hasContests"
        :status-options="statusOptions"
        :status-filter="statusFilter"
        :keyword-filter="keywordFilter"
        :contest-status-label="contestStatusLabel"
        @reload="emit('reload')"
        @open-contest="emit('openContest', $event)"
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
