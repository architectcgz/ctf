<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherClassReviewData, TeacherClassReviewItemData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  review: TeacherClassReviewData | null
  className?: string
}>()

const reviewItems = computed(() => props.review?.items ?? [])
const panelSubtitle = computed(() =>
  props.className
    ? `${props.className} 当前班级可直接执行的复盘结论与介入建议。`
    : '当前班级可直接执行的复盘结论与介入建议。'
)

function getAccentClass(accent: TeacherClassReviewItemData['accent']): string {
  if (accent === 'danger') return 'review-item review-item--danger'
  if (accent === 'warning') return 'review-item review-item--warning'
  if (accent === 'success') return 'review-item review-item--success'
  return 'review-item review-item--primary'
}
</script>

<template>
  <section class="teacher-panel">
    <header class="teacher-panel__header">
      <div class="journal-eyebrow">
        Review
      </div>
      <h2 class="teacher-panel__title">
        教学复盘结论
      </h2>
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

    <div
      v-else
      class="review-list"
    >
      <article
        v-for="item in reviewItems"
        :key="item.key"
        :class="getAccentClass(item.accent)"
      >
        <div class="review-item__title">
          {{ item.title }}
        </div>
        <div class="review-item__detail">
          {{ item.detail }}
        </div>

        <div
          v-if="item.students && item.students.length > 0"
          class="review-item__students"
        >
          <span
            v-for="student in item.students"
            :key="student.id"
            class="review-item__student-chip"
          >
            {{ student.name || student.username }}
          </span>
        </div>

        <div
          v-if="item.recommendation"
          class="review-item__recommendation"
        >
          <div class="review-item__recommendation-label">
            推荐训练题
          </div>
          <div class="review-item__recommendation-title">
            {{ item.recommendation.title }}
          </div>
          <div class="review-item__recommendation-meta">
            {{ item.recommendation.category }} / {{ item.recommendation.difficulty }}
          </div>
          <div class="review-item__recommendation-reason">
            {{ item.recommendation.reason }}
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
  gap: var(--space-3);
}

.review-item {
  --review-accent: var(--panel-accent);
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--review-accent) 18%, var(--panel-border));
  border-top-width: 3px;
  border-top-color: color-mix(in srgb, var(--review-accent) 58%, transparent);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--panel-surface) 94%, var(--color-bg-base)),
    color-mix(in srgb, var(--panel-surface-subtle) 96%, var(--color-bg-base))
  );
  padding: var(--space-4) var(--space-4) var(--space-4);
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
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--panel-ink);
}

.review-item__detail {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-0-85);
  line-height: 1.72;
  color: var(--panel-muted);
}

.review-item__students {
  margin-top: var(--space-3);
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1-5);
}

.review-item__student-chip {
  display: inline-flex;
  align-items: center;
  border: 1px solid color-mix(in srgb, var(--review-accent) 34%, transparent);
  background: color-mix(in srgb, var(--review-accent) 8%, transparent);
  padding: var(--space-0-5) var(--space-2);
  font-size: var(--font-size-0-74);
  color: color-mix(in srgb, var(--review-accent) 78%, var(--panel-ink));
}

.review-item__recommendation {
  margin-top: var(--space-3);
  border-top: 1px dashed color-mix(in srgb, var(--review-accent) 28%, var(--panel-border));
  padding-top: var(--space-3);
}

.review-item__recommendation-label {
  font-size: var(--font-size-0-69);
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--review-accent) 76%, var(--panel-muted));
}

.review-item__recommendation-title {
  margin-top: var(--space-1);
  font-size: var(--font-size-0-86);
  font-weight: 700;
  color: var(--panel-ink);
}

.review-item__recommendation-meta {
  margin-top: var(--space-0-5);
  font-size: var(--font-size-0-76);
  color: var(--panel-muted);
}

.review-item__recommendation-reason {
  margin-top: var(--space-1);
  font-size: var(--font-size-0-82);
  line-height: 1.68;
  color: var(--panel-muted);
}
</style>
