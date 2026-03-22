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
.teacher-panel {
  border-top: 1px solid var(--color-border-default);
  padding-top: 0.95rem;
}

.teacher-panel__header {
  margin-bottom: 0.72rem;
}

.teacher-panel__title {
  font-size: 1.04rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.teacher-panel__subtitle {
  margin-top: 0.3rem;
  font-size: 0.84rem;
  line-height: 1.65;
  color: var(--color-text-secondary);
}

.review-list {
  display: grid;
  gap: 0.7rem;
}

.review-item {
  --review-accent: var(--color-primary);
  border-bottom: 1px solid var(--color-border-subtle);
  border-left: 2px solid var(--review-accent);
  padding: 0.72rem 0.2rem 0.82rem 0.8rem;
}

.review-item--primary {
  --review-accent: var(--color-primary);
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
  font-size: 0.93rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.review-item__detail {
  margin-top: 0.36rem;
  font-size: 0.85rem;
  line-height: 1.72;
  color: var(--color-text-secondary);
}

.review-item__students {
  margin-top: 0.5rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.42rem;
}

.review-item__student-chip {
  display: inline-flex;
  align-items: center;
  border: 1px solid color-mix(in srgb, var(--review-accent) 34%, transparent);
  background: color-mix(in srgb, var(--review-accent) 8%, transparent);
  padding: 0.14rem 0.45rem;
  font-size: 0.74rem;
  color: color-mix(in srgb, var(--review-accent) 78%, var(--color-text-primary));
}

.review-item__recommendation {
  margin-top: 0.56rem;
  border-left: 2px solid color-mix(in srgb, var(--review-accent) 66%, var(--color-border-default));
  padding-left: 0.62rem;
}

.review-item__recommendation-label {
  font-size: 0.69rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--review-accent) 76%, var(--color-text-secondary));
}

.review-item__recommendation-title {
  margin-top: 0.24rem;
  font-size: 0.86rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.review-item__recommendation-meta {
  margin-top: 0.1rem;
  font-size: 0.76rem;
  color: var(--color-text-secondary);
}

.review-item__recommendation-reason {
  margin-top: 0.24rem;
  font-size: 0.82rem;
  line-height: 1.68;
  color: var(--color-text-secondary);
}
</style>
