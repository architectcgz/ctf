<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherClassTrendData } from '@/api/contracts'
import LineChart from '@/components/charts/LineChart.vue'

const props = defineProps<{
  trend: TeacherClassTrendData | null
  title?: string
  subtitle?: string
}>()

const panelTitle = computed(() => props.title || '近 7 天训练趋势')

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
    </header>

    <p v-if="!trend || trend.points.length === 0" class="teacher-panel__empty-copy">暂无趋势数据</p>

    <div v-else class="teacher-panel__chart">
      <LineChart :categories="categories" :series="series" />
    </div>
  </section>
</template>

<style scoped>
.teacher-panel {
  --panel-ink: var(--journal-ink, #0f172a);
  --panel-muted: var(--journal-muted, #64748b);
  --panel-border: color-mix(
    in srgb,
    var(--journal-border, var(--color-border-default)) 74%,
    transparent
  );
  --panel-surface: var(--journal-surface, var(--color-bg-surface));
  --panel-surface-subtle: var(--journal-surface-subtle, var(--color-bg-elevated));
  --panel-accent: var(--journal-accent, #4f46e5);
  --panel-accent-strong: var(--journal-accent-strong, #4338ca);
  border: 1px solid var(--panel-border);
  border-radius: 16px;
  background: var(--panel-surface-subtle);
  padding: var(--space-5) var(--space-5) var(--space-5-5);
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.teacher-panel__header {
  margin-bottom: var(--space-4);
}

.teacher-panel__title {
  margin-top: var(--space-3);
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--panel-ink);
}

.teacher-panel__subtitle {
  margin-top: var(--space-2);
  font-size: var(--font-size-0-84);
  line-height: 1.65;
  color: var(--panel-muted);
}

.teacher-panel__empty-copy {
  margin-top: var(--space-1);
  font-size: var(--font-size-0-84);
  line-height: 1.7;
  color: var(--panel-muted);
}

.teacher-panel__chart {
  overflow-x: auto;
  margin-top: var(--space-1);
  border-radius: 14px;
  border: 1px solid var(--panel-border);
  background: var(--panel-surface);
  padding: var(--space-3);
}
</style>
