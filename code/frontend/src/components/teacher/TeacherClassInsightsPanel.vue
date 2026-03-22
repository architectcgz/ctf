<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  students: TeacherStudentItem[]
  className?: string
}>()

const topStudents = computed(() =>
  [...props.students]
    .sort((left, right) => {
      const solvedGap = (right.solved_count ?? 0) - (left.solved_count ?? 0)
      if (solvedGap !== 0) return solvedGap
      const scoreGap = (right.total_score ?? 0) - (left.total_score ?? 0)
      if (scoreGap !== 0) return scoreGap
      return (left.username || '').localeCompare(right.username || '')
    })
    .slice(0, 5)
)

const weakDimensionStats = computed(() => {
  const counter = new Map<string, number>()
  for (const student of props.students) {
    const key = student.weak_dimension?.trim()
    if (!key) continue
    counter.set(key, (counter.get(key) ?? 0) + 1)
  }

  const maxCount = Math.max(...counter.values(), 0)
  return Array.from(counter.entries())
    .map(([dimension, count]) => ({
      dimension,
      count,
      width: maxCount > 0 ? `${Math.round((count / maxCount) * 100)}%` : '0%',
    }))
    .sort((left, right) => right.count - left.count)
})
</script>

<template>
  <section class="teacher-insight-layout">
    <section class="teacher-panel">
      <header class="teacher-panel__header">
        <h2 class="teacher-panel__title">
          班级 Top 学生
        </h2>
        <p class="teacher-panel__subtitle">
          {{
            className
              ? `${className} 当前按解题数和得分排序的前 5 名。`
              : '当前班级按解题数和得分排序的前 5 名。'
          }}
        </p>
      </header>

      <AppEmpty
        v-if="topStudents.length === 0"
        icon="GraduationCap"
        title="暂无学生数据"
        description="当前班级还没有可用于排序的学生记录。"
      />

      <div
        v-else
        class="top-student-list"
      >
        <article
          v-for="(student, index) in topStudents"
          :key="student.id"
          class="top-student-item"
        >
          <div class="top-student-item__main">
            <div class="top-student-item__name-wrap">
              <span class="top-student-item__rank">
                {{ index + 1 }}
              </span>
              <span class="top-student-item__name">{{ student.name || student.username }}</span>
            </div>
            <div class="top-student-item__meta">
              @{{ student.username }}
              <span v-if="student.weak_dimension"> · 薄弱项 {{ student.weak_dimension }}</span>
            </div>
          </div>
          <div class="top-student-item__stats">
            <div>{{ student.solved_count ?? 0 }} 题</div>
            <div>{{ student.total_score ?? 0 }} 分</div>
          </div>
        </article>
      </div>
    </section>

    <section class="teacher-panel">
      <header class="teacher-panel__header">
        <h2 class="teacher-panel__title">
          薄弱维度分布
        </h2>
        <p class="teacher-panel__subtitle">
          {{
            className ? `${className} 当前学生最弱维度的分布情况。` : '当前班级学生最弱维度的分布情况。'
          }}
        </p>
      </header>

      <AppEmpty
        v-if="weakDimensionStats.length === 0"
        icon="FileChartColumnIncreasing"
        title="暂无维度分布"
        description="当前班级还没有可用于聚合的能力画像数据。"
      />

      <div
        v-else
        class="dimension-list"
      >
        <div
          v-for="item in weakDimensionStats"
          :key="item.dimension"
          class="dimension-item"
        >
          <div class="dimension-item__head">
            <span class="dimension-item__name">{{ item.dimension }}</span>
            <span class="dimension-item__count">{{ item.count }} 人</span>
          </div>
          <div class="dimension-item__bar">
            <div
              class="dimension-item__bar-fill"
              :style="{ width: item.width }"
            />
          </div>
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
.teacher-insight-layout {
  display: grid;
  gap: 1.25rem;
}

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

.top-student-list {
  display: grid;
  gap: 0.56rem;
}

.top-student-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.9rem;
  border-bottom: 1px solid var(--color-border-subtle);
  padding: 0.58rem 0.2rem 0.62rem;
}

.top-student-item__main {
  min-width: 0;
}

.top-student-item__name-wrap {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.top-student-item__rank {
  display: inline-flex;
  min-width: 1.3rem;
  justify-content: center;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--color-primary);
}

.top-student-item__name {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.92rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.top-student-item__meta {
  margin-top: 0.2rem;
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.top-student-item__stats {
  flex-shrink: 0;
  text-align: right;
  font-size: 0.78rem;
  line-height: 1.6;
  color: var(--color-text-secondary);
}

.dimension-list {
  display: grid;
  gap: 0.72rem;
}

.dimension-item {
  border-bottom: 1px solid var(--color-border-subtle);
  padding: 0.5rem 0.2rem 0.68rem;
}

.dimension-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.7rem;
}

.dimension-item__name {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.dimension-item__count {
  font-size: 0.78rem;
  color: var(--color-text-secondary);
}

.dimension-item__bar {
  margin-top: 0.4rem;
  height: 0.35rem;
  overflow: hidden;
  background: var(--color-border-default);
}

.dimension-item__bar-fill {
  height: 100%;
  background: color-mix(in srgb, var(--color-primary) 85%, white);
}

@media (min-width: 1280px) {
  .teacher-insight-layout {
    grid-template-columns: 1.05fr 0.95fr;
    gap: 1rem;
  }
}
</style>
