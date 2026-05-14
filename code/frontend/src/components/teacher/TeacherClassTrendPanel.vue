<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherClassTrendData } from '@/api/contracts'
import LineChart from '@/components/charts/LineChart.vue'

const props = defineProps<{
  trend: TeacherClassTrendData | null
  title?: string
  subtitle?: string
  emptyDescription?: string
  bare?: boolean
}>()

const panelTitle = computed(() => props.title || '训练趋势')
const emptyDescription = computed(
  () => props.emptyDescription || '当前时间段还没有可展示的训练趋势'
)

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
      <h2 class="teacher-panel__title">
        {{ panelTitle }}
      </h2>
      <p v-if="props.subtitle" class="teacher-panel__copy">
        {{ props.subtitle }}
      </p>
    </header>

    <div
      v-if="!hasTrendPoints"
      class="teacher-panel__empty-state"
    >
      <p class="teacher-panel__empty-title">
        暂无
      </p>
      <p class="teacher-panel__empty-copy">
        {{ emptyDescription }}
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
        {{ emptyDescription }}
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

.teacher-panel__chart {
  background: linear-gradient(
    to bottom,
    color-mix(in srgb, var(--panel-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--panel-surface-subtle) 96%, var(--color-bg-base))
  );
  border-radius: 20px;
  padding: var(--space-6);
  border: 1px solid var(--panel-border);
  box-shadow: inset 0 2px 4px var(--color-shadow-soft);
}

.teacher-panel__chart--bare {
  border: 0;
  background: transparent;
  padding: 0;
  box-shadow: none;
}

.teacher-panel__copy {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-0-82);
  line-height: 1.7;
  color: var(--color-text-secondary);
}
</style>
