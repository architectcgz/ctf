<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherClassTrendData } from '@/api/contracts'
import LineChart from '@/components/charts/LineChart.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import SectionCard from '@/components/common/SectionCard.vue'

const props = defineProps<{
  trend: TeacherClassTrendData | null
  title?: string
  subtitle?: string
}>()

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
  <SectionCard
    :title="title || '近 7 天训练趋势'"
    :subtitle="subtitle || '按天查看训练事件、成功解题和活跃学生变化。'"
  >
    <AppEmpty
      v-if="!trend || trend.points.length === 0"
      icon="FileChartColumnIncreasing"
      title="暂无趋势数据"
      description="当前班级近 7 天还没有可用训练趋势。"
    />

    <LineChart v-else :categories="categories" :series="series" />
  </SectionCard>
</template>
