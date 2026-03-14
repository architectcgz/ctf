<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppCard from '@/components/common/AppCard.vue'
import SectionCard from '@/components/common/SectionCard.vue'

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
  <section class="grid gap-6 xl:grid-cols-[1.05fr_0.95fr]">
    <SectionCard
      title="班级 Top 学生"
      :subtitle="
        className
          ? `${className} 当前按解题数和得分排序的前 5 名。`
          : '当前班级按解题数和得分排序的前 5 名。'
      "
    >
      <AppEmpty
        v-if="topStudents.length === 0"
        icon="GraduationCap"
        title="暂无学生数据"
        description="当前班级还没有可用于排序的学生记录。"
      />

      <div v-else class="space-y-3">
        <AppCard
          v-for="(student, index) in topStudents"
          :key="student.id"
          variant="action"
          :accent="index === 0 ? 'warning' : 'primary'"
        >
          <div class="flex items-center justify-between gap-4">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span
                  class="inline-flex h-7 min-w-7 items-center justify-center rounded-full border border-primary/16 bg-primary/10 px-2 text-xs font-semibold text-primary"
                >
                  {{ index + 1 }}
                </span>
                <span class="truncate font-semibold text-text-primary">{{
                  student.name || student.username
                }}</span>
              </div>
              <div class="mt-2 text-sm text-text-secondary">
                @{{ student.username }}
                <span v-if="student.weak_dimension"> · 薄弱项 {{ student.weak_dimension }}</span>
              </div>
            </div>
            <div class="text-right">
              <div class="text-sm font-semibold text-text-primary">
                {{ student.solved_count ?? 0 }} 题
              </div>
              <div class="mt-1 text-xs text-text-secondary">{{ student.total_score ?? 0 }} 分</div>
            </div>
          </div>
        </AppCard>
      </div>
    </SectionCard>

    <SectionCard
      title="薄弱维度分布"
      :subtitle="
        className ? `${className} 当前学生最弱维度的分布情况。` : '当前班级学生最弱维度的分布情况。'
      "
    >
      <AppEmpty
        v-if="weakDimensionStats.length === 0"
        icon="FileChartColumnIncreasing"
        title="暂无维度分布"
        description="当前班级还没有可用于聚合的能力画像数据。"
      />

      <div v-else class="space-y-4">
        <div
          v-for="item in weakDimensionStats"
          :key="item.dimension"
          class="rounded-[22px] border border-border-subtle bg-[var(--color-bg-base)] px-4 py-4"
        >
          <div class="flex items-center justify-between gap-3 text-sm">
            <span class="font-medium text-text-primary">{{ item.dimension }}</span>
            <span class="text-text-secondary">{{ item.count }} 人</span>
          </div>
          <div class="mt-3 h-2 overflow-hidden rounded-full bg-[var(--color-border-default)]">
            <div
              class="h-full rounded-full bg-[var(--color-primary)]"
              :style="{ width: item.width }"
            />
          </div>
        </div>
      </div>
    </SectionCard>
  </section>
</template>
