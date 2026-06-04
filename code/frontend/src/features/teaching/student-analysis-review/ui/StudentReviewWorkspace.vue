<script setup lang="ts">
import { computed } from 'vue'

import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import type { AttackSessionQuery, AttackSessionResponseData, StudentEvidenceData } from '@/api/contracts'
import {
  buildChallengeFilterOptions,
  eventMetaItems,
  eventTypeLabel,
  formatDateTime,
  sessionResultClass,
  sessionModeLabel,
  sessionPathSummary,
  sessionResultLabel,
} from './studentInsightShared'

const props = defineProps<{
  evidence: StudentEvidenceData | null
  attackSessions: AttackSessionResponseData | null
  challengeOptions?: Array<{ value: string; label: string }>
  loading: boolean
  query: AttackSessionQuery
}>()

const emit = defineEmits<{
  updateFilters: [payload: Partial<AttackSessionQuery>]
}>()

const challengeOptions = computed(() =>
  props.challengeOptions && props.challengeOptions.length > 0
    ? props.challengeOptions
    : buildChallengeFilterOptions({
        evidence: props.evidence,
        attackSessions: props.attackSessions,
      })
)
</script>

<template>
  <AppEmpty
    v-if="!attackSessions || attackSessions.sessions.length === 0"
    title="暂无攻击会话"
    description="当前学员还没有可用于复盘的攻击过程记录。"
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
              mode: (($event.target as HTMLSelectElement).value || undefined) as AttackSessionQuery['mode'],
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
              result: (($event.target as HTMLSelectElement).value || undefined) as AttackSessionQuery['result'],
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
      class="attack-session-list"
    >
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
                <span>{{ formatDateTime(event.occurred_at) }}</span>
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

.attack-session-list {
  display: grid;
  gap: var(--space-3);
}

.attack-session {
  display: grid;
  gap: var(--space-3);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  padding-top: var(--space-3);
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
  margin-top: var(--space-1-5);
}

.attack-session__target {
  justify-content: flex-end;
}

.attack-session__main p,
.attack-event__body p {
  margin: var(--space-1-5) 0 0;
  color: var(--journal-muted);
  font-size: var(--font-size-0-84);
  line-height: 1.55;
}

.attack-event-list {
  display: grid;
  gap: var(--space-2-5);
  margin: 0;
  padding: 0;
  list-style: none;
}

.attack-event {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-2-5);
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
  padding-left: var(--space-2-5);
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
