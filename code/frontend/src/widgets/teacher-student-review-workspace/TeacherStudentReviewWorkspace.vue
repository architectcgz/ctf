<script setup lang="ts">
import { computed } from 'vue'

import AppEmpty from '@/components/common/AppEmpty.vue'
import type { TeacherAttackSessionQuery } from '@/api/teacher'
import type { TeacherAttackSessionResponseData, TeacherEvidenceData } from '@/api/contracts'
import {
  buildChallengeFilterOptions,
  buildReviewWorkspaceObservations,
  buildTeacherStudentReviewSummaryItems,
  eventMetaItems,
  eventTypeLabel,
  formatReviewWorkspaceDateTime,
  sessionModeLabel,
  sessionPathSummary,
  sessionResultClass,
  sessionResultLabel,
  TEACHER_STUDENT_REVIEW_WORKSPACE_COPY,
} from './model/presentation'

const props = defineProps<{
  evidence: TeacherEvidenceData | null
  attackSessions: TeacherAttackSessionResponseData | null
  challengeOptions?: Array<{ value: string; label: string }>
  loading: boolean
  query: TeacherAttackSessionQuery
}>()

const emit = defineEmits<{
  updateFilters: [payload: Partial<TeacherAttackSessionQuery>]
}>()

const summaryItems = computed(() =>
  buildTeacherStudentReviewSummaryItems({
    sessionSummary: props.attackSessions?.summary,
    evidenceSummary: props.evidence?.summary,
  })
)

const challengeOptions = computed(() =>
  props.challengeOptions && props.challengeOptions.length > 0
    ? props.challengeOptions
    : buildChallengeFilterOptions({
        evidence: props.evidence,
        attackSessions: props.attackSessions,
      })
)

const observations = computed(() =>
  buildReviewWorkspaceObservations({
    evidence: props.evidence,
    attackSessions: props.attackSessions,
  })
)
</script>

<template>
  <AppEmpty
    v-if="!attackSessions || attackSessions.sessions.length === 0"
    :title="TEACHER_STUDENT_REVIEW_WORKSPACE_COPY.emptyTitle"
    :description="TEACHER_STUDENT_REVIEW_WORKSPACE_COPY.emptyDescription"
    icon="NotebookText"
  />

  <template v-else>
    <div class="review-filter-bar">
      <label
        v-if="challengeOptions.length > 0"
        class="review-filter-field"
      >
        <span>题目</span>
        <select
          class="review-filter-select"
          :value="query.challenge_id || ''"
          :disabled="loading"
          @change="
            emit('updateFilters', {
              challenge_id: ($event.target as HTMLSelectElement).value || undefined,
            })
          "
        >
          <option value="">全部</option>
          <option
            v-for="option in challengeOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </option>
        </select>
      </label>

      <label class="review-filter-field">
        <span>模式</span>
        <select
          class="review-filter-select"
          :value="query.mode || ''"
          :disabled="loading"
          @change="
            emit('updateFilters', {
              mode: (($event.target as HTMLSelectElement).value || undefined) as TeacherAttackSessionQuery['mode'],
            })
          "
        >
          <option value="">全部</option>
          <option value="practice">训练</option>
          <option value="jeopardy">Jeopardy</option>
          <option value="awd">AWD</option>
        </select>
      </label>

      <label class="review-filter-field">
        <span>结果</span>
        <select
          class="review-filter-select"
          :value="query.result || ''"
          :disabled="loading"
          @change="
            emit('updateFilters', {
              result: (($event.target as HTMLSelectElement).value || undefined) as TeacherAttackSessionQuery['result'],
            })
          "
        >
          <option value="">全部</option>
          <option value="success">成功</option>
          <option value="failed">失败</option>
          <option value="in_progress">进行中</option>
          <option value="unknown">未知</option>
        </select>
      </label>

      <span
        v-if="loading"
        class="review-filter-status"
      >
        正在更新会话...
      </span>
    </div>

    <div
      v-if="observations.length > 0"
      class="review-observation-strip"
    >
      <article
        v-for="item in observations"
        :key="item.key"
        class="review-observation"
        :class="`review-observation--${item.level}`"
      >
        <div class="review-observation__label">
          {{ item.label }}
        </div>
        <p class="review-observation__summary">
          {{ item.summary }}
        </p>
      </article>
    </div>

    <div
      class="insight-kpi-grid progress-strip metric-panel-grid metric-panel-default-surface md:grid-cols-4"
    >
      <article
        v-for="item in summaryItems"
        :key="item.key"
        class="insight-kpi-card progress-card metric-panel-card"
      >
        <div class="insight-kpi-label progress-card-label metric-panel-label">
          {{ item.label }}
        </div>
        <div class="insight-kpi-value progress-card-value metric-panel-value">
          {{ item.value }}
        </div>
        <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
          {{ item.hint }}
        </div>
      </article>
    </div>

    <div class="attack-session-list">
      <article
        v-for="session in attackSessions.sessions"
        :key="session.id"
        class="attack-session"
      >
        <header class="attack-session__head">
          <div class="attack-session__main">
            <div class="attack-session__title-row">
              <h3>{{ session.title || '未命名目标' }}</h3>
              <span :class="sessionResultClass(session.result)">
                {{ sessionResultLabel(session.result) }}
              </span>
            </div>
            <div class="attack-session__meta">
              <span>{{ sessionModeLabel(session.mode) }}</span>
              <span>{{ session.event_count }} 个事件</span>
              <span>{{ formatReviewWorkspaceDateTime(session.started_at) }}</span>
            </div>
            <p>{{ sessionPathSummary(session) }}</p>
          </div>
          <div class="attack-session__target">
            <span v-if="session.challenge_id">题目 {{ session.challenge_id }}</span>
            <span v-if="session.round_id">轮次 {{ session.round_id }}</span>
            <span v-if="session.service_id">服务 {{ session.service_id }}</span>
          </div>
        </header>

        <ol class="attack-event-list">
          <li
            v-for="event in session.events ?? []"
            :key="event.id"
            class="attack-event"
          >
            <div class="attack-event__marker" aria-hidden="true" />
            <div class="attack-event__body">
              <div class="attack-event__head">
                <strong>{{ eventTypeLabel(event.type) }}</strong>
                <span>{{ formatReviewWorkspaceDateTime(event.occurred_at) }}</span>
              </div>
              <p>{{ event.summary }}</p>
              <div
                v-if="eventMetaItems(event).length > 0"
                class="attack-event__meta"
              >
                <span
                  v-for="item in eventMetaItems(event)"
                  :key="item.key"
                  class="insight-meta-pill"
                  :title="item.label"
                >
                  {{ item.label }}
                </span>
              </div>
            </div>
          </li>
        </ol>
      </article>
    </div>
  </template>
</template>

<style scoped>
.review-filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  align-items: end;
  margin-bottom: var(--space-5);
}

.review-filter-field {
  display: grid;
  gap: var(--space-1-5);
  min-width: min(100%, 11rem);
}

.review-filter-field span,
.review-filter-status {
  color: var(--journal-muted);
  font-size: var(--font-size-0-72);
}

.review-filter-select {
  min-height: var(--ui-control-height-md);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-ink);
  padding: 0 var(--space-3);
}

.review-filter-select:disabled {
  cursor: wait;
  opacity: 0.7;
}

.insight-meta-pill {
  display: inline-flex;
  max-width: 100%;
  align-items: center;
  border-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-width: 1px;
  border-style: solid;
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  padding: var(--space-1) var(--space-2-5);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.review-observation-strip {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
  gap: var(--space-3);
  margin-bottom: var(--space-5);
}

.review-observation {
  display: grid;
  gap: var(--space-1-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  padding: var(--space-3) var(--space-3-5);
}

.review-observation--good {
  border-color: color-mix(in srgb, var(--color-success-500) 40%, var(--journal-border));
}

.review-observation--attention {
  border-color: color-mix(in srgb, var(--color-warning-500) 40%, var(--journal-border));
}

.review-observation__label {
  color: var(--journal-muted);
  font-size: var(--font-size-0-72);
}

.review-observation__summary {
  margin: 0;
  color: var(--journal-ink);
  font-size: var(--font-size-0-86);
  line-height: 1.5;
}

.attack-session-list {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-5);
}

.attack-session {
  display: grid;
  gap: var(--space-4);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  padding-top: var(--space-4);
}

.attack-session__head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4);
  align-items: start;
}

.attack-session__main {
  min-width: 0;
}

.attack-session__title-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  align-items: center;
}

.attack-session__title-row h3 {
  min-width: 0;
  margin: 0;
  color: var(--journal-ink);
  font-size: var(--font-size-1);
  font-weight: 700;
}

.attack-session__meta,
.attack-session__target {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  color: var(--journal-muted);
  font-size: var(--font-size-0-78);
}

.attack-session__meta {
  margin-top: var(--space-2);
}

.attack-session__target {
  justify-content: flex-end;
}

.attack-session__main p,
.attack-event__body p {
  margin: var(--space-2) 0 0;
  color: var(--journal-muted);
  font-size: var(--font-size-0-88);
  line-height: 1.65;
}

.attack-event-list {
  display: grid;
  gap: var(--space-3);
  margin: 0;
  padding: 0;
  list-style: none;
}

.attack-event {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
}

.attack-event__marker {
  width: var(--space-2);
  height: var(--space-2);
  margin-top: var(--space-2);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-accent) 72%, var(--journal-ink));
}

.attack-event__body {
  min-width: 0;
  border-left: 1px solid color-mix(in srgb, var(--teacher-divider) 82%, transparent);
  padding-left: var(--space-3);
}

.attack-event__head {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  color: var(--journal-ink);
  font-size: var(--font-size-0-88);
}

.attack-event__head span {
  flex: 0 0 auto;
  color: var(--journal-muted);
  font-size: var(--font-size-0-72);
}

.attack-event__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-2);
  color: var(--journal-muted);
  font-size: var(--font-size-0-72);
}
</style>
