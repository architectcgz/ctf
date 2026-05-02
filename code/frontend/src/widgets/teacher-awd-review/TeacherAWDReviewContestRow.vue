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
    @click="emit('openContest', contest.id)"
  >
    <div :class="rowClassByKey.code">
      AWD-{{ contest.id }}
    </div>

    <div :class="rowClassByKey.name">
      <h4 class="teacher-directory-row-title">
        {{ contest.title }}
      </h4>
      <p class="teacher-directory-row-copy">
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
  display: grid;
  grid-template-columns: var(--awd-review-directory-columns);
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-4-5) 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background 160ms ease,
    border-color 160ms ease;
}

.teacher-directory-row:hover,
.teacher-directory-row:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
  box-shadow: inset 2px 0 0 color-mix(in srgb, var(--journal-accent) 62%, transparent);
  outline: none;
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

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-02);
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-copy {
  margin: 0;
  font-size: var(--font-size-0-84);
  line-height: 1.6;
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
