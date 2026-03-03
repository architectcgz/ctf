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

const option = computed<EChartsOption>(() => ({
  tooltip: { trigger: 'axis' },
  legend: { textStyle: { color: '#8b949e' } },
  grid: { left: 16, right: 16, bottom: 16, top: 32, containLabel: true },
  xAxis: {
    type: 'category',
    data: props.categories,
    axisLine: { lineStyle: { color: '#30363d' } },
    axisLabel: { color: '#8b949e' },
  },
  yAxis: {
    type: 'value',
    axisLine: { lineStyle: { color: '#30363d' } },
    splitLine: { lineStyle: { color: 'rgba(48, 54, 61, 0.5)' } },
    axisLabel: { color: '#8b949e' },
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
