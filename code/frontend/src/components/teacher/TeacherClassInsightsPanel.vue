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
  <section class="teacher-panel">
    <div class="teacher-insight-layout">
      <section class="teacher-subsection">
        <header class="teacher-subsection__header">
          <div class="journal-eyebrow">Students</div>
          <h2 class="teacher-panel__title">班级 Top 学生</h2>
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

        <div v-else class="top-student-list">
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

      <section class="teacher-subsection">
        <header class="teacher-subsection__header">
          <div class="journal-eyebrow">Weak Dimensions</div>
          <h2 class="teacher-panel__title">薄弱维度分布</h2>
          <p class="teacher-panel__subtitle">
            {{
              className
                ? `${className} 当前学生最弱维度的分布情况。`
                : '当前班级学生最弱维度的分布情况。'
            }}
          </p>
        </header>

        <AppEmpty
          v-if="weakDimensionStats.length === 0"
          icon="FileChartColumnIncreasing"
          title="暂无维度分布"
          description="当前班级还没有可用于聚合的能力画像数据。"
        />

        <div v-else class="dimension-list">
          <div v-for="item in weakDimensionStats" :key="item.dimension" class="dimension-item">
            <div class="dimension-item__head">
              <span class="dimension-item__name">{{ item.dimension }}</span>
              <span class="dimension-item__count">{{ item.count }} 人</span>
            </div>
            <div class="dimension-item__bar">
              <div class="dimension-item__bar-fill" :style="{ width: item.width }" />
            </div>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.teacher-insight-layout {
  display: grid;
  gap: 1.25rem;
}

.teacher-panel {
  --panel-ink: var(--journal-ink, #0f172a);
  --panel-muted: var(--journal-muted, #64748b);
  --panel-border: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 74%, transparent);
  --panel-divider: color-mix(in srgb, var(--panel-border) 76%, transparent);
  --panel-surface: var(--journal-surface, var(--color-bg-surface));
  --panel-surface-subtle: var(--journal-surface-subtle, var(--color-bg-elevated));
  --panel-accent: var(--journal-accent, #4f46e5);
  --panel-accent-strong: var(--journal-accent-strong, #4338ca);
  border: 1px solid var(--panel-border);
  border-radius: 16px;
  background: var(--panel-surface-subtle);
  padding: 1.25rem 1.25rem 1.35rem;
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.teacher-subsection + .teacher-subsection {
  border-top: 1px dashed var(--panel-divider);
  padding-top: 1.25rem;
}

.teacher-subsection__header {
  margin-bottom: 1rem;
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--panel-accent) 24%, transparent);
  background: color-mix(in srgb, var(--panel-accent) 10%, transparent);
  padding: 0.2rem 0.72rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--panel-accent-strong);
}

.teacher-panel__title {
  margin-top: 0.75rem;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--panel-ink);
}

.teacher-panel__subtitle {
  margin-top: 0.45rem;
  font-size: 0.84rem;
  line-height: 1.65;
  color: var(--panel-muted);
}

.top-student-list {
  display: grid;
  gap: 0.75rem;
}

.top-student-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.9rem;
  border-bottom: 1px dashed var(--panel-divider);
  padding: 0.2rem 0 0.95rem;
}

.top-student-item:last-child {
  border-bottom: 0;
  padding-bottom: 0;
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
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
    monospace;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--panel-accent);
}

.top-student-item__name {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.92rem;
  font-weight: 700;
  color: var(--panel-ink);
}

.top-student-item__meta {
  margin-top: 0.2rem;
  font-size: 0.8rem;
  color: var(--panel-muted);
}

.top-student-item__stats {
  flex-shrink: 0;
  text-align: right;
  font-size: 0.78rem;
  line-height: 1.6;
  color: var(--panel-muted);
}

.dimension-list {
  display: grid;
  gap: 0.82rem;
}

.dimension-item {
  border-bottom: 1px dashed var(--panel-divider);
  padding: 0.2rem 0 0.85rem;
}

.dimension-item:last-child {
  border-bottom: 0;
  padding-bottom: 0;
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
  color: var(--panel-ink);
}

.dimension-item__count {
  font-size: 0.78rem;
  color: var(--panel-muted);
}

.dimension-item__bar {
  margin-top: 0.4rem;
  height: 0.35rem;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--panel-border) 84%, var(--panel-surface));
}

.dimension-item__bar-fill {
  height: 100%;
  border-radius: 999px;
  background: color-mix(in srgb, var(--panel-accent) 85%, var(--panel-surface));
}

@media (min-width: 1280px) {
  .teacher-insight-layout {
    grid-template-columns: 1.05fr 0.95fr;
    gap: 1rem;
  }
}
</style>
