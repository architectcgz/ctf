<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherClassReviewData, TeacherClassReviewItemData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

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

function getAccentClass(accent: TeacherClassReviewItemData['accent']): string {
  if (accent === 'danger') return 'review-item review-item--danger'
  if (accent === 'warning') return 'review-item review-item--warning'
  if (accent === 'success') return 'review-item review-item--success'
  return 'review-item review-item--primary'
}
</script>

<template>
  <section
    class="teacher-panel"
    :class="{ 'teacher-panel--shellless': bare }"
  >
    <header
      v-if="!bare"
      class="teacher-panel__header"
    >
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
      class="review-list review-list--premium"
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
          class="review-item__students review-item__students--premium"
        >
          <span
            v-for="student in item.students"
            :key="student.id"
            class="review-item__student-chip review-item__student-chip--premium"
          >
            {{ student.name || student.username }}
          </span>
        </div>

        <div
          v-if="item.recommendation"
          class="review-item__recommendation review-item__recommendation--premium"
        >
          <div class="review-item__recommendation-label">
            推荐训练题
          </div>
          <div class="review-item__recommendation-body">
            <div class="recommendation-info">
              <div class="review-item__recommendation-title">
                {{ item.recommendation.title }}
              </div>
              <div class="review-item__recommendation-meta">
                {{ item.recommendation.category }} · {{ item.recommendation.difficulty }}
              </div>
            </div>
            <div class="review-item__recommendation-reason">
              {{ item.recommendation.reason }}
            </div>
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
  border-radius: 20px;
  border: 1px solid color-mix(in srgb, var(--review-accent) 18%, var(--panel-border));
  border-left: 4px solid color-mix(in srgb, var(--review-accent) 64%, transparent);
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--panel-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--panel-surface-subtle) 96%, var(--color-bg-base))
  );
  padding: var(--space-5) var(--space-6);
  box-shadow: 0 4px 12px var(--color-shadow-soft);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.review-item:hover {
  transform: translateX(4px);
  box-shadow: 0 8px 24px var(--color-shadow-soft);
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
  font-size: var(--font-size-17);
  font-weight: 800;
  color: var(--panel-ink);
}

.review-item__detail {
  margin-top: var(--space-2);
  font-size: var(--font-size-15);
  line-height: 1.75;
  color: var(--panel-muted);
}

.review-item__students--premium {
  margin-top: var(--space-4);
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.review-item__student-chip--premium {
  display: inline-flex;
  align-items: center;
  border: 1px solid color-mix(in srgb, var(--review-accent) 34%, transparent);
  background: color-mix(in srgb, var(--review-accent) 8%, transparent);
  padding: var(--space-1) var(--space-3);
  border-radius: 8px;
  font-size: var(--font-size-13);
  font-weight: 600;
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

@media (max-width: 768px) {
  .review-item__recommendation-body {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
}
</style>
