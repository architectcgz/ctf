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
      <div class="journal-eyebrow">Trend</div>
      <h2 class="teacher-panel__title">{{ panelTitle }}</h2>
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

    <div v-else class="teacher-panel__chart">
      <LineChart :categories="categories" :series="series" />
    </div>
  </section>
</template>

<style scoped>
.teacher-panel {
  --panel-ink: var(--journal-ink, #0f172a);
  --panel-muted: var(--journal-muted, #64748b);
  --panel-border: var(--journal-border, rgba(226, 232, 240, 0.8));
  --panel-surface: var(--journal-surface, rgba(248, 250, 252, 0.9));
  --panel-surface-subtle: var(--journal-surface-subtle, rgba(241, 245, 249, 0.7));
  --panel-accent: var(--journal-accent, #4f46e5);
  --panel-accent-strong: var(--journal-accent-strong, #4338ca);
  border: 1px solid var(--panel-border);
  border-radius: 16px;
  background: var(--panel-surface-subtle);
  padding: 1.25rem 1.25rem 1.35rem;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.035);
}

.teacher-panel__header {
  margin-bottom: 1rem;
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.18);
  background: rgba(99, 102, 241, 0.06);
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

.teacher-panel__chart {
  overflow-x: auto;
  margin-top: 0.25rem;
  border-radius: 14px;
  border: 1px solid var(--panel-border);
  background: var(--panel-surface);
  padding: 0.75rem;
}
</style>
