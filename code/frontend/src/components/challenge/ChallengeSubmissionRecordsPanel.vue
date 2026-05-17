<template>
  <section
    id="challenge-workspace-panel-records"
    class="workspace-panel panel"
    role="tabpanel"
    aria-labelledby="challenge-workspace-tab-records"
  >
    <section class="section section--flat">
      <div class="section-head workspace-tab-heading">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            Submissions
          </div>
          <h2 class="section-title workspace-tab-heading__title">
            提交记录
          </h2>
        </div>
        <div class="section-hint">
          最近提交
        </div>
      </div>

      <div v-if="submissionRecordsLoading" class="inline-note">正在加载提交记录...</div>

      <div
        v-else-if="submissionRecords.length === 0"
        class="inline-note"
      >
        还没有提交记录。你在右侧提交 Flag 后，新的提交结果会出现在这里。
      </div>

      <div
        v-else
        class="submission-records"
      >
        <div
          v-for="item in paginatedSubmissionRecords"
          :key="item.id"
          class="submission-record-item"
        >
          <div class="submission-record-time">
            {{ formatSubmissionTime(item.submittedAt) }}
          </div>
          <div class="submission-record-answer">
            {{ item.answer || submissionRecordMessage(item.status) }}
          </div>
          <div
            class="submission-record-status status-chip"
            :class="`submission-record-status--${item.status}`"
          >
            {{ submissionStatusText(item.status) }}
          </div>
        </div>
      </div>

      <div
        v-if="submissionRecordTotal > 0"
        class="submission-pagination workspace-directory-pagination"
      >
        <PagePaginationControls
          :page="submissionRecordPage"
          :total-pages="submissionRecordTotalPages"
          :total="submissionRecordTotal"
          :total-label="`共 ${submissionRecordTotal} 条提交`"
          @change-page="emit('change-page', $event)"
        />
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import type { ChallengeSubmissionRecordStatus } from '@/features/challenge-detail'

interface SubmissionRecordItem {
  id: string
  answer?: string
  status: ChallengeSubmissionRecordStatus
  submittedAt?: string
}

interface Props {
  submissionRecordsLoading: boolean
  submissionRecords: SubmissionRecordItem[]
  paginatedSubmissionRecords: SubmissionRecordItem[]
  submissionRecordPage: number
  submissionRecordTotal: number
  submissionRecordTotalPages: number
  formatSubmissionTime: (value?: string) => string
  submissionRecordMessage: (status: ChallengeSubmissionRecordStatus) => string
  submissionStatusText: (status: ChallengeSubmissionRecordStatus) => string
}

defineProps<Props>()

const emit = defineEmits<{
  'change-page': [page: number]
}>()
</script>

<style scoped>
.section--flat {
  padding-top: 0;
  border-top: 0;
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.section-hint {
  font-size: var(--font-size-13);
  line-height: 1.75;
  color: var(--text-faint);
}

.inline-note {
  padding-left: var(--space-4);
  border-left: 2px solid var(--line-soft);
  font-size: var(--font-size-0-90);
  line-height: 1.8;
  color: var(--text-subtle);
}

.submission-records {
  display: grid;
  gap: 0;
  border-top: 1px solid var(--line-soft);
}

.submission-record-item {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr) auto;
  gap: var(--space-4-5);
  align-items: center;
  padding: var(--space-4-5) 0;
  border-bottom: 1px solid var(--line-soft);
}

.submission-record-time {
  color: var(--text-faint);
  font: 500 13px/1.6 var(--font-mono);
}

.submission-record-answer {
  min-width: 0;
  font-family: var(--font-mono);
  font-size: var(--font-size-13);
  color: var(--text-subtle);
  word-break: break-all;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 var(--space-3-5);
  border: 1px solid var(--line-soft);
  border-radius: 999px;
  background: color-mix(in srgb, var(--bg-panel) 72%, transparent);
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--text-subtle);
}

.submission-record-status--correct {
  background: var(--journal-success-soft);
  color: var(--journal-success-ink);
}

.submission-record-status--incorrect,
.submission-record-status--error {
  background: var(--journal-danger-soft);
  color: var(--journal-danger-ink);
}

.submission-record-status--pending_review {
  background: var(--journal-warning-soft);
  color: var(--journal-warning-ink);
}

@media (max-width: 760px) {
  .submission-record-item {
    grid-template-columns: minmax(0, 1fr);
    gap: var(--space-2-5);
  }
}
</style>
