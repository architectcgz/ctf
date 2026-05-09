<script setup lang="ts">
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import { formatDate } from '@/utils/format'
import {
  AWD_REVIEW_DIRECTORY_COLUMN_SCHEMA,
  type AwdReviewDirectoryColumnKey,
} from './model/directory'
import TeacherAWDReviewContestRowCta from './TeacherAWDReviewContestRowCta.vue'
import TeacherAWDReviewContestRowMetrics from './TeacherAWDReviewContestRowMetrics.vue'
import TeacherAWDReviewContestRowStatusTags from './TeacherAWDReviewContestRowStatusTags.vue'

defineProps<{
  contest: TeacherAWDReviewContestItemData
  contestStatusLabel: (status: string) => string
}>()

const emit = defineEmits<{
  openContest: [contestId: string]
}>()

const rowClassByKey = AWD_REVIEW_DIRECTORY_COLUMN_SCHEMA.reduce(
  (result, column) => {
    result[column.key] = column.rowClass
    return result
  },
  {} as Record<AwdReviewDirectoryColumnKey, string>
)
</script>

<template>
  <button
    type="button"
    class="teacher-directory-row"
    :class="'workspace-directory-grid-row'"
    @click="emit('openContest', contest.id)"
  >
    <div :class="rowClassByKey.code">AWD-{{ contest.id }}</div>

    <div :class="rowClassByKey.name">
      <h4 class="teacher-directory-row-title" :class="'workspace-directory-row-title'">
        {{ contest.title }}
      </h4>
      <p class="workspace-directory-row-subtitle teacher-directory-row-copy">
        最近信号
        {{ contest.latest_evidence_at ? formatDate(contest.latest_evidence_at) : '暂无' }}
      </p>
    </div>

    <div :class="rowClassByKey.rounds">
      <TeacherAWDReviewContestRowMetrics
        :primary="contest.current_round ? `第 ${contest.current_round} 轮` : '未开始'"
        :secondary="`共 ${contest.round_count} 轮`"
      />
    </div>

    <div :class="rowClassByKey.teams">
      <TeacherAWDReviewContestRowMetrics
        :primary="`${contest.team_count} 支队伍`"
        :secondary="contest.mode.toUpperCase()"
      />
    </div>

    <div :class="rowClassByKey.status">
      <TeacherAWDReviewContestRowStatusTags
        :status-label="contestStatusLabel(contest.status)"
        :export-ready="contest.export_ready"
      />
    </div>

    <div :class="rowClassByKey.action">
      <TeacherAWDReviewContestRowCta />
    </div>
  </button>
</template>

<style scoped>
.teacher-directory-row {
  grid-template-columns: var(--awd-review-directory-columns);
  cursor: pointer;
}

.teacher-directory-cell {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
  align-content: center;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-cell-code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-78);
  font-weight: 700;
  color: var(--journal-muted);
}

.teacher-directory-row-copy {
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-directory-row-action {
  justify-self: end;
}

@media (max-width: 1080px) {
  .teacher-directory-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
    padding: var(--space-4) 0;
  }
}
</style>
