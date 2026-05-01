<script setup lang="ts">
import { computed } from 'vue'
import { GraduationCap, FileChartColumnIncreasing } from 'lucide-vue-next'

import type { TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  students: TeacherStudentItem[]
  className?: string
  stacked?: boolean
  splitCards?: boolean
  bare?: boolean
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
  <section
    class="teacher-panel"
    :class="{
      'teacher-panel--shellless': splitCards || bare,
      'teacher-panel--bare': bare,
    }"
  >
    <div
      class="teacher-insight-layout"
      :class="{
        'teacher-insight-layout--stacked': stacked,
        'teacher-insight-layout--split-cards': splitCards,
        'teacher-insight-layout--bare': bare,
      }"
    >
      <section
        class="teacher-subsection"
        :class="
          splitCards
            ? ['showcase-panel-card', 'showcase-panel-card--minimal-wire']
            : bare
              ? ['teacher-subsection--bare']
              : undefined
        "
      >
        <header class="teacher-subsection__header">
          <div class="journal-eyebrow">
            <GraduationCap class="inline-block w-3 h-3 mr-1 mb-0.5 opacity-60" />
            Top Students
          </div>
          <h2 class="teacher-panel__title">
            班级 Top 学生
          </h2>
          <p class="teacher-panel__subtitle">
            按解题效率与知识点掌握度综合评估。
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
          class="top-student-list top-student-list--premium"
        >
          <article
            v-for="(student, index) in topStudents"
            :key="student.id"
            class="top-student-item top-student-item--premium"
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
            <div class="top-student-item__stats top-student-item__stats--premium">
              <div class="stat-main">{{ student.solved_count ?? 0 }} <small>题</small></div>
              <div class="stat-sub">{{ student.total_score ?? 0 }} <small>分</small></div>
            </div>
          </article>
        </div>
      </section>

      <section
        class="teacher-subsection"
        :class="
          splitCards
            ? ['showcase-panel-card', 'showcase-panel-card--minimal-wire']
            : bare
              ? ['teacher-subsection--bare']
              : undefined
        "
      >
        <header class="teacher-subsection__header">
          <div class="journal-eyebrow">
            <FileChartColumnIncreasing class="inline-block w-3 h-3 mr-1 mb-0.5 opacity-60" />
            Skill Distribution
          </div>
          <h2 class="teacher-panel__title">
            薄弱维度分布
          </h2>
          <p class="teacher-panel__subtitle">
            当前班级在核心安全维度上的分布密度。
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
          class="dimension-list dimension-list--premium"
        >
          <div
            v-for="item in weakDimensionStats"
            :key="item.dimension"
            class="dimension-item dimension-item--premium"
          >
            <div class="dimension-item__head">
              <span class="dimension-item__name">{{ item.dimension }}</span>
              <span class="dimension-item__count">{{ item.count }} <small>人</small></span>
            </div>
            <div class="dimension-item__bar dimension-item__bar--premium">
              <div
                class="dimension-item__bar-fill"
                :style="{ width: item.width }"
              />
            </div>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
@import './teacher-panel-shell.css';

.teacher-insight-layout {
  display: grid;
  gap: var(--space-5);
}

.teacher-insight-layout--bare {
  gap: var(--space-10);
}

.teacher-subsection--bare {
  border: 1px solid var(--teacher-card-border);
  border-radius: 28px;
  background: linear-gradient(
    165deg,
    color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base))
  );
  padding: var(--space-8);
  box-shadow:
    0 4px 6px -1px rgb(0 0 0 / 0.03),
    0 2px 4px -2px rgb(0 0 0 / 0.03);
  transition: all 0.3s ease;
}

.teacher-subsection--bare:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 30%, var(--teacher-card-border));
  box-shadow: 0 12px 20px -10px rgb(0 0 0 / 0.08);
}

.teacher-insight-layout--split-cards {
  gap: var(--space-4);
  --showcase-panel-border: var(--panel-border);
  --showcase-panel-radius: 14px;
  --showcase-panel-background: transparent;
  --showcase-panel-shadow: none;
  --showcase-panel-padding: var(--space-4) var(--space-4-5);
}

.teacher-insight-layout--split-cards .teacher-subsection + .teacher-subsection {
  border-top: 0;
  padding-top: 0;
}

.teacher-insight-layout--split-cards .showcase-panel-card + .showcase-panel-card {
  border-top: 1px solid var(--showcase-panel-border, var(--panel-border));
}

.top-student-list--premium {
  display: grid;
  gap: var(--space-4);
}

.top-student-item--premium {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3-5);
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--panel-divider);
}

.top-student-item--premium:last-child {
  border-bottom: 0;
}

.top-student-item__main {
  min-width: 0;
}

.top-student-item__name-wrap {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.top-student-item__rank {
  display: inline-flex;
  min-width: 1.5rem;
  justify-content: center;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-14);
  font-weight: 800;
  color: var(--panel-accent-strong);
}

.top-student-item__name {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--panel-ink);
}

.top-student-item__meta {
  margin-top: var(--space-1);
  font-size: var(--font-size-13);
  color: var(--panel-muted);
}

.top-student-item__stats--premium {
  flex-shrink: 0;
  text-align: right;
  font-family: var(--font-family-mono);
}

.top-student-item__stats--premium .stat-main {
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--panel-ink);
}

.top-student-item__stats--premium .stat-sub {
  font-size: var(--font-size-12);
  color: var(--panel-muted);
}

.top-student-item__stats--premium small {
  font-size: var(--font-size-11);
  font-weight: 700;
  margin-left: 1px;
}

.dimension-list--premium {
  display: grid;
  gap: var(--space-5);
}

.dimension-item--premium {
  padding: var(--space-2) 0;
}

.dimension-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.dimension-item__name {
  font-size: var(--font-size-15);
  font-weight: 700;
  color: var(--panel-ink);
}

.dimension-item__count {
  font-size: var(--font-size-15);
  font-weight: 700;
  color: var(--panel-accent-strong);
}

.dimension-item__count small {
  font-size: var(--font-size-11);
  font-weight: 700;
  color: var(--panel-muted);
  margin-left: 1px;
}

.dimension-item__bar--premium {
  margin-top: var(--space-2-5);
  height: 0.5rem;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--panel-border) 60%, var(--panel-surface));
}

.dimension-item__bar-fill {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(to right, var(--panel-accent), var(--panel-accent-strong));
}

@media (min-width: 1280px) {
  .teacher-insight-layout:not(.teacher-insight-layout--stacked):not(
      .teacher-insight-layout--split-cards
    ) {
    grid-template-columns: 1.05fr 0.95fr;
    gap: var(--space-4);
  }
}

@media (min-width: 960px) {
  .teacher-insight-layout--split-cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
