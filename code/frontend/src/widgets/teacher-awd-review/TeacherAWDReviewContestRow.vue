<script setup lang="ts">
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import { formatDate } from '@/utils/format'
import TeacherAWDReviewContestRowCta from './TeacherAWDReviewContestRowCta.vue'

defineProps<{
  contest: TeacherAWDReviewContestItemData
  contestStatusLabel: (status: string) => string
}>()

const emit = defineEmits<{
  openContest: [contestId: string]
}>()
</script>

<template>
  <button
    type="button"
    class="teacher-directory-row"
    @click="emit('openContest', contest.id)"
  >
    <div class="teacher-directory-cell teacher-directory-cell-code">
      AWD-{{ contest.id }}
    </div>

    <div class="teacher-directory-cell teacher-directory-cell-name">
      <h4 class="teacher-directory-row-title">
        {{ contest.title }}
      </h4>
      <p class="teacher-directory-row-copy">
        最近信号
        {{ contest.latest_evidence_at ? formatDate(contest.latest_evidence_at) : '暂无' }}
      </p>
    </div>

    <div class="teacher-directory-row-metrics">
      <span>{{
        contest.current_round ? `第 ${contest.current_round} 轮` : '未开始'
      }}</span>
      <span>共 {{ contest.round_count }} 轮</span>
    </div>

    <div class="teacher-directory-row-metrics">
      <span>{{ contest.team_count }} 支队伍</span>
      <span>{{ contest.mode.toUpperCase() }}</span>
    </div>

    <div class="teacher-directory-row-tags">
      <span class="teacher-directory-chip">
        {{ contestStatusLabel(contest.status) }}
      </span>
      <span
        class="teacher-directory-chip"
        :class="contest.export_ready ? '' : 'teacher-directory-chip-muted'"
      >
        {{ contest.export_ready ? '可导出' : '实时复盘' }}
      </span>
    </div>

    <TeacherAWDReviewContestRowCta />
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

.teacher-directory-row-metrics {
  display: grid;
  gap: var(--space-1);
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.teacher-directory-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.teacher-directory-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.7rem;
  padding: 0 var(--space-2-5);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: var(--font-size-0-75);
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.teacher-directory-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

@media (max-width: 1080px) {
  .teacher-directory-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
    padding: var(--space-4) 0;
  }
}
</style>
