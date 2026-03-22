<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherClassTrendData } from '@/api/contracts'
import LineChart from '@/components/charts/LineChart.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  trend: TeacherClassTrendData | null
  title?: string
  subtitle?: string
}>()

const panelTitle = computed(() => props.title || '近 7 天训练趋势')
const panelSubtitle = computed(() => props.subtitle || '按天查看训练事件、成功解题和活跃学生变化。')

const categories = computed(() => (props.trend?.points ?? []).map((point) => point.date.slice(5)))

const series = computed(() => [
  {
    name: '训练事件',
    data: (props.trend?.points ?? []).map((point) => point.event_count),
  },
  {
    name: '成功解题',
    data: (props.trend?.points ?? []).map((point) => point.solve_count),
  },
  {
    name: '活跃学生',
    data: (props.trend?.points ?? []).map((point) => point.active_student_count),
  },
])
</script>

<template>
  <section class="teacher-panel">
    <header class="teacher-panel__header">
      <h2 class="teacher-panel__title">
        {{ panelTitle }}
      </h2>
      <p class="teacher-panel__subtitle">
        {{ panelSubtitle }}
      </p>
    </header>

    <AppEmpty
      v-if="!trend || trend.points.length === 0"
      icon="FileChartColumnIncreasing"
      title="暂无趋势数据"
      description="当前班级近 7 天还没有可用训练趋势。"
    />

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

.teacher-panel__chart {
  overflow-x: auto;
  padding-top: 0.3rem;
}
</style>
