<script setup lang="ts">
import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { LineChart as EChartsLineChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([GridComponent, TooltipComponent, LegendComponent, EChartsLineChart, CanvasRenderer])

interface SeriesItem {
  name: string
  data: number[]
}

const props = defineProps<{
  categories: string[]
  series: SeriesItem[]
}>()

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

const option = computed<EChartsOption>(() => ({
  tooltip: { trigger: 'axis' },
  legend: { textStyle: { color: cssVar('--color-text-secondary') } },
  grid: {
    left: 16,
    right: 16,
    bottom: 16,
    top: 32,
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
    axisLine: { lineStyle: { color: cssVar('--color-border-default') } },
    splitLine: { lineStyle: { color: cssVar('--color-border-subtle') } },
    axisLabel: { color: cssVar('--color-text-secondary') },
  },
  series: props.series.map((item) => ({
    name: item.name,
    type: 'line',
    smooth: true,
    data: item.data,
  })),
}))
</script>

<template>
  <VChart class="h-80 w-full" :option="option" autoresize />
</template>
