<script setup lang="ts">
import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { BarChart as EChartsBarChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import { GridComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { useEChartsMountGate } from '@/components/charts/echartsMountGate'

use([GridComponent, TooltipComponent, EChartsBarChart, CanvasRenderer])

const props = withDefaults(
  defineProps<{
    categories: string[]
    data: number[]
    seriesName?: string
  }>(),
  {
    seriesName: '统计值',
  }
)
const { containerRef, isChartReady } = useEChartsMountGate()

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

const option = computed<EChartsOption>(() => ({
  tooltip: { trigger: 'axis' },
  grid: {
    left: 16,
    right: 16,
    bottom: 16,
    top: 24,
    outerBoundsMode: 'same',
    outerBoundsContain: 'axisLabel',
  },
  xAxis: {
    type: 'category',
    data: props.categories,
    axisLine: { lineStyle: { color: cssVar('--color-border-default') } },
    axisLabel: { color: cssVar('--color-text-secondary') },
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: cssVar('--color-border-subtle') } },
    axisLabel: { color: cssVar('--color-text-secondary') },
  },
  series: [
    {
      name: props.seriesName,
      type: 'bar',
      data: props.data,
      itemStyle: {
        color: cssVar('--color-primary'),
        borderRadius: [8, 8, 0, 0],
      },
    },
  ],
}))
</script>

<template>
  <div ref="containerRef" class="h-80 w-full">
    <VChart
      v-if="isChartReady"
      class="h-full w-full"
      :option="option"
      autoresize
    />
  </div>
</template>
