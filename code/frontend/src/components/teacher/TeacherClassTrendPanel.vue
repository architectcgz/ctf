<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherClassTrendData } from '@/api/contracts'
import LineChart from '@/components/charts/LineChart.vue'

const props = defineProps<{
  trend: TeacherClassTrendData | null
  title?: string
  subtitle?: string
  bare?: boolean
}>()

const panelTitle = computed(() => props.title || '近 7 天训练趋势')

const trendPoints = computed(() => {
  const points = Array.isArray(props.trend?.points) ? props.trend.points : []
  return points.filter((point) => typeof point.date === 'string' && point.date.trim().length > 0)
})
const hasTrendPoints = computed(() =>
  trendPoints.value.some(
    (point) =>
      Number(point.event_count ?? 0) > 0 ||
      Number(point.solve_count ?? 0) > 0 ||
      Number(point.active_student_count ?? 0) > 0
  )
)

const categories = computed(() => trendPoints.value.map((point) => point.date.slice(5)))

const series = computed(() => [
  {
    name: '训练事件',
    data: trendPoints.value.map((point) => point.event_count),
  },
  {
    name: '成功解题',
    data: trendPoints.value.map((point) => point.solve_count),
  },
  {
    name: '活跃学生',
    data: trendPoints.value.map((point) => point.active_student_count),
  },
])
</script>

<template>
  <section
    v-if="!bare"
    class="teacher-panel"
  >
    <header class="teacher-panel__header">
      <div class="journal-eyebrow">
        Trend
      </div>
      <h2 class="teacher-panel__title">
        {{ panelTitle }}
      </h2>
    </header>

    <div
      v-if="!hasTrendPoints"
      class="teacher-panel__empty-state"
    >
      <p class="teacher-panel__empty-title">
        暂无
      </p>
      <p class="teacher-panel__empty-copy">
        近 7 天还没有可展示的训练趋势
      </p>
    </div>

    <div
      v-else
      class="teacher-panel__chart"
    >
      <LineChart
        :categories="categories"
        :series="series"
      />
    </div>
  </section>

  <template v-else>
    <div
      v-if="!hasTrendPoints"
      class="teacher-panel__empty-state teacher-panel__empty-state--bare"
    >
      <p class="teacher-panel__empty-title">
        暂无
      </p>
      <p class="teacher-panel__empty-copy">
        近 7 天还没有可展示的训练趋势
      </p>
    </div>

    <div
      v-else
      class="teacher-panel__chart teacher-panel__chart--bare"
    >
      <LineChart
        :categories="categories"
        :series="series"
      />
    </div>
  </template>
</template>

<style scoped>
@import './teacher-panel-shell.css';
</style>
