<template>
  <section class="archive-section-stack">
    <article class="archive-panel teacher-surface-section">
      <header class="archive-panel__header">
        <div>
          <div class="archive-panel__eyebrow">
            Practice Review
          </div>
          <h3 class="archive-panel__title">
            练习复盘
          </h3>
        </div>
        <p class="archive-panel__hint">
          以练习题为单位整理接入、利用、提交和复盘材料，优先还原训练闭环。
        </p>
      </header>
      <div
        v-if="caseGroups.practiceCases.length === 0"
        class="archive-panel__empty"
      >
        暂无练习复盘案例。
      </div>
      <div
        v-else
        class="case-list"
      >
        <article
          v-for="item in caseGroups.practiceCases"
          :key="item.id"
          :class="cardClass(item, 'practice')"
          data-testid="practice-case-card"
        >
          <div class="archive-case__summary">
            <div class="archive-case__identity">
              <div class="archive-case__title-row">
                <strong>{{ item.title }}</strong>
                <span class="archive-case__status">{{ item.statusLabel }}</span>
              </div>
              <p class="archive-case__subtitle">
                {{ item.subtitle }}
              </p>
            </div>
            <div class="archive-case__meta-grid">
              <div
                v-for="metric in item.metrics"
                :key="metric.label"
                class="archive-case__metric"
              >
                <span>{{ metric.label }}</span>
                <strong>{{ metric.value }}</strong>
              </div>
            </div>
          </div>
          <div class="archive-case__stages">
            <span
              v-for="stage in item.stages"
              :key="stage.key"
              class="archive-case__stage"
            >
              {{ stage.label }} {{ stage.count }}
            </span>
          </div>
          <button
            type="button"
            class="archive-case__toggle"
            :aria-expanded="isExpanded(item.id)"
            data-testid="practice-case-toggle"
            @click="toggleCase(item.id)"
          >
            {{ isExpanded(item.id) ? '收起案例' : '展开案例' }}
          </button>
          <div
            v-if="isExpanded(item.id)"
            class="archive-case__details"
          >
            <ol class="archive-event-list">
              <li
                v-for="event in item.events"
                :key="event.id"
                class="archive-event"
              >
                <div class="archive-event__head">
                  <strong>{{ event.label }}</strong>
                  <span>{{ event.timestamp }}</span>
                </div>
                <p class="archive-event__detail">
                  {{ event.detail }}
                </p>
                <div class="archive-event__meta">
                  <span>{{ event.stageLabel }}</span>
                  <span v-if="event.meta">{{ event.meta }}</span>
                </div>
              </li>
            </ol>
          </div>
        </article>
      </div>
    </article>

    <article class="archive-panel teacher-surface-section">
      <header class="archive-panel__header">
        <div>
          <div class="archive-panel__eyebrow">
            AWD Review
          </div>
          <h3 class="archive-panel__title">
            AWD 复盘
          </h3>
        </div>
        <p class="archive-panel__hint">
          以题目和目标队伍为单位收束对抗过程，突出命中结果、得分和攻击节奏。
        </p>
      </header>
      <div
        v-if="caseGroups.awdCases.length === 0"
        class="archive-panel__empty"
      >
        暂无 AWD 对抗案例。
      </div>
      <div
        v-else
        class="case-list"
      >
        <article
          v-for="item in caseGroups.awdCases"
          :key="item.id"
          :class="cardClass(item, 'awd')"
          data-testid="awd-case-card"
        >
          <div class="archive-case__summary">
            <div class="archive-case__identity">
              <div class="archive-case__title-row">
                <strong>{{ item.title }}</strong>
                <span class="archive-case__status">{{ item.statusLabel }}</span>
              </div>
              <p class="archive-case__subtitle">
                {{ item.subtitle }}
              </p>
            </div>
            <div class="archive-case__meta-grid">
              <div
                v-for="metric in item.metrics"
                :key="metric.label"
                class="archive-case__metric"
              >
                <span>{{ metric.label }}</span>
                <strong>{{ metric.value }}</strong>
              </div>
            </div>
          </div>
          <div class="archive-case__stages">
            <span
              v-for="stage in item.stages"
              :key="stage.key"
              class="archive-case__stage"
            >
              {{ stage.label }} {{ stage.count }}
            </span>
          </div>
          <button
            type="button"
            class="archive-case__toggle"
            :aria-expanded="isExpanded(item.id)"
            @click="toggleCase(item.id)"
          >
            {{ isExpanded(item.id) ? '收起案例' : '展开案例' }}
          </button>
          <div
            v-if="isExpanded(item.id)"
            class="archive-case__details"
          >
            <ol class="archive-event-list">
              <li
                v-for="event in item.events"
                :key="event.id"
                class="archive-event"
              >
                <div class="archive-event__head">
                  <strong>{{ event.label }}</strong>
                  <span>{{ event.timestamp }}</span>
                </div>
                <p class="archive-event__detail">
                  {{ event.detail }}
                </p>
                <div class="archive-event__meta">
                  <span>{{ event.stageLabel }}</span>
                  <span v-if="event.meta">{{ event.meta }}</span>
                </div>
              </li>
            </ol>
          </div>
        </article>
      </div>
    </article>
  </section>
</template>

<style scoped>
.archive-section-stack {
  display: grid;
  gap: var(--space-5);
}

.archive-panel {
  padding: var(--space-4-5) var(--space-4-5) var(--space-5);
}

.archive-panel__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  padding-bottom: var(--space-3-5);
  border-bottom: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.archive-panel__eyebrow {
  font-size: var(--font-size-0-72);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
  font-family: var(--font-family-sans);
}

.archive-panel__title {
  margin-top: var(--space-2);
  font-size: var(--font-size-1-18);
  color: var(--journal-ink);
}

.archive-panel__hint {
  max-width: 30rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-0-93);
  line-height: 1.72;
}

.archive-panel__empty {
  padding: var(--space-4) 0;
  color: var(--color-text-secondary);
}

.case-list {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-4);
}

.archive-case {
  padding: var(--space-4-5);
  border-radius: 20px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 74%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
}

.archive-case--practice {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 98%, var(--color-bg-base))
  );
}

.archive-case--awd {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-accent) 6%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
}

.archive-case--success {
  border-color: color-mix(in srgb, var(--color-success) 22%, var(--journal-border));
}

.archive-case--warning {
  border-color: color-mix(in srgb, var(--color-warning) 24%, var(--journal-border));
}

.archive-case__summary {
  display: flex;
  gap: var(--space-4);
  justify-content: space-between;
}

.archive-case__identity {
  min-width: 0;
}

.archive-case__title-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  align-items: center;
}

.archive-case__title-row strong {
  color: var(--journal-ink);
}

.archive-case__status {
  color: var(--journal-accent-strong);
  font-size: var(--font-size-0-78);
  font-family: var(--font-family-mono);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.archive-case__subtitle {
  margin-top: var(--space-2);
  color: var(--journal-muted);
}

.archive-case__meta-grid {
  display: grid;
  gap: var(--space-2);
  min-width: 12rem;
}

.archive-case__metric {
  display: flex;
  gap: var(--space-3);
  justify-content: space-between;
  color: var(--journal-muted);
  font-size: var(--font-size-0-86);
}

.archive-case__metric strong {
  color: var(--journal-ink);
}

.archive-case__stages {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-4);
}

.archive-case__stage {
  display: inline-flex;
  align-items: center;
  padding: var(--space-1-5) var(--space-3);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-border) 16%, transparent);
  color: var(--journal-muted);
  font-size: var(--font-size-0-78);
  font-family: var(--font-family-mono);
}

.archive-case__toggle {
  margin-top: var(--space-4);
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--journal-accent-strong);
  font-weight: 700;
  cursor: pointer;
}

.archive-case__details {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.archive-event-list {
  display: grid;
  gap: var(--space-3);
}

.archive-event {
  padding: var(--space-3-5) var(--space-4);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base));
}

.archive-event__head,
.archive-event__meta {
  display: flex;
  gap: var(--space-3);
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
}

.archive-event__head strong {
  color: var(--journal-ink);
}

.archive-event__head span,
.archive-event__meta {
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
  font-family: var(--font-family-mono);
}

.archive-event__detail {
  margin-top: var(--space-2);
  line-height: 1.75;
  color: color-mix(in srgb, var(--journal-muted) 80%, var(--journal-ink));
}

.archive-event__meta {
  margin-top: var(--space-2);
}

@media (max-width: 1023px) {
  .archive-panel__header,
  .archive-case__summary {
    flex-direction: column;
    align-items: flex-start;
  }

  .archive-case__meta-grid {
    width: 100%;
  }
}
</style>

<script setup lang="ts">
import { computed, ref } from 'vue'

import type {
  ReviewArchiveEvidenceItemData,
  ReviewArchiveManualReviewItemData,
  ReviewArchiveWriteupItemData,
  TimelineEvent,
} from '@/api/contracts'
import { buildReviewArchiveCaseGroups, type ReviewArchiveCase } from './reviewArchiveCases'

const props = defineProps<{
  timeline: TimelineEvent[]
  evidence: ReviewArchiveEvidenceItemData[]
  writeups: ReviewArchiveWriteupItemData[]
  manualReviews: ReviewArchiveManualReviewItemData[]
}>()

const expandedCases = ref<string[]>([])

const caseGroups = computed(() =>
  buildReviewArchiveCaseGroups({
    timeline: props.timeline,
    evidence: props.evidence,
    writeups: props.writeups,
    manualReviews: props.manualReviews,
  })
)

function isExpanded(caseId: string): boolean {
  return expandedCases.value.includes(caseId)
}

function toggleCase(caseId: string): void {
  expandedCases.value = isExpanded(caseId)
    ? expandedCases.value.filter((item) => item !== caseId)
    : [...expandedCases.value, caseId]
}

function cardClass(item: ReviewArchiveCase, section: 'practice' | 'awd'): string[] {
  return ['archive-case', `archive-case--${section}`, `archive-case--${item.tone}`]
}
</script>
