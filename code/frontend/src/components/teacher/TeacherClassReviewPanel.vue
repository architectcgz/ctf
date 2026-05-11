<script setup lang="ts">
import { computed } from 'vue'

import type {
  AdviceSeverity,
  TeacherClassReviewData,
  TeacherClassReviewItemData,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { ChallengeCategoryDifficultyPills } from '@/entities/challenge'

const props = defineProps<{
  review: TeacherClassReviewData | null
  className?: string
  bare?: boolean
}>()

const reviewItems = computed(() => props.review?.items ?? [])
const panelSubtitle = computed(() =>
  props.className
    ? `${props.className} 当前班级可直接执行的复盘结论与介入建议。`
    : '当前班级可直接执行的复盘结论与介入建议。'
)

function getSeverityClass(severity: TeacherClassReviewItemData['severity']): string {
  if (severity === 'danger') return 'review-item review-item--danger'
  if (severity === 'warning') return 'review-item review-item--warning'
  if (severity === 'good') return 'review-item review-item--success'
  return 'review-item review-item--primary'
}

function severityLabel(severity: AdviceSeverity): string {
  if (severity === 'danger') return '高风险'
  if (severity === 'warning') return '需尽快处理'
  if (severity === 'attention') return '建议跟进'
  return '表现稳定'
}
</script>

<template>
  <section class="teacher-panel" :class="{ 'teacher-panel--shellless': bare }">
    <header v-if="!bare" class="teacher-panel__header">
      <div class="journal-eyebrow">Review</div>
      <h2 class="teacher-panel__title">教学复盘结论</h2>
      <p class="teacher-panel__subtitle">
        {{ panelSubtitle }}
      </p>
    </header>

    <AppEmpty
      v-if="reviewItems.length === 0"
      icon="FileChartColumnIncreasing"
      title="暂无复盘结论"
      description="当前班级还没有足够的训练数据形成稳定结论。"
    />

    <div v-else class="review-list review-list--premium">
      <article
        v-for="item in reviewItems"
        :key="item.code"
        :class="getSeverityClass(item.severity)"
      >
        <div class="review-item__head">
          <span class="review-item__severity">
            {{ severityLabel(item.severity) }}
          </span>
          <span v-if="item.dimension" class="review-item__dimension">
            {{ item.dimension }}
          </span>
        </div>
        <div class="review-item__title">
          {{ item.summary }}
        </div>
        <div v-if="item.evidence" class="review-item__detail">
          {{ item.evidence }}
        </div>
        <div v-if="item.action" class="review-item__action">
          {{ item.action }}
        </div>

        <div
          v-if="item.students && item.students.length > 0"
          class="review-item__students review-item__students--premium"
        >
          <span
            v-for="student in item.students"
            :key="student.id"
            class="review-item__student-chip review-item__student-chip--premium"
            :class="'workspace-directory-status-pill'"
          >
            {{ student.name || student.username }}
          </span>
        </div>

        <div
          v-if="item.recommendation"
          class="review-item__recommendation review-item__recommendation--premium"
        >
          <div class="review-item__recommendation-label">推荐训练题</div>
          <div class="review-item__recommendation-body">
            <div class="recommendation-info">
              <div class="review-item__recommendation-title">
                {{ item.recommendation.title }}
              </div>
              <ChallengeCategoryDifficultyPills
                class="review-item__recommendation-meta"
                :category="item.recommendation.category"
                :difficulty="item.recommendation.difficulty"
              />
            </div>
            <div class="review-item__recommendation-reason">
              {{ item.recommendation.summary }}
            </div>
          </div>
          <div v-if="item.recommendation.evidence" class="review-item__recommendation-evidence">
            {{ item.recommendation.evidence }}
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
@import './teacher-panel-shell.css';

.review-list {
  display: grid;
  gap: var(--space-4);
}

.review-item {
  --review-accent: var(--panel-accent);
  border-radius: 24px;
  border: 1px solid color-mix(in srgb, var(--review-accent) 12%, var(--panel-border));
  border-left: 5px solid color-mix(in srgb, var(--review-accent) 64%, transparent);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--panel-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--panel-surface-subtle) 96%, var(--color-bg-base))
  );
  padding: var(--space-6) var(--space-7);
  box-shadow:
    0 1px 3px 0 rgb(0 0 0 / 0.1),
    0 1px 2px -1px rgb(0 0 0 / 0.1);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.review-item:hover {
  transform: translateX(6px);
  box-shadow:
    0 10px 15px -3px rgb(0 0 0 / 0.1),
    0 4px 6px -4px rgb(0 0 0 / 0.1);
  border-color: color-mix(in srgb, var(--review-accent) 30%, var(--panel-border));
}

.review-item--primary {
  --review-accent: var(--panel-accent);
}

.review-item--warning {
  --review-accent: var(--color-warning);
}

.review-item--danger {
  --review-accent: var(--color-danger);
}

.review-item--success {
  --review-accent: var(--color-success);
}

.review-item__title {
  margin-top: var(--space-3);
  font-size: var(--font-size-17);
  font-weight: 800;
  color: var(--panel-ink);
}

.review-item__head {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  align-items: center;
}

.review-item__severity,
.review-item__dimension {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: var(--space-1) var(--space-2-5);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.review-item__severity {
  border: 1px solid color-mix(in srgb, var(--review-accent) 30%, transparent);
  color: color-mix(in srgb, var(--review-accent) 82%, var(--panel-ink));
  background: color-mix(in srgb, var(--review-accent) 8%, transparent);
}

.review-item__dimension {
  border: 1px solid color-mix(in srgb, var(--panel-border) 88%, transparent);
  color: var(--panel-muted);
  background: color-mix(in srgb, var(--panel-border) 44%, transparent);
}

.review-item__detail {
  margin-top: var(--space-2);
  font-size: var(--font-size-15);
  line-height: 1.75;
  color: var(--panel-muted);
}

.review-item__action {
  margin-top: var(--space-3);
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: color-mix(in srgb, var(--review-accent) 74%, var(--panel-ink));
}

.review-item__students--premium {
  margin-top: var(--space-4);
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.review-item__student-chip--premium {
  border: 1px solid color-mix(in srgb, var(--review-accent) 34%, transparent);
  background: color-mix(in srgb, var(--review-accent) 8%, transparent);
  color: color-mix(in srgb, var(--review-accent) 78%, var(--panel-ink));
}

.review-item__recommendation--premium {
  margin-top: var(--space-5);
  border-top: 1px solid color-mix(in srgb, var(--review-accent) 12%, var(--panel-border));
  padding-top: var(--space-4);
}

.review-item__recommendation-body {
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: var(--space-5);
  margin-top: var(--space-2);
}

.review-item__recommendation-label {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--review-accent) 76%, var(--panel-muted));
}

.review-item__recommendation-title {
  margin-top: var(--space-1);
  font-size: var(--font-size-15);
  font-weight: 800;
  color: var(--panel-ink);
}

.review-item__recommendation-meta {
  margin-top: var(--space-0-5);
  font-size: var(--font-size-13);
  color: var(--panel-muted);
}

.review-item__recommendation-reason {
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--panel-muted);
}

.review-item__recommendation-evidence {
  margin-top: var(--space-2);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: color-mix(in srgb, var(--panel-muted) 86%, var(--panel-ink));
}

@media (max-width: 768px) {
  .review-item__recommendation-body {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
}
</style>
